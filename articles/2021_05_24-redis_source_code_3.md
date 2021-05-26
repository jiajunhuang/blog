# Redis源码阅读：字典是怎么实现的

dict，也就是哈希表这个数据结构，在Redis中的作用非常广泛，比如，Redis用它来存储支持的命令，这篇文章我们会看一下Redis是
如何实现dict的。

上一篇我们讲到，Redis读取网络请求的内容，解析出命令后，开始处理。其中就有一个函数叫做 `processCommand` ，就是用来处理客户端请求的：

```c
int processCommand(client *c) {
    // ...

    /* Now lookup the command and check ASAP about trivial error conditions
     * such as wrong arity, bad command name and so forth. */
    c->cmd = c->lastcmd = lookupCommand(c->argv[0]->ptr);

    // ...
}
```

这里的 `lookupCommand` 就是我开始讲的，Redis用dict来存储命令。那么是在哪里存储的呢？在 `initServerConfig` 里有这么一行：

```c
    populateCommandTable();

```

点进去看：

```c
/* Populates the Redis Command Table starting from the hard coded list
 * we have on top of server.c file. */
void populateCommandTable(void) {
    int j;
    int numcommands = sizeof(redisCommandTable)/sizeof(struct redisCommand);

    for (j = 0; j < numcommands; j++) {
        struct redisCommand *c = redisCommandTable+j;
        int retval1, retval2;

        /* Translate the command string flags description into an actual
         * set of flags. */
        if (populateCommandTableParseFlags(c,c->sflags) == C_ERR)
            serverPanic("Unsupported command flag");

        c->id = ACLGetCommandID(c->name); /* Assign the ID used for ACL. */
        retval1 = dictAdd(server.commands, sdsnew(c->name), c);
        /* Populate an additional dictionary that will be unaffected
         * by rename-command statements in redis.conf. */
        retval2 = dictAdd(server.orig_commands, sdsnew(c->name), c);
        serverAssert(retval1 == DICT_OK && retval2 == DICT_OK);
    }
}

```

其实就是相当于在Go里，声明一个 `map[string]Command{}` 然后把所有的命令名字作为key，Command结构体作为value存进去。
最上面的 `redisCommandTable` 就是Redis支持的所有命令的一个列表：

```c
struct redisCommand redisCommandTable[] = {
    {"module",moduleCommand,-2,
     "admin no-script",
     0,NULL,0,0,0,0,0,0},

    {"get",getCommand,2,
     "read-only fast @string",
     0,NULL,1,1,1,0,0,0},

    // ...
```

现在我们知道dict在Redis中是非常有用的一个数据结构了，偷偷告诉你，Redis中的expire命令也用到了dict：

```c
/* Redis database representation. There are multiple databases identified
 * by integers from 0 (the default database) up to the max configured
 * database. The database number is the 'id' field in the structure. */
typedef struct redisDb {
    dict *dict;                 /* The keyspace for this DB */
    dict *expires;              /* Timeout of keys with a timeout set */
    dict *blocking_keys;        /* Keys with clients waiting for data (BLPOP)*/
    dict *ready_keys;           /* Blocked keys that received a PUSH */
    dict *watched_keys;         /* WATCHED keys for MULTI/EXEC CAS */
    int id;                     /* Database ID */
    long long avg_ttl;          /* Average TTL, just for stats */
    unsigned long expires_cursor; /* Cursor of the active expire cycle. */
    list *defrag_later;         /* List of key names to attempt to defrag one by one, gradually. */
} redisDb;
```

不过这都扯远了，我们以后再来探索这些。回到正题，从上面的用法来看，我们知道，Redis的dict实现是类型无关的，那么
它到底是怎么实现的呢？让我们深入探索一下：

```c
typedef struct dict {
    dictType *type;
    void *privdata;
    dictht ht[2];
    long rehashidx; /* rehashing not in progress if rehashidx == -1 */
    int16_t pauserehash; /* If >0 rehashing is paused (<0 indicates coding error) */
} dict;
```

可以看到，dict的结构体里，有 `dictType *type`，`dictht ht[2]`，`long rehashidx` 三个很重要的结构，其余两个我暂时
还不知道是干啥用的，但是先不管，我们继续看，先看 `dictType` 到底是什么：

```c
typedef struct dictType {
    uint64_t (*hashFunction)(const void *key);
    void *(*keyDup)(void *privdata, const void *key);
    void *(*valDup)(void *privdata, const void *obj);
    int (*keyCompare)(void *privdata, const void *key1, const void *key2);
    void (*keyDestructor)(void *privdata, void *key);
    void (*valDestructor)(void *privdata, void *obj);
    int (*expandAllowed)(size_t moreMem, double usedRatio);
} dictType;
```

