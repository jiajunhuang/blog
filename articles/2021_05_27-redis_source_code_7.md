# Redis源码阅读：RDB是怎么实现的

Redis中提供的另外一种持久化方式就是RDB，全称是 Redis Database，其实就是把Redis进程中存储的内容全部dump到
磁盘上，因此 RDB 文件是以二进制形式保存的，这一点与 AOF 相反。

在 Redis 中，我们可以通过 `SAVE` 和 `BGSAVE` 两个命令去主动触发保存 RDB。Redis 也可以通过在 `redis.conf`
中配置每隔多久自动dump一次。如果不配置，那么Redis将会：

Unless specified otherwise, by default Redis will save the DB:

- After 3600 seconds (an hour) if at least 1 key changed
- After 300 seconds (5 minutes) if at least 100 keys changed
- After 60 seconds if at least 10000 keys changed

这三个触发dump的入口，分别在 `saveCommand`, `bgsaveCommand` 和 `serverCron` 里，但是其实他们最终都是调用
`rdbSave` 这个函数去处理。

```c
void saveCommand(client *c) {
    if (server.child_type == CHILD_TYPE_RDB) {
        addReplyError(c,"Background save already in progress");
        return;
    }
    rdbSaveInfo rsi, *rsiptr;
    rsiptr = rdbPopulateSaveInfo(&rsi);
    if (rdbSave(server.rdb_filename,rsiptr) == C_OK) { // 这里触发
        addReply(c,shared.ok);
    } else {
        addReplyErrorObject(c,shared.err);
    }
}

void bgsaveCommand(client *c) {
    int schedule = 0;

    // ...
    } else if (rdbSaveBackground(server.rdb_filename,rsiptr) == C_OK) { // 这里触发
        addReplyStatus(c,"Background saving started");
    } else {
        addReplyErrorObject(c,shared.err);
    }
}

int rdbSaveBackground(char *filename, rdbSaveInfo *rsi) {
    pid_t childpid;

    if (hasActiveChildProcess()) return C_ERR;

    server.dirty_before_bgsave = server.dirty;
    server.lastbgsave_try = time(NULL);

    if ((childpid = redisFork(CHILD_TYPE_RDB)) == 0) {
        int retval;

        /* Child */
        redisSetProcTitle("redis-rdb-bgsave");
        redisSetCpuAffinity(server.bgsave_cpulist);
        retval = rdbSave(filename,rsi); // 这里触发
        if (retval == C_OK) {
            sendChildCowInfo(CHILD_INFO_TYPE_RDB_COW_SIZE, "RDB");
        }
        exitFromChild((retval == C_OK) ? 0 : 1);
    } else {
        /* Parent */
        if (childpid == -1) {
            server.lastbgsave_status = C_ERR;
            serverLog(LL_WARNING,"Can't save in background: fork: %s",
                strerror(errno));
            return C_ERR;
        }
        serverLog(LL_NOTICE,"Background saving started by pid %ld",(long) childpid);
        server.rdb_save_time_start = time(NULL);
        server.rdb_child_type = RDB_CHILD_TYPE_DISK;
        return C_OK;
    }
    return C_OK; /* unreached */
}

// serverCron里
int serverCron(struct aeEventLoop *eventLoop, long long id, void *clientData) {
    // ...
    /* Save if we reached the given amount of changes,
        * the given amount of seconds, and if the latest bgsave was
        * successful or if, in case of an error, at least
        * CONFIG_BGSAVE_RETRY_DELAY seconds already elapsed. */
    if (server.dirty >= sp->changes &&
        server.unixtime-server.lastsave > sp->seconds &&
        (server.unixtime-server.lastbgsave_try >
            CONFIG_BGSAVE_RETRY_DELAY ||
            server.lastbgsave_status == C_OK))
    {
        serverLog(LL_NOTICE,"%d changes in %d seconds. Saving...",
            sp->changes, (int)sp->seconds);
        rdbSaveInfo rsi, *rsiptr;
        rsiptr = rdbPopulateSaveInfo(&rsi);
        rdbSaveBackground(server.rdb_filename,rsiptr);
        break;
    }

    // ...
    /* Start a scheduled BGSAVE if the corresponding flag is set. This is
     * useful when we are forced to postpone a BGSAVE because an AOF
     * rewrite is in progress.
     *
     * Note: this code must be after the replicationCron() call above so
     * make sure when refactoring this file to keep this order. This is useful
     * because we want to give priority to RDB savings for replication. */
    if (!hasActiveChildProcess() &&
        server.rdb_bgsave_scheduled &&
        (server.unixtime-server.lastbgsave_try > CONFIG_BGSAVE_RETRY_DELAY ||
         server.lastbgsave_status == C_OK))
    {
        rdbSaveInfo rsi, *rsiptr;
        rsiptr = rdbPopulateSaveInfo(&rsi);
        if (rdbSaveBackground(server.rdb_filename,rsiptr) == C_OK)
            server.rdb_bgsave_scheduled = 0;
    }

    // ...
}
```

从这里我们分别看到了三个触发 `rdbSave` 的入口，同时也看到，RDB保存的步骤：

- Redis 执行fork
- 子进程将数据库写到临时RDB文件
- 子进程完成之后，替换老的RDB文件

接下来我们去代码里求证，fork的逻辑已经在上面体现了，我们主要看是否子进程写完之后，替换RDB文件：

