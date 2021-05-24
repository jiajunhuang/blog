# Redis源码阅读：key是怎么过期的

我们经常用到Redis的expire这个命令，比如我们设置一个缓存，通常会这样用：

```bash
SETEX mykey 10 "Hello"
```

如官网文档所说，这个命令相当于：

```bash
SET mykey value
EXPIRE mykey seconds
```

我们直接翻代码求证：

```c
{"setex",setexCommand,4,
    "write use-memory @string",
    0,NULL,1,1,1,0,0,0},

void setexCommand(client *c) {
    c->argv[3] = tryObjectEncoding(c->argv[3]);
    setGenericCommand(c,OBJ_EX,c->argv[1],c->argv[3],c->argv[2],UNIT_SECONDS,NULL,NULL);
}

void setGenericCommand(client *c, int flags, robj *key, robj *val, robj *expire, int unit, robj *ok_reply, robj *abort_reply) {
    long long milliseconds = 0, when = 0; /* initialized to avoid any harmness warning */

    // ...
    genericSetKey(c,c->db,key, val,flags & OBJ_KEEPTTL,1);

    // ...
    if (expire) {
        // ...
        setExpire(c,c->db,key,when);
        // ...
    }
    // ...
}
```

接下来我们来看看 `EXPIRE` 命令，这个命令的作用是为某个key设置超时时间：

```c
/* EXPIRE key seconds */
void expireCommand(client *c) {
    expireGenericCommand(c,mstime(),UNIT_SECONDS);
}

void expireGenericCommand(client *c, long long basetime, int unit) {
    robj *key = c->argv[1], *param = c->argv[2];
    long long when; /* unix time in milliseconds when the key will expire. */

    if (getLongLongFromObjectOrReply(c, param, &when, NULL) != C_OK)
        return;
    int negative_when = when < 0;
    if (unit == UNIT_SECONDS) when *= 1000;
    when += basetime;
    if (((when < 0) && !negative_when) || ((when-basetime > 0) && negative_when)) {
        /* EXPIRE allows negative numbers, but we can at least detect an
         * overflow by either unit conversion or basetime addition. */
        addReplyErrorFormat(c, "invalid expire time in %s", c->cmd->name);
        return;
    }
    /* No key, return zero. */
    if (lookupKeyWrite(c->db,key) == NULL) {
        addReply(c,shared.czero);
        return;
    }

    if (checkAlreadyExpired(when)) {
        // 如果已经超时，那么就删掉
        robj *aux;

        int deleted = server.lazyfree_lazy_expire ? dbAsyncDelete(c->db,key) :
                                                    dbSyncDelete(c->db,key);
        serverAssertWithInfo(c,key,deleted);
        server.dirty++;

        /* Replicate/AOF this as an explicit DEL or UNLINK. */
        aux = server.lazyfree_lazy_expire ? shared.unlink : shared.del;
        rewriteClientCommandVector(c,2,aux,key);
        signalModifiedKey(c,c->db,key);
        notifyKeyspaceEvent(NOTIFY_GENERIC,"del",key,c->db->id);
        addReply(c, shared.cone);
        return;
    } else {
        // 如果没有超时，就设置超时时间
        setExpire(c,c->db,key,when);
        addReply(c,shared.cone);
        signalModifiedKey(c,c->db,key);
        notifyKeyspaceEvent(NOTIFY_GENERIC,"expire",key,c->db->id);
        server.dirty++;
        return;
    }
}
```

其实我们主要还是想知道，Redis到底是怎么记住超时信息的。我们去 `setExpire` 看看：

```c
/* Set an expire to the specified key. If the expire is set in the context
 * of an user calling a command 'c' is the client, otherwise 'c' is set
 * to NULL. The 'when' parameter is the absolute unix time in milliseconds
 * after which the key will no longer be considered valid. */
void setExpire(client *c, redisDb *db, robj *key, long long when) {
    dictEntry *kde, *de;

    /* Reuse the sds from the main dict in the expire dict */
    kde = dictFind(db->dict,key->ptr);
    serverAssertWithInfo(NULL,key,kde != NULL);
    de = dictAddOrFind(db->expires,dictGetKey(kde));
    dictSetSignedIntegerVal(de,when);

    int writable_slave = server.masterhost && server.repl_slave_ro == 0;
    if (c && writable_slave && !(c->flags & CLIENT_MASTER))
        rememberSlaveKeyWithExpire(db,key);
}
```

可以看到，先去 `db->dict` 里找，确定key一定存在，然后去 `db->expires` 里找或者新增，然后 `dictSetSignedIntegerVal` 设置
value为超时时间：

