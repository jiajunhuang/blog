# Hello World!

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

## Windows用户看这里

如果你是Windows用户，首先你需要按 `Win` + `r` 键，输入 `python`，然后回车，接着和下面的 macOS和Linux用户一样的流程了。

## macOS和Linux用户看这里

如果你是macOS或者Linux用户，那么你应该直接打开命令行终端，例如 `gnome-terminal`，输入 `python`，然后回车：

```bash
$ python
Python 3.6.0 (default, Dec 26 2018, 20:58:03)
[GCC 8.2.1 20181127] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> print("Hello World!")
Hello World!
```

> 怎么退出呢？输入 `exit()` 然后回车即可。

## 保存代码然后运行

上面的方式虽然我们也能输出 `Hello World!`，但是有一个问题，就是当我们退出终端之后，源码就没有了。所以我们需要介绍我们平时工作的时候用的第二种模式，即把源代码保存下来。

首先我们使用自己熟悉的编辑器，新建一个文件，保存为 `hello.py`，例如我是用 `Vim`：

```bash
$ vim hello.py
```

输入以下内容：

```python
print("Hello World!")
```

保存好，然后打开终端，输入 `python <hello.py 所在的路径>`，就可以看到输出：

```bash
$ python hello.py
Hello World!
```

我们的第一个Python程序到此就大功告成了！

---

- 上一篇：[安装](./linux.md)
- 下一篇：[基本类型](./basic_types.md)
