# Collections 源码阅读

## deque

deque并不是普通的教科书式的双链表实现。经典实现是：

```c
struct node {
    struct node *prev;
    struct node *next;
    void *data;
} node;

struct list {
    struct node *leftnode;
    struct node *rightnode;
} list;
```

这样每个节点都保存了前一个和后一个节点的指针，64位机器上每个节点空间使用量
为（不计算内存对齐）：`8 + 8 + 8 = 24(bytes)`。

而deque的实现为：

```c
#define BLOCKLEN 64

typedef struct BLOCK {
    struct BLOCK *leftlink;
    PyObject *data[BLOCKLEN];
    struct BLOCK *rightlink;
} block;

typedef struct {
    PyObject_VAR_HEAD
    block *leftblock;
    block *rightblock;
    Py_ssize_t leftindex;       /* 0 <= leftindex < BLOCKLEN */
    Py_ssize_t rightindex;      /* 0 <= rightindex < BLOCKLEN */
    size_t state;               /* incremented whenever the indices move */
    Py_ssize_t maxlen;          /* maxlen is -1 for unbounded deques */
    PyObject *weakreflist;
} dequeobject;
```

一个块的内存容量为：`8 + 8 * 64 + 8 = 528(bytes)`，平均到data数组中的每一个
成员，内存使用量为：`528/64 = 8.25`。是不是大大的节省了内存空间？不过对于
`dequeobject` 我没搞懂的是里面的 `state` ，每次操作他都会 `+1` 但是却没有看到
具体用途。

对于deque，`pop popleft append appendleft` 操作都是 `O(1)` 的时间效率，因为
只要一动一下指针就行了。而 `insert` 时间复杂度为 `O(N)`，其底层实现在
`_deque_rotate` 函数中（`_deque_rotate`函数的具体实现方式是每次移动一个块，
直到移动完n个数据为止），insert操作的实现方式是，先把 `insert(index, object)`
index左边的数据rotate到右边，然后插入，然后再把刚才的数据rotate回来。

## ordereddict

ordereddict实现方式为继承dict，然后底层用一个双向链表保存顺序。而双向链表的
实现比较有趣，是：

```python
class _Link(object):
    __slots__ = 'prev', 'next', 'key', '__weakref__'

class OrderedDict(dict):
    def __init__(self):
        # ...
        try:
            self.__root
        except AttributeError:
            self.__hardroot = _Link()
            self.__root = root = _proxy(self.__hardroot)
            root.prev = root.next = root
            self.__map = {}
        # ...
```

其中 `_proxy` 返回的弱引用。我们再看一下 `__setitem__` 操作：

```python
def __setitem__(self, key, value, dict_setitem=dict.__setitem__, proxy=_proxy, Link=_Link):
    if key not in self:
        self.__map[key] = link = Link()
        root = self.__root
        last = root.prev
        link.prev, link.next, link.key = last, root, key
        last.next = link
        root.prev = proxy(link)
    dict_setitem(self, key, value)
```

`self.__map` 里用 `key-value` 形式保存每个key的前后节点。

## namedtuple

`namedtuple` 返回的是 `class tuple` 的子类。所以使用层面上和tuple一致，包括
支持index等。实现原理是，通过上面定义的模板，把typename传进去当做namedtuple
的名字，然后exec生成，放到 `__name__` 命名空间下。

```python
In [1]: from collections import namedtuple

In [2]: a = namedtuple("Point", ['x', 'y'])

In [3]: a(1, 2)
Out[3]: Point(x=1, y=2)

In [4]: print(a(1, 2)._source)
from builtins import property as _property, tuple as _tuple
from operator import itemgetter as _itemgetter
from collections import OrderedDict

class Point(tuple):
    'Point(x, y)'

    __slots__ = ()

    _fields = ('x', 'y')

    def __new__(_cls, x, y):
        'Create new instance of Point(x, y)'
        return _tuple.__new__(_cls, (x, y))

    @classmethod
    def _make(cls, iterable, new=tuple.__new__, len=len):
        'Make a new Point object from a sequence or iterable'
        result = new(cls, iterable)
        if len(result) != 2:
            raise TypeError('Expected 2 arguments, got %d' % len(result))
        return result

    def _replace(_self, **kwds):
        'Return a new Point object replacing specified fields with new values'
        result = _self._make(map(kwds.pop, ('x', 'y'), _self))
        if kwds:
            raise ValueError('Got unexpected field names: %r' % list(kwds))
        return result

    def __repr__(self):
        'Return a nicely formatted representation string'
        return self.__class__.__name__ + '(x=%r, y=%r)' % self

    def _asdict(self):
        'Return a new OrderedDict which maps field names to their values.'
        return OrderedDict(zip(self._fields, self))

    def __getnewargs__(self):
        'Return self as a plain tuple.  Used by copy and pickle.'
        return tuple(self)

    x = _property(_itemgetter(0), doc='Alias for field number 0')

    y = _property(_itemgetter(1), doc='Alias for field number 1')



In [5]:
```

## counter

counter也是继承自dict，主要实现原理是这么一句话：

```python
for elem, count in iterable.items():
    self[elem] = count + self_get(elem, 0)
```

## chainmap

chainmap实现其实就是一个proxy，我们看看 `__init__` 和 `__getitem__` 就知道了：

```python
class ChainMap(MutableMapping):
    def __init__(self, *maps):
        self.maps = list(maps) or [{}]

    def __getitem__(self, key):
        for mapping in self.maps:
            try:
                return mapping[key]
            except KeyError:
                pass
        return self.__missing__(key)
```

其他的几个例如 `UserDict`，`UserString` 就不多说了。
