# 指针

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

Go语言中也有指针，指针这个词其实用得不好，让人感觉他是指来指去的，反而让人迷糊搞不清楚。不如我们以后碰到 "指针" 这个词语，
就自动把它在脑子里替换成 "地址" 吧。接下来我们开始讲解。首先我们看一个例子：

> 请读者自动在脑子里替换为地址。

```go
package main

import (
	"fmt"
)

func main() {
	var x = 1
	var px *int = &x

	fmt.Printf("%T, %T\n", x, px)
}
```

首先，指针本身并不是什么奇怪的东西，只是表示上与众不同。例如 `*int` 是表示 `int` 型的指针，实际上，有 `*int` 类型的变量，
不过就是一个普通变量而已，他存储了一个 `int` 变量的地址罢了。

Go语言中由于不能对指针进行指针运算，所以指针的内容也就只有这些了：

- `*int` 表示 `int` 型的指针，也就是说上面的 `px` 这个变量，存储的是一个内存地址，而所存储的内存地址所在的那个变量，存储的是一个 `int`
- `&x` 表示取 `x` 的内存地址

---

- 上一篇：[面向接口编程](./interface.md)
- 下一篇：[Goroutine](./goroutine.md)
