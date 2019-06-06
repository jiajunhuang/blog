# Linux安装Python

## 目录

- 安装
    - [Linux](./linux.md)
    - [Windows](./windows.md)
    - [macOS](./macos.md)
- [Hello World!](./hello_world.md)
- [基本类型](./basic_types.md)
- [容器类型](./composite_types.md)
- [控制流](./flow.md)
- [函数](./function.md)
- [面向对象编程](./oo.md)
- [面向接口编程](./interface.md)
- [模块和包](./module_and_package.md)
- [异常处理](./exception.md)

绝大部分Linux都会自带Python，但是有可能带的是Python 2，我们需要安装Python 3。

## ArchLinux

直接在终端执行：

```bash
$ sudo pacman -S python
```

## Ubuntu/Debian

在终端执行：

```bash
$ sudo apt-get install python3
```

然后在 `~/.bashrc` 的最后添加一行：`alias python=python3`，当然，也可以选择不添加，那么你需要记住，
在本教程的后续篇章中，只要碰到了在命令行使用 `python`，那么你需要自己替换成 `python3`。

## CentOS

在终端执行：

```bash
$ sudo yum install python3
```

然后在 `~/.bashrc` 的最后添加一行：`alias python=python3`，当然，也可以选择不添加，那么你需要记住，
在本教程的后续篇章中，只要碰到了在命令行使用 `python`，那么你需要自己替换成 `python3`。

## 验证

在终端输入 `python --version` 如果打印出了Python的版本是 `3.x` 那么就是ok的！

```bash
$ python --version
Python 3.6.0
```

---

- 上一篇：这是第一篇
- 下一篇：[基本类型](./basic_types.md)
