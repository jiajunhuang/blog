# Python中实现单例模式的n种方式和原理

在Python中如何实现单例模式？这可以说是一个经典的Python面试题了。这回我们讲讲实现Python中实现单例模式的n种方式，和它的原理。

## 什么是单例模式

[维基百科](https://en.wikipedia.org/wiki/Singleton_pattern) 中说：

> 单例模式，也叫单子模式，是一种常用的软件设计模式。在应用这个模式时，单例对象的类必须保证只有一个实例存在。许多时候整个系统只需要拥有一个的全局对象，这样有利于我们协调系统整体的行为。比如在某个服务器程序中，该服务器的配置信息存放在一个文件中，这些配置数据由一个单例对象统一读取，然后服务进程中的其他对象再通过这个单例对象获取这些配置信息。这种方式简化了在复杂环境下的配置管理。

在日常编程中，最常用的地方就在于配置类了。举个例子：

```python
from config import config

print(config.SQLALCHEMY_DB_URI)
```

我们当然是希望 `config` 在全局中都是唯一的，那么最简单的实现单例的方式就出来了：使用一个全局变量。

## 实现单例的方式

### 全局变量

我们在一个模块中实现配置类：

```python
# config.py

class Config:
    def __init__(self, SQLALCHEMY_DB_URI):
        self.SQLALCHEMY_DB_URI = SQLALCHEMY_DB_URI

config = Config("mysql://xxx")
```
当然这只是一个例子。真正实现的时候我们肯定不会这样做，因为 `__init__` 太难写了。也许我们可以考虑 Python 3.7 中引入的 `dataclass`:

```python
# config.py

from dataclasses import dataclass

@dataclass
class Config:
        SQLALCHEMY_DB_URI = SQLALCHEMY_DB_URI

config = Config(SQLALCHEMY_DB_URI ="mysql://")
```

通过使用全局变量，我们在所有需要引用配置的地方，都使用 `from config import config` 来导入，这样就达到了全局唯一的目的。

### 使用metaclass

```python
class Singleton(type):
    _instances = {}
    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            cls._instances[cls] = super(Singleton, cls).__call__(*args, **kwargs)
        return cls._instances[cls]

class Config(metaclass=Singleton):
    def __init__(self, SQLALCHEMY_DB_URI):
        self.SQLALCHEMY_DB_URI = SQLALCHEMY_DB_URI
```

metaclass 是类的类，在Python中，instance是实例，class是类，metaclass是类的类。instance是class实例化的结果，而class是metaclass实例化的结果。因此，`Config` 在被实例化的时候，就会调用 `Singleton.__call__`， 所以所有 `Config()` 的地方，最后都会返回同一个对象。

### 重写 `__new__`

```python
class Singleton(object):
    _instance = None
    def __new__(class_, *args, **kwargs):
        if not isinstance(class_._instance, class_):
            class_._instance = object.__new__(class_, *args, **kwargs)
        return class_._instance

class Config(Singleton, BaseClass):
    pass
```

Python中，类实例化的过程是先执行 `Config.__new__` 生成实例，然后执行 `实例.__init__` 进行初始化的，所以通过重写 `__new__` 也可以达到所有调用 `Config()` 的地方都返回同一个对象。

### 使用装饰器

```python
def singleton(class_):
    class class_w(class_):
        _instance = None
        def __new__(class_, *args, **kwargs):
            if class_w._instance is None:
                class_w._instance = super(class_w, class_).__new__(class_, *args, **kwargs)
                class_w._instance._sealed = False
            return class_w._instance
        def __init__(self, *args, **kwargs):
            if self._sealed:
                return
            super(class_w, self).__init__(*args, **kwargs)
            self._sealed = True
    class_w.__name__ = class_.__name__
    return class_w

@singleton
class Config(BaseClass):
    pass
```

使用装饰器也能达到这样的目的，即：有闭包存储了实例，在每次调用 `Config()` 之前，检查该实例，如果已经初始化过，那么就直接返回，否则则调用 `Config()` 进行初始化，然后存储。

## 总结

看完了这四种实现单例的方式，不知道你有没有发现他们都有一个共同点，即：在真正调用 `Config()` 之前进行一些拦截操作，来保证返回的对象都是同一个：

- 全局变量：不直接调用 `Config()`，而使用同一个全局变量
- 使用metaclass：metaclass重写 `__call__` 来保证每次调用 `Config()` 都会返回同一个对象
- 重写 `__new__`：重写 `__new__` 来保证每次调用 `Config()` 都会返回同一个对象
- 使用装饰器：使用装饰器来保证每次调用 `Config()` 都会返回同一个对象
