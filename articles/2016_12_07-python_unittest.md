# Python和单元测试

以前我是不写任何测试的，后来偶尔写单元测试，现在我主动写单元测试 ----- 不得
不承认，测试是有其存在必要性的，要说为什么的话，大概又会引发语言的强弱类型和
是否静态语言之争了吧。

就目前而言，个人认为写单元测试的好处有以下几点：

- 当修改了代码之后，单元测试可以保证API不会发生变化（假设原需求就不需API发生
  变化）。这点可能一般情况下没什么感觉，但是当你去修改前辈留下的代码的时候，
  你就会感谢他写了单元测试，最少让你知道了从功能上，这个函数是干什么的，而且
  能保证你修改了函数内部实现，但是不影响函数功能。

- 写单元测试的时候会回想函数的作用，从而自动对函数进行回想和 `review`。

缺点嘛：耗费时间。单元测试和文档一样，属于非常重要，但是非常耗费时间的工作，
因为要考虑齐全，考虑到的边界条件越多，测试覆盖率越高，程序越可靠，而想这些东
西是很耗费时间精力的。

吐槽完毕，我们来说说目前我知道的几个和测试有关的东西(全程 `Python 3`)。

## [Mock](https://docs.python.org/3/library/unittest.mock.html)

Mock是个好东西呀，遇到测试中出现的不可预知的或者不稳定因素，就用 `Mock` 来代
替。例如查询数据库（当然像目前我们用的MongoDB，由于特别灵活，可以直接在代码里
把相应的collection替换掉），例如异步任务等。举个例子：

```python
import logging
from unittest.mock import Mock

logging.basicConfig(level=logging.DEBUG)


#  code
class ASpecificException(Exception):
    pass


def foo():
    pass


def bar():
    try:
        logging.info("enter function <foo> now")
        foo()
    except ASpecificException:
        logging.exception("we caught a specific exception")


#  unittest
def test_foo():
    foo = Mock(side_effect=ASpecificException())  # noqa

    logging.info("enter function <bar> now")
    bar()
    logging.info("everything just be fine")


if __name__ == "__main__":
    test_foo()
```

运行一下：

```python
root@arch tests: python test_demo.py
INFO:root:enter function <bar> now
INFO:root:enter function <foo> now
INFO:root:everything just be fine
```

duang，一个简单的测试就这么写好了。来，跟我念，`Mock` 大法好呀！

## [doctest](https://docs.python.org/3/library/doctest.html#module-doctest)

doctest属于比较简单的测试，写在 `docstring` 里，这样既能测试用，又能当文档
示例，是在是好用之极啊。缺点是，如果测试太复杂，doctest就显得太臃肿了（例如
如果测试之前要导入一堆东西）。举个例子：

```python
import logging

logging.basicConfig(level=logging.DEBUG)


def foo():
    """A utility function that returns True

    >>> foo()
    True
    """
    return True


if __name__ == "__main__":
    import doctest
    logging.debug("start of test...")
    doctest.testmod()
    logging.debug("end of test...")
```

测试结果：

```bash
root@arch tests: python test_demo.py
DEBUG:root:start of test...
DEBUG:root:end of test...
```

## [unittest](https://docs.python.org/3/library/unittest.html#module-unittest)

这个文档确实有点长，我感觉还是仔细去读一下文档比较好（虽然我也没读完）。

```python
import unittest


class TestStringMethods(unittest.TestCase):
    def setUp(self):
        self.alist = []

    def tearDown(self):
        print(self.alist)

    def test_list(self):
        for i in range(5):
            self.alist.append(i)


if __name__ == '__main__':
    unittest.main()
```

```bash
root@arch tests: python test_demo.py
[0, 1, 2, 3, 4]
.
----------------------------------------------------------------------
Ran 1 test in 0.001s

OK
```

unittest框架配合上Mock，单元测试基本无忧啦。

## [pytest](http://doc.pytest.org/en/latest/)

上面的单元测试跑起来比较麻烦，当然也可以写一个脚本遍历所有的单元测试文件，然
后执行。不过 `pytest` 对unittest有比较好的支持。

pytest默认支持的是 函数 风格的测试，但是我们可以不用这一块嘛（而且很多时候
还是很有用的）。走进项目根目录，输入 `pytest` 就可以啦。它会自动发现 `test_`
开头的文件，然后执行其中 `test_` 开头的函数和 `unittest` 的 `test_` 开头的
方法。

```bash
root@arch tests: pytest
========================================================= test session starts =========================================================
platform linux -- Python 3.5.2, pytest-3.0.5, py-1.4.31, pluggy-0.4.0
rootdir: /root/tests, inifile:
collected 1 items

test_afunc.py .

====================================================== 1 passed in 0.03 seconds =======================================================
root@arch tests:
```

## 总结

编译器没给python做检查，就只有靠我们手写测试了 `:(`
