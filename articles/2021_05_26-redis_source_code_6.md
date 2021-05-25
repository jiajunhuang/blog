# Redis源码阅读：AOF重写

Redis会自动进行AOF重写，也可以由 `BGREWRITEAOF` 命令手动触发重写。我们来看看，从 `BGREWRITEAOF` 开始入手：

```c
{"bgrewriteaof",bgrewriteaofCommand,1,
    "admin no-script",
    0,NULL,0,0,0,0,0,0},

void bgrewriteaofCommand(client *c) {
    if (server.child_type == CHILD_TYPE_AOF) {
        addReplyError(c,"Background append only file rewriting already in progress");
    } else if (hasActiveChildProcess()) {
        server.aof_rewrite_scheduled = 1;
        addReplyStatus(c,"Background append only file rewriting scheduled");
    } else if (rewriteAppendOnlyFileBackground() == C_OK) {
        addReplyStatus(c,"Background append only file rewriting started");
    } else {
        addReplyError(c,"Can't execute an AOF background rewriting. "
                        "Please check the server logs for more information.");
    }
}

/* This is how rewriting of the append only file in background works:
 *
 * 1) The user calls BGREWRITEAOF
 * 2) Redis calls this function, that forks():
 *    2a) the child rewrite the append only file in a temp file.
 *    2b) the parent accumulates differences in server.aof_rewrite_buf.
 * 3) When the child finished '2a' exists.
 * 4) The parent will trap the exit code, if it's OK, will append the
 *    data accumulated into server.aof_rewrite_buf into the temp file, and
 *    finally will rename(2) the temp file in the actual file name.
 *    The the new file is reopened as the new append only file. Profit!
 */
int rewriteAppendOnlyFileBackground(void) {
    pid_t childpid;

    if (hasActiveChildProcess()) return C_ERR;
    if (aofCreatePipes() != C_OK) return C_ERR;
    if ((childpid = redisFork(CHILD_TYPE_AOF)) == 0) {
        // fork，子进程负责重写AOF
        char tmpfile[256];

        /* Child */
        redisSetProcTitle("redis-aof-rewrite");
        redisSetCpuAffinity(server.aof_rewrite_cpulist);
        snprintf(tmpfile,256,"temp-rewriteaof-bg-%d.aof", (int) getpid());
        if (rewriteAppendOnlyFile(tmpfile) == C_OK) { // 重写AOF文件
            sendChildCowInfo(CHILD_INFO_TYPE_AOF_COW_SIZE, "AOF rewrite");
            exitFromChild(0);  // 写完以后，退出
        } else {
            exitFromChild(1);
        }
    } else {
        // 父进程返回后继续执行其余命令
        /* Parent */
        if (childpid == -1) {
            serverLog(LL_WARNING,
                "Can't rewrite append only file in background: fork: %s",
                strerror(errno));
            aofClosePipes();
            return C_ERR;
        }
        serverLog(LL_NOTICE,
            "Background append only file rewriting started by pid %ld",(long) childpid);
        server.aof_rewrite_scheduled = 0;
        server.aof_rewrite_time_start = time(NULL);

        /* We set appendseldb to -1 in order to force the next call to the
         * feedAppendOnlyFile() to issue a SELECT command, so the differences
         * accumulated by the parent into server.aof_rewrite_buf will start
         * with a SELECT statement and it will be safe to merge. */
        server.aof_selected_db = -1;
        replicationScriptCacheFlush();
        return C_OK;
    }
    return C_OK; /* unreached */
}

/* Write a sequence of commands able to fully rebuild the dataset into
 * "filename". Used both by REWRITEAOF and BGREWRITEAOF.
 *
 * In order to minimize the number of commands needed in the rewritten
 * log Redis uses variadic commands when possible, such as RPUSH, SADD
 * and ZADD. However at max AOF_REWRITE_ITEMS_PER_CMD items per time
 * are inserted using a single command. */
int rewriteAppendOnlyFile(char *filename) {
    // ...
    if (rewriteAppendOnlyFileRio(&aof) == C_ERR) goto werr;
    // ...


int rewriteAppendOnlyFileRio(rio *aof) {
    dictIterator *di = NULL;
    dictEntry *de;
    size_t processed = 0;
    int j;
    long key_count = 0;
    long long updated_time = 0;

    for (j = 0; j < server.dbnum; j++) {
        char selectcmd[] = "*2\r\n$6\r\nSELECT\r\n";

    // ...
    if (o->type == OBJ_STRING) {
        /* Emit a SET command */
        char cmd[]="*3\r\n$3\r\nSET\r\n";
        if (rioWrite(aof,cmd,sizeof(cmd)-1) == 0) goto werr;
        /* Key and value */
        if (rioWriteBulkObject(aof,&key) == 0) goto werr;
        if (rioWriteBulkObject(aof,o) == 0) goto werr;
    } else if (o->type == OBJ_LIST) {
        if (rewriteListObject(aof,&key,o) == 0) goto werr;
    } else if (o->type == OBJ_SET) {
        if (rewriteSetObject(aof,&key,o) == 0) goto werr;
    } else if (o->type == OBJ_ZSET) {
        if (rewriteSortedSetObject(aof,&key,o) == 0) goto werr;
    } else if (o->type == OBJ_HASH) {
        if (rewriteHashObject(aof,&key,o) == 0) goto werr;
    } else if (o->type == OBJ_STREAM) {
        if (rewriteStreamObject(aof,&key,o) == 0) goto werr;
    } else if (o->type == OBJ_MODULE) {
        if (rewriteModuleObject(aof,&key,o) == 0) goto werr;
    } else {
        serverPanic("Unknown object type");
    }

    // ...
}
```

