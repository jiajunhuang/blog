# unittest 源代码阅读

又是和单元测试有关的代码的阅读 :smile:

## unittest 源码以及注释

https://github.com/jiajunhuang/cpython/tree/annotation/Lib/unittest

## 简单讲解

unittest里的概念很简单，几乎和文件组织一致

```
tree
.
├── case.py
├── __init__.py
├── loader.py
├── __main__.py
├── main.py
├── result.py
├── runner.py
├── signals.py
├── suite.py
└── util.py
```

其中mock和unittest的test被我移除了。

- `case.py` 就是我们平时继承的 `unittest.TestCase` 所在处

- `loader.py` 加载测试用例

- `result.py` 保存结果的基类

- `runner.py` 实际跑单测的时候直接加载的 `TextTestResult` 和 `TextTestRunner`
    所在地

- `signals.py` 处理相关信号

- `suite.py` TestSuite所在地，TestSuite是TestCase的集合

## 造个小轮子

```python
# coding: utf-8

import importlib
import logging


class TestCase(object):
    def __init__(self, name):
        self.name = name

    def setup(self):
        pass

    def teardown(self):
        pass


class Loader(object):
    def __init__(self):
        self.cases = {}

    def load(self, path):
        module = importlib.import_module(path)
        for test_class_name in dir(module):
            test_class = getattr(module, test_class_name)
            if (
                    isinstance(test_class, type) and
                    issubclass(test_class, TestCase)
            ):
                self.cases.update({
                    test_class: self.find_test_method(test_class) or []
                })

    def find_test_method(self, test_class):
        test_methods = []

        for method in dir(test_class):
            if method.startswith("test_"):
                test_methods.append(
                    getattr(test_class, method)
                )

        return test_methods

    def __iter__(self):
        for test_class, test_cases in self.cases.items():
            yield test_class, test_cases


class Runner(object):
    def __init__(self, path):
        self.path = path

    def run(self):
        loader = Loader()
        loader.load(self.path)

        for test_class, test_cases in loader:
            test_instance = test_class(test_class.__name__)
            test_instance.setup()

            try:
                for test_case in test_cases:
                    test_case(test_instance)
            except:
                logging.exception("error occured, skip this method")

            test_instance.teardown()
```

用上面的 `TestCase` 写个测试用例：

```python
from myunittest import TestCase


class DemoTestCase(TestCase):
    def setup(self):
        print("setup")

    def teardown(self):
        print("teardown")

    def test_normal(self):
        print("test normal function")

    def test_exception(self):
        raise Exception("haha, exception here!")
```

启动文件：

```python
from myunittest import Runner


if __name__ == "__main__":
    runner = Runner("test_demo")
    runner.run()
```

执行结果：

```
$ python main.py
setup
ERROR:root:error occured, skip this method
Traceback (most recent call last):
  File "/home/jiajun/Code/tests/myunittest/myunittest.py", line 64, in run
    test_case(test_instance)
  File "/home/jiajun/Code/tests/myunittest/test_demo.py", line 15, in test_exception
    raise Exception("haha, exception here!")
Exception: haha, exception here!
teardown
```
