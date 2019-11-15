# Python中优雅的处理文件路径

写代码（尤其是脚本的时候）经常会遇到要处理文件路径的问题，通常有这么几个考量：

- 简单易懂
- 跨平台（Unix使用 `/` 而Windows使用 `\` ）
- 容易拼接

个人此前最常用的就是 `os.path.join` 了，现在介绍一种更直观更高端的方式，那就是pathlib里的 `Path`：

```python
In [1]: from pathlib import Path  # 首先导入Path

In [2]: current_path = Path(".")  # 获取当前路径

In [3]: current_path.home()  # 打印家目录的路径
Out[3]: PosixPath('/home/jiajun')

In [4]: current_path.resolve()  # 获取绝对路径
Out[4]: PosixPath('/home/jiajun/Code/blog')

In [5]: current_path.glob("*.py")  # 使用glob来匹配文件或者文件夹
Out[5]: <generator object Path.glob at 0x7f88b533a840>

In [6]: [i for i in current_path.glob("*.py")]
Out[6]:
[PosixPath('config.py'),
 PosixPath('gen_catalog.py'),
 PosixPath('utils.py'),
 PosixPath('models.py')]

In [7]: fake_path = current_path / "helloworld"  # 使用 / 来增加层级，是不是比 os.path.join 好看些

In [8]: fake_path.resolve()
Out[8]: PosixPath('/home/jiajun/Code/blog/helloworld')

In [9]: fake_path.exists()  # 判断是否存在
Out[9]: False

```

通过上述操作可以看出来，`Path` 的操作比 `os.path` 中的操作简单明了的多，上面只是其中一部分操作，
全面的 `os` 和 `os.path` 中的操作 和 `pathlib` 中的操作的对比表格 [在这里](https://docs.python.org/3/library/pathlib.html#correspondence-to-tools-in-the-os-module)

---

参考资料：

- [官方文档](https://docs.python.org/3/library/pathlib.html)
