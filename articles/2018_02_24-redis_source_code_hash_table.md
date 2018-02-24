# Redis源码阅读三：哈希表

直接上代码：

```c
typedef struct dictEntry {
    void *key;
    union {
        void *val;
        uint64_t u64;
        int64_t s64;
        double d;
    } v;
    struct dictEntry *next;
} dictEntry;

typedef struct dictType {
    uint64_t (*hashFunction)(const void *key);
    void *(*keyDup)(void *privdata, const void *key);
    void *(*valDup)(void *privdata, const void *obj);
    int (*keyCompare)(void *privdata, const void *key1, const void *key2);
    void (*keyDestructor)(void *privdata, void *key);
    void (*valDestructor)(void *privdata, void *obj);
} dictType;

/* This is our hash table structure. Every dictionary has two of this as we
 * implement incremental rehashing, for the old to the new table. */
typedef struct dictht {
    dictEntry **table;
    unsigned long size;
    unsigned long sizemask;
    unsigned long used;
} dictht;

typedef struct dict {
    dictType *type;
    void *privdata;
    dictht ht[2];
    long rehashidx; /* rehashing not in progress if rehashidx == -1 */
    unsigned long iterators; /* number of iterators currently running */
} dict;
```

- 采用面向接口编程，把函数定义放到 `struct dictType` 里，这里是对字典中所有
元素进行操作的函数

- `struct dictEntry` 是字典中K-V真正存放的地方， `void *key` 是K，而 `union ... v`是V

- 每个字典有两张哈希表，在 `dictht ht[2]` 这里定义。 `rehashidx` 是一个标记位，代表是否处于
重新哈希。其中rehash采用的是渐进式rehash，可以看到代码中有好几个地方有这样的代码：

`if (dictIsRehashing(d)) _dictRehashStep(d);`

- dict使用的哈希函数见：https://en.wikipedia.org/wiki/SipHash

- 解决冲突是用链式

- 迭代，字典在有安全跌带器的时候不会进行rehash，见代码：

```c
/* This function performs just a step of rehashing, and only if there are
 * no safe iterators bound to our hash table. When we have iterators in the
 * middle of a rehashing we can't mess with the two hash tables otherwise
 * some element can be missed or duplicated.
 *
 * This function is called by common lookup or update operations in the
 * dictionary so that the hash table automatically migrates from H1 to H2
 * while it is actively used. */
static void _dictRehashStep(dict *d) {
    if (d->iterators == 0) dictRehash(d,1);
}
```
