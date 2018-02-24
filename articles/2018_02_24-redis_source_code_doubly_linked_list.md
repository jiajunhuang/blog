# Redis源码阅读二：双链表

Redis中的双向链表实现属于比较经典的实现，我们直接看代码：

```c
typedef struct listNode {
    struct listNode *prev;
    struct listNode *next;
    void *value;
} listNode;

typedef struct list {
    listNode *head;
    listNode *tail;
    void *(*dup)(void *ptr);
    void (*free)(void *ptr);
    int (*match)(void *ptr, void *key);
    unsigned long len;
} list;
```

从上面我们可以看出，整个链表存储的元素属于同一类型(此处我的意思是可以使用同一套函数操作)，因为对链表进行操作的三个函数
在 `struct list` 里而非`struct listNode` 里。

另外我们可以看出作者在C中使用面向接口编程的方式：在结构体中保存操作元素的函数，值要给定这几个方法，那么
任何元素都可以存储在链表中。

其中 `listNode` 是真正存储元素的地方，而 `list` 是表头。实现非常的经典。

## 扩展：Python中的deque是怎么实现的

https://github.com/jiajunhuang/cpython/blob/eb81795d7d3a8c898fa89a376d63fc3bbfb9a081/Modules/_collectionsmodule.c#L71-L86

```c
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

Python中deque的实现就不那么教科书式的经典了，它糅合了双链表和数组两种形式。把一定数量的元素组成一个block，然后
把多个block串起来，其主要优点是节省了每个节点之间的前后两个指针的空间。
