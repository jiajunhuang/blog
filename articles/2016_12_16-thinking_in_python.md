# Thinking in Python

每学习一门语言，就总会想找到它的 `core`。对于Python，我认为是它的 `core` 是
字典。无论从 `CPython` 实现上或者是平时使用时的类机制，都能体会到Python中字典
的作用。但是这篇文章不是什么教程或者学习体会。只是一些杂言。TL;DR

## 0和1，一切都来自两种状态

在内存中所有的东西都是0和1，每一个位我们称为一个bit，因为一个bit只能表示两种
状态，可能不够用，所以我们规定把8个bit组成一个byte。同时我们为了表示更多的信息
需要把byte也组装起来，例如一个int有4个byte，也就是32个bit。就是通过这样一层
一层的抽象和打包，于是我们用不同长度的bit表示出 `int`, `long`, `char` 等等。

## 从C看来，也许是打包

我们进一步将 `int`, `long`, `char` 等基本数据类型进一步封装，通过 `struct`,
`union` 等，就可以将数据捆绑在一起。例如：

> [代码出处](https://github.com/jiajunhuang/cpython/blob/annotation/Include/object.h#L106-L110)

```c
typedef struct _object {
    int ob_refcnt;
    struct _typeobject *ob_type;
} PyObject;

typedef struct {
    PyObject ob_base;
    Py_ssize_t ob_size; /* Number of items in variable part */
} PyVarObject;
```

其中，在release状态下， `_PyObject_HEAD_EXTRA` 展开为空, `ob_refcnt` 在 `configure`
时确定，根据各种情况可能为 `ssize_t`, `long`, `int`。`ob_type` 为指向另一个
struct结构的指针。

接下来我们看一眼 `PyLongObject`：

```c
struct _longobject {
	PyObject_VAR_HEAD
	digit ob_digit[1];
};
```

其中 `PyObject_VAR_HEAD` 为 `#define PyObject_VAR_HEAD PyVarObject ob_base`，
所以这三个struct就构成了如下面的关系：

```
+-----------------------------------------------------------------------+
|                                                                       |
|    struct _longobject {                                               |
|                                                                       |
|  +--------------------------------------------------------------+     |
|  |                                                              |     |
|  |  typedef struct {                                            |     |
|  | +-----------------------------------+                        |     |
|  | |typedef struct _object {           |                        |     |
|  | |    _PyObject_HEAD_EXTRA           |                        |     |
|  | |    Py_ssize_t ob_refcnt;          |                        |     |
|  | |    struct _typeobject *ob_type;   |                        |     |
|  | |} PyObject;                        |                        |     |
|  | |                                   |                        |     |
|  | +-----------------------------------+                        |     |
|  |                                                              |     |
|  |  Py_ssize_t ob_size; /* Number of items in variable part */  |     |
|  |  } PyVarObject;                                              |     |
|  |                                                              |     |
|  +--------------------------------------------------------------+     |
|                                                                       |
|    	digit ob_digit[1];                                              |
|    };                                                                 |
|                                                                       |
+-----------------------------------------------------------------------+
```

通过一层一层的抽象，组成了一个对象系统。

> 这里需要解释一下，我所说的抽象可能与平时所看到的也许有区别。因为抽象
> 这个词语本身就挺抽象的，每个人理解可能都不一样。但是我个人对抽象的一个
> 简单解释是：装作看不见，我只管这一块内存是 `struct _typeobject` 类型的，
> 我不管里面具体是什么数据类型。通过如此一层一层的抽象，我们就能站在一个
> 比较高的地方来看程序。

## 透过字典，绑定变量名

用数组和合适的hash函数，以及rehash方案，组成字典。

这里所说的字典，就是 `key-value` 对，也许是一对，也许是一堆。通过字典，
我们可以将变量名和一个相对应的对象（也就是一块相对应的内存）对应起来。
借此还可以实现变量名作用域：通过将 `local`, `global`, `built-in` 存在
不同的字典里，我们可以实现作用域查找的规则，例如 `Python` 的 `LEGB` -
`local -> enclosing(闭包) -> global -> built-in`。

## 抽象数据结构，模拟计算机

当函数(caller)调用新的函数(callee)，那么在当前栈空间压入新的 `PyFrameObject`，
其中包括该函数的相关参数，回退指针，虚拟机字节码等。当函数执行完毕，根据回退
指针，执行出栈操作，还原到调用该函数的函数(caller)执行到的地方。

用链表将标记为删除的 `PyObject` 连起来，备下次使用，以节省申请和销毁内存次数。

## 属性，继承，实例，MRO

每个对象都有自己的属性存放处(`__dict__`)，有自己的类型，对于类，它有自己的
基类列表，有查找属性规则(MRO)，通过字典保存对象的状态，借此构成Python的对象
系统。

完（杂文嘛，就是这么杂，杂乱无章的杂）。
