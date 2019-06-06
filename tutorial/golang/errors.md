# 错误处理

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

错误处理一直是Go语言中饱受诟病的一点，类似这样：

```go
func OpenFile(path string) (*os.File, error) {
    if f, err := os.Open(path); err != nil {
        return nil, err
    } else {
        return f, nil
    }
}
```

> 当然，上面的例子可以写的更简单，为了展示异常处理，因此写成这样。

而Python中是这样写：

```python
try:
    open(path)
except FileNotFoundError:
    pass
```

表面上看不出什么太大的区别，但实际上当调用层次深了之后，`try...except...` 的这种形式会比Go语言返回error的形式简单明了的多。

---

- 上一篇：[流程控制](./flow.md)
- 下一篇：[面向对象编程](./oo.md)
