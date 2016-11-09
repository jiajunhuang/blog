阅读Flask源码
================

看到很多人推荐新手去阅读Flask，但其实个人并不推荐，因为单说Flask的话，代码量
确实很少，但是Flask完全建立在Werkzeug之上，如果把Werkzeug的代码加起来，代码量
可就不少了，而要完全弄懂Flask，就一定少不了看Werkzeug。

接下来，我们就一起来读一下Flask0.1的代码吧 :)

首先，我们把代码clone下来:

.. code:: bash

    $ git clone https://github.com/pallets/flask
    $ cd flask
    $ git checkout 0.1

打开 ``flask.py`` 你就能看到，其实没多少行代码，按照惯例，找一下看有没有入口
函数，拉到最下面，没有 ``__main__`` ，但是我们发现了这个:

.. code:: python

    # context locals
    _request_ctx_stack = LocalStack()
    current_app = LocalProxy(lambda: _request_ctx_stack.top.app)
    request = LocalProxy(lambda: _request_ctx_stack.top.request)
    session = LocalProxy(lambda: _request_ctx_stack.top.session)
    g = LocalProxy(lambda: _request_ctx_stack.top.g)

``LocalStack`` ??? ``LocalProxy`` ??? 这是什么东西。。。看一下文件，原来是从
Werkzeug导入的::

    from werkzeug import Request as RequestBase, Response as ResponseBase, \
        LocalStack, LocalProxy, create_environ, cached_property, \
        SharedDataMiddleware

嗯，看来终究我们都逃脱不了Werkzeug的魔爪，不过，我们暂时先不理他，我们先来看看
flask的helloworld程序的调用栈是怎样的，这样有利于我们有看代码的目标，而不是拿着
代码毫无目的的随便看，首先我们新建一个virtualenv，然后装上jinja2和werkzeug，并
且把之前0.1版本的 ``flask.py`` 拷贝到那个目录，激活当前venv(这个我就不贴出来了):

.. code:: python

    # coding: utf-8


    from flask import Flask

    app = Flask(__name__)


    @app.route("/")
    def hello():
        import pdb; pdb.set_trace()  # TODO remove it
        return "hello world"


    if __name__ == "__main__":
        app.run()

然后运行，并且新开一个终端或者浏览器的tab，来访问::

    $ python2 main.py
    $ # 另一个终端
    $ http localhost:5000

然后就会执行到断点，此时我们再pdb里打印出当前调用栈:

.. code:: python

    (py2kenv) ➜  src python main.py
    * Running on http://localhost:5000/ (Press CTRL+C to quit)
    > /home/jiajun/Code/python/py2kenv/src/main.py(12)hello()
    -> return "hello world"
    (Pdb) w
    /home/jiajun/Code/python/py2kenv/src/main.py(16)<module>()
    -> app.run()
    /home/jiajun/Code/python/py2kenv/src/flask.py(331)run()
    -> return run_simple(host, port, self, **options)
    /home/jiajun/Code/python/py2kenv/lib/python2.7/site-packages/werkzeug/serving.py(694)run_simple()
    -> inner()
    /home/jiajun/Code/python/py2kenv/lib/python2.7/site-packages/werkzeug/serving.py(659)inner()
    -> srv.serve_forever()
    /home/jiajun/Code/python/py2kenv/lib/python2.7/site-packages/werkzeug/serving.py(499)serve_forever()
    -> HTTPServer.serve_forever(self)
    /usr/lib64/python2.7/SocketServer.py(233)serve_forever()
    -> self._handle_request_noblock()
    /usr/lib64/python2.7/SocketServer.py(290)_handle_request_noblock()
    -> self.process_request(request, client_address)
    /usr/lib64/python2.7/SocketServer.py(318)process_request()
    -> self.finish_request(request, client_address)
    /usr/lib64/python2.7/SocketServer.py(331)finish_request()
    -> self.RequestHandlerClass(request, client_address, self)
    /usr/lib64/python2.7/SocketServer.py(652)__init__()
    -> self.handle()
    /home/jiajun/Code/python/py2kenv/lib/python2.7/site-packages/werkzeug/serving.py(216)handle()
    -> rv = BaseHTTPRequestHandler.handle(self)
    /usr/lib64/python2.7/BaseHTTPServer.py(340)handle()
    -> self.handle_one_request()
    /home/jiajun/Code/python/py2kenv/lib/python2.7/site-packages/werkzeug/serving.py(251)handle_one_request()
    -> return self.run_wsgi()
    /home/jiajun/Code/python/py2kenv/lib/python2.7/site-packages/werkzeug/serving.py(193)run_wsgi()
    -> execute(self.server.app)
    /home/jiajun/Code/python/py2kenv/lib/python2.7/site-packages/werkzeug/serving.py(181)execute()
    -> application_iter = app(environ, start_response)
    /home/jiajun/Code/python/py2kenv/src/flask.py(655)__call__()
    -> return self.wsgi_app(environ, start_response)
    /home/jiajun/Code/python/py2kenv/lib/python2.7/site-packages/werkzeug/wsgi.py(599)__call__()
    -> return self.app(environ, start_response)
    /home/jiajun/Code/python/py2kenv/src/flask.py(626)wsgi_app()
    -> rv = self.dispatch_request()
    /home/jiajun/Code/python/py2kenv/src/flask.py(544)dispatch_request()
    -> return self.view_functions[endpoint](**values)
    > /home/jiajun/Code/python/py2kenv/src/main.py(12)hello()
    -> return "hello world"

