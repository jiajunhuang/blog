# Redis是如何工作的？

黄健宏老师的《Redis源码剖析》采用自底向上的讲述方法，先从Redis的数据结构讲起。但是，Redis是怎么从监听端口到执行这些
命令的呢？本文记录了我探索这一问题的过程(Redis 6.0)。

如果是一个阻塞版本的echo服务器，大概长成这样：

```python
# https://pymotw.com/3/socket/tcp.html
import socket
import sys

# Create a TCP/IP socket
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

# Bind the socket to the port
server_address = ('localhost', 10000)
print('starting up on {} port {}'.format(*server_address))
sock.bind(server_address)

# Listen for incoming connections
sock.listen(1)

while True:
    # Wait for a connection
    print('waiting for a connection')
    connection, client_address = sock.accept()
    try:
        print('connection from', client_address)

        # Receive the data in small chunks and retransmit it
        while True:
            data = connection.recv(16)
            print('received {!r}'.format(data))
            if data:
                print('sending data back to the client')
                connection.sendall(data)
            else:
                print('no data from', client_address)
                break

    finally:
        # Clean up the connection
        connection.close()
```

可以看到，server是经历了 bind - listen - accept - recv 这几个步骤，然后才能拿到数据，再通过sendall写入数据，但是这是
一个阻塞版本的，也就是说，由于这里没有使用线程池，一次最多只能处理一个请求，大家排成队，一个一个来。

Redis可不能这样设计，为了处理高并发请求，Redis采用了I/O多路复用，在Linux下，就是epoll，但是Redis设计的时候就是想要在
多个系统下可以使用，因此它对I/O多路复用进行了一个封装，这个抽象对程序来说增加了跨平台能力，但是对于读代码的人来说却
增加了心智负担。总的来说，由于以下两点，增加了阅读难度：

- 使用I/O多路复用，把本来连贯的代码拆成了各种回调函数，本来一整块的代码，变成了碎片
- 抽象提高了软件健壮性，但是(相比面条式代码从上往下读过去即可)增加了心智负担

不过我们还是要克服这些困难，去理解Redis到底咋工作的。

## 读起来

首先我们要大概的了解一下Redis源码的结构，可以看一下README。然后我们要找到 `main` 函数所在处，因为在C里，main函数是
用户程序的入口，可以通过搜索，也可以通过gdb，搜索的话，因为Redis里有很多测试代码里都有main函数，所以可能不好找：

```bash
$ ack -Q 'int main' | wc
     31     174    1911
```

所以我们用gdb：

```bash
$ make -j8
$ gdb --args ./src/redis-server --port 6380
GNU gdb (Debian 8.2.1-2+b3) 8.2.1
Copyright (C) 2018 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
Type "show copying" and "show warranty" for details.
This GDB was configured as "x86_64-linux-gnu".
Type "show configuration" for configuration details.
For bug reporting instructions, please see:
<http://www.gnu.org/software/gdb/bugs/>.
Find the GDB manual and other documentation resources online at:
    <http://www.gnu.org/software/gdb/documentation/>.

For help, type "help".
Type "apropos word" to search for commands related to "word"...
Reading symbols from ./src/redis-server...done.
(gdb) 
(gdb) b main
Breakpoint 1 at 0x3a0c0: file server.c, line 4874.
 
```

就知道了，在 `server.c` 的 4874 行(行数可能不同，下同，因为我自己加了点私货代码在里面)。

可以喵一喵 main 函数的代码，不过没关系，肯定喵不出什么特别多的东西，但是大概知道就可以了：

- 最后几行是 `aeMain(server.el);` 这里是事件循环，那么在此之前肯定有地方是初始化server
- 往前翻可以翻到 `initServer();`

这两个都很长，说实话，我是看不出什么细节来的。所以，我得找一个命令，然后下个断点，接着来看函数调用栈，翻了一下 `server.c`
发现上面有一个这样的代码：

```c
struct redisCommand redisCommandTable[] = {
    {"module",moduleCommand,-2,
     "admin no-script",
     0,NULL,0,0,0,0,0,0},

    {"get",getCommand,2,
     "read-only fast @string",
     0,NULL,1,1,1,0,0,0},

    /* Note that we can't flag set as fast, since it may perform an
     * implicit DEL of a large key. */
    {"set",setCommand,-3,
     "write use-memory @string",
     0,NULL,1,1,1,0,0,0},

```

