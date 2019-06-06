# 异常处理

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

Python中的异常使用用法如下：

```python
try:
    open("xxxxxxxxx")
except FileNotFoundError as e:
    print(e)
```

---

- 上一篇：[模块和包](./module_and_package.md)
- 下一篇：[I/O处理](./io.md)
