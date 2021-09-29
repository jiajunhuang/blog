# gevent不是黑魔法(一): greenlet 实现

最近粗略的读了一下 gevent 的实现，毕竟用了这么多年的 gevent，之前没去看过怎么实现，心里没底呀。

gevent 是基于 greenlet 之上，结合 eventloop 实现的一套 Python 协程库，通过 gevent monkey patch，可以用同步的方式写出
异步的 Python 代码，这就和写 Go 一样，完全不需要担心阻塞的问题，但是坊间流传 gevent 是黑魔法，我以前也不懂，就跟着瞎
起哄，看了以后，发现其实并不黑魔法。

## greenlet 定义

greenlet 是 Python 中一个协程的实现，核心是用C写的，代码在 [这里](https://github.com/python-greenlet/greenlet) 。主要实现
就在 `src/greenlet/greenlet.c` 和 `src/greenlet/greenlet.h` 中，其中协程切换是用汇编写的，在 `src/greenlet/platform/` 里。

首先我们看一个 greenlet 长什么样：

```c
typedef struct _greenlet {
    PyObject_HEAD
    char* stack_start;  // 栈顶部
    char* stack_stop;  // 栈底部
    char* stack_copy; // 栈在堆中的位置
    intptr_t stack_saved;  // 栈在堆中保存的大小
    struct _greenlet* stack_prev;
    struct _greenlet* parent; // 父 greenlet
    PyObject* run_info;  // 执行相关的信息，是一个字典
    struct _frame* top_frame; // 栈顶的帧, struct _frame 定义在 Python.h 里
    int recursion_depth; // 递归的深度
    PyObject* weakreflist;
#if PY_VERSION_HEX >= 0x030700A3
    _PyErr_StackItem* exc_info;
    _PyErr_StackItem exc_state;
#else
    PyObject* exc_type;
    PyObject* exc_value;
    PyObject* exc_traceback;
#endif
    PyObject* dict;
#if PY_VERSION_HEX >= 0x030700A3
    PyObject* context;
#endif
#if PY_VERSION_HEX >= 0x30A00B1
    CFrame* cframe;
#endif
} PyGreenlet;
```

其中的 `stack_start` 和 `stack_stop` 分别指向协程的栈的顶部与底部，可以参考代码中的示意图：

```
A PyGreenlet is a range of C stack addresses that must be
saved and restored in such a way that the full range of the
stack contains valid data when we switch to it.

Stack layout for a greenlet:

               |     ^^^       |
               |  older data   |
               |               |
  stack_stop . |_______________|
        .      |               |
        .      | greenlet data |
        .      |   in stack    |
        .    * |_______________| . .  _____________  stack_copy + stack_saved
        .      |               |     |             |
        .      |     data      |     |greenlet data|
        .      |   unrelated   |     |    saved    |
        .      |      to       |     |   in heap   |
 stack_start . |     this      | . . |_____________| stack_copy
               |   greenlet    |
               |               |
               |  newer data   |
               |     vvv       |

```

`stack_stop` 在上面，`stack_start` 在下面，是因为 Linux 进程模型中，栈是从高地址往低地址增长的。

至此我们就知道了，Python的协程实现也是一样，定义一个结构体，这个结构体用来保存当前执行的代码的各种信息，例如栈、寄存器信息，
然后如果当前协程将要执行一个阻塞的调用，就想办法切换到另外一个协程去执行。接下来我们就要看看纯粹的 greenlet 需要如何使用。

## PyGreenlet 初始化

刚才说了，greenlet 是用 C 语言实现的，而我们实际使用的时候，是通过 Python 代码调用的，我们来看看他们是如何对应的，Python
有提供一套 C API，用于在 C 语言中定义类型，我们来看看 PyGreenlet：

```c
PyTypeObject PyGreenlet_Type = {
    PyVarObject_HEAD_INIT(NULL, 0)
    "greenlet.greenlet", /* tp_name */
    sizeof(PyGreenlet),  /* tp_basicsize */
    0,                   /* tp_itemsize */
    /* methods */
    (destructor)green_dealloc, /* tp_dealloc */
    0,                         /* tp_print */
    0,                         /* tp_getattr */
    0,                         /* tp_setattr */
    0,                         /* tp_compare */
    (reprfunc)green_repr,      /* tp_repr */
    &green_as_number,          /* tp_as _number*/
    0,                         /* tp_as _sequence*/
    0,                         /* tp_as _mapping*/
    0,                         /* tp_hash */
    0,                         /* tp_call */
    0,                         /* tp_str */
    0,                         /* tp_getattro */
    0,                         /* tp_setattro */
    0,                         /* tp_as_buffer*/
    Py_TPFLAGS_DEFAULT | Py_TPFLAGS_BASETYPE |
        GREENLET_GC_FLAGS, /* tp_flags */
    "greenlet(run=None, parent=None) -> greenlet\n\n"
    "Creates a new greenlet object (without running it).\n\n"
    " - *run* -- The callable to invoke.\n"
    " - *parent* -- The parent greenlet. The default is the current "
    "greenlet.",                        /* tp_doc */
    (traverseproc)GREENLET_tp_traverse, /* tp_traverse */
    (inquiry)GREENLET_tp_clear,         /* tp_clear */
    0,                                  /* tp_richcompare */
    offsetof(PyGreenlet, weakreflist),  /* tp_weaklistoffset */
    0,                                  /* tp_iter */
    0,                                  /* tp_iternext */
    green_methods,                      /* tp_methods */
    0,                                  /* tp_members */
    green_getsets,                      /* tp_getset */
    0,                                  /* tp_base */
    0,                                  /* tp_dict */
    0,                                  /* tp_descr_get */
    0,                                  /* tp_descr_set */
    offsetof(PyGreenlet, dict),         /* tp_dictoffset */
    (initproc)green_init,               /* tp_init */
    GREENLET_tp_alloc,                  /* tp_alloc */
    green_new,                          /* tp_new */
    GREENLET_tp_free,                   /* tp_free */
    (inquiry)GREENLET_tp_is_gc,         /* tp_is_gc */
};
```

这个，我们需要参考一下官方文档：https://docs.python.org/3/c-api/typeobj.html ，把几个重要的方法看一看。对于一个声明的
类型而言，最重要的几个方法，无非是 `__new__`, `__init__` 和实例提供的方法，参考文档，我们可以看到:

- `__new__` 对应的是 `tp_new`，在上述结构体中，就是 `green_new`
- `__init__` 对应的是 `tp_init`，在上述结构体中，就是 `green_init`
- 此外，`tp_methods` 声明了类型中的方法，在上述结构体中，就是 `green_methods`

我们来看官方给的示例，然后顺着例子，我们去看 greenlet 对象是如何初始化的，以及它提供了哪些方法。

```python
>>> from greenlet import greenlet

>>> def test1():
...     print("[gr1] main  -> test1")
...     gr2.switch()
...     print("[gr1] test1 <- test2")
...     return 'test1 done'

>>> def test2():
...     print("[gr2] test1 -> test2")
...     gr1.switch()
...     print("This is never printed.")

>>> gr1 = greenlet(test1)
>>> gr2 = greenlet(test2)
>>> gr1.switch()
[gr1] main  -> test1
[gr2] test1 -> test2
[gr1] test1 <- test2
'test1 done'
>>> gr1.dead
True
>>> gr2.dead
False
```

看起来，实例化的时候，就是和普通模块一致，Python 中，实例化会先调用 `__new__`，然后把对象传入 `__init__` 调用，这样就可以
得到一个已经初始化好了的对象，我们来看看 `green_new`：

```c
// 相当于 __new__
static PyObject*
green_new(PyTypeObject* type, PyObject* args, PyObject* kwds)
{
    PyObject* o =
        PyBaseObject_Type.tp_new(type, ts_empty_tuple, ts_empty_dict);
    if (o != NULL) {
        if (!STATE_OK) {
            Py_DECREF(o);
            return NULL;
        }
        Py_INCREF(ts_current);
        ((PyGreenlet*)o)->parent = ts_current; // parent 默认等于当前的greenlet
#if GREENLET_USE_CFRAME
        ((PyGreenlet*)o)->cframe = &PyThreadState_GET()->root_cframe;
#endif
    }
    return o;
}
```

注意，`STATE_OK` 是一个宏：

```c
#define STATE_OK                                          \
    (ts_current->run_info == PyThreadState_GET()->dict || \
     !green_updatecurrent())

/* Strong reference to the current greenlet in this thread state */
static PyGreenlet* volatile ts_current = NULL;

static int
green_updatecurrent(void)
{
    PyObject *exc, *val, *tb;
    PyThreadState* tstate;
    PyGreenlet* current;
    PyGreenlet* previous;
    PyObject* deleteme;

green_updatecurrent_restart:
    /* save current exception */
    PyErr_Fetch(&exc, &val, &tb);

    /* get ts_current from the active tstate */
    tstate = PyThreadState_GET();
    if (tstate->dict &&
        (current = (PyGreenlet*)PyDict_GetItem(tstate->dict, ts_curkey))) {
        /* found -- remove it, to avoid keeping a ref */
        Py_INCREF(current);
        PyDict_DelItem(tstate->dict, ts_curkey);
    }
    else {
        /* first time we see this tstate */
        current = green_create_main();
        if (current == NULL) {
            Py_XDECREF(exc);
            Py_XDECREF(val);
            Py_XDECREF(tb);
            return -1;
        }
    }
    assert(current->run_info == tstate->dict);

green_updatecurrent_retry:
    /* update ts_current as soon as possible, in case of nested switches */
    Py_INCREF(current);
    previous = ts_current;
    ts_current = current;

    /* save ts_current as the current greenlet of its own thread */
    if (PyDict_SetItem(previous->run_info, ts_curkey, (PyObject*)previous)) {
        Py_DECREF(previous);
        Py_DECREF(current);
        Py_XDECREF(exc);
        Py_XDECREF(val);
        Py_XDECREF(tb);
        return -1;
    }
    Py_DECREF(previous);

    /* green_dealloc() cannot delete greenlets from other threads, so
       it stores them in the thread dict; delete them now. */
    deleteme = PyDict_GetItem(tstate->dict, ts_delkey);
    if (deleteme != NULL) {
        /* The only reference to these greenlets should be in this list, so
           clearing the list should let them be deleted again, triggering
           calls to green_dealloc() in the correct thread. This may run
           arbitrary Python code?
         */
        PyList_SetSlice(deleteme, 0, INT_MAX, NULL);
    }

    if (ts_current != current) {
        /* some Python code executed above and there was a thread switch,
         * so ts_current points to some other thread again. We need to
         * delete ts_curkey (it's likely there) and retry. */
        PyDict_DelItem(tstate->dict, ts_curkey);
        goto green_updatecurrent_retry;
    }

    /* release an extra reference */
    Py_DECREF(current);
    /* restore current exception */
    PyErr_Restore(exc, val, tb);

    /* thread switch could happen during PyErr_Restore, in that
       case there's nothing to do except restart from scratch. */
    if (ts_current->run_info != tstate->dict) {
        goto green_updatecurrent_restart;
    }
    return 0;
}
```

其中 `green_updatecurrent` 的作用是更新 `ts_current` 的值，而 `ts_current` 存储的是当前线程绑定的 greenlet。

接着我们看看 `green_init`：

```c
// 相当于 __init__
static int
green_init(PyGreenlet* self, PyObject* args, PyObject* kwargs)
{
    PyObject* run = NULL;
    PyObject* nparent = NULL;
    static char* kwlist[] = {"run", "parent", 0};
    if (!PyArg_ParseTupleAndKeywords(
            args, kwargs, "|OO:green", kwlist, &run, &nparent)) {
        return -1;
    }

    if (run != NULL) {
        if (green_setrun(self, run, NULL)) {
            return -1;
        }
    }
    if (nparent != NULL && nparent != Py_None) {
        return green_setparent(self, nparent, NULL);
    }
    return 0;
}
```

没有什么新颖的，但是可以看得出来，`run` 和 `parent` 参数是选填的。接下来看看提供的方法：

```c
static PyMethodDef green_methods[] = {
    {"switch",
     (PyCFunction)green_switch,
     METH_VARARGS | METH_KEYWORDS,
     green_switch_doc},
    {"throw", (PyCFunction)green_throw, METH_VARARGS, green_throw_doc},
    {"__getstate__", (PyCFunction)green_getstate, METH_NOARGS, NULL},
    {NULL, NULL} /* sentinel */
};
```

也就是说greenlet提供了三个方法。

## greenlet 是怎么运行的？

回到上面的例子，可以看到，如果我们只使用 greenlet 的话，是需要手动让出执行权的，也就是调用你想要执行的 greenlet 的
`switch` 方法：

```python
>>> from greenlet import greenlet

>>> def test1():
...     print("[gr1] main  -> test1")
...     gr2.switch()
...     print("[gr1] test1 <- test2")
...     return 'test1 done'

>>> def test2():
...     print("[gr2] test1 -> test2")
...     gr1.switch()
...     print("This is never printed.")

>>> gr1 = greenlet(test1)
>>> gr2 = greenlet(test2)
>>> gr1.switch()
[gr1] main  -> test1
[gr2] test1 -> test2
[gr1] test1 <- test2
'test1 done'
>>> gr1.dead
True
>>> gr2.dead
False
```

我们来看看 `switch` 是怎么实现的吧，从上一节的方法声明可以看到，对应 `green_switch` 函数：

```c
static PyObject*
green_switch(PyGreenlet* self, PyObject* args, PyObject* kwargs)
{
    Py_INCREF(args);
    Py_XINCREF(kwargs);
    return single_result(g_switch(self, args, kwargs)); // 先看下 single_result 干啥的
}

static PyObject*
single_result(PyObject* results)
{
    if (results != NULL && PyTuple_Check(results) &&
        PyTuple_GET_SIZE(results) == 1) {
        PyObject* result = PyTuple_GET_ITEM(results, 0);
        Py_INCREF(result);
        Py_DECREF(results);
        return result;
    }
    else {
        return results;
    }
}  // 返回结果，所以这里不是重点，继续看 g_switch

static PyObject*
g_switch(PyGreenlet* target, PyObject* args, PyObject* kwargs)
{
    /* _consumes_ a reference to the args tuple and kwargs dict,
       and return a new tuple reference */
    int err = 0;
    PyObject* run_info;

    /* check ts_current */
    if (!STATE_OK) {
        Py_XDECREF(args);
        Py_XDECREF(kwargs);
        return NULL;
    }
    run_info = green_statedict(target);
    if (run_info == NULL || run_info != ts_current->run_info) {
        Py_XDECREF(args);
        Py_XDECREF(kwargs);
        PyErr_SetString(PyExc_GreenletError,
                        run_info ?
                            "cannot switch to a different thread" :
                            "cannot switch to a garbage collected greenlet");
        return NULL;
    }

    ts_passaround_args = args;
    ts_passaround_kwargs = kwargs;

    /* find the real target by ignoring dead greenlets,
       and if necessary starting a greenlet. */
    while (target) {
        if (PyGreenlet_ACTIVE(target)) {
            ts_target = target;
            err = g_switchstack();  // 发生栈切换的地方
            break;
        }
        if (!PyGreenlet_STARTED(target)) {
            void* dummymarker;
            ts_target = target;
            err = g_initialstub(&dummymarker); // 如果是一个新的协程，就执行这个
            if (err == 1) {
                continue; /* retry the switch */
            }
            break;
        }
        target = target->parent;
    }

    /* For a very short time, immediately after the 'atomic'
       g_switchstack() call, global variables are in a known state.
       We need to save everything we need, before it is destroyed
       by calls into arbitrary Python code. */
    args = ts_passaround_args;
    ts_passaround_args = NULL;
    kwargs = ts_passaround_kwargs;
    ts_passaround_kwargs = NULL;
    if (err < 0) {
        /* Turn switch errors into switch throws */
        assert(ts_origin == NULL);
        Py_CLEAR(kwargs);
        Py_CLEAR(args);
    }
    else {
        PyGreenlet* origin;
        PyGreenlet* current;
        PyObject* tracefunc;
        origin = ts_origin;
        ts_origin = NULL;
        current = ts_current;
        if ((tracefunc = PyDict_GetItem(current->run_info, ts_tracekey)) != NULL) {
            Py_INCREF(tracefunc);
            if (g_calltrace(tracefunc,
                            args ? ts_event_switch : ts_event_throw,
                            origin,
                            current) < 0) {
                /* Turn trace errors into switch throws */
                Py_CLEAR(kwargs);
                Py_CLEAR(args);
            }
            Py_DECREF(tracefunc);
        }

        Py_DECREF(origin);
    }

    /* We need to figure out what values to pass to the target greenlet
       based on the arguments that have been passed to greenlet.switch(). If
       switch() was just passed an arg tuple, then we'll just return that.
       If only keyword arguments were passed, then we'll pass the keyword
       argument dict. Otherwise, we'll create a tuple of (args, kwargs) and
       return both. */
    if (kwargs == NULL) {
        return args;
    }
    else if (PyDict_Size(kwargs) == 0) {
        Py_DECREF(kwargs);
        return args;
    }
    else if (PySequence_Length(args) == 0) {
        Py_DECREF(args);
        return kwargs;
    }
    else {
        PyObject* tuple = PyTuple_New(2);
        if (tuple == NULL) {
            Py_DECREF(args);
            Py_DECREF(kwargs);
            return NULL;
        }
        PyTuple_SET_ITEM(tuple, 0, args);
        PyTuple_SET_ITEM(tuple, 1, kwargs);
        return tuple;
    }
}
```

切换的核心代码，其实就是 `slp_switch`，他们是用汇编写的，做的事情基本上就是保存当前的栈信息，把目标栈替换当前栈。

我们还需要看一下 `g_initialstub` 函数：

```c
static int GREENLET_NOINLINE(g_initialstub)(void* mark)
{
    int err;
    PyObject *o, *run;
    PyObject *exc, *val, *tb;
    PyObject* run_info;
    PyGreenlet* self = ts_target;
    PyObject* args = ts_passaround_args;
    PyObject* kwargs = ts_passaround_kwargs;
#if GREENLET_USE_CFRAME
    /*
      See green_new(). This is a stack-allocated variable used
      while *self* is in PyObject_Call().
      We want to defer copying the state info until we're sure
      we need it and are in a stable place to do so.
    */
    CFrame trace_info;
#endif
    /* save exception in case getattr clears it */
    PyErr_Fetch(&exc, &val, &tb);
    /* self.run is the object to call in the new greenlet */
    run = PyObject_GetAttrString((PyObject*)self, "run");
    if (run == NULL) {
        Py_XDECREF(exc);
        Py_XDECREF(val);
        Py_XDECREF(tb);
        return -1;
    }
    /* restore saved exception */
    PyErr_Restore(exc, val, tb);

    /* recheck the state in case getattr caused thread switches */
    if (!STATE_OK) {
        Py_DECREF(run);
        return -1;
    }

    /* recheck run_info in case greenlet reparented anywhere above */
    run_info = green_statedict(self);
    if (run_info == NULL || run_info != ts_current->run_info) {
        Py_DECREF(run);
        PyErr_SetString(PyExc_GreenletError,
                        run_info ?
                            "cannot switch to a different thread" :
                            "cannot switch to a garbage collected greenlet");
        return -1;
    }

    /* by the time we got here another start could happen elsewhere,
     * that means it should now be a regular switch
     */
    if (PyGreenlet_STARTED(self)) {
        Py_DECREF(run);
        ts_passaround_args = args;
        ts_passaround_kwargs = kwargs;
        return 1;
    }

#if GREENLET_USE_CFRAME
    /* OK, we need it, we're about to switch greenlets, save the state. */
    trace_info = *PyThreadState_GET()->cframe;
    /* Make the target greenlet refer to the stack value. */
    self->cframe = &trace_info;
    /*
      And restore the link to the previous frame so this one gets
      unliked appropriately.
    */
    self->cframe->previous = &PyThreadState_GET()->root_cframe;
#endif
    /* start the greenlet */
    self->stack_start = NULL;
    self->stack_stop = (char*)mark;
    if (ts_current->stack_start == NULL) {
        /* ts_current is dying */
        self->stack_prev = ts_current->stack_prev;
    }
    else {
        self->stack_prev = ts_current;
    }
    self->top_frame = NULL;
    green_clear_exc(self);
    self->recursion_depth = PyThreadState_GET()->recursion_depth;

    /* restore arguments in case they are clobbered */
    ts_target = self;
    ts_passaround_args = args;
    ts_passaround_kwargs = kwargs;

    /* perform the initial switch */
    err = g_switchstack();  // 栈切换

    /* returns twice!
       The 1st time with ``err == 1``: we are in the new greenlet
       The 2nd time with ``err <= 0``: back in the caller's greenlet

       注意，这里所谓的返回两次，其实和fork是一样的。切换之后，由于有两个协程，所以一个返回0，一个返回1
       1就是新的协程，0就是旧的。
    */
    if (err == 1) {
        /* in the new greenlet */
        PyGreenlet* origin;
        PyObject* tracefunc;
        PyObject* result;
        PyGreenlet* parent;
        self->stack_start = (char*)1; /* running */

        /* grab origin while we still can */
        origin = ts_origin;
        ts_origin = NULL;

        /* now use run_info to store the statedict */
        o = self->run_info;
        self->run_info = green_statedict(self->parent);
        Py_INCREF(self->run_info);
        Py_XDECREF(o);

        if ((tracefunc = PyDict_GetItem(self->run_info, ts_tracekey)) != NULL) {
            Py_INCREF(tracefunc);
            if (g_calltrace(tracefunc,
                            args ? ts_event_switch : ts_event_throw,
                            origin,
                            self) < 0) {
                /* Turn trace errors into switch throws */
                Py_CLEAR(kwargs);
                Py_CLEAR(args);
            }
            Py_DECREF(tracefunc);
        }

        Py_DECREF(origin);

        if (args == NULL) {
            /* pending exception */
            result = NULL;
        }
        else {
            /* call g.run(*args, **kwargs) */
            // 主动调用协程的 run 方法
            result = PyObject_Call(run, args, kwargs);
            Py_DECREF(args);
            Py_XDECREF(kwargs);
        }
        Py_DECREF(run);
        result = g_handle_exit(result);

        /* jump back to parent */
        self->stack_start = NULL; /* dead */
        // 执行完了，就向上切换
        for (parent = self->parent; parent != NULL; parent = parent->parent) {
            result = g_switch(parent, result, NULL);
            /* Return here means switch to parent failed,
             * in which case we throw *current* exception
             * to the next parent in chain.
             */
            assert(result == NULL);
        }
        /* We ran out of parents, cannot continue */
        PyErr_WriteUnraisable((PyObject*)self);
        Py_FatalError("greenlets cannot continue");
    }
    /* back in the parent */
    if (err < 0) {
        /* start failed badly, restore greenlet state */
        self->stack_start = NULL;
        self->stack_stop = NULL;
        self->stack_prev = NULL;
    }
    return err;
}
```

到目前我们知道：

- greenlet 使用线程内的全局变量来存储信息，因此，一个线程会对应多个greenlet，因此一个线程对应的greenlet其实是有限制的，并不是越多越好。
- 默认情况下，greenlet 的 parent 会是当前的 ts_current
- greenlet 通过 parent 这个属性，来组成了一个树状的协程组织关系
- 如果是一个新的协程，那么就会自动执行 run 函数

greenlet 文档中有这么一段话：

Every greenlet, except the main greenlet, has a “parent” greenlet. The parent greenlet defaults to being the one in
which the greenlet was created (this can be changed at any time). In this way, greenlets are organized in a tree.
Top-level code that doesn’t run in a user-created greenlet runs in the implicit main greenlet, which is the root of
the tree.

The parent is where execution continues when a greenlet dies, whether by explicitly returning from its function,
“falling off the end” of its function, or by raising an uncaught exception.

In the above example, both gr1 and gr2 have the main greenlet as a parent. Whenever one of them dies,
the execution comes back to “main”.

由此，也就引入了 main greenlet 这个东西，这个东西的特点是，它的 parent 是 None。main greenlet 永远不会退出，因为它是
协程关系树的根节点。

文档中，有这么一个例子：

```python
>>> from greenlet import getcurrent
>>> def am_i_main():
...     current = getcurrent()
...     return current.parent is None
>>> am_i_main()
True
>>> glet = greenlet(am_i_main)
>>> glet.switch()
False
```

也就是说，main greenlet 不是我们手动创建的，看上面的代码，没有手动初始化 main greenlet，手动初始化的那个也不是 main greenlet。
那么，是咋做到的呢？那就只有模块初始化的时候能做到了。

## main greenlet 是怎么初始化的？

我们来看代码：

```c
PyMODINIT_FUNC
init_greenlet(void)
{
    // ...

    ts_current = green_create_main(); // 创建 main greenlet
    if (ts_current == NULL) {
        INITERROR;
    }

    // ...
}
```

由上我们可以看到，在模块初始化的时候，会创建一个 main greenlet 并且赋值给 ts_current。而后续创建 greenlet 的时候，
会自动把 parent 设置成 ts_current。因此，当创建第二个的时候，parent自然就是最开始模块初始化的时候创建的 main greenlet了。

看一下 `green_create_main`：

```c
static PyGreenlet*
green_create_main(void)
{
    PyGreenlet* gmain;
    PyObject* dict = PyThreadState_GetDict();
    if (dict == NULL) {
        if (!PyErr_Occurred()) {
            PyErr_NoMemory();
        }
        return NULL;
    }

    /* create the main greenlet for this thread */
    gmain = (PyGreenlet*)PyType_GenericAlloc(&PyGreenlet_Type, 0);
    if (gmain == NULL) {
        return NULL;
    }
    gmain->stack_start = (char*)1;
    gmain->stack_stop = (char*)-1;
    /* GetDict() returns a borrowed reference. Make it strong. */
    gmain->run_info = dict;
    Py_INCREF(dict);
    return gmain;
}
```

## 总结

这篇文章我们大概看了一下 greenlet 这个 Python 中用的比较广的协程库是如何实现的，总体看下来，其实原理和 Go 没有差太多，
只是很多实现细节上不同，尤其是调度方面。通过这篇文章，我们就可以对 greenlet 有一个粗浅的认识，接下来，就可以站在
greenlet 的基础之上，去研究一下 gevent 是怎么实现的了。

---

Ref:

- https://greenlet.readthedocs.io/en/latest/greenlet.html
- https://docs.python.org/3/c-api/index.html
- https://docs.python.org/3/c-api/typeobj.html