我们从下往上看，或者从上往下看，都能推理出我们的调用链::

    app.run[main.py] -> run_simple[flask.py] -> srv.serve_forever[werkzeug] \
    -> HTTPServer.serve_forever -> SocketServer._handle_request_noblock -> \
    SocketServer.process_request -> werkzeug.run_wsgi -> flask.dispatch_request \
    -> app.view_functions[endpoint](**values)[也就是我们的 `def hello`]

哇，好长，不行，如果一个一个深究，恐怕是要很长时间，无论是看代码还是写代码，我们
都要学会抽象，也就是说，我们要把某一部分东西当做一个模块，我们不管他里面是怎么实现
的，只管，他会完成这样的功能。所以为了顺利的看完 flask 的代码，我们不管 werkzeug
以下的调用链，也就是不管 SocketServer, HTTPServer 之类的。简化一下上面的调用链
(其实在这里我们所说的抽象，就是自动忽略细节)::

    app.run[main.py] -> run_simple[flask.py] -> srv.serve_forever[werkzeug] \
    -> werkzeug.run_wsgi -> flask.dispatch_request \
    -> app.view_functions[endpoint](**values)[也就是我们的 `def hello`]

现在我们是不是已经站在一个更高的角度来看待flask了，werkzeug会自动调用好底层的
SocketServer，有请求来了，就会发到werkzeug上，然后就会处理请求。不过，我们说过，
暂时不管werkzeug，所以我们再来进行一次“抽象”::

    app.run[main.py] -> run_simple[flask.py] -> flask.dispatch_request \
    -> app.view_functions[endpoint](**values)[也就是我们的 `def hello`]

哦～原来flask这么简单，没错，因为我们忽略了werkzeug做了那么多事情嘛。
好，接下来我们照着调用链，深入到代码里去，首先我们来看 ``app.run`` :

.. code:: python

    def run(self, host='localhost', port=5000, **options):
        from werkzeug import run_simple
        if 'debug' in options:
            self.debug = options.pop('debug')
        options.setdefault('use_reloader', self.debug)
        options.setdefault('use_debugger', self.debug)
        return run_simple(host, port, self, **options)

哦，又是调用werkzeug。。。哎，反正是监听服务器就对了，好，接下来我们看下一个:

.. code:: python

    def dispatch_request(self):
        """Does the request dispatching.  Matches the URL and returns the
        return value of the view or error handler.  This does not have to
        be a response object.  In order to convert the return value to a
        proper response object, call :func:`make_response`.
        """
        try:
            endpoint, values = self.match_request()
            return self.view_functions[endpoint](**values)
        except HTTPException, e:
            handler = self.error_handlers.get(e.code)
            if handler is None:
                return e
            return handler(e)
        except Exception, e:
            handler = self.error_handlers.get(500)
            if self.debug or handler is None:
                raise
            return handler(e)

嗯，这个函数我没有把注释去掉，是因为注释还是很有说明性的，一般读代码，有注释
都要先看看注释，就像上面的run函数其实也有注释，但是为了篇幅，我把注释删掉了。
真正读代码的时候，请优先读注释。

这个函数，作用就是，看请求是否有匹配到，如果没有的话，就报错。这一步，同时看完了
调用链的后两部分，但是同时为我们引入了新的问题，那就是，WTF is endpoint???

为了调查真相，再一次进入pdb，然后看看这些都是啥。


.. code:: python

    (Pdb) app.view_functions
    {'hello': <function hello at 0x7fb9a3f5ce60>}

也就是说，endpoint是一个字符串，然后对应了一个函数，那endpoint是在哪里设置的呢？

