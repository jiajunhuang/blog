:Date: 02/14/2016

weakref
========

读 `PEP205`_ 和 `weakref`__ 模块文档, 了解了一下weakref.

在python中,正常的引用是会增加引用计数的,从而会影响到垃圾回收机制.
而weakref则不会增加.文档上说weakref主要是用作cache:

.. code:: python

    import weakref


    class lazy(object):
        def __init__(self, f):
            self.data = weakref.WeakKeyDictionary()
            self.f = f

        def __get__(self, obj, cls):
            print(self, obj, cls)
            if obj not in self.data:
                self.data[obj] = self.f(obj)
            return self.data[obj]


    class Foo(object):
        @lazy
        def foo(self):
            print("Being lazy in foo")
            return 42

        @lazy
        def bar(self):
            print("Being lazy in bar")
            return 41

    f = Foo()

    print(f.foo, f.bar)
    print(f.foo, f.bar)

下面是运行结果:

.. code:: bash

    $ python fun.py
    <__main__.lazy object at 0x7fa001e487f0> <__main__.Foo object at 0x7fa001e5a550> <class '__main__.Foo'>
    Being lazy in foo
    <__main__.lazy object at 0x7fa001e488d0> <__main__.Foo object at 0x7fa001e5a550> <class '__main__.Foo'>
    Being lazy in bar
    42 41
    <__main__.lazy object at 0x7fa001e487f0> <__main__.Foo object at 0x7fa001e5a550> <class '__main__.Foo'>
    <__main__.lazy object at 0x7fa001e488d0> <__main__.Foo object at 0x7fa001e5a550> <class '__main__.Foo'>
    42 41


分析: ``weakref.WeakKeyDictionary`` 的特性就是当key的引用计数为0时,整个item将会
"失效". 上面代码的工作原理就是每个lazy实例化以后作为key, 当Foo的实例销毁后,
lazy.data里的对应的item就会失效.

不过我个人还是觉得直接存放在实例的dict里的cache好用点(抄自django)::

.. code: python

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

: _`PEP205`: https://www.python.org/dev/peps/pep-0205/