```c
/* Add or Find:
 * dictAddOrFind() is simply a version of dictAddRaw() that always
 * returns the hash entry of the specified key, even if the key already
 * exists and can't be added (in that case the entry of the already
 * existing key is returned.)
 *
 * See dictAddRaw() for more information. */
dictEntry *dictAddOrFind(dict *d, void *key) {
    dictEntry *entry, *existing;
    entry = dictAddRaw(d,key,&existing);
    return entry ? entry : existing;
}

#define dictSetSignedIntegerVal(entry, _val_) \
    do { (entry)->v.s64 = _val_; } while(0)

// 为了看懂上面的宏，我们得知道具体结构。这样就能理解了。
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
```

到这里我们知道了，原来Redis的过期，就是通过一个字典，key为key，value为过期时间，存好来实现的，唉，原来如此。不过，真的
是这么简单吗？这样，我们再来看看 `TTL` 这个命令是怎么实现的，如果它也只读了这两个，那就可以证明是这么简单了：

```c
/* TTL key */
void ttlCommand(client *c) {
    ttlGenericCommand(c, 0);
}

/* Implements TTL and PTTL */
void ttlGenericCommand(client *c, int output_ms) {
    long long expire, ttl = -1;

    /* If the key does not exist at all, return -2 */
    if (lookupKeyReadWithFlags(c->db,c->argv[1],LOOKUP_NOTOUCH) == NULL) {
        addReplyLongLong(c,-2);
        return;
    }
    /* The key exists. Return -1 if it has no expire, or the actual
     * TTL value otherwise. */
    expire = getExpire(c->db,c->argv[1]);
    if (expire != -1) {
        ttl = expire-mstime();
        if (ttl < 0) ttl = 0;
    }
    if (ttl == -1) {
        addReplyLongLong(c,-1);
    } else {
        addReplyLongLong(c,output_ms ? ttl : ((ttl+500)/1000));
    }
}

/* Return the expire time of the specified key, or -1 if no expire
 * is associated with this key (i.e. the key is non volatile) */
long long getExpire(redisDb *db, robj *key) {
    dictEntry *de;

    /* No expire? return ASAP */
    if (dictSize(db->expires) == 0 ||
        // 还真是！
       (de = dictFind(db->expires,key->ptr)) == NULL) return -1;

    /* The entry was found in the expire dict, this means it should also
     * be present in the main dict (safety check). */
    serverAssertWithInfo(NULL,key,dictFind(db->dict,key->ptr) != NULL);
    return dictGetSignedIntegerVal(de);
}
```

还真的是从 `db-expires` 里读取。。。好吧，我们来看下db的结构体长啥样：

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

原来如此，看来连 `MULTI` 之类的，都是靠一个字典来实现的，妙啊。

等等，还有一个问题，这里我们看到，expire执行的时候，如果key已经过期，会删除。但是前提是，client执行了expire这个
命令呀，如果没有执行咋办？岂不是内存就要被过期的key塞爆？当然不是，Redis的过期策略，文档上有如下描述：

##### How Redis expires keys

Redis keys are expired in two ways: a passive way, and an active way.

A key is passively expired simply when some client tries to access it, and the key is found to be timed out.

Of course this is not enough as there are expired keys that will never be accessed again. These keys should
be expired anyway, so periodically Redis tests a few keys at random among keys with an expire set. All the
keys that are already expired are deleted from the keyspace.

Specifically this is what Redis does 10 times per second:

Test 20 random keys from the set of keys with an associated expire.
Delete all the keys found expired.
If more than 25% of keys were expired, start again from step 1.
This is a trivial probabilistic algorithm, basically the assumption is that our sample is representative of
the whole key space, and we continue to expire until the percentage of keys that are likely to be expired is under 25%

This means that at any given moment the maximum amount of keys already expired that are using memory is at
max equal to max amount of write operations per second divided by 4.

我们去源码里翻翻，我搜索cron，找到了 `serverCron`：

```c
int serverCron(struct aeEventLoop *eventLoop, long long id, void *clientData) {
    // ...

    /* Handle background operations on Redis databases. */
    databasesCron();

    // ...
}

/* This function handles 'background' operations we are required to do
 * incrementally in Redis databases, such as active key expiring, resizing,
 * rehashing. */
void databasesCron(void) {
    /* Expire keys by random sampling. Not required for slaves
     * as master will synthesize DELs for us. */
    if (server.active_expire_enabled) {
        if (iAmMaster()) {
            activeExpireCycle(ACTIVE_EXPIRE_CYCLE_SLOW);
        } else {
            expireSlaveKeys();
        }
    }

    // ...
}

void activeExpireCycle(int type) {
    // ...
}
```

最后的这个 `activeExpireCycle` 里做的事情，就是上述文档中描述的逻辑，这个函数很长，我也没细看，知道就可以了。

## 总结

这篇文章中我们看到了Redis是如何处理Key过期的，如何保存，以及过期的策略，其实就是通过创建一个dict来保存哪些key是
要过期的，然后通过主动检查key是否应当删除，加上 server cron 定期检查并且删除已经过期的key来实现的。

---

ref:

- https://redis.io/commands/setex
- https://redis.io/commands/expire
- https://redis.io/commands/ttl
