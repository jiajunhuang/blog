# Python中用tuple作为key

其实当我得到这个结论的时候，我觉得很简单很正常，但是我第一次看到这个用法时，我觉得很新奇：

```python
class Solution(object):
    def lenLongestFibSubseq(self, A):
        index = {x: i for i, x in enumerate(A)}
        longest = collections.defaultdict(lambda: 2)

        ans = 0
        for k, z in enumerate(A):
            for j in xrange(k):
                i = index.get(z - A[j], None)
                if i is not None and i < j:
                    cand = longest[j, k] = longest[i, j] + 1
                    ans = max(ans, cand)

        return ans if ans >= 3 else 0
```

注意倒数低三行代码中的 `longest[j, k]`，居然可以这样用？一开始我以为是 `defaultdict` 提供的功能，因此我去翻了一下
`collections` 的实现，在 `_collectionsmodule.c` 里，但是没有找到诸如 `__getitem__` 的方法，但是我看到 `defaultdict`
的定义如下：

```c
typedef struct {
    PyDictObject dict;
    PyObject *default_factory;
} defdictobject;

```

当 `dict` 里没有key时，就会去调用 `default_factory` 拿到默认的key，这么说来，`dict` 也支持这种用法？我试试：

```bash
>>> a = {}
>>> a[1, 2] = 3
>>> a[1,2]
3
>>> a
{(1, 2): 3}
```

还真的是。看到这里，我大概知道为啥了，因为 `1, 2` 其实是一个tuple，tuple是不可变对象，是hashable的，来验证一下。但是我没有
找到 `__getitem__` 的实现，不过我找到了 `key in dict` 这个操作的实现：

```c
/* Return 1 if `key` is in dict `op`, 0 if not, and -1 on error. */
int
PyDict_Contains(PyObject *op, PyObject *key)
{
    Py_hash_t hash;
    Py_ssize_t ix;
    PyDictObject *mp = (PyDictObject *)op;
    PyObject *value;

    if (!PyUnicode_CheckExact(key) ||
        (hash = ((PyASCIIObject *) key)->hash) == -1) {
        hash = PyObject_Hash(key);
        if (hash == -1)
            return -1;
    }
    ix = (mp->ma_keys->dk_lookup)(mp, key, hash, &value);
    if (ix == DKIX_ERROR)
        return -1;
    return (ix != DKIX_EMPTY && value != NULL);
}

/* Internal version of PyDict_Contains used when the hash value is already known */
int
_PyDict_Contains(PyObject *op, PyObject *key, Py_hash_t hash)
{
    PyDictObject *mp = (PyDictObject *)op;
    PyObject *value;
    Py_ssize_t ix;

    ix = (mp->ma_keys->dk_lookup)(mp, key, hash, &value);
    if (ix == DKIX_ERROR)
        return -1;
    return (ix != DKIX_EMPTY && value != NULL);
}

```

可以看出来，只要 `hash = PyObject_Hash(key);` 这一步是成功的，就可以。也就是说，只要能算出hash值即可，那么我们可以来试试
把一个class，重写 `__hash__` 方法，来看看是否可以做key：

```python
class A(list):
    pass

a = A()

d = {}

d[a] = 1

```

原本这样是会报错的：

```bash
$ python test.py 
Traceback (most recent call last):
  File "test.py", line 8, in <module>
    d[a] = 1
TypeError: unhashable type: 'A'

```

因为list并不是immutable的，所以没法计算hash值，但是我们可以给他加上方法：

```python
class A(list):
    def __hash__(self):
        return 1

a = A()

d = {}

d[a] = 1

```

这样就可以了。

好了，到这里终于弄明白为啥可以这样做了。那么 `dict[1, 2, 3]` 这种写法的好处是啥呢？那就是，原本如果我们遇到了二位矩阵、
三维矩阵时，我们得弄一个矩阵：`[[1, 2, 3], [4, 5, 6]]` 然后访问其中的数据时，就这样访问：`a[1][2]`，但是有了这个特性
之后，我们可以像最开始的代码那样，结合 `defaultdict`，然后 `a[1, 2]` 来访问和赋值，感觉方便了不少。
