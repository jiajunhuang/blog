# Redis源码阅读：AOF持久化

都说Redis是内存数据库，其实 Redis 也有持久化机制，就是我们在 `redis.conf` 里配置的如下几行：

```c
appendonly no

# The name of the append only file (default: "appendonly.aof")

appendfilename "appendonly.aof"

# The fsync() call tells the Operating System to actually write data on disk
# instead of waiting for more data in the output buffer. Some OS will really flush
# data on disk, some other OS will just try to do it ASAP.
#
# Redis supports three different modes:
#
# no: don't fsync, just let the OS flush the data when it wants. Faster.
# always: fsync after every write to the append only log. Slow, Safest.
# everysec: fsync only one time every second. Compromise.
#
# The default is "everysec", as that's usually the right compromise between
# speed and data safety. It's up to you to understand if you can relax this to
# "no" that will let the operating system flush the output buffer when
# it wants, for better performances (but if you can live with the idea of
# some data loss consider the default persistence mode that's snapshotting),
# or on the contrary, use "always" that's very slow but a bit safer than
# everysec.
#
# More details please check the following article:
# http://antirez.com/post/redis-persistence-demystified.html
#
# If unsure, use "everysec".

# appendfsync always
appendfsync everysec
# appendfsync no
```

如上面的文档所说，AOF的全称就是 `Append Only File`，三个单词的首字母，也就是说，我们只追加到文件里，而不去更改历史内容，
但是 UNIX 中的 `write` 系统调用只保证内容会被加到内核维护的文件写入缓冲区，并不保证一定刷到了磁盘上，所以还需要配合
一个 `fsync` 或者 `fdatasync` 系统调用确保数据已经写到磁盘上，对于 Redis 来说，有三个选项，一个是 `no`，表示 Redis 不主动
刷盘，让操作系统负责刷盘；一个是 `everysec` 表示每秒刷一次；一个是 `always`，表示每次写完AOF文件以后就刷盘一次。

这三种策略中，不刷盘的性能最好，但是安全性最差，比如如果系统挂了，那么上一次落盘之后的数据就没了；如果每秒一次，那么
上一次落盘之后的数据就没了，区别是由于每秒都会落盘一次，损失不会很大，最多会损失2秒的数据（为啥不是1秒？后面我们会看到源码），
这是一种性能和安全性相对都不错的方案；还有就是每次都落盘，数据是安全了，但是性能就上不去了。

接下来我们来看看代码，一开始我想通过搜索来找代码，但是找了半天没有找到，然后我去 `aof.c` 里瞎逛了一下，发现一个函数叫做
`feedAppendOnlyFile`，看函数名称，我严重怀疑就是它，搜索一下用到它的地方：

```c
/* Propagate the specified command (in the context of the specified database id)
 * to AOF and Slaves.
 *
 * flags are an xor between:
 * + PROPAGATE_NONE (no propagation of command at all)
 * + PROPAGATE_AOF (propagate into the AOF file if is enabled)
 * + PROPAGATE_REPL (propagate into the replication link)
 *
 * This should not be used inside commands implementation since it will not
 * wrap the resulting commands in MULTI/EXEC. Use instead alsoPropagate(),
 * preventCommandPropagation(), forceCommandPropagation().
 *
 * However for functions that need to (also) propagate out of the context of a
 * command execution, for example when serving a blocked client, you
 * want to use propagate().
 */
void propagate(struct redisCommand *cmd, int dbid, robj **argv, int argc,
               int flags)
{
    if (!server.replication_allowed)
        return;

    /* Propagate a MULTI request once we encounter the first command which
     * is a write command.
     * This way we'll deliver the MULTI/..../EXEC block as a whole and
     * both the AOF and the replication link will have the same consistency
     * and atomicity guarantees. */
    if (server.in_exec && !server.propagate_in_transaction)
        execCommandPropagateMulti(dbid);

    /* This needs to be unreachable since the dataset should be fixed during 
     * client pause, otherwise data may be lossed during a failover. */
    serverAssert(!(areClientsPaused() && !server.client_pause_in_transaction));

    if (server.aof_state != AOF_OFF && flags & PROPAGATE_AOF)
        feedAppendOnlyFile(cmd,dbid,argv,argc);
    if (flags & PROPAGATE_REPL)
        replicationFeedSlaves(server.slaves,dbid,argv,argc);
}

// 我们继续找哪里用到了这个 `propagate` 函数

void call(client *c, int flags) {
    // ...
    propagate(rop->cmd,rop->dbid,rop->argv,rop->argc,target);
    // ...
}

int processCommand(client *c) {
    // ...

    /* Exec the command */
    if (c->flags & CLIENT_MULTI &&
        c->cmd->proc != execCommand && c->cmd->proc != discardCommand &&
        c->cmd->proc != multiCommand && c->cmd->proc != watchCommand &&
        c->cmd->proc != resetCommand)
    {
        queueMultiCommand(c);
        addReply(c,shared.queued);
    } else {
        call(c,CMD_CALL_FULL);
        c->woff = server.master_repl_offset;
        if (listLength(server.ready_keys))
            handleClientsBlockedOnKeys();
    }

    // ...
}
```