父进程在fork之后，在哪里去检测子进程是否退出呢？我猜测是在 `serverCron` 里，然后就去找，果然找到了：

```c
    /* Check if a background saving or AOF rewrite in progress terminated. */
    if (hasActiveChildProcess() || ldbPendingChildren())
    {
        run_with_period(1000) receiveChildInfo();
        checkChildrenDone();
    } else {


// 说明子进程退出之前有保存一些信息
/* Receive info data from child. */
void receiveChildInfo(void) {
    if (server.child_info_pipe[0] == -1) return;

    size_t cow;
    monotime cow_updated;
    size_t keys;
    double progress;
    childInfoType information_type;

    /* Drain the pipe and update child info so that we get the final message. */
    while (readChildInfo(&information_type, &cow, &cow_updated, &keys, &progress)) {
        updateChildInfo(information_type, cow, cow_updated, keys, progress);
    }
}

void checkChildrenDone(void) {
    int statloc = 0;
    pid_t pid;

    if ((pid = waitpid(-1, &statloc, WNOHANG)) != 0) {
        // ...

    if (pid == -1) {
        serverLog(LL_WARNING,"waitpid() returned an error: %s. "
            "child_type: %s, child_pid = %d",
            strerror(errno),
            strChildType(server.child_type),
            (int) server.child_pid);
    } else if (pid == server.child_pid) {
        if (server.child_type == CHILD_TYPE_RDB) {
            backgroundSaveDoneHandler(exitcode, bysignal);
        } else if (server.child_type == CHILD_TYPE_AOF) { // 处理子进程重写AOF的函数在这里
            backgroundRewriteDoneHandler(exitcode, bysignal);
        } else if (server.child_type == CHILD_TYPE_MODULE) {
            ModuleForkDoneHandler(exitcode, bysignal);
        } else {
            serverPanic("Unknown child type %d for child pid %d", server.child_type, server.child_pid);
            exit(1);
        }
        if (!bysignal && exitcode == 0) receiveChildInfo();
        resetChildState();
    } else {
    // ...
}

/* A background append only file rewriting (BGREWRITEAOF) terminated its work.
 * Handle this. */
void backgroundRewriteDoneHandler(int exitcode, int bysignal) {
    if (!bysignal && exitcode == 0) {
        // ...
        newfd = open(tmpfile,O_WRONLY|O_APPEND);
        if (newfd == -1) {
            serverLog(LL_WARNING,
                "Unable to open the temporary AOF produced by the child: %s", strerror(errno));
            goto cleanup;
        }

        // 把重写期间没有写完的命令写入到新的AOF文件里
        if (aofRewriteBufferWrite(newfd) == -1) {
            serverLog(LL_WARNING,
                "Error trying to flush the parent diff to the rewritten AOF: %s", strerror(errno));
            close(newfd);
            goto cleanup;
        }

        // ...
        // 刷盘
        if (server.aof_fsync == AOF_FSYNC_EVERYSEC) {
            aof_background_fsync(newfd);
        } else if (server.aof_fsync == AOF_FSYNC_ALWAYS) {
            latencyStartMonitor(latency);
            if (redis_fsync(newfd) == -1) {
                serverLog(LL_WARNING,
                    "Error trying to fsync the parent diff to the rewritten AOF: %s", strerror(errno));
                close(newfd);
                goto cleanup;
            }
            latencyEndMonitor(latency);
            latencyAddSampleIfNeeded("aof-rewrite-done-fsync",latency);
        }

        // 把重写后的AOF文件重命名为aof文件
        if (rename(tmpfile,server.aof_filename) == -1) {

        // ...
}
```

