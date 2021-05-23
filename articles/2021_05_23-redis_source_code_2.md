# Redis源码阅读：执行命令

上一篇我们读到，Redis是怎么从启动服务，到开始读取来自socket的字节流。这一篇我们继续看看，如何处理字节流，然后变成命令，
到返回对应数据。

在开始之前，我们得先看看Redis服务端与客户端的通信协议，也就是 [RESP](https://redis.io/topics/protocol)。简单来说，就是：

传输的内容分为5大类，分别以：

- `+` 开头，代表这是一个字符串，simple string，也就是非二进制安全的字符串。然后以 `\r\n` 结尾，比如如果服务端返回 `OK`，那么实际返回的内容是 `+OK\r\n`。
- `$` 开头，代表这是一个字符串，但是是二进制安全的字符串。当然，也可以传输简单的字符串，比如上面的OK，会被传输为 `$2\r\nOK\r\n`，可以看出来，和简单字符串的不同之处在于，最前面告诉了我们内容到底有多长。有一个特例，那就是NULL，表示为 `$-1\r\n`。
- `-` 开头，代表这是一个错误，比如 `-Error message\r\n`，实际上要显示的错误就是 `Error message`，也就是说，中间的部分就是错误信息。
- `:` 开头，代表这是一个数字。比如 `:10000\r\n` 代表10000，而 `:0\r\n` 就是0。
- `*` 开头，代表这是一个数组，比如由foo和bar两个字符串组成的数组，就应该返回为：`"*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"` 。`*2` 代表这个数组有两个元素，后续的内容其实就是上面几种内容的组合。有两个特例：长度为0的数组表示为 `*0\r\n`，空数组表示为 `*-1\r\n`。

注意，文档中有一句话：Clients send commands to the Redis server using RESP Arrays. Similarly certain Redis commands returning collections of elements to the client use RESP Arrays are reply type.
也就是说，Redis服务端和客户端交互，基本上都是用array来装数据的。对协议有了基本的了解之后，我们写一个简单的Go程序来求证，
我们起一个TCP服务，然后打印读到的所有字节流，并且返回NULL给客户端，用Go来写：

```go
package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
)

func handleConn(conn net.Conn) {
	defer conn.Close()

	buf := bufio.NewReader(conn)
	sb := strings.Builder{}

	for {
		// 读取第一个\r\n结尾的
		bs, err := buf.ReadBytes('\n')
		if err != nil {
			log.Printf("err: %s", err)
			break
		}

		// 如果不是数组，我们就直接panic了
		if bs[0] != '*' {
			log.Panicf("bad bs: %s", bs)
		}

		// 数组里有多少个元素，我们要解析出来，然后读取
		length, err := strconv.ParseUint(string(bs[1:len(bs)-2]), 10, 64)
		if err != nil {
			log.Panicf("bad length: %s", err)
		}

		// 把最开始读到的命令写进去
		sb.Write(bs)
		var i uint64 = 0
		for ; i < length; i++ {
			bs, err = buf.ReadBytes('\n')
			if err != nil {
				log.Printf("err: %s", err)
				break
			}

			if bs[0] == '$' {
				// 如果是复杂字符串，那么就有两个\r\n
				sb.Write(bs)
				bs, _ = buf.ReadBytes('\n')
			}

			sb.Write(bs)
		}

		// 打印出来
		log.Printf("content: %#v", sb.String())
		conn.Write([]byte("+OK\r\n"))

		// 重置
		sb.Reset()
	}
}

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:6389")
	if err != nil {
		log.Panicf("error: %s", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept: %s", err)
			continue
		}

		// 对每一个接受到的请求，都起一个goroutine处理
		go handleConn(conn)
	}
}
```

对于下列输入：

```bash
$ redis-cli -p 6389
127.0.0.1:6389> get hello
OK
127.0.0.1:6389> set hello world EX 10
OK
127.0.0.1:6389> get hello
OK
127.0.0.1:6389> 

```

输出是：

```bash
$ go run main.go 
2021/05/22 17:14:05 content: "*1\r\n$7\r\nCOMMAND\r\n"
2021/05/22 17:14:09 content: "*2\r\n$3\r\nget\r\n$5\r\nhello\r\n"
2021/05/22 17:14:14 content: "*5\r\n$3\r\nset\r\n$5\r\nhello\r\n$5\r\nworld\r\n$2\r\nEX\r\n$2\r\n10\r\n"
2021/05/22 17:14:18 content: "*2\r\n$3\r\nget\r\n$5\r\nhello\r\n"
2021/05/22 17:14:22 err: EOF
^Csignal: interrupt

```

可以看到，`redis-cli` 启动的时候，自动发送了一个 `COMMAND` 命令，然后后面就是我输入的三个命令。

现在我们已经证明了，Redis服务端与客户端之间，的确是用Array+bulk string来传输命令及其参数的，那么接下来我们就来看看
Redis是怎么解析命令，然后执行命令的。上一篇，我们追踪到了 `readQueryFromClient` 函数，我们来大概看一下流程：

```c
void readQueryFromClient(connection *conn) {
    // ...
    qblen = sdslen(c->querybuf);
    if (c->querybuf_peak < qblen) c->querybuf_peak = qblen;
    // 准备空间
    c->querybuf = sdsMakeRoomFor(c->querybuf, readlen);
    // 读取
    nread = connRead(c->conn, c->querybuf+qblen, readlen);
    // ...
    } else if (c->flags & CLIENT_MASTER) {
        // 如果读到了，并且是master的话
        /* Append the query buffer to the pending (not applied) buffer
         * of the master. We'll use this buffer later in order to have a
         * copy of the string applied by the last command executed. */
        c->pending_querybuf = sdscatlen(c->pending_querybuf,
                                        c->querybuf+qblen,nread);
    }

    // ...

    /* There is more data in the client input buffer, continue parsing it
     * in case to check if there is a full command to execute. */
     processInputBuffer(c); // 跟进去看
}

/* This function is called every time, in the client structure 'c', there is
 * more query buffer to process, because we read more data from the socket
 * or because a client was blocked and later reactivated, so there could be
 * pending query buffer, already representing a full command, to process. */
void processInputBuffer(client *c) {
    /* Keep processing while there is something in the input buffer */
    while(c->qb_pos < sdslen(c->querybuf)) {
        // ...
        /* Determine request type when unknown. */
        if (!c->reqtype) {
            if (c->querybuf[c->qb_pos] == '*') {
                c->reqtype = PROTO_REQ_MULTIBULK;
            } else {
                c->reqtype = PROTO_REQ_INLINE;
            }
        }

        if (c->reqtype == PROTO_REQ_INLINE) {
            if (processInlineBuffer(c) != C_OK) break;
            /* If the Gopher mode and we got zero or one argument, process
             * the request in Gopher mode. To avoid data race, Redis won't
             * support Gopher if enable io threads to read queries. */
            if (server.gopher_enabled && !server.io_threads_do_reads &&
                ((c->argc == 1 && ((char*)(c->argv[0]->ptr))[0] == '/') ||
                  c->argc == 0))
            {
                processGopherRequest(c);
                resetClient(c);
                c->flags |= CLIENT_CLOSE_AFTER_REPLY;
                break;
            }
        } else if (c->reqtype == PROTO_REQ_MULTIBULK) {
            // 读取参数
            if (processMultibulkBuffer(c) != C_OK) break;
        } else {
            serverPanic("Unknown request type");
        }

        /* Multibulk processing could see a <= 0 length. */
        if (c->argc == 0) {
            resetClient(c);
        } else {
            /* If we are in the context of an I/O thread, we can't really
             * execute the command here. All we can do is to flag the client
             * as one that needs to process the command. */
            if (c->flags & CLIENT_PENDING_READ) {
                c->flags |= CLIENT_PENDING_COMMAND;
                break;
            }

            // 执行命令
            /* We are finally ready to execute the command. */
            if (processCommandAndResetClient(c) == C_ERR) {
                /* If the client is no longer valid, we avoid exiting this
                 * loop and trimming the client buffer later. So we return
                 * ASAP in that case. */
                return;
            }
        }
    }

    /* Trim to pos */
    if (c->qb_pos) {
        sdsrange(c->querybuf,c->qb_pos,-1);
        c->qb_pos = 0;
    }
}

/* Process the query buffer for client 'c', setting up the client argument
 * vector for command execution. Returns C_OK if after running the function
 * the client has a well-formed ready to be processed command, otherwise
 * C_ERR if there is still to read more buffer to get the full command.
 * The function also returns C_ERR when there is a protocol error: in such a
 * case the client structure is setup to reply with the error and close
 * the connection.
 *
 * This function is called if processInputBuffer() detects that the next
 * command is in RESP format, so the first byte in the command is found
 * to be '*'. Otherwise for inline commands processInlineBuffer() is called. */
int processMultibulkBuffer(client *c) {
    // 读取命令及其参数
}

/* This function calls processCommand(), but also performs a few sub tasks
 * for the client that are useful in that context:
 *
 * 1. It sets the current client to the client 'c'.
 * 2. calls commandProcessed() if the command was handled.
 *
 * The function returns C_ERR in case the client was freed as a side effect
 * of processing the command, otherwise C_OK is returned. */
int processCommandAndResetClient(client *c) {
    int deadclient = 0;
    client *old_client = server.current_client;
    server.current_client = c;
    if (processCommand(c) == C_OK) {
        commandProcessed(c);
    }
    if (server.current_client == NULL) deadclient = 1;
    /*
     * Restore the old client, this is needed because when a script
     * times out, we will get into this code from processEventsWhileBlocked.
     * Which will cause to set the server.current_client. If not restored
     * we will return 1 to our caller which will falsely indicate the client
     * is dead and will stop reading from its buffer.
     */
    server.current_client = old_client;
    /* performEvictions may flush slave output buffers. This may
     * result in a slave, that may be the active client, to be
     * freed. */
    return deadclient ? C_ERR : C_OK;
}

int processCommand(client *c) {
    // ...

    /* Check if the user is authenticated. This check is skipped in case
     * the default user is flagged as "nopass" and is active. */
    int auth_required = (!(DefaultUser->flags & USER_FLAG_NOPASS) ||
                          (DefaultUser->flags & USER_FLAG_DISABLED)) &&
                        !c->authenticated;
    if (auth_required) {

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

    return C_OK;
}
```

如果不是 `MULTI` 的命令，那么就会调用 `call(c, CMD_CALL_FULL)`：

```c
/* Call() is the core of Redis execution of a command.
 *
 * ...
 */
void call(client *c, int flags) {
    // ...

    /* Call the command. */
    dirty = server.dirty;
    prev_err_count = server.stat_total_error_replies;
    updateCachedTime(0);
    elapsedStart(&call_timer);
    c->cmd->proc(c);
    const long duration = elapsedUs(call_timer);
    c->duration = duration;

    // ...
}
```

我们来看看 `c->cmd->proc(c)`：

```c
struct redisCommand {
    char *name;
    redisCommandProc *proc;
    int arity;
    char *sflags;   /* Flags as string representation, one char per flag. */
    uint64_t flags; /* The actual flags, obtained from the 'sflags' field. */
    /* Use a function to determine keys arguments in a command line.
     * Used for Redis Cluster redirect. */
    redisGetKeysProc *getkeys_proc;
    /* What keys should be loaded in background when calling this command? */
    int firstkey; /* The first argument that's a key (0 = no keys) */
    int lastkey;  /* The last argument that's a key */
    int keystep;  /* The step between first and last key */
    long long microseconds, calls, rejected_calls, failed_calls;
    int id;     /* Command ID. This is a progressive ID starting from 0 that
                   is assigned at runtime, and is used in order to check
                   ACLs. A connection is able to execute a given command if
                   the user associated to the connection has this command
                   bit set in the bitmap of allowed commands. */
};
```

这个 `struct redisCommand` 就是Redis里每一个命令了。他们每一个都有一个 `proc` 函数，写明了那个命令应当如何执行，比如
我们来看看 `GET`：

```c
struct redisCommand redisCommandTable[] = {
    {"module",moduleCommand,-2,
     "admin no-script",
     0,NULL,0,0,0,0,0,0},

    {"get",getCommand,2,
     "read-only fast @string",
     0,NULL,1,1,1,0,0,0},

     // ...
}
```

我们看看 `getCommand` 实现：

```c
void getCommand(client *c) {
    getGenericCommand(c);
}

int getGenericCommand(client *c) {
    robj *o;

    if ((o = lookupKeyReadOrReply(c,c->argv[1],shared.null[c->resp])) == NULL)
        return C_OK;

    if (checkType(c,o,OBJ_STRING)) {
        return C_ERR;
    }

    addReplyBulk(c,o);
    return C_OK;
}
```

就是这样。

## 总结

到目前为止，我们已经知道了Redis是如何启动并且准备好接受命令，同时也知道Redis是如何解析命令并且执行的。

---

ref:

- https://redis.io/topics/protocol
