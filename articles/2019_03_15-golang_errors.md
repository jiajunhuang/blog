# Golang中的错误处理

## error是个啥

Golang的error是一个接口：

```go
type error interface {
        Error() string
}
```

也就是说，实现了 `Error() string` 就可以。

## errors

看看标准库里的errors咋用：

```go
package main

import (
	"errors"
	"fmt"
)

// MyError 就是一个自定义的错误
type MyError struct{}

func (e MyError) Error() string {
	return "my error instance"
}

func oops() error {
	return MyError{}
}

func oops2() error {
	return errors.New("hoho")
}

func checkError(err error) {
	switch err.(type) {
	case MyError:
		fmt.Printf("MyError happened: %s\n", err)
	default:
		fmt.Printf("errors happend: %#v\n", err)
	}
}

func main() {
	checkError(oops())
	checkError(oops2())
}
```

执行一下：

```bash
$ go run main.go
MyError happened: my error instance
errors happend: &errors.errorString{s:"hoho"}
```

我们发现，标准库里的errors只能保存一个字符串，看下errors的实现就知道了：

```go
package errors

// New returns an error that formats as the given text.
func New(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
```

有时候我们想要把调用栈也打印出来，如果只使用errors的话，我们就要每次自己把栈封装进去：

```go
package main

import (
	"fmt"
	"runtime/debug"
)

// MyError 就是一个自定义的错误
type MyError struct {
	s     string
	stack []byte
}

// NewMyError 新建错误
func NewMyError(s string) error {
	return MyError{s, debug.Stack()}
}

func (e MyError) Error() string {
	return fmt.Sprintf("%s\n%s", e.s, e.stack)
}

func main() {
	fmt.Printf("%s\n", NewMyError("hello"))
}
```

不过已经有库了，推荐使用 https://github.com/pkg/errors，也可以坐等 Go2: https://go.googlesource.com/proposal/+/master/design/go2draft.md

---

- https://golang.org/pkg/errors/
- https://github.com/pkg/errors
- https://go.googlesource.com/proposal/+/master/design/go2draft.md
