# Cython! Python和C两个世界的交叉点

最近一周都没有发博客，因为发现一个好玩的东西---Cython！这周一直在研究这个。
虽然了解C和Python之后学Cython，语法上很简单，但是为了探究它为何能快起来，
还是翻了蛮多的代码并且做了测试的。

## 开始

Cython的资料不多，主要有三个（中文的就不用看了，包括这篇博客，我也不会讲
详细的Cython的语法等，只是大概感慨一下）：

- [官方文档](http://docs.cython.org/en/latest/)
- [官方wiki](https://github.com/cython/cython/wiki)
- [Cython: A Guide for Python Programmers](https://book.douban.com/subject/26250831/)

它的文件扩展名有三种, `.pyx`, `.pxd`, `.pxi`（虽然UNIX下文件扩展名无意义，但是
对人来说还是有意义的）：

- `pyx` 主要是implementation file，实现写在这里面，相当于c里面的 `.c` 文件。
- `pxd` 声明文件，d代表declaration，相当于c里面的头文件的作用。
- `pxi` include files，主要是用来包含其他文件，但是我还没用过。

我们先来看一段代码和性能比较，我选择的性能比较的代码很简单，就是递归计算斐波那契
数列第36位，然后我们来看时间。首先看纯c版本的代码：

```c
#include <time.h>
#include <stdio.h>

int fib(int n) {
    if (n == 0 || n == 1)
        return n;
    return fib(n - 1) + fib(n - 2);
}

int main() {
    clock_t begin = clock();

    fib(36);

    clock_t end = clock();
    double time_spent = (double)(end - begin) / CLOCKS_PER_SEC;
    printf("spent time: %.2fms\n", time_spent * 1000);
}
```

执行时间：

```bash
root@arch fib: cc fib.c && ./a.out
spent time: 136.85ms
root@arch fib: cc fib.c && ./a.out
spent time: 135.86ms
root@arch fib: cc fib.c && ./a.out
spent time: 137.15ms
root@arch fib: cc fib.c && ./a.out
spent time: 135.95ms
```

然后我们看纯Python版本：

```python
In [1]: def fib(n):
   ...:     if n in (0, 1):
   ...:         return n
   ...:     return fib(n - 1) + fib(n - 2)
   ...:

In [2]: %timeit -n 3 fib(36)
3 loops, best of 3: 7.23 s per loop
```

（其实我有测过Java的性能，竟然比C还快，有JIT也不能这样啊！）

接下来看看Cython和Cython包装第一个纯c版本的代码和运行时间：

```cython
cdef extern from "fib.c":
    cdef int fib(int n)


cpdef int cfib(int n):
    return fib(n)


cpdef int cyfib(int n):
    if n == 0 or n == 1:
        return n
    return cyfib(n - 1) + cyfib(n - 2)


cpdef int pure_cython(int n):
    if n in (0, 1):
        return n
    return pure_cython(n - 1) + pure_cython(n - 2)
```

运行时间：

```python
In [1]: from cyfib import cfib, cyfib, pure_cython

In [2]: %timeit -n 3 cfib(36)
3 loops, best of 3: 85.9 ms per loop

In [3]: %timeit -n 3 cyfib(36)
3 loops, best of 3: 69.9 ms per loop

In [4]: %timeit -n 3 pure_cython(36)
3 loops, best of 3: 87.3 ms per loop
```

和纯python版本是不是百倍的速度之差 :doge:

## Cython为何能提速？

Cython的速度来源于何处？我们看到了上面的cython代码，都有标注类型。在Python中
所有的东西都是一个object，在其实现里，就是所有的东西都是一个 `PyObject`，然后
里面都是指针指来指去。每个对象想要确定其类型，都至少要通过对指针进行一次解引用，
看一下PyObject的定义：

```c
typedef struct _object {
    _PyObject_HEAD_EXTRA
    Py_ssize_t ob_refcnt;
    struct _typeobject *ob_type;
} PyObject;
```

其中的 `ob_type` 就是其类型。再例如属性查找，完整的C代码看 [这里](https://github.com/jiajunhuang/cpython/blob/annotation/Objects/object.c#L1028-L1122)
我简化了一下：

```c
/* Generic GetAttr functions - put these in your tp_[gs]etattro slot */

PyObject *
_PyObject_GenericGetAttrWithDict(PyObject *obj, PyObject *name, PyObject *dict) {
    // 初始化变量
    PyTypeObject *tp = Py_TYPE(obj);
    PyObject *descr = NULL;
    PyObject *res = NULL;
    descrgetfunc f = NULL;
    Py_ssize_t dictoffset;
    PyObject **dictptr;

    // 先从MRO中找出描述符
    descr = _PyType_Lookup(tp, name);
    if (descr != NULL) {
        // 如果描述符不为空
        f = descr->ob_type->tp_descr_get;
        if (f != NULL && PyDescr_IsData(descr)) { // 如果是data描述符, 使用 __get__
            return f(descr, obj, (PyObject *)obj->ob_type);
        }
    }
    if (dict != NULL) { // 找对象的 __dict__
        return PyDict_GetItem(dict, name);
    }
    if (f != NULL) {  // 不是data描述符，使用 __get__
        return f(descr, obj, (PyObject *)Py_TYPE(obj));
    }

    if (descr != NULL) {
        return descr;
    }

    raise AttributeError();
}
```

这都是要经过很多步骤的，而对于C这样的静态语言来说，在编译的时候就确定了
类型，如果对于struct这样的结构体，进行属性查找，其实就是计算出某个属性
相对于struct起始位置的内存大小偏移量，然后直接跑过去访问就行。

还有一点消耗，在于Python VM处理时的切换。不信我们来做个测试，写一个fib.py
然后用cython把该文件编译成动态链接库，然后进行测速：

```python
def fib(n):
    if n in (0, 1):
        return n
    return fib(n - 1) + fib(n - 2)
```

为了测试方便，把原文件重命名为 `pyfib.py` （懒的写setup.py，其实可以通过写
Extension来指定编译成啥名儿的）。

```python
In [1]: import fib, pyfib

In [2]: %timeit -n 3 fib.fib(36)
3 loops, best of 3: 2.3 s per loop

In [3]: %timeit -n 3 pyfib.fib(36)
3 loops, best of 3: 7.34 s per loop
```

可以打开cython生成的 `fib.c` 来看看，有好几千行，但是定位到相关代码，首先
就可以看到函数声明：

```
static PyObject *__pyx_pf_3fib_fib(CYTHON_UNUSED PyObject *__pyx_self, PyObject *__pyx_v_n);
...
__pyx_t_3 = __Pyx_PyInt_EqObjC(__pyx_t_1, __pyx_int_0, 0, 0);
```

可以看出Cython对于生成的代码，进行了优化动作，例如先假设n是整型。

## 分析，调试

- `cythonize -a` 会生成一个html文件，里面分析了哪些地方大量调用了Python提供的
C-API
- `cygdb` 我暂时还没用过

## 完结

一个类型系统对于语言的优化来说非常的重要，我一直感慨要是有一门语言能和
Python一样写起来爽，但是速度又能和C一样快就好了！我想我找到了！

不过Cython仍然在发展中，很多地方都还可以改进，例如生成更可读的c代码等。

Cython Rocks!
