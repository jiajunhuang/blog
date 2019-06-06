# macOS安装Python

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

## 使用homebrew安装

如果你是 `homebrew` 的用户，那么直接在

> 首先需要学会使用 [终端](https://zh.wikihow.com/%E5%9C%A8Mac%E7%94%B5%E8%84%91%E4%B8%8A%E6%89%93%E5%BC%80%E7%BB%88%E7%AB%AF)

如果你是 `homebrew` 的用户，那么很简单，在终端中输入下面一行即可安装：

> 了解如何 [如何使用homebrew](https://brew.sh/index_zh-cn)

```bash
$ brew install python
```

> 请注意，$ 和 # 都是命令行的提示符，$ 代表你是普通账号，# 代表你是root(管理员)账号，下同。

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
