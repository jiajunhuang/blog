# contextlib代码阅读

首先我们要知道 [with协议](https://www.python.org/dev/peps/pep-0343/)。
然后我们看看 `class _GeneratorContextManager(ContextDecorator, AbstractContextManager):`：

```python
class _GeneratorContextManager(ContextDecorator, AbstractContextManager):
    def __init__(self, func, args, kwds):
        self.gen = func(*args, **kwds)
        self.func, self.args, self.kwds = func, args, kwds

    def __enter__(self):
        try:
            return next(self.gen)
        except StopIteration:
            raise RuntimeError("generator didn't yield") from None


def contextmanager(func):
    def inner(*args, **kwds):
        return _GeneratorContextManager(func, args, kwds)
    return inner
```

所以当我们调用的时候，例如：

```
In [1]: from contextlib import contextmanager

In [2]: @contextmanager
   ...: def foo(arg1, arg2, kwd=None):
   ...:     print('enter function foo with args and kwargs: %s, %s, %s' % (arg1, arg2, kwd))
   ...:     yield None
   ...:     print('leave function foo')
   ...:

In [3]: with foo(1, 2, "hello") as f:
   ...:     print('after yield, execute some ops')
   ...:
enter function foo with args and kwargs: 1, 2, hello
after yield, execute some ops
leave function foo
```

首先contextmanager将原本的foo函数替换成 `foo=contextmanager(foo)`，其实就相当于
`foo = inner`，当执行 `with foo(1, 2, "hello") as f`的时候，首先执行 `foo(1, 2, "hello")`
相当于执行 `inner(1, 2, "hello")`，也就是执行 `_GeneratorContextManager(foo, 1, 2, kwd="hello")`，
然后会执行 `_GeneratorContextManager` 的 `__enter__` 返回给 `f`。这就是
这个decorator的作用。用代码里的注释来解释：

```
Typical usage:

    @contextmanager
    def some_generator(<arguments>):
        <setup>
        try:
            yield <value>
        finally:
            <cleanup>

This makes this:

    with some_generator(<arguments>) as <variable>:
        <body>

equivalent to this:

    <setup>
    try:
        <variable> = <value>
        <body>
    finally:
        <cleanup>
```

此外 contextlib 里还有 `closing`, `redirect_stderr` 等几个帮助函数，其实现原理
都和上面类似，[打开代码](https://github.com/jiajunhuang/cpython/blob/annotation/Lib/contextlib.py)
看看就知道了:)
