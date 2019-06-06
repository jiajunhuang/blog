# 模块和包

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

---

如大多数语言一样，Python中也有包和模块的概念。

首先我们来说模块，Python中的一个模块，就是一个 `.py` 文件，模块的名字就是文件的名字，举个例子，我们有一个文件叫 `sayhi.py`，
它的内容是：

```python
def sayhi():
    print("hi")
```

我们可以说，我们有个模块叫做 `sayhi`，这个模块里有个函数叫做 `sayhi`。

而包的概念，就是指把多个模块组合在一起，放在一个文件夹里，与此同时，这个包里一定要有一个 `__init__.py` 的文件，`__init__.py`
可以是空文件。举个例子，有这么一个包：

```bash
$ pwd
$ tree
.
├── __init__.py
├── sayhi.py

1 directory, 2 files
```

就是一个包，如果所在文件夹叫做 `say`，那么这个包的名字就是 `say`。

---

- 上一篇：[面向接口编程](./interface.md)
- 下一篇：[异常处理](./exception.md)