```c
int rdbSave(char *filename, rdbSaveInfo *rsi) {
    char tmpfile[256];
    char cwd[MAXPATHLEN]; /* Current working dir path for error messages. */
    FILE *fp = NULL;
    rio rdb;
    int error = 0;

    snprintf(tmpfile,256,"temp-%d.rdb", (int) getpid());
    fp = fopen(tmpfile,"w");
    //...
    /* Use RENAME to make sure the DB file is changed atomically only
     * if the generate DB file is ok. */
    if (rename(tmpfile,filename) == -1) {
        char *cwdp = getcwd(cwd,MAXPATHLEN);
        serverLog(LL_WARNING,
            "Error moving temp DB file %s on the final "
            "destination %s (in server root dir %s): %s",
            tmpfile,
            filename,
            cwdp ? cwdp : "unknown",
            strerror(errno));
        unlink(tmpfile);
        stopSaving(0);
        return C_ERR;
    }
    // ...
}
```

看来确实如此。最后我们来看一下，RDB大概是怎么写入的，当然我们并不会去细究具体格式，因为意义不是特别大：

```c
/* Produces a dump of the database in RDB format sending it to the specified
 * Redis I/O channel. On success C_OK is returned, otherwise C_ERR
 * is returned and part of the output, or all the output, can be
 * missing because of I/O errors.
 *
 * When the function returns C_ERR and if 'error' is not NULL, the
 * integer pointed by 'error' is set to the value of errno just after the I/O
 * error. */
int rdbSaveRio(rio *rdb, int *error, int rdbflags, rdbSaveInfo *rsi) {
    // ...
    snprintf(magic,sizeof(magic),"REDIS%04d",RDB_VERSION);
    if (rdbWriteRaw(rdb,magic,9) == -1) goto werr; // 写入魔数
    if (rdbSaveInfoAuxFields(rdb,rdbflags,rsi) == -1) goto werr;
    if (rdbSaveModulesAux(rdb, REDISMODULE_AUX_BEFORE_RDB) == -1) goto werr;

    for (j = 0; j < server.dbnum; j++) {
        // 遍历数据库，写入其中的内容
        /* Write the SELECT DB opcode */
        if (rdbSaveType(rdb,RDB_OPCODE_SELECTDB) == -1) goto werr;
        if (rdbSaveLen(rdb,j) == -1) goto werr;

        /* Iterate this DB writing every entry */
        while((de = dictNext(di)) != NULL) { // 拿到每一个Key Value，写入
            sds keystr = dictGetKey(de);
            robj key, *o = dictGetVal(de);
            long long expire;

            initStaticStringObject(key,keystr);
            expire = getExpire(db,&key);
            if (rdbSaveKeyValuePair(rdb,&key,o,expire) == -1) goto werr;

    // ...

    /* CRC64 checksum. It will be zero if checksum computation is disabled, the
     * loading code skips the check in this case. */
    cksum = rdb->cksum;
    memrev64ifbe(&cksum); // 写入校验值
    if (rioWrite(rdb,&cksum,8) == 0) goto werr;

    // ...
}

int rdbSaveKeyValuePair(rio *rdb, robj *key, robj *val, long long expiretime) {
    // ...
    /* Save type, key, value */
    // 写入类型，key以string的方式写，写入value
    if (rdbSaveObjectType(rdb,val) == -1) return -1;
    if (rdbSaveStringObject(rdb,key) == -1) return -1;
    if (rdbSaveObject(rdb,val,key) == -1) return -1;
    // ...
}

/* Save a Redis object.
 * Returns -1 on error, number of bytes written on success. */
ssize_t rdbSaveObject(rio *rdb, robj *o, robj *key) {
    ssize_t n = 0, nwritten = 0;

    if (o->type == OBJ_STRING) {
        /* Save a string value */
        if ((n = rdbSaveStringObject(rdb,o)) == -1) return -1;
        nwritten += n;
    } else if (o->type == OBJ_LIST) {
        // ...
    }
}
```

可以看到，最后不同类型的值，以不同方式写入，具体的RDB文件的格式，可以参考
[这篇文章](https://github.com/sripathikrishnan/redis-rdb-tools/wiki/Redis-RDB-Dump-File-Format)。

## 总结

这就是对RDB的具体介绍，RDB可以定期将数据库中的内容dump到磁盘，但是及时性与AOF还是差得比较远，但是一般来说，
我们可以同时打开AOF和RDB，这样就可以获得一个比较不错的备份效果。

最后提一句，Redis在启动的时候，也就是 `main` 函数里，会调用 `loadDataFromDisk()` 从磁盘恢复数据，而恢复的逻辑是
如果有AOF，那么优先从AOF获取，否则从RDB获取：

```c
void loadDataFromDisk(void) {
    long long start = ustime();
    if (server.aof_state == AOF_ON) {
        if (loadAppendOnlyFile(server.aof_filename) == C_OK)
            serverLog(LL_NOTICE,"DB loaded from append only file: %.3f seconds",(float)(ustime()-start)/1000000);
    } else {
        rdbSaveInfo rsi = RDB_SAVE_INFO_INIT;
        errno = 0; /* Prevent a stale value from affecting error checking */
        if (rdbLoad(server.rdb_filename,&rsi,RDBFLAGS_NONE) == C_OK) {
    // ...
```

---

ref:

- https://redis.io/topics/persistence
- https://github.com/sripathikrishnan/redis-rdb-tools/wiki/Redis-RDB-Dump-File-Format
