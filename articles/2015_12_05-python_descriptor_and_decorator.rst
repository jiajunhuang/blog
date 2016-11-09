关于python的decorator和descriptor
=================================

一提到python，大家都会说，这门语言很简单。对有编程经验的人来说，python的语法
确实能够在两小时内掌握，但是python这么强大，必然是有一些不那么容易掌握的东西的。

刚学python的时候descriptor，decorator让我感觉有点晕。
不过现在清晰了，于是分享一下我的理解，如有错误，还望不吝赐教。

decorator
~~~~~~~~~

去flask这个web框架的首页一看，就能看到一个例子，例子的内容是这样的：

.. code:: python

    from flask import Flask
    app = Flask(__name__)

    @app.route("/")
    def hello():
        return "Hello World!"

    if __name__ == "__main__":
        app.run()

| 聪明的你一定一眼就看出来这个例子什么意思，不过上面这个\ ``@``\ 是什么意思呢？学过java
| 的朋友可能觉得好像有点眼熟。要注意，我觉得学习一门语言的时候，最大的忌讳就是
| 拿其他语言的语法糖往这门语言上套。

| python中的\ ``@``\ 是一个语法糖，什么叫语法糖？聪明的你此时一定已经打开google搜索了，
| 语法糖的存在主要就是为了少打字（懒惰的程序员，哼）。不过上面这个例子的decorator
| 有点复杂，在此之前，我们先来看一个简单的decorator：

.. code:: python

    def say_hi(func):
        print("hi")

        def wrapper(*args, **kwargs):
            func(args, **kwargs)

        return wrapper


    @say_hi
    def foo(astr):
        print(astr)

    if __name__ == "__main__":
        foo("foo")

执行结果是啥呢？你可以先猜一下再往下看：

.. code:: python

    $ python ~/Code/python/fun.py
    hi
    ('foo',)

| 我们知道在python中，函数是可以当做参数传来传去的，就跟C语言里可以把函数地址
| 传来传去一样。之所以说\ ``@``\ 是一个语法糖，是因为根据结果我们可以推测出，执行
| 函数的时候的foo，绝不是我们定义函数时的foo。在\ `PEP318 <https://www.python.org/dev/peps/pep-0318/>`__
| 有对decorator的详细描述，总结成一句话，就是\ ``@``\ 在上面的例子里相当于
| ``foo = say_hi(foo)``\ ，foo被重新binding成了say\_hi里返回的wrapper,
| 所以我们会按照say\_hi里定义的顺序执行，所以最开始会先打印出"hi"。

| 那flask那个例子是怎么回事呢？能接收参数的decorator。我们就想，\ ``@app.route``
| 自己可以接收参数，但同时又要装饰一下下面的hello函数，这怎么做到？hello和app.route
| 怎么组合？事实上，可以接收参数的decorator是这样子的:

.. code:: python

    def say_sth(arg):
        print(arg)
        def real_decorator(func):
            def wrapper(*args, **kwargs):
                func(*args, **kwargs)
            return wrapper
        return real_decorator

    @say_sth("hi")
    def foo(astr):
        print(astr)

    foo("foo")

执行结果:

.. code:: bash

    $ python ~/Code/python/fun.py
    hi
    foo

另外值得一说，descriptor可以写成class的形式:

.. code:: python

    class Dec(object):
        def __init__(self, *args):
            print("__init__")

        def __call__(self, *args):
            print("__call__")


    @Dec
    def hello():
        print("hello")


    if __name__ == "__main__":
        hello()

运行结果是:

.. code:: bash

    $ python fun.py
    __init__
    __call__

| 为什么这里没有打印出hello呢？按照上面所述，\ ``@``\ 那里相当于\ ``hello = Dec(hello)``
| 当\ ``__main__``\ 里调用hello()的时候，相当于调用\ ``Dec.__call__(hello)``\ 于是就只执行了
| ``print("__call__")``\ 。可能有同学会有疑问，没有\ ``__call__``\ 行吗？当然可以啦，调用了
| ``__call__``\ 是因为hello后面的括号。你可以把hello()去掉，换成print(hello)试试。

我想看到这里应该已经对decorator有一定的了解了。

descriptor
~~~~~~~~~~

| `Python Descriptor
  HOWTO <https://docs.python.org/3/howto/descriptor.html>`__
| 上写的非常清楚，我就不再“抄写”一遍了。

关键点就在:

-  | If the looked-up value is an object defining one of the
   | descriptor methods, then Python may override the default behavior
     and invoke
   | the descriptor method instead. Where this occurs in the precedence
     chain
   | depends on which descriptor methods were defined.

-  | Data and non-data descriptors differ in how overrides are
     calculated with
   | respect to entries in an instance’s dictionary.
   | If an instance’s dictionary has an entry with the same name as a
     data
   | descriptor, the data descriptor takes precedence.
   | If an instance’s dictionary has an entry with the same name as a
     non-data
   | descriptor, the dictionary entry takes precedence.

所以我们来分析当descriptor和decorator组合起来的例子。

组合
~~~~

| 在常见的web框架中，为了避免cpu重复计算，一般都会使用或者自己实现一个
| 缓存机制，避免重复计算同一个东西。

下面的代码来自django1.8:

.. code:: python

    class cached_property(object):
        """
        Decorator that converts a method with a single self argument into a
        property cached on the instance.

        Optional ``name`` argument allows you to make cached properties of other
        methods. (e.g.  url = cached_property(get_absolute_url, name='url') )
        """
        def __init__(self, func, name=None):
            self.func = func
            self.__doc__ = getattr(func, '__doc__')
            self.name = name or func.__name__

        def __get__(self, instance, cls=None):
            if instance is None:
                return self
            res = instance.__dict__[self.name] = self.func(instance)
            return res

| 看到代码，\ ``cached_property``\ 是怎么实现的呢？由descriptor一节我们知道，因为没有
| 定义\ ``__set__``\ 方法，所以如果\ ``obj.__dict__``\ 里有名为foo的属性和名为foo被cached\_property
| 装饰的方法，foo会被优先选择。

| 看上面代码中的\ ``__get__``\ 方法，\ ``res = instance.__dict__[self.name] = self.func(instance)``
| 这一行，如果当前实例不为空而且没有叫做name的属性，就会调用到这一行，在实例的
| ``__dict__``\ 中增加一个叫做name的属性，值为\ ``self.func(instance)``\ ，并且同时返回
| 计算结果，当同一个实例再次取叫做name的属性的值的时候，因为已经在\ ``__dict__``\ 中存在，
| 回直接取\ ``__dict__``\ 中的值，不需要再计算一次，从而达到了cache的目的
| (当然了，django中的这个cached\_property还可以用做其他用途，请自己看注释)。

总结
~~~~

| 好了，简单的总结了“再介绍”了一遍descriptor和decorator，希望能帮到对此感到迷惑的
| 朋友。