.. code:: python

    def route(self, rule, **options):
        """很多注释，值得一读"""
        def decorator(f):
            self.add_url_rule(rule, f.__name__, **options)
            self.view_functions[f.__name__] = f
            return f
        return decorator

也就是说，flask的url匹配是这样的: ``url -> endpoint -> func`` 要问我为什么？
我也不清楚，可以看看 [#]_ 和 [#]_ ，分别是stackoverflow的一个解释，和我提出的
关于endpoint的作用的问题。如果你能找到endpoint的存在必要性，请务必告知我，谢谢。

看到这里，好像已经看完了flask是怎么工作的了？no，刚刚开始而已。提个问题，
我们知道flask里读取参数是这样的::

    from flask import request
    test = request.args.get("test", None)

如果有多线程存在，他是怎么做到线程安全的？暂且不说线程，我们是怎么从一个导入的
模块里读取我们当前请求的参数的？

还记得上面我们最开始贴的那段代码嘛？

.. code:: python

    # context locals
    _request_ctx_stack = LocalStack()
    current_app = LocalProxy(lambda: _request_ctx_stack.top.app)
    request = LocalProxy(lambda: _request_ctx_stack.top.request)
    session = LocalProxy(lambda: _request_ctx_stack.top.session)
    g = LocalProxy(lambda: _request_ctx_stack.top.g)

原来 flask.request 是这样一个东西，看来我们有必要深入到werkzeug里看看 ``LocalProxy``
是如何工作的了。不过在此之前我猜我们需要先看看 ``LocalStack`` 这个东西，因为上面
的代码显示，他是作为一个参数传到 ``LocalProxy`` 里的。

好吧，接下来我们把werkzeug的代码搞下来，然后搜一下这货在哪:

.. code:: bash

	(py2kenv) ➜  werkzeug git:(master) ack 'class LocalProxy'
	werkzeug/local.py
	254:class LocalProxy(object):
	(py2kenv) ➜  werkzeug git:(master) ack 'class LocalStack'
	werkzeug/local.py
	89:class LocalStack(object):
	(py2kenv) ➜  werkzeug git:(master)

所以接下来我们打开 ``local.py`` 来看看，看到 ``LocalStack`` 的注释，这就是一个
栈(名字其实就说明了它是栈)。看到 ``__init__`` 里最终存储数据的还是 ``Local`` 类，
所以接下来我们来看 ``Local`` 类。

.. code:: python

    class Local(object):
        __slots__ = ('__storage__', '__ident_func__')

        def __init__(self):
            object.__setattr__(self, '__storage__', {})
            object.__setattr__(self, '__ident_func__', get_ident)

        def __iter__(self):
            return iter(self.__storage__.items())

        def __call__(self, proxy):
            """Create a proxy for a name."""
            return LocalProxy(self, proxy)

        def __release_local__(self):
            self.__storage__.pop(self.__ident_func__(), None)

        def __getattr__(self, name):
            try:
                return self.__storage__[self.__ident_func__()][name]
            except KeyError:
                raise AttributeError(name)

        def __setattr__(self, name, value):
            ident = self.__ident_func__()
            storage = self.__storage__
            try:
                storage[ident][name] = value
            except KeyError:
                storage[ident] = {name: value}

        def __delattr__(self, name):
            try:
                del self.__storage__[self.__ident_func__()][name]
            except KeyError:
                raise AttributeError(name)

这个类的作用就是，存东西的时候，实际上存在 ``__storage__`` 里，而它是一个字典。
字典的key是线程id，value是另外的dict。而我们的 ``LocalStack`` 就是，之前说的value
里，以 "stack" 为key，list为value的一个键值对，我们来验证一下:

.. code:: python

    (Pdb) from flask import _request_ctx_stack as test
    (Pdb) test._local
    <werkzeug.local.Local object at 0x7fec126179d0>
    (Pdb) for i in test._local: print i
    (140651991053184, {'stack': [<flask._RequestContext object at 0x7fec11d2d2d0>]})
    (Pdb)

通过 ``LocalProxy`` ，我们执行如下代码的时候::

    from flask import request

    @app.route("/")
    def hello():
        name = request.args.get("name", None)
        if name:
            return "Hello %s" % name
        else:
            return "Hello World"

request总是能正确的指向当前所压入的请求。


小结
------

好了，Flask的代码我们暂时看到这里。下一篇，我准备探索一下，
Gunicorn + Gevent的工作原理(也就是，他们俩使怎样让Flask的代码能够异步执行的)。

.. [#] http://stackoverflow.com/questions/19261833/what-is-an-endpoint-in-flask/19262349#19262349
.. [#] https://www.v2ex.com/t/304941#reply11
