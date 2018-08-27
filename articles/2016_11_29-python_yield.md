# Python的yield关键字有什么作用？

我是从 `python 2.7` 开始接触的，要说python的各种历史，我还真的说不出来。
不过好在有 `peps` 记录了python的发展和进化。今天我们来挖挖 `yield` 的
历史。

[PEP-255](https://www.python.org/dev/peps/pep-0255/): 在2001年的时候，
yield首次在python中出现。这时候的用法比较简单，只是单纯的作为 `generator`:

```ipython
In [1]: def foo():
   ...:     for i in range(10):
   ...:         yield i
   ...:

In [2]: gen = foo()

In [3]: for i in gen:
   ...:     print(i, end=', ')
   ...:
0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
```

这个时候yield的主要作用应该在于节省内存，例如查询数据库的时候，结果的数量非常
多，此时yield就可以发挥作用，在用到它的时候才产生数据。而不是像list一样保存
所有的数据。

[PEP-380](https://www.python.org/dev/peps/pep-0380/): 这个pep介绍了新的语法
`yield from`。他的作用主要是把对原本的generator产生作用的语句指向另外一个
generator，例如：

```ipython
In [1]: def foo():
   ...:     for i in range(10):
   ...:         yield i
   ...:

In [2]: def bar():
   ...:     yield from foo()
   ...:

In [3]: gen = bar()

In [4]: for i in gen:
   ...:     print(i, end=', ')
   ...:
0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
```

通过 `yield from foo()` 我们能直接从 `bar()` 这个generaotor里拿到 `foo()` 产
生的值。这也是传说中的代理模式的应用。

---------------------------------------------

## yield 和 yield from 有区别吗？什么时候该用哪个？

我们分别用一下：

```python
In [1]: def foo():
   ...:     yield [1, 2, 3]
   ...:

In [2]: def bar():
   ...:     yield from [1, 2, 3]
   ...:

In [3]: for i in foo():
   ...:     print(i)
   ...:
[1, 2, 3]

In [4]: for i in bar():
   ...:     print(i)
   ...:
1
2
3
```

可以看出来，如果直接yield那是会直接把整个 `[1, 2, 3]` 给yield出来。什么时候
用yield什么时候用yield from傻傻分不清怎么办？有一个trick，但是如果你用这个
被主管打了可千万别怪我：

```ipython
In [1]: class Future:
   ...:     pass
   ...:

In [2]: def foo():
   ...:     yield Future()
   ...:

In [3]: def bar():
   ...:     yield from Future()
   ...:

In [4]: for i in foo():
   ...:     pass
   ...:

In [5]: for i in bar():
   ...:     pass
   ...:
-------------------------------------------------------------------
TypeError                         Traceback (most recent call last)
<ipython-input-5-c0dd2fd4aa2d> in <module>()
----> 1 for i in bar():
      2     pass

<ipython-input-3-1c731241676c> in bar()
      1 def bar():
----> 2     yield from Future()
      3

TypeError: 'Future' object is not iterable

In [6]: class Future:
   ...:     def __iter__(self):
   ...:         yield self
   ...:

In [7]: for i in bar():
   ...:     pass
   ...:

In [8]:
```

为什么？因为

```python
def foo():
    yield from bar()
```

几乎相当于

```python
def foo():
    for i in bar():
        yield i
```

但是不同的是，`yield from` 会把对 foo() 作用的 `next`, `.send`, `.throw` 全部
传递给 bar() 但是后面一种形式不会，也就是说后面一种形式把代理这个动作几乎丢掉
了。
