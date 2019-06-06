# 面向接口编程

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
- [I/O处理](./io.md)
- [进程和线程](./process_and_thread.md)
- [常见库源码阅读](./source_code_analysis.md)

---

Python中没有对接口的直接支持，但是实际上面向接口编程是一种编程思维、编程范式，与语言是否有关键字支持无关。Python中面向
接口编程我们一般使用 [abc](https://docs.python.org/3/library/abc.html)：

```python
In [1]: import abc

In [2]: class Bird(abc.ABC):
   ...:     @abc.abstractmethod
   ...:     def fly(self):
   ...:         pass
   ...:

In [3]: class Parrot(Bird):
   ...:     pass
   ...:

In [4]: Parrot().fly()
---------------------------------------------------------------------------
TypeError                                 Traceback (most recent call last)
<ipython-input-4-beb5d307db79> in <module>
----> 1 Parrot().fly()

TypeError: Can't instantiate abstract class Parrot with abstract methods fly

In [5]: class Parrot(Bird):
   ...:     def fly(self):
   ...:         print("flying")
   ...:

In [6]: Parrot().fly()
flying
```

可以看到，如果我们继承了 `Bird` 却没有实现接口的话，直接调用就会报错。

---

- 上一篇：[面向对象编程](./oo.md)
- 下一篇：[模块和包](./module_and_package.md)
