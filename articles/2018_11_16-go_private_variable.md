# Go访问私有变量

> 工程上当然不能这么干，所以这篇文章呢，just for fun :)

Go语言里，小写的变量，包外不可访问。

前些天，同事说，Ruby有办法直接访问私有变量，我说其实Python也可以。那么问题来了，Go可以吗？答案当然是可以。正常情况下Go没法
直接访问私有变量只是因为编译器不让你这么干，我们绕开它就好了。

虽然Go没有指针运算，不能直接根据指针运算来取出私有变量，但是有指针和type cast，就有办法访问的。看代码（为了方便，我没有把
这两个放到不同的包里，而是直接放到一个文件里了）：

```go
package main

import (
	"fmt"
	"unsafe"
)

type Demo struct {
	private        string
	youCannotSeeMe int
	Trick          bool
}

func main() {
	d := Demo{private: "hahaha", youCannotSeeMe: 110, Trick: true}
	p := unsafe.Pointer(&d)

	type Header struct {
		NotPrivate  string
		YouCanSeeMe int
	}

	fmt.Printf("%+v", *(*Header)(p))
}
```

执行一下：

```bash
$ go run main.go  && echo
{NotPrivate:hahaha YouCanSeeMe:110}
```

原理就是，侵入Demo的实现，取出d的结构体起始地址，然后强转为 Header 类型，这样就可以读出里边的值了。

😯，最后再强调一遍，要是你在生产的代码里这么干，被同事打死了可不要说是我教的哈哈哈哈。

----

- https://golang.org/doc/faq#no_pointer_arithmetic