原来是一个结构体，这个结构体里全是函数指针，这是C里实现类型无关代码的常用方式，也就是不管具体类型，调用者传入处理具体
数据的函数指针，然后数据都以 `void *` 的指针来传递。我们知道 `type` 是一个 `dictType *` 的类型就可以了，我们继续看
`dictht ht[2]` 是什么：

```c
/* This is our hash table structure. Every dictionary has two of this as we
 * implement incremental rehashing, for the old to the new table. */
typedef struct dictht {
    dictEntry **table;
    unsigned long size;
    unsigned long sizemask;
    unsigned long used;
} dictht;
```

注释非常重要，它说了，这是我们哈希表的真正结构，每一个dict都有两个表，因为这样才可以实现渐进式哈希，渐进式哈希是什么呢？
试想，如果你是数据结构的实现者，当你的哈希表容量不够，冲突率严重上升时，你是不是应该选择扩容，那么扩容怎么做呢？当然是
创建一个容量更大的哈希表，然后把数据搬过去，替换老的表，然后把老表free。

这样可以吗？这样当然是一种方式，可是这样做有一个缺点，如果这个哈希表的数据量非常大，那么处理就会很耗时间，Redis的设计
是一个高并发的数据结构服务器，如果那样做的话，就势必突然有一些客户端的某些请求特别慢，因为恰巧触发了rehash。所以Redis
没有采用这种方式，而是采用了渐进式哈希，也就是每一次都搬一部分数据，直到搬完为止。

我们继续看Redis代码，`dictht` 的第一个属性，`dictEntry **table` 就是存储真正的数据的地方，`size` 是当前所申请的存储
数据的大小，哈希表说白了，就是用一个大数组，然后计算哈希值，对数组长度取余，放到对应的slot上，对于C来说，`dictEntry **table`
的意思就是，`table` 是一个指针，指向一块内存，而这块内存存储的数据的类型是 `dictEntry *`，可能有点绕，但是习惯了就好了。
`sizemask` 是用来做位运算的掩码，`used` 是已经存储的数据的大小。为啥要有 `sizemask` 呢？这是因为，当大小为2的幂时，
取余操作 `num % size` 可以直接用位运算 `num & (size - 1)` 来做，这样会更快，我们来看看：

```c
#include <stdio.h>
#include <time.h>
#include <sys/time.h>

#define LOOP_TIME (8589934592 * 3)

double timeit(void (*f)()) {
    struct timeval start, end;

    gettimeofday(&start, NULL);
    (*f)();
    gettimeofday(&end, NULL);

    return end.tv_sec + end.tv_usec / 1e6 - start.tv_sec - start.tv_usec / 1e6; // in seconds
}

void use_bitwise(void) {
    long size = 64;
    long mask = size - 1;

    long num = 4294967296;
    for (long i = 0; i < LOOP_TIME; i++) {
        num & mask;
    }
}

void use_mod(void) {
    long size = 64;
    long mask = size - 1;

    long num = 4294967296;
    for (long i = 0; i < LOOP_TIME; i++) {
        num % size;
    }
}

int main(void) {
    for (int i = 0; i < 3; i++) {
        printf("use_bitwise took %f seconds, use_mod took %f seconds\n", timeit(use_bitwise), timeit(use_mod));
    }
}
```

执行一下：

```c
$ cc -O0 main.c && ./a.out 
use_bitwise took 10.128805 seconds, use_mod took 10.356709 seconds
use_bitwise took 10.151692 seconds, use_mod took 10.334191 seconds
use_bitwise took 10.304045 seconds, use_mod took 10.360759 seconds
```

well，还是更快一丢丢的。我们再次回到主题，看下dict是如何使用的，上面说了，`populateCommandTable()` 把命令从列表存到
dict里，但是这个dict并不是在这里创建的，而是 `populateCommandTable()` 上面：

```c
    /* Command table -- we initialize it here as it is part of the
     * initial configuration, since command names may be changed via
     * redis.conf using the rename-command directive. */
    server.commands = dictCreate(&commandTableDictType,NULL);
```

我们来看看 `dictCreate`：

```c
/* Create a new hash table */
dict *dictCreate(dictType *type,
        void *privDataPtr)
{
    dict *d = zmalloc(sizeof(*d));

    _dictInit(d,type,privDataPtr);
    return d;
}

/* Initialize the hash table */
int _dictInit(dict *d, dictType *type,
        void *privDataPtr)
{
    _dictReset(&d->ht[0]);
    _dictReset(&d->ht[1]);
    d->type = type;
    d->privdata = privDataPtr;
    d->rehashidx = -1;
    d->pauserehash = 0;
    return DICT_OK;
}

/* Reset a hash table already initialized with ht_init().
 * NOTE: This function should only be called by ht_destroy(). */
static void _dictReset(dictht *ht)
{
    ht->table = NULL;
    ht->size = 0;
    ht->sizemask = 0;
    ht->used = 0;
}
```

