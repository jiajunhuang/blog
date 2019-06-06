# 函数

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

Python中的函数使用 `def` 来进行声明。Python中的函数可以接受两种类型的参数：普通参数和命名参数(named arguments)。

```python
In [1]: def check_num(num, ignore=False):
   ...:     if ignore:
   ...:         return
   ...:
   ...:     if num < 0:
   ...:         print("num < 0")
   ...:     elif num == 0:
   ...:         print("num == 0")
   ...:     else:
   ...:         print("num > 0")
   ...:

In [2]: check_num()
---------------------------------------------------------------------------
TypeError                                 Traceback (most recent call last)
<ipython-input-2-e2de90bec545> in <module>
----> 1 check_num()

TypeError: check_num() missing 1 required positional argument: 'num'

In [3]: check_num(1)
num > 0

In [4]: check_num(2)
num > 0

In [5]: check_num(2, ignore=True)

In [6]: check_num(2, True)

```

如果 `return` 后面不接返回值，那么默认将会返回 `None`。

---

- 上一篇：[控制流](./flow.md)
- 下一篇：[面向对象编程](./oo.md)
