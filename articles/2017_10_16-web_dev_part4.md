# Web开发系列(四)：Flask, Tornado和WSGI

如 [上篇](https://jiajunhuang.com/articles/2017_10_14-web_dev_part2.md.html) 所讲，作为一个web服务器，
我们需要建立socket连接，并且监听在指定端口，当请求来到时我们要解析请求的内容，做出判断，给出响应，
然后关闭连接，进行下一个服务。

可是谁也不愿意需要做web开发的时候都从头开始，然后每次都处理这么多事情，那简直是太麻烦了。于是便有了框架，
今天我们讲两个框架，Flask，Tornado。

## Flask

Flask似乎很受欢迎，emm。。。实际上我目前所在公司和上一家都是用Flask，不过我个人并不喜欢Flask的设计，最讨厌两点：

- proxy
- 借助proxy提供伪全局变量

那么，怎么写一个简单的Flask应用呢？在开始之前我们先来讲讲一个web框架大概需要哪些东西。

首先我们需要一个路由器，不是发射Wi-Fi的那个路由器，而是将URL里，不同的URL指向不同的函数或者其他能处理并且作出响应的
东西，我们叫做路由器，或者叫路由。此外我们需要一个东西，包含一些默认的配置，一般情况下都会叫做 "app"，一般都会把router
放在app里。这样我们就可以做出一个简单的web框架。

但是框架之所以叫做框架，是因为它规定了一系列流程，定义好了一系列接口，应用程序员只需要按照给定的接口写出符合接口的代码，
便可以做出web服务来。比如吧，我们决定我们的web框架有这样一系列动作：

```python
def before_request(app, request):
    pass

def handle_request(app, request):
    pass

def after_request(app, request):
    pass
```

我们的应用将会从上至下依次调用函数，那么我们只要实现具体的函数，便可以完成指定的功能。

我们来看一个简单的Flask示例，来自官网：

```python
from flask import Flask
app = Flask(__name__)


@app.route("/")
def hello():
    return "Hello World!"


if __name__ == "__main__":
    app.run()
```

保存并运行，就可以了。Flask的核心之一在于 `@app.route` 这个装饰器，他有一个概念叫做Blueprint，是什么呢？就是把一伙URL
集结在一起，比如，凡是 `/api/say/v1` 开头的URL都放在 `say_bp_v1` 下，那么便可以这样使用：

```python
@say_bp_v1.route("/hello")
def foo():
    pass


@say_bp_v1.route("/world")
def bar():
    pass
```

其作用吧。。。其实是可以少些很多重复的代码，差不多就这样。我们来看看 `@app.route` 的源码：

```python
def route(self, rule, **options):
    def decorator(f):
        endpoint = options.pop('endpoint', None)
        self.add_url_rule(rule, endpoint, f, **options)
        return f
    return decorator
```

所以我们应该追下去看 [add_url_rule](https://github.com/pallets/flask/blob/master/flask/app.py#L1058)，这里我们暂不继续
展开。

我们再看一下Flask中最最核心的东西，proxy：

https://github.com/pallets/flask/blob/master/flask/globals.py#L14

继续追到Werkzeug的代码中看 `LocalProxy`和`LocalStack`:

```python
class LocalStack(object):
    def __init__(self):
        self._local = Local()

    def __release_local__(self):
        self._local.__release_local__()

    def _get__ident_func__(self):
        return self._local.__ident_func__

    def _set__ident_func__(self, value):
        object.__setattr__(self._local, '__ident_func__', value)
    __ident_func__ = property(_get__ident_func__, _set__ident_func__)
    del _get__ident_func__, _set__ident_func__

    def __call__(self):
        def _lookup():
            rv = self.top
            if rv is None:
                raise RuntimeError('object unbound')
            return rv
        return LocalProxy(_lookup)

    def push(self, obj):
        """Pushes a new item to the stack"""
        rv = getattr(self._local, 'stack', None)
        if rv is None:
            self._local.stack = rv = []
        rv.append(obj)
        return rv

    def pop(self):
        """Removes the topmost item from the stack, will return the
        old value or `None` if the stack was already empty.
        """
        stack = getattr(self._local, 'stack', None)
        if stack is None:
            return None
        elif len(stack) == 1:
            release_local(self._local)
            return stack[-1]
        else:
            return stack.pop()

    @property
    def top(self):
        """The topmost item on the stack.  If the stack is empty,
        `None` is returned.
        """
        try:
            return self._local.stack[-1]
        except (AttributeError, IndexError):
            return None

@implements_bool
class LocalProxy(object):
    __slots__ = ('__local', '__dict__', '__name__', '__wrapped__')

    def __init__(self, local, name=None):
        object.__setattr__(self, '_LocalProxy__local', local)
        object.__setattr__(self, '__name__', name)
        if callable(local) and not hasattr(local, '__release_local__'):
            # "local" is a callable that is not an instance of Local or
            # LocalManager: mark it as a wrapped function.
            object.__setattr__(self, '__wrapped__', local)

    def _get_current_object(self):
        """Return the current object.  This is useful if you want the real
        object behind the proxy at a time for performance reasons or because
        you want to pass the object into a different context.
        """
        if not hasattr(self.__local, '__release_local__'):
            return self.__local()
        try:
            return getattr(self.__local, self.__name__)
        except AttributeError:
            raise RuntimeError('no object bound to %s' % self.__name__)

    @property
    def __dict__(self):
        try:
            return self._get_current_object().__dict__
        except RuntimeError:
            raise AttributeError('__dict__')

    # 略略略
```

可以看出（当然，肯定不是这样随随便便看两眼，其实Flask的源码还是可以研究研究的），Flask的本质就是，如果你执行以下代码：

```python
from flask import request

@app.route("/")
def foo():
    print(request.args.get("hello"))
```

request本来是一个导入的object，但实际上从中获取属性或者值时，会从栈顶的ctx里，再取出来，所以他是个代理，哎，不多说了，
等你看过Flask源码之后你就知道这个设计有多么不科学了（虽然很多人都似乎比较喜欢Flask。。。）。

对Flask有兴趣的可以看看我的这篇博客：https://jiajunhuang.com/articles/2016_09_15-flask_source_code.rst.html

## Tornado

Tornado我还是比较喜欢，可惜除了web框架之外，数据库或者其他几乎都是阻塞的。Tornado与Flask的函数形式的写法不一样，Tornado
属于class形式的写法，我认为这个设计比较科学，举个例子：


```python
import tornado.ioloop
import tornado.web


class MainHandler(tornado.web.RequestHandler):
    def get(self):
        self.write("Hello, world")

    def post(self):
        self.write("post :)")


def make_app():
    return tornado.web.Application([
        (r"/", MainHandler),
    ])


if __name__ == "__main__":
    app = make_app()
    app.listen(8888)
    tornado.ioloop.IOLoop.current().start()
```

于是乎，对同一个URL的get，post，put等请求，我们都可以只要实现对应方法便可以，有人要说了，Flask也有class based view，emm，
是有，你去看看源码看看你会想用吗？

Tornado比我们最上面所说的框架所需要的东西还多了什么呢？还多了一个IOLoop，I/O多路复用，此外借助yield，用同步的方式写异步代码，
当然，前提是带上病毒式传播的decorator----只要想写非阻塞代码，那么这个decorator便一加到底。

关于yield是如何把本来回调式的代码连接起来编程同步式的代码，可以看这篇博客：https://jiajunhuang.com/articles/2016_11_29-python_yield.md.html

此外我写了一个类似的Tornado的代码，基于Cython和UVLoop: https://github.com/jiajunhuang/storm 有兴趣的话可以看看。

既然Tornado这么好用，性能又高，为什么好像还没有Flask受欢迎呢？因为Web开发虽然看起来就是分析一下请求，给一下响应，但是远
不是这么简单，还需要和数据库打交道，Tornado自身可以写出非阻塞的代码，但是连数据库，想用ORM的时候却不行，所以也不是特别方便。

因此很多人选择使用Flask或者是Django，然后Gunicorn挡在前面，加上Gevent加持，于是又可以愉快的用写同步的方式写异步。说起Gunicorn，
我们就得说说WSGI了。

## WSGI

WSGI全名Web Server Gateway Interface，Python界web框架百家争鸣，怎么统一一下呢？于是便有了WSGI这种，定义接口，而非定义实现的方式。
具体需要看看这里：https://www.python.org/dev/peps/pep-3333/#specification-details 实现了对应的接口，便可以接入针对WSGI的
应用例如Gunicorn。

讲完，收工 :)
