# 从源码看Python的descriptor

descriptor是Python初学者比较不好懂的概念之一，再加上官网文档说的也不是很清楚，
就更容易让人误解了。但是源码总不会说不清楚，所以我们从源码看清楚descriptor到底
在干啥。

```c
PyObject *
_PyObject_GenericGetAttrWithDict(PyObject *obj, PyObject *name, PyObject *dict)
{
    /* Make sure the logic of _PyObject_GetMethod is in sync with
       this method.
    */

    PyTypeObject *tp = Py_TYPE(obj);
    PyObject *descr = NULL;
    PyObject *res = NULL;
    descrgetfunc f;
    Py_ssize_t dictoffset;
    PyObject **dictptr;

    if (!PyUnicode_Check(name)){
        PyErr_Format(PyExc_TypeError,
                     "attribute name must be string, not '%.200s'",
                     name->ob_type->tp_name);
        return NULL;
    }
    Py_INCREF(name);

    if (tp->tp_dict == NULL) {
        if (PyType_Ready(tp) < 0)
            goto done;
    }

    descr = _PyType_Lookup(tp, name);  // 从自身包括继承关系中找descriptor

    f = NULL;
    if (descr != NULL) {
        Py_INCREF(descr);
        f = descr->ob_type->tp_descr_get;
        if (f != NULL && PyDescr_IsData(descr)) {
        // 如果有descriptor并且是data descriptor，那么就调用并返回
            res = f(descr, obj, (PyObject *)obj->ob_type);
            goto done;
        }
    }

    if (dict == NULL) {
        /* Inline _PyObject_GetDictPtr */
        dictoffset = tp->tp_dictoffset;
        if (dictoffset != 0) {
            if (dictoffset < 0) {
                Py_ssize_t tsize;
                size_t size;

                tsize = ((PyVarObject *)obj)->ob_size;
                if (tsize < 0)
                    tsize = -tsize;
                size = _PyObject_VAR_SIZE(tp, tsize);
                assert(size <= PY_SSIZE_T_MAX);

                dictoffset += (Py_ssize_t)size;
                assert(dictoffset > 0);
                assert(dictoffset % SIZEOF_VOID_P == 0);
            }
            dictptr = (PyObject **) ((char *)obj + dictoffset);
            dict = *dictptr;
        }
    }
    if (dict != NULL) {
        Py_INCREF(dict);
        // 从__dict__里查找属性并返回
        res = PyDict_GetItem(dict, name);
        if (res != NULL) {
            Py_INCREF(res);
            Py_DECREF(dict);
            goto done;
        }
        Py_DECREF(dict);
    }

    if (f != NULL) {
        // 非data descriptor，返回
        res = f(descr, obj, (PyObject *)Py_TYPE(obj));
        goto done;
    }

    if (descr != NULL) {
        res = descr;
        descr = NULL;
        goto done;
    }

    PyErr_Format(PyExc_AttributeError,
                 "'%.50s' object has no attribute '%U'",
                 tp->tp_name, name);
  done:
    Py_XDECREF(descr);
    Py_DECREF(name);
    return res;
}

PyObject *
PyObject_GenericGetAttr(PyObject *obj, PyObject *name)
{
    // 调用此函数，往上看
    return _PyObject_GenericGetAttrWithDict(obj, name, NULL);
}
```

所以可以得知：

    - descriptor是作为属性时才会生效，如下demo中，Foo是作为属性时，才会生效
    - 属性查找顺序为 data descriptor -> __dict__ -> descriptor


```python
In [1]: class Foo:
   ...:     def __get__(self, obj, type):
   ...:         print("__get__")
   ...:     def __set__(self, obj, val):
   ...:         print("__set__")
   ...:

In [2]: class Test:
   ...:     bar = Foo()
   ...:

In [3]: t = Test()

In [4]: t.bar
__get__

In [5]: foo = Foo()

In [6]: foo.bar
---------------------------------------------------------------------------
AttributeError                            Traceback (most recent call last)
<ipython-input-6-673afadb7b8e> in <module>()
----> 1 foo.bar

AttributeError: 'Foo' object has no attribute 'bar'
```
