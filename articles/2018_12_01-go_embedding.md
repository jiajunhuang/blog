# Go类型嵌套

Go没有继承，但是有类型嵌套，主要有三种使用方式，使用类型嵌套，wrapper可以自动获得被嵌套类型的所有方法。接下来我们分别看
三种情况下的例子：

- struct中嵌套struct

```go
package main

import (
	"fmt"
)

type Foo struct{}

func (f Foo) SayFoo() {
	fmt.Println("foo")
}

type Bar struct {
	Foo
}

func main() {
	b := Bar{}
	b.SayFoo()
}
```

- interface中嵌套interface，对于接口来说，则是自动获得被嵌套的接口规定的方法，所以实现ReadWriter的实例必须同时有 `Read` 和 `Write` 方法。

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type ReadWriter interface {
	Reader
	Writer
}
```

- struct中嵌套interface，这种情况比较特殊，struct中嵌套interface之后，struct会自动获得接口规定的方法：

```go
package main

import (
	"fmt"
)

type Foo interface {
	SayFoo()
}

type foo struct{}

func (f foo) SayFoo() {
	fmt.Println("foo")
}

type Bar struct {
	Foo
}

func main() {
	b := Bar{foo{}} // 传入一个符合 Foo 这个接口的实例

	b.SayFoo()
}
```

---

- https://golang.org/doc/effective_go.html#embedding