可以看到，`dictCreate` 的时候，没怎么申请内存，那么肯定是在添加第一个元素的时候申请的，我们来看下，我翻到 `populateCommandTable`
函数里，添加是调用的 `dictAdd`：

```c
/* Add an element to the target hash table */
int dictAdd(dict *d, void *key, void *val)
{
    dictEntry *entry = dictAddRaw(d,key,NULL);

    if (!entry) return DICT_ERR;
    dictSetVal(d, entry, val);
    return DICT_OK;
}

dictEntry *dictAddRaw(dict *d, void *key, dictEntry **existing)
{
    long index;
    dictEntry *entry;
    dictht *ht;

    if (dictIsRehashing(d)) _dictRehashStep(d);  // 渐进式哈希在这里

    /* Get the index of the new element, or -1 if
     * the element already exists. */
    if ((index = _dictKeyIndex(d, key, dictHashKey(d,key), existing)) == -1) // 计算应当落在哪个slot
        return NULL;

    /* Allocate the memory and store the new entry.
     * Insert the element in top, with the assumption that in a database
     * system it is more likely that recently added entries are accessed
     * more frequently. */
    // 放进去
    ht = dictIsRehashing(d) ? &d->ht[1] : &d->ht[0];
    entry = zmalloc(sizeof(*entry));
    entry->next = ht->table[index];
    ht->table[index] = entry;
    ht->used++;

    /* Set the hash entry fields. */
    dictSetKey(d, entry, key);
    return entry;
}
```

上面我们看到了是如何添加，那么啥时候会检查是否需要扩容以及开始渐进式哈希呢？原来这个逻辑在 `_dictKeyIndex` 里：

```c
static long _dictKeyIndex(dict *d, const void *key, uint64_t hash, dictEntry **existing)
{
    unsigned long idx, table;
    dictEntry *he;
    if (existing) *existing = NULL;

    /* Expand the hash table if needed */
    if (_dictExpandIfNeeded(d) == DICT_ERR)
        return -1;
    // ...
}

/* Expand the hash table if needed */
static int _dictExpandIfNeeded(dict *d)
{
    /* Incremental rehashing already in progress. Return. */
    if (dictIsRehashing(d)) return DICT_OK;

    /* If the hash table is empty expand it to the initial size. */
    if (d->ht[0].size == 0) return dictExpand(d, DICT_HT_INITIAL_SIZE);

    /* If we reached the 1:1 ratio, and we are allowed to resize the hash
     * table (global setting) or we should avoid it but the ratio between
     * elements/buckets is over the "safe" threshold, we resize doubling
     * the number of buckets. */
    if (d->ht[0].used >= d->ht[0].size &&
        (dict_can_resize ||
         d->ht[0].used/d->ht[0].size > dict_force_resize_ratio) &&
        dictTypeExpandAllowed(d))
    {
        return dictExpand(d, d->ht[0].used + 1);
    }
    return DICT_OK;
}

/* return DICT_ERR if expand was not performed */
int dictExpand(dict *d, unsigned long size) {
    return _dictExpand(d, size, NULL);
}

/* Expand or create the hash table,
 * when malloc_failed is non-NULL, it'll avoid panic if malloc fails (in which case it'll be set to 1).
 * Returns DICT_OK if expand was performed, and DICT_ERR if skipped. */
int _dictExpand(dict *d, unsigned long size, int* malloc_failed)
{
    if (malloc_failed) *malloc_failed = 0;

    /* the size is invalid if it is smaller than the number of
     * elements already inside the hash table */
    if (dictIsRehashing(d) || d->ht[0].used > size)
        return DICT_ERR;

    dictht n; /* the new hash table */
    unsigned long realsize = _dictNextPower(size);

    /* Rehashing to the same table size is not useful. */
    if (realsize == d->ht[0].size) return DICT_ERR;

    /* Allocate the new hash table and initialize all pointers to NULL */
    n.size = realsize;
    n.sizemask = realsize-1;
    if (malloc_failed) {
        n.table = ztrycalloc(realsize*sizeof(dictEntry*));
        *malloc_failed = n.table == NULL;
        if (*malloc_failed)
            return DICT_ERR;
    } else
        n.table = zcalloc(realsize*sizeof(dictEntry*));

    n.used = 0;

    /* Is this the first initialization? If so it's not really a rehashing
     * we just set the first hash table so that it can accept keys. */
    if (d->ht[0].table == NULL) {
        d->ht[0] = n;
        return DICT_OK;
    }

    /* Prepare a second hash table for incremental rehashing */
    d->ht[1] = n;
    d->rehashidx = 0;
    return DICT_OK;
}
```

剩下的命令怎么实现的我就不讲了，其实道理都差不多。

## 总结

这篇文章我们看了一下Redis是如何实现dict也就是哈希表的，哦对了，Redis的冲突解决是用链接法去避免的。
