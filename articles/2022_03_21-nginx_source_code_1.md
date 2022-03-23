# Nginx 源码阅读（一）: 启动流程

去年读了一部分Nginx源码，后来耽搁了，最近决定捡起来做完这件事情。说起Nginx源码，
我主要是想了解Nginx的代码组织、架构、master/worker分工、以及一个请求的处理流程。
对其它的我就不是那么感兴趣了。因此这一系列Nginx源码阅读也主要是围绕这些话题来写的。

> 我读的是 Nginx 0.1.0

- [Nginx 源码阅读（一）: 启动流程](https://jiajunhuang.com/articles/2022_03_21-nginx_source_code_1.md.html)
- [Nginx 源码阅读（二）: 请求处理](https://jiajunhuang.com/articles/2022_03_22-nginx_source_code_2.md.html)
- [Nginx 源码阅读（三）: 连接池、内存池、buf](https://jiajunhuang.com/articles/2022_03_23-nginx_source_code_3.md.html)

## 源码组织

首先把代码拉下来，然后切到 `release-0.1.0` 这个tag，这就是我要读的源码。
要了解一份代码，首先我们要了解它的组织结构，然后我们要找到入口，顺着入口一路跟，
通常来说，入口就是 `main` 函数。

先来看看源码组织：

```bash
$ tree -d 
.
├── auto
│   ├── fmt
│   ├── lib
│   │   ├── md5
│   │   ├── openssl
│   │   ├── pcre
│   │   └── zlib
│   ├── os
│   └── types
├── conf
├── docs
│   ├── dtd
│   ├── html
│   ├── text
│   ├── xml
│   │   └── nginx
│   ├── xsls
│   └── xslt
├── objs
│   └── src
│       ├── core
│       ├── event
│       │   └── modules
│       ├── http
│       │   └── modules
│       │       └── proxy
│       ├── imap
│       └── os
│           ├── unix
│           └── win32
└── src
    ├── core
    ├── event
    │   └── modules
    ├── http
    │   └── modules
    │       └── proxy
    ├── imap
    └── os
        ├── unix
        └── win32

41 directories
```

首先 `docs`, `conf`, `objs` 目录我们不管，分别是文档、配置文件和编译出来的目标
文件存放处。`auto` 目录下是各种编译脚本，`src` 下就是源码。

`src` 下主要是5个子目录：

- `core` 一些基本类型和各种数据结构、函数，比如 `string`, `array`, `pool`等
- `event` 事件库封装，`epoll`, `kqueue`, `select` 等
- `http` HTTP相关的所有代码和模块都在这里
- `imap` 邮件相关，我们忽略
- `os` 操作系统相关的函数封装

由上可知，我们主要的关注点都将落在 `core`, `http` 这两个文件夹下，其它的我们
就在要用到的时候，再去考察。

接下来就是找到 `main` 函数，这个简单，搜索一下便是。

```bash
$ ag 'int main'
event/ngx_event_connectex.c
62:int ngx_iocp_new_thread(int main)
120:void ngx_iocp_wait_events(int main)

core/nginx.c
99:int main(int argc, char *const *argv)
```

## 启动流程分析

找到了 `main` 函数，接下来要做的事情就是跟着 `main` 函数一步一步看启动流程，不过我读源码的时候，还需要用gdb
定位一些函数的位置，所以我费了一些时间在本地把Nginx 0.1.0 编译起来，十几年前的代码，还是需要做一些修改才能
跑起来的。我们来看 `main` 函数：

```c
int main(int argc, char *const *argv)
{
    // 声明一些变量
    ngx_int_t          i;
    ngx_log_t         *log;
    ngx_cycle_t       *cycle, init_cycle;
    ngx_core_conf_t   *ccf;
    ngx_master_ctx_t   ctx;

    /* TODO */ ngx_max_sockets = -1;

    ngx_time_init(); // 跟踪代码看了一下，是初始化Nginx内部缓存的时间变量

#if (HAVE_PCRE)
    ngx_regex_init(); // 这个很明显，初始化PCRE，也就是正则表达式
#endif

    ngx_pid = ngx_getpid(); // 获取进程ID

    if (!(log = ngx_log_init_stderr())) {
        return 1;
    }

#if (NGX_OPENSSL)
    ngx_ssl_init(log); // 初始化SSL
#endif

    /* init_cycle->log is required for signal handlers and ngx_getopt() */
    // 初始化 cycle。cycle是一个很重要的变量，简单理解，就是Nginx各种乱七八糟的运行时上下文都存在这里。
    // 参考 https://nginx.org/en/docs/dev/development_guide.html#cycle
    ngx_memzero(&init_cycle, sizeof(ngx_cycle_t));
    init_cycle.log = log;
    ngx_cycle = &init_cycle;

    ngx_memzero(&ctx, sizeof(ngx_master_ctx_t));
    ctx.argc = argc;
    ctx.argv = argv;

    // 内存池：https://nginx.org/en/docs/dev/development_guide.html#pool
    if (!(init_cycle.pool = ngx_create_pool(1024, log))) {
        return 1;
    }

    if (ngx_getopt(&ctx, &init_cycle) == NGX_ERROR) {
        return 1;
    }

    if (ngx_test_config) {
        log->log_level = NGX_LOG_INFO;
    }

    // src/os/unix/ngx_linux_init.c 初始化一些系统相关的东西，我的系统是Linux，所以调用 ngx_linux_init.c
    if (ngx_os_init(log) == NGX_ERROR) {
        return 1;
    }

    // 这里是从环境变量里提取要继承的fd
    if (ngx_add_inherited_sockets(&init_cycle) == NGX_ERROR) {
        return 1;
    }

    ngx_max_module = 0;
    // ngx_modules 在 objs/ngx_modules.c 里
    for (i = 0; ngx_modules[i]; i++) {
        ngx_modules[i]->index = ngx_max_module++;
    }

    // 初始化 cycle。里面会初始化各个模块
    cycle = ngx_init_cycle(&init_cycle);
    if (cycle == NULL) {
        if (ngx_test_config) {
            ngx_log_error(NGX_LOG_EMERG, log, 0,
                          "the configuration file %s test failed",
                          init_cycle.conf_file.data);
        }

        return 1;
    }

    if (ngx_test_config) {
        ngx_log_error(NGX_LOG_INFO, log, 0,
                      "the configuration file %s was tested successfully",
                      init_cycle.conf_file.data);
        return 0;
    }

    ngx_os_status(cycle->log);

    ngx_cycle = cycle;

    ccf = (ngx_core_conf_t *) ngx_get_conf(cycle->conf_ctx, ngx_core_module);

    ngx_process = ccf->master ? NGX_PROCESS_MASTER : NGX_PROCESS_SINGLE;

    if (ngx_create_pidfile(cycle, NULL) == NGX_ERROR) {
        return 1;
    }

    if (ngx_process == NGX_PROCESS_MASTER) {
        // master/worker 模式下要执行的代码
        ngx_master_process_cycle(cycle, &ctx);
    } else {
        // 单进程模式下要执行的代码
        ngx_single_process_cycle(cycle, &ctx);
    }

    return 0;
}
```

看到这里，我们大概知道Nginx是怎么启动的，首先各种初始化，然后判断是否是 master/worker 模式，是的话，就执行
`ngx_master_process_cycle` 函数去处理，但是我们还不知道具体里面发生了什么，所以继续跟下去。

## master 是如何工作的

```c
void ngx_master_process_cycle(ngx_cycle_t *cycle, ngx_master_ctx_t *ctx)
{
    char              *title;
    u_char            *p;
    size_t             size;
    ngx_int_t          i;
    sigset_t           set;
    struct timeval     tv;
    struct itimerval   itv;
    ngx_uint_t         live;
    ngx_msec_t         delay;
    ngx_core_conf_t   *ccf;

    // 设置感兴趣的信号，信号处理函数已经在 ngx_os_init 里设置了
    sigemptyset(&set);
    // ...

    // 设置进程名
    ngx_setproctitle(title);


    ccf = (ngx_core_conf_t *) ngx_get_conf(cycle->conf_ctx, ngx_core_module);

    // 此处创建worker进程
    ngx_start_worker_processes(cycle, ccf->worker_processes,
                               NGX_PROCESS_RESPAWN);

    ngx_new_binary = 0;
    delay = 0;
    live = 1;

    for ( ;; ) {
        sigsuspend(&set);

        ngx_gettimeofday(&tv);
        ngx_time_update(tv.tv_sec);

        // 然后根据全局变量来执行对应动作

        ngx_log_debug0(NGX_LOG_DEBUG_EVENT, cycle->log, 0, "wake up");

        if (ngx_reap) {
            // ...
        }

        if (!live && (ngx_terminate || ngx_quit)) {
            // ...
        }

        if (ngx_terminate) {
            // ...
        }

        if (ngx_quit) {
            // ...
        }

        if (ngx_timer) {
            // ...
        }

        if (ngx_reconfigure) {
            // ...
        }

        if (ngx_restart) {
            // ...
        }

        if (ngx_reopen) {
            // ...
        }

        if (ngx_change_binary) {
            // ...
        }

        if (ngx_noaccept) {
            // ...
        }
    }
}
```

现在我们知道了，master其实不会处理任何连接。master首先进行初始化，设置好感兴趣的信号，然后创建worker，之后master本身
就进入一个无限循环，通过 `sigsuspend` 函数阻塞自身，当收到信号时，通过几个全局变量来决定自己的行为，例如重启，rotate
日志，退出等等。

```c
sig_atomic_t  ngx_reap;
sig_atomic_t  ngx_timer;
sig_atomic_t  ngx_sigio;
sig_atomic_t  ngx_terminate;
sig_atomic_t  ngx_quit;
sig_atomic_t  ngx_reconfigure;
sig_atomic_t  ngx_reopen;
sig_atomic_t  ngx_change_binary;
sig_atomic_t  ngx_noaccept;
```

Nginx通过使用这几个原子变量来指示 `master` 收到信号以后，应该如何处理。我们在很多Go的代码里，也看到过类似的逻辑，例如
用一个 atomic 来存储 `exit` 以指示是否要退出(以前看 Thrift Go 实现就有这样的用法)。

那么，master 是怎么改变这些变量的呢？答案其实就是通过信号。我们上面只看到了 master 设置感兴趣的信号，但是没有看到哪里
设置了信号处理函数，经过搜索，发现其实就在 `ngx_os_init` 函数里，它调用了 `ngx_posix_init`：

```c
ngx_int_t ngx_posix_init(ngx_log_t *log)
{
    ngx_signal_t      *sig;
    struct sigaction   sa;

    ngx_pagesize = getpagesize();

    if (ngx_ncpu == 0) {
        ngx_ncpu = 1;
    }

    // 在这里初始化信号处理函数，看代码，统一都是 ngx_signal_handler
    for (sig = signals; sig->signo != 0; sig++) {
        ngx_memzero(&sa, sizeof(struct sigaction));
        sa.sa_handler = sig->handler;
        sigemptyset(&sa.sa_mask);
        if (sigaction(sig->signo, &sa, NULL) == -1) {
            ngx_log_error(NGX_LOG_EMERG, log, ngx_errno,
                          "sigaction(%s) failed", sig->signame);
            return NGX_ERROR;
        }
    }

    // ...
}
```

秘密就藏在 `sigaction(sig->signo, &sa, NULL)` 函数调用里，`sa.sa_handler` 赋值于 `sig->handler`，`sig` 来自于迭代 `signals`。

```c
ngx_signal_t  signals[] = {
    { ngx_signal_value(NGX_RECONFIGURE_SIGNAL),
      "SIG" ngx_value(NGX_RECONFIGURE_SIGNAL),
      ngx_signal_handler },

    { ngx_signal_value(NGX_REOPEN_SIGNAL),
      "SIG" ngx_value(NGX_REOPEN_SIGNAL),
      ngx_signal_handler },

    { ngx_signal_value(NGX_NOACCEPT_SIGNAL),
      "SIG" ngx_value(NGX_NOACCEPT_SIGNAL),
      ngx_signal_handler },

    { ngx_signal_value(NGX_TERMINATE_SIGNAL),
      "SIG" ngx_value(NGX_TERMINATE_SIGNAL),
      ngx_signal_handler },

    { ngx_signal_value(NGX_SHUTDOWN_SIGNAL),
      "SIG" ngx_value(NGX_SHUTDOWN_SIGNAL),
      ngx_signal_handler },

    { ngx_signal_value(NGX_CHANGEBIN_SIGNAL),
      "SIG" ngx_value(NGX_CHANGEBIN_SIGNAL),
      ngx_signal_handler },

    { SIGALRM, "SIGALRM", ngx_signal_handler },

    { SIGINT, "SIGINT", ngx_signal_handler },

    { SIGIO, "SIGIO", ngx_signal_handler },

    { SIGCHLD, "SIGCHLD", ngx_signal_handler },

    { SIGPIPE, "SIGPIPE, SIG_IGN", SIG_IGN },

    { 0, NULL, NULL }
};
```

可以看到，基本所有的信号处理函数，都是 `ngx_signal_handler`：

```c
void ngx_signal_handler(int signo)
{
    /*
     * 收到对应信号，就设置对应的变量，比如 ngx_reconfigure。它是一个atomic值，
     * 然后 master 进程就会从 sigsuspend 苏醒，然后检测变量进行处理
     */
    char            *action;
    struct timeval   tv;
    ngx_int_t        ignore;
    ngx_err_t        err;
    ngx_signal_t    *sig;

    ignore = 0;

    err = ngx_errno;

    for (sig = signals; sig->signo != 0; sig++) {
        if (sig->signo == signo) {
            break;
        }
    }

    ngx_gettimeofday(&tv);
    ngx_time_update(tv.tv_sec);

    action = "";

    switch (ngx_process) {

    case NGX_PROCESS_MASTER:
    case NGX_PROCESS_SINGLE:
        switch (signo) {

        case ngx_signal_value(NGX_SHUTDOWN_SIGNAL):
            ngx_quit = 1;
            action = ", shutting down";
            break;

        case ngx_signal_value(NGX_TERMINATE_SIGNAL):
        case SIGINT:
            ngx_terminate = 1;
            action = ", exiting";
            break;

        // ...
        }
    }
}
```

至此，我们大概就知道了 master 是如何工作的，它本身接收一些信号，然后通过比较信号值，改变一些内部的全局 atomic 变量，
之后在循环中，通过判断这些 atomic 变量，来做出对应的行为。

## worker 是如何工作的

要看 `worker`，那我们就得回到 `ngx_start_worker_processes`：

```c
static void ngx_start_worker_processes(ngx_cycle_t *cycle, ngx_int_t n,
                                       ngx_int_t type)
{
    ngx_int_t         i;
    ngx_channel_t     ch;
    struct itimerval  itv;

    ngx_log_error(NGX_LOG_INFO, cycle->log, 0, "start worker processes");

    ch.command = NGX_CMD_OPEN_CHANNEL;

    while (n--) {
        // 起n个worker进程，执行 ngx_worker_process_cycle
        ngx_spawn_process(cycle, ngx_worker_process_cycle, NULL,
                          "worker process", type);

        ch.pid = ngx_processes[ngx_process_slot].pid;
        ch.slot = ngx_process_slot;
        ch.fd = ngx_processes[ngx_process_slot].channel[0];

        for (i = 0; i < ngx_last_process; i++) {

            if (i == ngx_process_slot
                || ngx_processes[i].pid == -1
                || ngx_processes[i].channel[0] == -1)
            {
                continue;
            }

            ngx_log_debug6(NGX_LOG_DEBUG_CORE, cycle->log, 0,
                           "pass channel s:%d pid:" PID_T_FMT
                           " fd:%d to s:%d pid:" PID_T_FMT " fd:%d",
                           ch.slot, ch.pid, ch.fd,
                           i, ngx_processes[i].pid,
                           ngx_processes[i].channel[0]);

            /* TODO: NGX_AGAIN */

            ngx_write_channel(ngx_processes[i].channel[0],
                              &ch, sizeof(ngx_channel_t), cycle->log);
        }
    }

    /*
     * we have to limit the maximum life time of the worker processes
     * by 10 days because our millisecond event timer is limited
     * by 24 days on 32-bit platforms
     */

    itv.it_interval.tv_sec = 0;
    itv.it_interval.tv_usec = 0;
    itv.it_value.tv_sec = 10 * 24 * 60 * 60;
    itv.it_value.tv_usec = 0;

    if (setitimer(ITIMER_REAL, &itv, NULL) == -1) {
        ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                      "setitimer() failed");
    }
}
```

这个函数里，有两个很重要的东西，第一个是 `ngx_spawn_process(cycle, ngx_worker_process_cycle, NULL, "worker process", type)`，
第二个是 `ngx_write_channel(ngx_processes[i].channel[0], &ch, sizeof(ngx_channel_t), cycle->log);`。我们先来跟踪第一个：

```c
ngx_pid_t ngx_spawn_process(ngx_cycle_t *cycle,
                            ngx_spawn_proc_pt proc, void *data,
                            char *name, ngx_int_t respawn)
{
    // ...
    pid = fork();

    switch (pid) {

    case -1:
        ngx_log_error(NGX_LOG_ALERT, cycle->log, ngx_errno,
                      "fork() failed while spawning \"%s\"", name);
        ngx_close_channel(ngx_processes[s].channel, cycle->log);
        return NGX_ERROR;

    case 0:
        ngx_pid = ngx_getpid();
        proc(cycle, data); // 看起来这里最终也是要返回然后设置下面的这些东西。此处的 proc 是 ngx_worker_process_cycle
        break;

    default:
        break;
    }
}
```

这里的 `proc` 就是调用 `ngx_spawn_process` 时传入的 `ngx_worker_process_cycle`：

```c
static void ngx_worker_process_cycle(ngx_cycle_t *cycle, void *data)
{
    // ...

    for (i = 0; ngx_modules[i]; i++) {
        if (ngx_modules[i]->init_process) {
            // 这里调用 init_process，此处就会初始化事件模块等
            if (ngx_modules[i]->init_process(cycle) == NGX_ERROR) {
                /* fatal */
                exit(2);
            }
        }
    }

    // ...

    for ( ;; ) {
        // worker 真正执行处理事件的地方
        ngx_process_events(cycle);

        // ...

        if (ngx_terminate) {
            // ...
        }

        if (ngx_quit) {
            // ...
        }

        if (ngx_reopen) {
            // ...
        }

        // ...
    }
}
```

这就是 `worker` 大概的启动流程，由 `master` fork之后，初始化各个模块，然后进入自身的工作循环，执行 `ngx_process_events`
处理各种事件，`ngx_process_events` 的定义为：

```c
// src/event/ngx_event.h 里
#define ngx_process_events   ngx_event_actions.process_events

// src/event/ngx_event.c 里
ngx_event_actions_t               ngx_event_actions;

// src/event/ngx_event.h 里
typedef struct {
    // 添加/删除事件
    ngx_int_t  (*add)(ngx_event_t *ev, int event, u_int flags);
    ngx_int_t  (*del)(ngx_event_t *ev, int event, u_int flags);

    ngx_int_t  (*enable)(ngx_event_t *ev, int event, u_int flags);
    ngx_int_t  (*disable)(ngx_event_t *ev, int event, u_int flags);

    // 添加和删除连接，也就是对读和写都进行处理
    ngx_int_t  (*add_conn)(ngx_connection_t *c);
    ngx_int_t  (*del_conn)(ngx_connection_t *c, u_int flags);

    ngx_int_t  (*process_changes)(ngx_cycle_t *cycle, ngx_uint_t try);
    // 工作循环中处理事件
    ngx_int_t  (*process_events)(ngx_cycle_t *cycle);

    ngx_int_t  (*init)(ngx_cycle_t *cycle);
    void       (*done)(ngx_cycle_t *cycle);
} ngx_event_actions_t;


extern ngx_event_actions_t   ngx_event_actions;
```

可以看到，`ngx_event_actions` 就是事件模块，而其中的 `process_events` 就是处理事件的函数，而且 `ngx_event_actions` 是
一个全局变量。试着想想，Nginx要支持跨平台，而且多个 I/O 多路复用库，这里又是一个全局变量，我怀疑是编译的时候生成，然后
具体初始化某个 I/O 多路复用 的时候赋值的，搜索一下代码：

```bash
$ ag ngx_event_actions
src/event/ngx_event.c
59:ngx_event_actions_t               ngx_event_actions;

src/event/ngx_event.h
219:} ngx_event_actions_t;
222:extern ngx_event_actions_t   ngx_event_actions;
386:#define ngx_process_changes  ngx_event_actions.process_changes
387:#define ngx_process_events   ngx_event_actions.process_events
388:#define ngx_done_events      ngx_event_actions.done
390:#define ngx_add_event        ngx_event_actions.add
391:#define ngx_del_event        ngx_event_actions.del
392:#define ngx_add_conn         ngx_event_actions.add_conn
393:#define ngx_del_conn         ngx_event_actions.del_conn
435:    ngx_event_actions_t     actions;

src/event/modules/ngx_aio_module.c
78:    ngx_event_actions = ngx_aio_module_ctx.actions;

src/event/modules/ngx_epoll_module.c
169:    ngx_event_actions = ngx_epoll_module_ctx.actions;

src/event/modules/ngx_kqueue_module.c
189:    ngx_event_actions = ngx_kqueue_module_ctx.actions;

src/event/modules/ngx_iocp_module.c
112:    ngx_event_actions = ngx_iocp_module_ctx.actions;

src/event/modules/ngx_poll_module.c
101:    ngx_event_actions = ngx_poll_module_ctx.actions;

src/event/modules/ngx_select_module.c
113:    ngx_event_actions = ngx_select_module_ctx.actions;

src/event/modules/ngx_rtsig_module.c
163:    ngx_event_actions = ngx_rtsig_module_ctx.actions;
529:        ngx_event_actions.process_events = ngx_rtsig_process_overflow;
765:    ngx_event_actions.process_events = ngx_rtsig_process_events;

src/event/modules/ngx_devpoll_module.c
168:    ngx_event_actions = ngx_devpoll_module_ctx.actions;
```

可以看到，代码上确实是在多个 I/O 多路复用的模块里分别赋值的，我的系统是Linux，所以肯定是用 epoll，跳到 `ngx_epoll_module.c`
看一下：

```c
static int ngx_epoll_init(ngx_cycle_t *cycle)
{
    // ...
    ngx_event_actions = ngx_epoll_module_ctx.actions;
    // ...
}
```

好，就此打住。我们大概知道了worker的流程，首先由 `master` fork，然后初始化各个模块，这里就包括了 I/O 多路复用模块，
然后进入自身的工作循环，执行 `ngx_process_events` 处理各种事件，`ngx_process_events` 就是由 I/O 多路复用模块提供的。

每次执行完之后，也是通过判断几个全局变量来决定worker的行为。

## master/worker 通信

刚才我们说到 `ngx_write_channel(ngx_processes[i].channel[0], &ch, sizeof(ngx_channel_t), cycle->log)`，我们现在来看看它是
干啥的：

```c
// master/worker 之间进程间通信
ngx_int_t ngx_write_channel(ngx_socket_t s, ngx_channel_t *ch, size_t size,
                            ngx_log_t *log) 
{
    ssize_t             n;
    ngx_err_t           err;
    struct iovec        iov[1];
    struct msghdr       msg;

#if (HAVE_MSGHDR_MSG_CONTROL)

    union {
        struct cmsghdr  cm;
        char            space[CMSG_SPACE(sizeof(int))];
    } cmsg;

    if (ch->fd == -1) {
        msg.msg_control = NULL;
        msg.msg_controllen = 0;

    } else {
        msg.msg_control = (caddr_t) &cmsg;
        msg.msg_controllen = sizeof(cmsg);

        cmsg.cm.cmsg_len = sizeof(cmsg);
        cmsg.cm.cmsg_level = SOL_SOCKET; 
        cmsg.cm.cmsg_type = SCM_RIGHTS;
        *(int *) CMSG_DATA(&cmsg.cm) = ch->fd;
    }

#else

    if (ch->fd == -1) {
        msg.msg_accrights = NULL;
        msg.msg_accrightslen = 0;

    } else {
        msg.msg_accrights = (caddr_t) &ch->fd;
        msg.msg_accrightslen = sizeof(int);
    }

#endif

    iov[0].iov_base = (char *) ch;
    iov[0].iov_len = size;

    msg.msg_name = NULL;
    msg.msg_namelen = 0;
    msg.msg_iov = iov;
    msg.msg_iovlen = 1;

    n = sendmsg(s, &msg, 0);

    if (n == -1) {
        err = ngx_errno;
        if (err == NGX_EAGAIN) {
            return NGX_AGAIN;
        }

        ngx_log_error(NGX_LOG_ALERT, log, err, "sendmsg() failed");
        return NGX_ERROR;
    }

    return NGX_OK;
}
```

这里其实是 master 和 worker 之间通信的代码，也是通过网络，传输的结构体定义为：

```c
typedef struct {
     ngx_uint_t  command;
     ngx_pid_t   pid;
     ngx_int_t   slot;
     ngx_fd_t    fd;
} ngx_channel_t;
```

可以看到，`ch` 其实是一个写往通道里的命令。而 `channel` 的定义则是 `ngx_processes[i].channel[0]`，
其实它是一个 socket 的 fd。那么这个 `ngx_processes` 是什么时候初始化的呢？
答案就在 `ngx_spawn_process` 中：

```c
ngx_pid_t ngx_spawn_process(ngx_cycle_t *cycle,
                            ngx_spawn_proc_pt proc, void *data,
                            char *name, ngx_int_t respawn)
{
    u_long     on;
    ngx_pid_t  pid;
    ngx_int_t  s;

    if (respawn >= 0) {
        s = respawn;

    } else {
        for (s = 0; s < ngx_last_process; s++) {
            if (ngx_processes[s].pid == -1) {
                break;
            }
        }

        if (s == NGX_MAX_PROCESSES) {
            ngx_log_error(NGX_LOG_ALERT, cycle->log, 0,
                          "no more than %d processes can be spawned",
                          NGX_MAX_PROCESSES);
            return NGX_ERROR;
        }
    }
    // ...
        if (socketpair(AF_UNIX, SOCK_STREAM, 0, ngx_processes[s].channel) == -1)
    // ...
    pid = fork();
    // ...
}
```

在 fork 之前，Nginx会在 `ngx_processes` 里找到一个最小的位置，然后用 `socketpair`
创建一个 socket pair 用于 master 和 worker 通信。

## 总结

看到这里，这第一篇终于可以收尾了，在这一篇源码阅读中，我们从 `main` 函数入手，依次跟踪了 `main` 函数的启动过程，
然后我们抵达了 `ngx_master_process_cycle`，从这里开始，Nginx便会进入 master/worker 模型的代码。接着我们了解了
master 大概的启动流程，worker的启动流程，以及他们之间是如何通信的，以及master和worker大致的工作模型。

接下来的文章，我们将聚焦到 worker ，看 worker 中是如何初始化模块的，以及worker是如何处理请求的。

---

参考资料：

- https://nginx.org/en/docs/dev/development_guide.html#introduction
