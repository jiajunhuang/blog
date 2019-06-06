# Hello, World!

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

`Hello, World!` 是我们学习编程语言的第一个步骤，通过这个简单的程序，我们可以大概看到Go语言是什么样子的，以及如何使用
Go语言的工具，如何运行等等。

首先打开我们最常用的编辑器，或者如果你使用的是IDE(例如Goland)的话，然后新建一个文件，我们把它保存为 `hello_world.go`，
然后我们输入下面的内容：

```go
package main

import (
    "fmt"
)

func main() {
    fmt.Println("Hello, World!")
}
```

然后我们打开终端，编译我们的第一个程序：

```bash
$ go build hello_world.go
```

然后你会看到代码所在目录下多了一个可执行文件(使用ls查看):

```bash
$ ls
hello_world.go hello_world
$ ./hello_world
Hello, World!
```

我们使用 `./hello_world` 来执行它(Windows是 `.\hello_world.exe`)，就会发现输出了 `Hello, World!`。

我们简单的看看这个例子：

- `package main` 是说明，这个文件属于 `main` 这个包。包的意思是，将一堆Go文件组合在一起，提供一些功能。
- `import ("fmt")` 这三行，是导入了 `fmt` 这个包，刚才我们说了，一堆Go文件组合在一起叫做一个包，`fmt` 包提供的功能就是打印字符串相关。
- `func main {}` 声明了一个函数，函数的名字是 `main`，以后我们声明函数都是用 `func xxx {}` 这样的形式。
- `fmt.Println("Hello, World!")` 是调用了一个函数，函数的名字是 `Println`，它属于 `fmt` 这个包。我们导入了 `fmt` 之后，
如果想使用它里面提供的函数，就是这样子使用。而圆括号里 `("Hello, World!")` 则是传递给这个函数的参数，它可以是其它值，例如 `"World, Hello!"`

---

## 练习题

1. 写一个程序，执行之后输出 `World, Hello!`，即执行时，在终端里有如下效果：

```bash
$ ./hello_world
World, Hello!
```

2. 尝试修改 `main` 函数，把函数名改成其他的，再编译试试看效果。

---


- 上一篇：[安装](./installation_linux.md)
- 下一篇：[Go语言简介](./intro.md)