到此，`BGREWRITEAOF` 的流程就结束了。那么，刚才我们看到，`aofRewriteBufferWrite` 上的注释说，把重写AOF期间，没有
写完的命令写入到新的AOF文件是什么意思呢？原来，在fork之后，父进程还会不断的把新的命令追加到 `server.aof_rewrite_buf_blocks`
这个链表，并且通过 pipe 传输给子进程，这段代码比较不容易看到，在 `rewriteAppendOnlyFileBackground` 里，fork前面有
调用 `aofCreatePipes` 函数创建pipe：

```c
    if (aofCreatePipes() != C_OK) return C_ERR;

/* Create the pipes used for parent - child process IPC during rewrite.
 * We have a data pipe used to send AOF incremental diffs to the child,
 * and two other pipes used by the children to signal it finished with
 * the rewrite so no more data should be written, and another for the
 * parent to acknowledge it understood this new condition. */
int aofCreatePipes(void) {
    int fds[6] = {-1, -1, -1, -1, -1, -1};
    int j;

    if (pipe(fds) == -1) goto error; /* parent -> children data. */
    if (pipe(fds+2) == -1) goto error; /* children -> parent ack. */
    if (pipe(fds+4) == -1) goto error; /* parent -> children ack. */
    /* Parent -> children data is non blocking. */
    if (anetNonBlock(NULL,fds[0]) != ANET_OK) goto error;
    if (anetNonBlock(NULL,fds[1]) != ANET_OK) goto error;
    // 有可读事件时，就会调用 `aofChildPipeReadable`
    if (aeCreateFileEvent(server.el, fds[2], AE_READABLE, aofChildPipeReadable, NULL) == AE_ERR) goto error;

    // 这里有一堆的pipe，用来父子进程间通信
    server.aof_pipe_write_data_to_child = fds[1];
    server.aof_pipe_read_data_from_parent = fds[0];
    server.aof_pipe_write_ack_to_parent = fds[3];
    server.aof_pipe_read_ack_from_child = fds[2];
    server.aof_pipe_write_ack_to_child = fds[5];
    server.aof_pipe_read_ack_from_parent = fds[4];
    server.aof_stop_sending_diff = 0;
    return C_OK;

error:
    serverLog(LL_WARNING,"Error opening /setting AOF rewrite IPC pipes: %s",
        strerror(errno));
    for (j = 0; j < 6; j++) if(fds[j] != -1) close(fds[j]);
    return C_ERR;
}

/* This event handler is called when the AOF rewriting child sends us a
 * single '!' char to signal we should stop sending buffer diffs. The
 * parent sends a '!' as well to acknowledge. */
void aofChildPipeReadable(aeEventLoop *el, int fd, void *privdata, int mask) {
    char byte;
    UNUSED(el);
    UNUSED(privdata);
    UNUSED(mask);

    if (read(fd,&byte,1) == 1 && byte == '!') {
        serverLog(LL_NOTICE,"AOF rewrite child asks to stop sending diffs.");
        server.aof_stop_sending_diff = 1;
        if (write(server.aof_pipe_write_ack_to_child,"!",1) != 1) {
            /* If we can't send the ack, inform the user, but don't try again
             * since in the other side the children will use a timeout if the
             * kernel can't buffer our write, or, the children was
             * terminated. */
            serverLog(LL_WARNING,"Can't send ACK to AOF child: %s",
                strerror(errno));
        }
    }
    /* Remove the handler since this can be called only one time during a
     * rewrite. */
    aeDeleteFileEvent(server.el,server.aof_pipe_read_ack_from_child,AE_READABLE);
}
```

而同时，在 父进程中，每执行完一堆命令之后，都会写AOF，在 `feedAppendOnlyFile` 中的尾部，有这么一段代码：

