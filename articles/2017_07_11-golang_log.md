# Golang log库 源码阅读

Golang的log库。。。还是太简单，简单瞄了一下实现，差不多就是这样：

```go
package main

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type Logger struct {
	mu  sync.Mutex
	out io.Writer
}

func New(out io.Writer) *Logger {
	return &Logger{out: out}
}

func (l *Logger) output(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.out.Write([]byte(fmt.Sprintf(format, v...)))
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.output(format, v...)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	l.output(format, v...)
	panic("traceback:\n")
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.output(format, v...)
	os.Exit(1)
}

var std = New(os.Stderr)

func Printf(format string, v ...interface{}) {
	std.Printf(format, v...)
}

func Panicf(format string, v ...interface{}) {
	std.Panicf(format, v...)
}

func Fatalf(format string, v ...interface{}) {
	std.Fatalf(format, v...)
}
```

测试用例:

```go
package main

func main() {
    std.Printf("this is: %d\n", 1)

    std.Panicf("bye\n")
}
```

运行结果：

```bash
root@arch test: ./main          
this is: 1                      
bye                             
panic: traceback:               


goroutine 1 [running]:          
main.(*Logger).Panicf(0xc42006a060, 0x4a7115, 0x4, 0x0, 0x0, 0x0)                                                                
        /root/test/mylog.go:32 +0xa8                            
main.main()                     
        /root/test/main.go:6 +0xeb                              
root@arch test:                 

```
