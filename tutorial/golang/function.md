# 函数

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

Go语言的函数声明非常简单，我们看一个例子：

```go
func Add(a int, b int) int {
    return a + b
}
```

上面就是Go语言的函数的声明，Add是函数名，a和b是Add函数所需要的两个参数，他们的类型分别是int，int。Add函数的返回值是int。
而Add函数的函数体，也就是实现则是 `return a + b`。

> 注意，上面的Add函数中的参数类型是一样的，因此也可以写成 `func Add(a, b int)`

而Go语言中还有一类函数叫做 "匿名函数"，他们没有名字，但是可以赋值给变量，当然也可以不赋值直接使用，例如：

```go
func (a, b int) int {
    return a + b
}
```

这就是一个匿名函数，但是你不能直接使用它，因为没有办法通过名字去调用那个函数，因此，
你有两种选择：

- 把这个匿名函数赋值给一个变量

```go
var x = func(a, b int) int { return a + b }
x(1, 2)
```

- 把这个匿名函数直接当做参数传给 `go` 关键字

```go
go func(a, b int) {
    fmt.Println(a + b)
}()
```
---

- 上一篇：[容器类型](./composite_types.md)
- 下一篇：[流程控制](./flow.md)
