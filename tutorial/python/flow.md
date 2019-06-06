# 控制流

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

Python中的控制流与其他语言非常接近，接下来我们分别看几个例子：

## if...elif...else

```python
In [1]: def check_num(num):
   ...:     if num < 0:
   ...:         print("num < 0")
   ...:     elif num == 0:
   ...:         print("num == 0")
   ...:     else:
   ...:         print("num > 0")
   ...:

In [2]: check_num(0)
num == 0
```

## while

```python
In [3]: while True:
   ...:     print("infinite loop")
```

## for

```python
In [3]: for i in range(10):
   ...:     print(i)
   ...:
0
1
2
3
4
5
6
7
8
9
```

## continue, break

这两者与其他语言一致，都是用于控制循环里的跳转。

最后，Python没有 `switch` 语句。

---

- 上一篇：[容器类型](./composite_types.md)
- 下一篇：[函数](./function.md)
