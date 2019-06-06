# Linux上安装Golang

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

首先，打开Linux的终端模拟器，例如 `gnome-terminal`，然后选择对应的发行版，输入如下命令安装：

## Ubuntu/Debian

```bash
$ sudo apt-get update
$ sudo apt-get install golang-go
```

## ArchLinux

```bash
$ sudo pacman -Syu go
```

---

然后输入以下命令来进行确认：

```bash
$ go version
go version go1.12.4 linux/amd64
$ which go
/usr/bin/go
```

---

- 上一篇：这是第一篇
- 下一篇：[Hello, World](./hello_world.md)
