# 流程控制

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

常见的语言中，都有命令分支结构，我们今天依次看看Go语言的流程控制。

## if...else...

这是最常见的流程控制，Go语言的 `if...else...` 与其他语言并没有什么太大的不同，唯一需要注意的地方是它不需要括号：

```go
if x > 1 {
    fmt.Println("x > 1")
}
```

## switch...case...

Go语言中的 `switch...case...` 与C语言非常像，不同的地方在于，每一条case语句默认是带了 `break` 的，也就是说，当前的 `case`
执行完成之后，整个 `switch` 就会退出：

```go
switch x {
    case x == 1:
        fmt.Println("x == 1")
    case x == 2:
        fmt.Println("x == 2")
    default:
        fmt.Println("x != 1 && x != 2")
}
```

`default` 分支就是用于所有 `case` 都没有匹配时执行的，我们可以不提供 `default` 分支。

## for

Go语言的 `for` 循环也是循规蹈矩，唯一需要注意的地方在于没有括号：

```go
for i := 0; i < 10; i++ {
    fmt.Println(i)
}
```

> 注意，Go语言里没有while，也没有 do...while...

## continue和break

`continue` 和 `break` 就是用于控制 `for` 循环的，例如：

```go
for {
    if x == 1 {
        break
    }

    if x == 2 {
        continue
    }
}
```

## goto

一般我们不用 `goto`。Go语言里的 `goto` 与C语言的一致，我们需要提供一个标签，所以我们也可以用 `goto` 写一个死循环：

```go
loop:
    goto loop
```

它与下面这样是等效的：

```go
for {

}
```

---

- 上一篇：[函数](./function.md)
- 下一篇：[错误处理](./errors.md)