诶？这不就是前面处理命令的函数吗？对的，我们继续看 `call` 函数：

```c
void call(client *c, int flags) {
    // ...
    
    /* Call the command. */
    dirty = server.dirty;
    prev_err_count = server.stat_total_error_replies;
    updateCachedTime(0);
    elapsedStart(&call_timer);
    c->cmd->proc(c);

    // ..
    /* Propagate the command into the AOF and replication link */
    if (flags & CMD_CALL_PROPAGATE &&
        (c->flags & CLIENT_PREVENT_PROP) != CLIENT_PREVENT_PROP)
    {
        // ...

        /* Call propagate() only if at least one of AOF / replication
         * propagation is needed. Note that modules commands handle replication
         * in an explicit way, so we never replicate them automatically. */
        if (propagate_flags != PROPAGATE_NONE && !(c->cmd->flags & CMD_MODULE))
            propagate(c->cmd,c->db->id,c->argv,c->argc,propagate_flags);
    }
}
```

这里我们可以看出一点，AOF是在函数执行成功之后才开始去保存的。

```c
void feedAppendOnlyFile(struct redisCommand *cmd, int dictid, robj **argv, int argc) {
    // ...

    buf = catAppendOnlyGenericCommand(buf,argc,argv);

    /* Append to the AOF buffer. This will be flushed on disk just before
     * of re-entering the event loop, so before the client will get a
     * positive reply about the operation performed. */
     // 追加到AOF的缓冲区
    if (server.aof_state == AOF_ON)
        server.aof_buf = sdscatlen(server.aof_buf,buf,sdslen(buf));

    /* If a background append only file rewriting is in progress we want to
     * accumulate the differences between the child DB and the current one
     * in a buffer, so that when the child process will do its work we
     * can append the differences to the new append only file. */
     // 如果正在重写AOF，同时也追加到重写AOF缓冲区
    if (server.child_type == CHILD_TYPE_AOF)
        aofRewriteBufferAppend((unsigned char*)buf,sdslen(buf));

}
```

这里其实是把AOF的内容加到内存里，仍然没有写到磁盘，那么写磁盘的操作在哪里呢？我继续在 `aof.c` 里寻找，发现了这么一个函数：

```c
/* Write the append only file buffer on disk.
 *
 * Since we are required to write the AOF before replying to the client,
 * and the only way the client socket can get a write is entering when the
 * the event loop, we accumulate all the AOF writes in a memory
 * buffer and write it on disk using this function just before entering
 * the event loop again.
 *
 * About the 'force' argument:
 *
 * When the fsync policy is set to 'everysec' we may delay the flush if there
 * is still an fsync() going on in the background thread, since for instance
 * on Linux write(2) will be blocked by the background fsync anyway.
 * When this happens we remember that there is some aof buffer to be
 * flushed ASAP, and will try to do that in the serverCron() function.
 *
 * However if force is set to 1 we'll write regardless of the background
 * fsync. */
#define AOF_WRITE_LOG_ERROR_RATE 30 /* Seconds between errors logging. */
void flushAppendOnlyFile(int force) {
    ssize_t nwritten;
    int sync_in_progress = 0;
    mstime_t latency;

    if (sdslen(server.aof_buf) == 0) {
        /* Check if we need to do fsync even the aof buffer is empty, */
        // ...
    }
}
```

