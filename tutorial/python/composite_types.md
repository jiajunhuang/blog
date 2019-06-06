# 基本类型

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

Python在标准库里提供了很多种数据结构，包括 dict, list, set, tuple。在 `collections` 这个库里还有更多的数据结构。接下来
我们讲解以下 dict, list, set 和 tuple。

> 所谓标准库，就是Python默认带的一些库，所谓的库，就是一堆Python代码作为一个合集来提供一些功能。

## list

可能你用过其他语言中类似的数据结构，例如数组、链表等等，但是Python中的list与他们不太一样，list我们通常翻译为 列表。

首先Python中的列表是有顺序的，然后Python中的列表里可以包含任意东西，举个例子：

```python
In [1]: [1, "Hello", None, 3.14]
Out[1]: [1, 'Hello', None, 3.14]
```

可以看到，Python中的列表的表示，是在中括号里加入列表中的元素，而列表的元素没有类型要求，而且，Python中的列表是长度不固定
的，举个例子：

```python
In [1]: a = [1, "Hello", None, 3.14]

In [2]: a.append("World")

In [3]: a
Out[3]: [1, 'Hello', None, 3.14, 'World']

In [4]: a.remove(1)

In [5]: a
Out[5]: ['Hello', None, 3.14, 'World']

In [6]: a[0]
Out[6]: 'Hello'

In [7]: a[-1]
Out[7]: 'World'

In [8]: a[:]
Out[8]: ['Hello', None, 3.14, 'World']

In [9]: a[1:]
Out[9]: [None, 3.14, 'World']
```

在上面的例子里，我们还展示了如何用下标来取Python中列表里的值，还有就是如何取其中的一部分，我们把这种操作叫做切片，
例如 `a[:]`是去全部，`a[1:]` 是取第一个元素及其后所有元素，`a[0]` 和 `a[-1]` 分别是取第一个和最后一个，如果下标是正数，
就是从左往右取值，而如果下标是负数，就是从右往左取值。

## tuple

`tuple` 与 `list` 很多方面都是一样的，例如下标取值，切片等等，最大的区别在于，tuple是不可变的，也就是说，里面的元素
不可以替换，也不可以对tuple进行删除或者追加元素的操作：

```python
In [1]: a = (1, "Hello", None, 3.14)

In [2]: a[0]
Out[2]: 1

In [3]: a[-1]
Out[3]: 3.14

In [4]: a[:]
Out[4]: (1, 'Hello', None, 3.14)

In [5]: a[1:]
Out[5]: ('Hello', None, 3.14)

In [6]: a.append("World")
---------------------------------------------------------------------------
AttributeError                            Traceback (most recent call last)
<ipython-input-6-be23f54a6d34> in <module>
----> 1 a.append("World")

AttributeError: 'tuple' object has no attribute 'append'
```

那么为什么不直接使用list呢？答案是省内存，而且有的时候我们就是要保证元素的不可变。

> tuple为什么比list更省内存呢？正是由于list的长度可变，为了容纳更多的元素，list需要在空间不够用的时候申请更多的
> 内存来保存元素，而申请内存是很慢的，因此一般的策略都是申请当前空间的两倍，也就是说，很有可能list里会有很多没有
> 用到，只是等待使用的空间，而tuple不需要，因为他的长度是不可变的，所以创建的时候是几个位置，就一直都是，因此tuple
> 比list 更加省内存。

## dict

dict，很多编程语言里都有这个数据结构，他就是字典（其他语言一般叫哈希表、map）。它的作用也是把Key和Value关联起来，而
Python中的dict又比较特别了，所有的有 `__hash__` 方法的对象，都可以作为key，例如：

```python
In [1]: a = {}

In [2]: a[None] = 1

In [3]: a[1] = 2

In [4]: a[2] = "hello"

In [5]: a["hello"] = 3.14

In [6]: class World:
   ...:     pass
   ...:

In [7]: a[3.14] = World

In [8]: a[World] = World()

In [9]: a
Out[9]:
{None: 1,
 1: 2,
 2: 'hello',
 'hello': 3.14,
 3.14: __main__.World,
 __main__.World: <__main__.World at 0x10c40b2b0>}
```

## set

`set` 就是传说中的集合，Python中，一般用集合来进行各种集合操作，例如取交集，取并集。`set` 需要注意的是初始化方式与 `dict`
很容易搞混，如下：

```python
In [1]: a = {}

In [2]: type(a)
Out[2]: dict

In [3]: b = {1}

In [4]: type(b)
Out[4]: set

In [5]: a = {1: 2}

In [6]: type(a)
Out[6]: dict
```

看到了吗？`set` 和 `dict` 都是使用大括号来初始化，如果给的值是空的或者键值对，那么就会被初始化为 `dict`，如果只给Key，那么
就会被初始化为 `set`。接下来我们看看 `set` 的常见用法：

```python
In [1]: a = {1, 1, 2, 3, 4, 5, 6, 6, 8}

In [2]: b = {"Hello", "World", 3}

In [3]: 2 in a
Out[3]: True

In [4]: "World" in b
Out[4]: True

In [5]: "World" in a
Out[5]: False

In [6]: None in a
Out[6]: False

In [7]: a | b
Out[7]: {1, 2, 3, 4, 5, 6, 8, 'Hello', 'World'}

In [8]: a & b
Out[8]: {3}

In [9]: a
Out[9]: {1, 2, 3, 4, 5, 6, 8}

In [10]: b
Out[10]: {3, 'Hello', 'World'}
```

---

- 上一篇：[基本类型](./basic_types.md)
- 下一篇：[控制流](./flow.md)