```c
    /* If a background append only file rewriting is in progress we want to
     * accumulate the differences between the child DB and the current one
     * in a buffer, so that when the child process will do its work we
     * can append the differences to the new append only file. */
    if (server.child_type == CHILD_TYPE_AOF)
        aofRewriteBufferAppend((unsigned char*)buf,sdslen(buf));


/* Append data to the AOF rewrite buffer, allocating new blocks if needed. */
void aofRewriteBufferAppend(unsigned char *s, unsigned long len) {
    listNode *ln = listLast(server.aof_rewrite_buf_blocks);
    aofrwblock *block = ln ? ln->value : NULL;

    while(len) {
        /* If we already got at least an allocated block, try appending
         * at least some piece into it. */
        if (block) {
        // ...
        // 追加到 server.aof_rewrite_buf_blocks 链表中
        // ...
    }

    /* Install a file event to send data to the rewrite child if there is
     * not one already. */
    if (aeGetFileEvents(server.el,server.aof_pipe_write_data_to_child) == 0) {
        aeCreateFileEvent(server.el, server.aof_pipe_write_data_to_child,
            AE_WRITABLE, aofChildWriteDiffData, NULL);
        // 当 server.aof_pipe_write_data_to_child 可写时，执行 aofChildWriteDiffData
    }
}

/* Event handler used to send data to the child process doing the AOF
 * rewrite. We send pieces of our AOF differences buffer so that the final
 * write when the child finishes the rewrite will be small. */
void aofChildWriteDiffData(aeEventLoop *el, int fd, void *privdata, int mask) {
    // 发送数据到pipe，让子进程去读
    listNode *ln;
    aofrwblock *block;
    ssize_t nwritten;
    UNUSED(el);
    UNUSED(fd);
    UNUSED(privdata);
    UNUSED(mask);

    while(1) {
        ln = listFirst(server.aof_rewrite_buf_blocks);
        block = ln ? ln->value : NULL;
        if (server.aof_stop_sending_diff || !block) {
            aeDeleteFileEvent(server.el,server.aof_pipe_write_data_to_child,
                              AE_WRITABLE);
            return;
        }
        if (block->used > 0) {
            nwritten = write(server.aof_pipe_write_data_to_child,
                             block->buf,block->used);
            if (nwritten <= 0) return;
            memmove(block->buf,block->buf+nwritten,block->used-nwritten);
            block->used -= nwritten;
            block->free += nwritten;
        }
        if (block->used == 0) listDelNode(server.aof_rewrite_buf_blocks,ln);
    }
}
```

回到最开始，如果子进程退出了，那么剩余的数据就只会在 `server.aof_rewrite_buf_blocks` 链表里，否则就会不断的往pipe里写。
另外我们最开始说到，Redis自己也会触发AOF重写，只要满足一定的条件，其实这段代码就在 `serverCron` 里：

```c
        /* Trigger an AOF rewrite if needed. */
        if (server.aof_state == AOF_ON &&
            !hasActiveChildProcess() &&
            server.aof_rewrite_perc &&
            server.aof_current_size > server.aof_rewrite_min_size)
        {
            long long base = server.aof_rewrite_base_size ?
                server.aof_rewrite_base_size : 1;
            long long growth = (server.aof_current_size*100/base) - 100;
            if (growth >= server.aof_rewrite_perc) {
                serverLog(LL_NOTICE,"Starting automatic rewriting of AOF on %lld%% growth",growth);
                rewriteAppendOnlyFileBackground();
            }
        }

```

可以看到，自动触发的4个条件，必须全部满足，才会触发，分别是：

- AOF 是开的
- 没有正在重写AOF的子进程
- aof_rewrite_perc 不等于0，aof_rewrite_perc 的注释是：Rewrite AOF if % growth is > M and...，也就是说是一个比率
- server.aof_current_size 大于 server.aof_rewrite_min_size

## 总结

这一篇文章中，我们看到了Redis如何进行重写。首先重写有两种方式，一种是用户手动触发，一种是Redis自动触发。

触发AOF重写以后，Redis首先创建一堆pipe用于父子进程通信，然后fork，父进程返回后继续执行命令以及定期执行 `serverCron`,
子进程进行重写AOF，重写完成后，子进程设置了退出后要保存的信息，然后 `exit(0)` 退出；父进程在 `serverCron` 里会去收集
子进程退出的状态。子进程在重写时，父进程还会不断的将fork之后的AOF分别写到老的AOF文件，以及 `server.aof_rewrite_buf_blocks`
里，以链表的形式保存，并且不断的往子进程pipe里同步；当子进程退出之后，父进程将剩余的 `server.aof_rewrite_buf_blocks`
里的内容写到临时文件，然后将文件重命名，替代原来的AOF文件。

这就是 Redis 重写AOF的整个流程。