这里其实就AOF实际上发生 `write` 系统调用的地方，其中有一段逻辑：

```c
    if (server.aof_fsync == AOF_FSYNC_EVERYSEC && !force) {
        /* With this append fsync policy we do background fsyncing.
         * If the fsync is still in progress we can try to delay
         * the write for a couple of seconds. */
        if (sync_in_progress) {
            if (server.aof_flush_postponed_start == 0) {
                /* No previous write postponing, remember that we are
                 * postponing the flush and return. */
                server.aof_flush_postponed_start = server.unixtime;
                return;
            } else if (server.unixtime - server.aof_flush_postponed_start < 2) {
                /* We were already waiting for fsync to finish, but for less
                 * than two seconds this is still ok. Postpone again. */
                return;
            }
            /* Otherwise fall trough, and go write since we can't wait
             * over two seconds. */
            server.aof_delayed_fsync++;
            serverLog(LL_NOTICE,"Asynchronous AOF fsync is taking too long (disk is busy?). Writing the AOF buffer without waiting for fsync to complete, this may slow down Redis.");
        }
    }
```

可以看到，如果当前正在写，并且 `server.aof_flush_postponed_start == 0` 的话，就会跳过这一次；或者如果上一次跳过的时间
和现在相差不到2秒的话，也跳过，但是如果超过2秒，就要写入。这也是上面说最多会丢2秒数据的缘故。

继续看：

```c
    latencyStartMonitor(latency);
    nwritten = aofWrite(server.aof_fd,server.aof_buf,sdslen(server.aof_buf));
    latencyEndMonitor(latency);

    // ...
try_fsync:
    /* Don't fsync if no-appendfsync-on-rewrite is set to yes and there are
     * children doing I/O in the background. */
    if (server.aof_no_fsync_on_rewrite && hasActiveChildProcess())
        return;

    /* Perform the fsync if needed. */
    if (server.aof_fsync == AOF_FSYNC_ALWAYS) {
        /* redis_fsync is defined as fdatasync() for Linux in order to avoid
         * flushing metadata. */
        latencyStartMonitor(latency);
        /* Let's try to get this data on the disk. To guarantee data safe when
         * the AOF fsync policy is 'always', we should exit if failed to fsync
         * AOF (see comment next to the exit(1) after write error above). */
        if (redis_fsync(server.aof_fd) == -1) {
            serverLog(LL_WARNING,"Can't persist AOF for fsync error when the "
              "AOF fsync policy is 'always': %s. Exiting...", strerror(errno));
            exit(1);
        }
        latencyEndMonitor(latency);
        latencyAddSampleIfNeeded("aof-fsync-always",latency);
        server.aof_fsync_offset = server.aof_current_size;
        server.aof_last_fsync = server.unixtime;
    } else if ((server.aof_fsync == AOF_FSYNC_EVERYSEC &&
                server.unixtime > server.aof_last_fsync)) {
        if (!sync_in_progress) {
            aof_background_fsync(server.aof_fd);
            server.aof_fsync_offset = server.aof_current_size;
        }
        server.aof_last_fsync = server.unixtime;
    }

```

这里就是 `write` 调用，以及触发 `fsync` 的逻辑，我们看看 `aofWrite`：

```c
/* This is a wrapper to the write syscall in order to retry on short writes
 * or if the syscall gets interrupted. It could look strange that we retry
 * on short writes given that we are writing to a block device: normally if
 * the first call is short, there is a end-of-space condition, so the next
 * is likely to fail. However apparently in modern systems this is no longer
 * true, and in general it looks just more resilient to retry the write. If
 * there is an actual error condition we'll get it at the next try. */
ssize_t aofWrite(int fd, const char *buf, size_t len) {
    ssize_t nwritten = 0, totwritten = 0;

    while(len) {
        nwritten = write(fd, buf, len);

        if (nwritten < 0) {
            if (errno == EINTR) continue;
            return totwritten ? totwritten : -1;
        }

        len -= nwritten;
        buf += nwritten;
        totwritten += nwritten;
    }

    return totwritten;
}
```