这是Redis的命令，set，就是我们要找的，看看 `setCommand` 是个啥：

```c
/* SET key value [NX] [XX] [KEEPTTL] [EX <seconds>] [PX <milliseconds>] */
void setCommand(client *c) {
    int j;
    ...
```

是个函数，非常好，给它下断点！

```bash
$ gdb --args ./src/redis-server --port 6380
GNU gdb (Debian 8.2.1-2+b3) 8.2.1
Copyright (C) 2018 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
Type "show copying" and "show warranty" for details.
This GDB was configured as "x86_64-linux-gnu".
Type "show configuration" for configuration details.
For bug reporting instructions, please see:
<http://www.gnu.org/software/gdb/bugs/>.
Find the GDB manual and other documentation resources online at:
    <http://www.gnu.org/software/gdb/documentation/>.

For help, type "help".
Type "apropos word" to search for commands related to "word"...
Reading symbols from ./src/redis-server...done.
(gdb) b setCommand 
Breakpoint 1 at 0x6a5d0: file t_string.c, line 103.
(gdb) run
Starting program: /home/jiajun/Code/redis/src/redis-server --port 6380
[Thread debugging using libthread_db enabled]
Using host libthread_db library "/lib/x86_64-linux-gnu/libthread_db.so.1".
32392:C 21 May 2020 14:08:59.432 # oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
32392:C 21 May 2020 14:08:59.432 # Redis version=5.9.101, bits=64, commit=96851c4e, modified=1, pid=32392, just started
32392:C 21 May 2020 14:08:59.432 # Configuration loaded
32392:M 21 May 2020 14:08:59.433 * Increased maximum number of open files to 10032 (it was originally set to 1024).

```

起一个 `redis-cli` 连上这个服务器，然后执行 `set a b`，这个时候就会卡在断点上：

```bash
Thread 1 "redis-server" hit Breakpoint 1, setCommand (c=0x7ffff791c300) at t_string.c:103
103	    for (j = 3; j < c->argc; j++) {
(gdb) bt
#0  setCommand (c=0x7ffff791c300) at t_string.c:103
#1  0x000055555559786d in call (c=0x7ffff791c300, flags=15) at server.c:3195
#2  0x000055555559813c in processCommand (c=c@entry=0x7ffff791c300) at server.c:3549
#3  0x00005555555a57e0 in processCommandAndResetClient (c=c@entry=0x7ffff791c300) at networking.c:1655
#4  0x00005555555a974f in processInputBuffer (c=0x7ffff791c300) at networking.c:1750
#5  0x0000555555623083 in callHandler (handler=0x5555555aa000 <readQueryFromClient>, conn=0x7ffff7815100) at connhelpers.h:78
#6  connSocketEventHandler (el=<optimized out>, fd=<optimized out>, clientData=0x7ffff7815100, mask=<optimized out>) at connection.c:276
#7  0x0000555555591654 in aeProcessEvents (eventLoop=eventLoop@entry=0x7ffff780b480, flags=flags@entry=11) at ae.c:459
#8  0x000055555559189b in aeMain (eventLoop=0x7ffff780b480) at ae.c:519
#9  0x000055555558e59b in main (argc=<optimized out>, argv=0x7fffffffe0b8) at server.c:5048

```

非常好，这就是我们要的东西，我们从上往下，依次找一下是怎么被调用的，然后再反过来看一下，就知道Redis是怎么依次把它们
设置好的(步骤我以代码注释的方式写)：

