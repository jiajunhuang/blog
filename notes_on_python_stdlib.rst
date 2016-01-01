:Date: 12/31/2015

Python标准库阅读笔记
====================

str
---

- str.title()::

    >>> 'Hello world'.title()
    'Hello World'

- str.swapcase()::

    s.swapcase().swapcase() == s

Special Attributes
-------------------

The implementation adds a few special read-only attributes to several object
types, where they are relevant.  Some of these are not reported by the
`dir` built-in function.


- object.__dict__::

   A dictionary or other mapping object used to store an object's (writable)
   attributes.


- instance.__class__::

   The class to which a class instance belongs.


- class.__bases__::

   The tuple of base classes of a class object.


- class.__name__::

   The name of the class or type.

- class.__qualname__::

   The :term:`qualified name` of the class or type.

- class.__mro__::

   This attribute is a tuple of classes that are considered when looking for
   base classes during method resolution.

- class.mro()::

   This method can be overridden by a metaclass to customize the method
   resolution order for its instances.  It is called at class instantiation, and
   its result is stored in :attr:`~class.__mro__`.

- class.__subclasses__::

   Each class keeps a list of weak references to its immediate subclasses.  This
   method returns a list of all those references still alive.
   Example::

      >>> int.__subclasses__()
      [<class 'bool'>]

exceptions
-----------

`exceptions <https://docs.python.org/3/library/exceptions.html>`__

`exception hierarchy <https://docs.python.org/3/library/exceptions.html#exception-hierarchy>`__

re
---

细节请参考 `精通正则表达式 <http://book.douban.com/subject/2154713/>`__ ，里面什么都有，包括环视等

readline
--------

补全大法:

.. code:: python

    >>> import rlcompleter
    >>> import readline
    >>> readline.parse_and_bind("tab: complete")
    >>> readline. <TAB PRESSED>
    readline.__doc__          readline.get_line_buffer(  readline.read_init_file(
    readline.__file__         readline.insert_text(      readline.set_completer(
    readline.__name__         readline.parse_and_bind(
    >>> readline.

struct
-------

把数据打包起来， `看这里 <https://docs.python.org/3/library/struct.html>`__

.. code:: python

    >>> from struct import *
    >>> pack('hhl', 1, 2, 3)
    b'\x00\x01\x00\x02\x00\x00\x00\x03'
    >>> unpack('hhl', b'\x00\x01\x00\x02\x00\x00\x00\x03')
    (1, 2, 3)
    >>> calcsize('hhl')
    8

data types - datetime, collections
-----------------------------------

datetime.date 不带时区信息，datetime.datetime, datetime.time选择性带时区信息

strftime()格式化参数见 `这里 <https://docs.python.org/3/library/datetime.html#strftime-and-strptime-behavior>`__

collections.ChainMap把多个字典链在一起，当作一个用

collections.Counter可以统计一篇文章中每个词语的出现次数

collections.deque对popleft()和popright()进行了优化，比list更高效（在这一方面）

collections.defaultdict提供了key-value的默认value，例如defaultdict(list), value会是list

collections.namedtuple如名字，是有名字的tuple，参数见 `这里 <https://docs.python.org/3/library/collections.html#namedtuple-factory-function-for-tuples-with-named-fields>`__

collections.ordereddict记住插入顺序

userdict, userlist, usersting见文档

collections.abc `好东西 <https://docs.python.org/3/library/collections.abc.html>`__

`heapq <https://docs.python.org/3/library/heapq.html>`__ 堆，文档写的挺有意思，还有优先队列等，建议看一看

bisect 二分查找

array 想念C里的数组吗？

weakref 弱引用

    A weak reference to an object is not enough to keep the object alive:
    when the only remaining references to a referent are weak references,
    garbage collection is free to destroy the referent and reuse its memory
    for something else. However, until the object is actually destroyed the
    weak reference may return the object even if there are no strong
    references to it.

types ``isinstance`` 的时候应该很有用

copy 深拷贝，浅拷贝

pprint ``pretty-print``

reprlib 还不知道具体能用在什么地方

enum 枚举

functional programming
-----------------------

`functools <https://docs.python.org/3/library/functional.html>`__

concurrent
-----------

`Here <https://docs.python.org/3/library/concurrency.html>`__

threading
~~~~~~~~~

Thread-local data is data whose values are thread specific. To manage
thread-local data, just create an instance of `local` (or a subclass)
and store attributes on it:

.. code:: python

    mydata = threading.local()
    mydata.x = 1

the instance's values will be diffrent for separate threads.

`threading <https://docs.python.org/3/library/threading.html>`__ 模块可以一看，
包括锁，可重入锁等，信号，竞争条件等。

multiprocessing
~~~~~~~~~~~~~~~

`Pool <https://docs.python.org/3/library/multiprocessing.html#multiprocessing.pool.Pool>`__
对象可以方便的进行并行计算:

.. code:: python

    from multiprocessing import Pool

    def f(x):
        return x*x

    if __name__ == "__main__":
        with Pool(5) as p:
            print(p.map(f, [1,2,3]))

    [1,4,9]  # result

`Queue <https://docs.python.org/3/library/multiprocessing.html#multiprocessing.Queue>`__
队列操作，api和threading很相似。管道，队列，锁等。

network programming
--------------------

`Here <https://docs.python.org/3/library/ipc.html>`__

development tools - unit test
-----------------------------

`Here <https://docs.python.org/3/library/debug.html>`__

debugging
---------

`Here <https://docs.python.org/3/library/development.html>`__

runtime
-------

`Here <https://docs.python.org/3/library/python.html>`__