然后看看 `aof_background_fsync`：

```c
/* Starts a background task that performs fsync() against the specified
 * file descriptor (the one of the AOF file) in another thread. */
void aof_background_fsync(int fd) {
    bioCreateFsyncJob(fd);
}

void bioCreateFsyncJob(int fd) {
    struct bio_job *job = zmalloc(sizeof(*job));
    job->fd = fd;

    bioSubmitJob(BIO_AOF_FSYNC, job);
}

void bioSubmitJob(int type, struct bio_job *job) {
    job->time = time(NULL);
    pthread_mutex_lock(&bio_mutex[type]);
    listAddNodeTail(bio_jobs[type],job);
    bio_pending[type]++;
    pthread_cond_signal(&bio_newjob_cond[type]);
    pthread_mutex_unlock(&bio_mutex[type]);
}
```

可以看到，创建了一个类型为 `BIO_AOF_FSYNC` 的异步任务，由另外一个线程去执行，我搜索了一下 `BIO_AOF_FSYNC`：

```c
void *bioProcessBackgroundJobs(void *arg) {
    struct bio_job *job;
    unsigned long type = (unsigned long) arg;
    sigset_t sigset;

    // ...
    } else if (type == BIO_AOF_FSYNC) {
    /* The fd may be closed by main thread and reused for another
        * socket, pipe, or file. We just ignore these errno because
        * aof fsync did not really fail. */
    if (redis_fsync(job->fd) == -1 &&
        errno != EBADF && errno != EINVAL)
    {

```

可以看到，`fdatasync` 函数调用，实际上是由另外一个线程完成的。而如果 `aof_fsync` 设置的是 `always` ，则直接在主线程
就把这个事情给做了。

最后一个问题，`bioProcessBackgroundJobs` 是由哪里触发的呢？继续搜索：

```c
/* Initialize the background system, spawning the thread. */
void bioInit(void) {
    pthread_attr_t attr;
    pthread_t thread;
    size_t stacksize;
    int j;

    // ...
    for (j = 0; j < BIO_NUM_OPS; j++) {
        void *arg = (void*)(unsigned long) j;
        if (pthread_create(&thread,&attr,bioProcessBackgroundJobs,arg) != 0) {
    // ...


void InitServerLast() {
    bioInit();
    initThreadedIO();
    set_jemalloc_bg_thread(server.jemalloc_bg_thread);
    server.initial_memory_usage = zmalloc_used_memory();
}

int main(int argc, char **argv) {
    // ...
    initServer();
    // ...
    moduleInitModulesSystemLast();
    // ...
    moduleLoadFromQueue();
    // ...
    ACLLoadUsersAtStartup();
    // ...
    InitServerLast();
    // ...
    loadDataFromDisk();
    // ...
}
```

所以服务启动的时候，会初始化线程，等待执行任务。那么除了刚才那里会下发任务，还有别的地方会下发任务吗？有的，搜索
`flushAppendOnlyFile`，可以发现在 `server.c` 的 `serverCron` 里有：

```c
    /* AOF postponed flush: Try at every cron cycle if the slow fsync
     * completed. */
    if (server.aof_state == AOF_ON && server.aof_flush_postponed_start)
        flushAppendOnlyFile(0);

    /* AOF write errors: in this case we have a buffer to flush as well and
     * clear the AOF error in case of success to make the DB writable again,
     * however to try every second is enough in case of 'hz' is set to
     * a higher frequency. */
    run_with_period(1000) {
        if (server.aof_state == AOF_ON && server.aof_last_write_status == C_ERR)
            flushAppendOnlyFile(0);
    }

```

此外，还在 `beforeSleep` 里也有，这就确保了每一次执行完发生的事件之后，都会判断是否需要把AOF数据落盘。

## 总结

这一篇文章里，我们看了 Redis 的AOF是如何工作的。Redis执行完命令之后，又从命令重新生成符合 RESP 格式的字符串形式的命令，
保存到 `server.aof_buf` 里，然后在 `serverCron` 里，有定期检查是否要调用 `write` 和 `fsync`，在 `beforeSleep` 里，也有
执行落盘。这就是 Redis AOF的一些基本操作。
