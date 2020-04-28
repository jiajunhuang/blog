# Golang log 源码阅读

来看看Go标准库里的log，怎么实现的，首先我们来看一个例子：

```go
package main

import (
	"log"
)

func main() {
	log.Printf("hello %s", "world")
	log.Panicf("this")
}
```

运行一下：

```go
$ go build && ./test
2020/04/28 22:24:25 hello world
2020/04/28 22:24:25 this
panic: this

goroutine 1 [running]:
log.Panicf(0x4cc48d, 0x4, 0x0, 0x0, 0x0)
	/snap/go/5646/src/log/log.go:358 +0xc0
main.main()
	/home/jiajun/Code/test/main.go:9 +0xa0
```

我们从 `log.Printf` 来看看咋实现的：

```go
func Printf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf(format, v...))
}
```

我们跟进 `Output`：

```go
func (l *Logger) Output(calldepth int, s string) error {
	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, line)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
}
```

可以看到里面的实现，就是往 `l.buf` 里写入，然后呢，往 `l.out` 写出，下面的代码可以看到，标准库里其实就是往
`os.Stderr` 写：

```go
var std = New(os.Stderr, "", LstdFlags)

func New(out io.Writer, prefix string, flag int) *Logger {
	return &Logger{out: out, prefix: prefix, flag: flag}
}

type Logger struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	prefix string     // prefix on each line to identify the logger (but see Lmsgprefix)
	flag   int        // properties
	out    io.Writer  // destination for output
	buf    []byte     // for accumulating text to write
}
```

最后我们来看一下 `Panicf`，其实很简单，就是输出log，然后来一个 `panic`：

```go
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	std.Output(2, s)
	panic(s)
}
```
