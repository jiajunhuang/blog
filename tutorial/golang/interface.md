# 面向接口编程

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

如果你听说过面向对象编程，那么你很有可能也听说过面向接口编程。面向接口编程的核心思想是，只要你实现了这个方法，你就符合这个
接口。听起来有点玄乎，看个例子吧：

```go
package main

import (
	"fmt"
)

type Walker interface {
	Walk()
}

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
	var x Walker = &Duck{}
	x.Walk()
}
```

> Go语言中不需要像Java那样显示的声明是否实现了某个接口，编译器会自动检测，如果实现了接口所要求的所有函数，那么这个实例就
> 符合该接口。

请注意，倒数第三行，为什么要写成 `var x Walker = &Duck{}` 而不是 `var x Walker = Duck{}` 呢？

正如上面所说，实现了 `Walk` 这个方法的类型，是 `*Animal`，而不是 `Animal`，因此，只有 `*Animal`, `*Duck` 符合这个接口的定义，
`Animal` 和 `Duck` 却不符合。Go并没有在这种情况下帮我们做自动转换。

---

- 上一篇：[面向对象编程](./oo.md)
- 下一篇：[指针](./pointers.md)
