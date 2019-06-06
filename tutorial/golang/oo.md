# 面向对象编程

## 目录

- [安装](./installation_linux.md)
    - [Windows](./installation_windows.md)
    - [Linux](./installation_linux.md)
    - [macOS](./installation_mac_os.md)
- [Hello, World](./hello_world.md)
- [Go语言简介](./intro.md)
- [基本类型](./basic_types.md)
- [容器类型](./composite_types.md)
- [函数](./function.md)
- [流程控制](./flow.md)
- [错误处理](./errors.md)
- [面向对象编程](./oo.md)
- [面向接口编程](./interface.md)
- [指针](./pointers.md)
- [Goroutine](./goroutine.md)
- [Channel](./channel.md)
- [并发编程](./concurrency.md)
- [go tools](./go_tool.md)

Go语言中没有 `class` 关键字，但是Go语言仍然可以使用面向对象的方式来编程。

```go
package main

import (
	"fmt"
)

type Animal struct {
}

func (a *Animal) Walk() {
	fmt.Println("Animal walk")
}

type Duck struct {
	Animal
	Color string
}

func main() {
	d := Duck{}
	d.Walk()
}
```

可以看到，Duck继承了Animal的方法 `Walk`，因此可以直接 `d.Walk` 来调用。

此处需要注意的是，`Walk` 方法，`Walk` 是一个方法，与函数的区别在于, `func` 关键字之后，`Walk` 函数名之前有一个 `(a *Animal)`，
这是说，把实例传给a这个参数，而a的类型是 `*Animal`。

另外需要提一点的是，在 `main` 函数中，`d` 的类型是 `Duck`，但是 `Walk` 函数的实例的类型必须是 `*Animal`，既然类型不匹配，
为什么编译器没有报错呢？这是因为Go的编译器会自动为我们转换，即上面的代码相当于：

```go
package main

import (
	"fmt"
)

type Animal struct {
}

func (a *Animal) Walk() {
	fmt.Println("Animal walk")
}

type Duck struct {
	Animal
	Color string
}

func main() {
	d := Duck{}
	(&d).Walk()
}
```

> &的作用是取地址，关于这个我们会在指针一节讲述

---

- 上一篇：[错误处理](./errors.md)
- 下一篇：[面向接口编程](./interface.md)
