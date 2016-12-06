# Python 的继承

## MRO

Python在2.3之后引入了新的继承方式，同时也带来了新的 `MRO - the C3 method resolution order`。
为啥用英文呢，因为我不知道中文是啥。另外鉴于不是讲历史，所以我们只介绍 `C3 MRO`。

假设我们有一个类，叫C，他们继承自B1, B2, B3, ..., BN

生成MRO的规则如下：

> the linearization of C is the sum of C plus the merge of the
> linearizations of the parents and the list of the parents.

我们用 `L[C]`来表示C的MRO：

> L[C] = C + merge(L[B1], L[B2], ..., L[BN], B1...BN) 如果C是object的话，停止搜索

取 `L[B1]` 中的第一个结果，然后看看是否在后续列表的tail中，如果不在，那就把它加到
结果里(并且从所有列表中移除它)，否则，就看 `L[B2]` 的第一个结果，如此循环。

> 我们来定义一下tail: tail 就是取列表中除去第一个之后的所有元素，例如 [1, 2, 3]
> 的tail就是 [2, 3]；相反，head就是第一个元素，例如上面的例子，head是 1。

接下来我们来举例子：

```python
>>> O = object
>>> class F(O): pass
>>> class E(O): pass
>>> class D(O): pass
>>> class C(D,F): pass
>>> class B(D,E): pass
>>> class A(B,C): pass
```

画成ASCII表示：

```
                          6
                         ---
Level 3                 | O |                  (more general)
                      /  ---  \
                     /    |    \                      |
                    /     |     \                     |
                   /      |      \                    |
                  ---    ---    ---                   |
Level 2        3 | D | 4| E |  | F | 5                |
                  ---    ---    ---                   |
                   \  \ _ /       |                   |
                    \    / \ _    |                   |
                     \  /      \  |                   |
                      ---      ---                    |
Level 1            1 | B |    | C | 2                 |
                      ---      ---                    |
                        \      /                      |
                         \    /                      \ /
                           ---
Level 0                 0 | A |                (more specialized)
                           ---
```

我们先来计算他们各自的MRO：

```
L[O] = O 因为他是object自己
L[D] = D O
L[E] = E O
L[F] = F O
```

B的MRO为：

`L[B] = B + merge(L[D], L[E], DE)` 也就是 `L[B] = B + merge(DO, EO, DE)`

首先我们取 `DO` 里的 `D`，它不在后面串里的tail部分，所以它会被加入到MRO中，
并且会从后面的列表中去掉。

`L[B] = B + D + merge(O, EO, E)`

然后我们进行下一步，选择 `O`，但是发现它存在于后面的列表中，`EO` 包含了它。
所以跳过 `O`。选择 `E`：

`L[B] = B + D + E + merge(O, O)` 也就是 `L[B] = B + D + E + O` 我们写成

```
L[B] = B D E O
```

同样我们来算 `L[C]`：

```
L[C] = C + merge(DO, FO, DF)
    = C + D + merge(O, FO, F)
    = C + D + F + merge(O, O)
    = C D F O
```

然后我们算 `L[A]`：

```
L[A] = A + merge(BDEO,CDFO,BC)
    = A + B + merge(DEO,CDFO,C)
    = A + B + C + merge(DEO,DFO)
    = A + B + C + D + merge(EO,FO)
    = A + B + C + D + E + merge(O,FO)
    = A + B + C + D + E + F + merge(O,O)
    = A B C D E F O
```

所以最后我们调用 `A.a_method` 的寻找顺序就是 `A B C D E F O`。

我们来验证一下：

```python
In [1]: O = object

In [2]: class F(O): pass

In [3]: class E(O): pass

In [4]: class D(O): pass

In [5]: class C(D, F): pass

In [6]: class B(D, E): pass

In [7]: class A(B, C): pass

In [8]: A.__mro__
Out[8]:
(__main__.A,
 __main__.B,
 __main__.C,
 __main__.D,
 __main__.E,
 __main__.F,
 object)

In [9]:
```

## 应用

利用Python的多继承，我们可以使用 `Mixin`。这是啥呢，就是相当于一个补丁，比如:

```python
class Parent(object):
    internal_str = "from internal"

    def foo(self):
        print("foo", self.internal_str)

    def bar(self):
        print("bar", self.internal_str)


class FooMixin(object):
    def foo(self):
        print("FooMixin.foo, and Parent's internal_str is %s" % self.internal_str)


class Son(FooMixin, Parent):
    pass


Son().foo()
Son().bar()
```

```bash
root@arch (master): python test.py
FooMixin.foo, and Parent's internal_str is from internal
bar from internal
```

FooMixin中没有 `internal_str` 但是也能正常打印。

--------------------------------------

[Python 2.3 MRO](https://www.python.org/download/releases/2.3/mro/)