```c
/* SET key value [NX] [XX] [KEEPTTL] [EX <seconds>] [PX <milliseconds>] */
void setCommand(client *c) {
    int j;
    robj *expire = NULL;
    int unit = UNIT_SECONDS;
    int flags = OBJ_SET_NO_FLAGS;

    for (j = 3; j < c->argc; j++) {

// TODO 所以我们要看哪里调用的 setCommand，看调用栈：at server.c:3549

int processCommand(client *c) {
    moduleCallCommandFilters(c);
    ...
    } else {
    call(c,CMD_CALL_FULL);
    ...

// TODO 所以我们继续往上翻 networking.c:1655

int processCommandAndResetClient(client *c) {
    int deadclient = 0;
    server.current_client = c;
    if (processCommand(c) == C_OK) {

// TODO 继续 networking.c:1750

void processInputBuffer(client *c) {
    /* Keep processing while there is something in the input buffer */
    while(c->qb_pos < sdslen(c->querybuf)) {
    ...
    /* We are finally ready to execute the command. */
    if (processCommandAndResetClient(c) == C_ERR) {

// TODO 继续 connhelpers.h:78

static inline int callHandler(connection *conn, ConnectionCallbackFunc handler) {
    printf("=== callHandler\n");
    conn->flags |= CONN_FLAG_IN_HANDLER;
    if (handler) handler(conn);
    conn->flags &= ~CONN_FLAG_IN_HANDLER;
    if (conn->flags & CONN_FLAG_CLOSE_SCHEDULED) {
        connClose(conn);
        return 0;
    }
    return 1;
}

// 继续 connection.c:276

static void connSocketEventHandler(struct aeEventLoop *el, int fd, void *clientData, int mask)
{
    printf("=== connSocketEventHandler: fd: %d\n", fd);
    UNUSED(el);
    UNUSED(fd);
    connection *conn = clientData;
    ...
    /* Handle normal I/O flows */
    if (!invert && call_read) {
        if (!callHandler(conn, conn->read_handler)) return;
    }

// 继续 ae.c:459
            if (!invert && fe->mask & mask & AE_READABLE) {
                printf("=== handle readable event of fd %d\n", fd);
                fe->rfileProc(eventLoop,fd,fe->clientData,mask);
                fired++;
            }


// 到这里，似乎断掉了联系。我们回去看看 connSocketEventHandler 在哪些地方被调用了，搜索：
ConnectionType CT_Socket = {
    .ae_handler = connSocketEventHandler,
    .close = connSocketClose,
    .write = connSocketWrite,
    .read = connSocketRead,
    .accept = connSocketAccept,
    .connect = connSocketConnect,
    .set_write_handler = connSocketSetWriteHandler,
    .set_read_handler = connSocketSetReadHandler,
    .get_last_error = connSocketGetLastError,
    .blocking_connect = connSocketBlockingConnect,
    .sync_write = connSocketSyncWrite,
    .sync_read = connSocketSyncRead,
    .sync_readline = connSocketSyncReadLine
};

// 搜索 CT_Socket
connection *connCreateSocket() {
    connection *conn = zcalloc(sizeof(connection));
    conn->type = &CT_Socket;
    conn->fd = -1;

    return conn;
}

// 搜索 connCreateSocket
connection *connCreateAcceptedSocket(int fd) {
    connection *conn = connCreateSocket();
    conn->fd = fd;
    conn->state = CONN_STATE_ACCEPTING;
    return conn;
}

// 搜索 connCreateAcceptedSocket
void acceptTcpHandler(aeEventLoop *el, int fd, void *privdata, int mask) {
    int cport, cfd, max = MAX_ACCEPTS_PER_CALL;
    char cip[NET_IP_STR_LEN];
    UNUSED(el);
    UNUSED(mask);
    UNUSED(privdata);

    while(max--) {
        cfd = anetTcpAccept(server.neterr, fd, cip, sizeof(cip), &cport);
        if (cfd == ANET_ERR) {
            if (errno != EWOULDBLOCK)
                serverLog(LL_WARNING,
                    "Accepting client connection: %s", server.neterr);
            return;
        }
        serverLog(LL_VERBOSE,"Accepted %s:%d", cip, cport);
        acceptCommonHandler(connCreateAcceptedSocket(cfd),0,cip);
    }
}

// 搜索 acceptTcpHandler

void initServer(void) {
    int j;
    ...
    for (j = 0; j < server.ipfd_count; j++) {
        if (aeCreateFileEvent(server.el, server.ipfd[j], AE_READABLE,
            acceptTcpHandler,NULL) == AE_ERR)

```

到这里，就比较明了了，Redis首先在 `initServer` 里设置好 `acceptTcpHandler`，然后这个里面提供了回调，在建立好了连接之后，
根据不同的命令，调用不同的 xxxCommand 函数去处理。

好了，分析到此为止，至于Redis的数据结构嘛，还是看《Redis源码分析》或者算法书吧 :)
