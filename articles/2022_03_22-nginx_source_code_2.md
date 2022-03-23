# Nginx 源码阅读（二）: 请求处理

上一篇我们了解了Nginx的启动流程，master和worker的分工。这一篇我们主要聚焦在Nginx是怎么处理请求的。

- [Nginx 源码阅读（一）: 启动流程](https://jiajunhuang.com/articles/2022_03_21-nginx_source_code_1.md.html)
- [Nginx 源码阅读（二）: 请求处理](https://jiajunhuang.com/articles/2022_03_22-nginx_source_code_2.md.html)
- [Nginx 源码阅读（三）: 连接池、内存池、buf](https://jiajunhuang.com/articles/2022_03_23-nginx_source_code_3.md.html)

## worker 是怎么处理请求的

上一篇我们看到了 worker 其实最后其实就是执行 `ngx_worker_process_cycle` 函数：

```c
// worker 处理循环在这里
static void ngx_worker_process_cycle(ngx_cycle_t *cycle, void *data)
{
    // ...
    for (i = 0; ngx_modules[i]; i++) {
        if (ngx_modules[i]->init_process) {
            // 这里调用每一个模块的 init_process
            if (ngx_modules[i]->init_process(cycle) == NGX_ERROR) {
                /* fatal */
                exit(2);
            }
        }
    }
    // ...
    if (ngx_add_channel_event(cycle, ngx_channel, NGX_READ_EVENT,
                                             ngx_channel_handler) == NGX_ERROR)
    {
        /* fatal */
        exit(2);
    }
    // ...
    for ( ;; ) {
        if (ngx_exiting
            && ngx_event_timer_rbtree == &ngx_event_timer_sentinel)
        {
            // ...
            // worker 真正执行处理事件的地方
            ngx_process_events(cycle);
            // ...
        }
        // ...
    }
    // ...
}
```

可以看到，首先worker会调用每一个模块里的 `init_process` 来执行一些代码，表示目前
worker进程已经开始启动了。接着，worker把和master用于通信的socket加入了 I/O 多路复用
监听，然后就开始了worker的循环，处理事件，并且根据全局变量来决定是否退出等。

跟踪 `ngx_process_events` 发现是一个宏定义：

```c
#define ngx_process_events   ngx_event_actions.process_events
```

对于我的系统来说，肯定就是epoll了，找一下：

```c
ngx_event_module_t  ngx_epoll_module_ctx = {
    &epoll_name,
    ngx_epoll_create_conf,               /* create configuration */
    ngx_epoll_init_conf,                 /* init configuration */

    {
        ngx_epoll_add_event,             /* add an event */
        ngx_epoll_del_event,             /* delete an event */
        ngx_epoll_add_event,             /* enable an event */
        ngx_epoll_del_event,             /* disable an event */
        ngx_epoll_add_connection,        /* add an connection */
        ngx_epoll_del_connection,        /* delete an connection */
        NULL,                            /* process the changes */
        ngx_epoll_process_events,        /* process the events */
        ngx_epoll_init,                  /* init the events */
        ngx_epoll_done,                  /* done the events */
    }
};

// epoll 处理事件
int ngx_epoll_process_events(ngx_cycle_t *cycle)
{
    // ...

    for ( ;; ) { 
        // 找到最小的超时时间
        timer = ngx_event_find_timer();

        if (timer != 0) {
            break;
        }

        ngx_event_expire_timers((ngx_msec_t) (ngx_elapsed_msec - ngx_old_elapsed_msec));

        if (ngx_posted_events && ngx_threaded) {
            ngx_wakeup_worker_thread(cycle);
        }
    }

    // ...

    // 用timer作为最小超时时间，调用epoll_wait开始被挂起
    events = epoll_wait(ep, event_list, nevents, timer);

    if (events == -1) {
        err = ngx_errno;
    } else {
        err = 0;
    }

    // ...

    for (i = 0; i < events; i++) {
        c = event_list[i].data.ptr;

        instance = (uintptr_t) c & 1;
        c = (ngx_connection_t *) ((uintptr_t) c & (uintptr_t) ~1);

        rev = c->read;
        // ...
        wev = c->write;

        // 处理写事件
        if ((event_list[i].events & (EPOLLOUT|EPOLLERR|EPOLLHUP))
            && wev->active)
        {
            if (ngx_threaded) {
                wev->posted_ready = 1;
                ngx_post_event(wev);

            } else {
                wev->ready = 1;

                if (!ngx_accept_mutex_held) {
                    wev->event_handler(wev);

                } else {
                    ngx_post_event(wev);
                }
            }
        }

        /*
         * EPOLLIN must be handled after EPOLLOUT because we use
         * the optimization to avoid the unnecessary mutex locking/unlocking
         * if the accept event is the last one.
         */

         // 处理读事件
        if ((event_list[i].events & (EPOLLIN|EPOLLERR|EPOLLHUP))
            && rev->active)
        {
            if (ngx_threaded && !rev->accept) {
                rev->posted_ready = 1;

                ngx_post_event(rev);

                continue;
            }

            rev->ready = 1;

            if (!ngx_threaded && !ngx_accept_mutex_held) {
                rev->event_handler(rev);

            } else if (!rev->accept) {
                ngx_post_event(rev);

            } else if (ngx_accept_disabled <= 0) {
                rev->event_handler(rev);

                // ...
            }
        }
    }

    // ...

    return NGX_OK;
}
```

这个函数很长，此处已经删减了很多代码，只保留了核心代码。我们可以看到，`ngx_epoll_process_events` 做的事情，
首先就是找到最小的超时时间 `timer = ngx_event_find_timer`，然后把这个值作为 `epoll_wait` 的超时时间，这样，
worker就可以执行一些定时任务，比如多少秒之后要是没有收到请求就超时。

在 `events = epoll_wait(ep, event_list, nevents, timer);` 之后，就拿到了事件，紧接着依次迭代事件进行处理，
分别使用 `wev->event_handler(wev);` 和 `rev->event_handler(rev);` 处理写和读，可是，event_handler 此刻究竟
是什么呢？为什么没有看到Nginx先bind地址，然后listen并且accept呢？如果没有这几步，Nginx就无法和客户端建立
TCP 连接，就无法接受请求并且读取内容了。

## bind/listen/accept 在哪里执行

我分别搜索了一下代码：

- `bind` 和 `listen` 是在 `src/core/ngx_connection.c` 的 `ngx_open_listening_sockets` 里执行的
- `accept` 是在 `src/event/ngx_event_accept.c` 的 `ngx_event_accept` 里执行的

我们先来看 `bind` 和 `listen`：

```c
ngx_int_t ngx_open_listening_sockets(ngx_cycle_t *cycle)
{
    for (tries = /* STUB */ 5; tries; tries--) {
        failed = 0;

        /* for each listening socket */

        ls = cycle->listening.elts;
        for (i = 0; i < cycle->listening.nelts; i++) {
            // ...

            // 创建socket
            s = ngx_socket(ls[i].family, ls[i].type, ls[i].protocol, ls[i].flags);

            if (s == -1) {
                ngx_log_error(NGX_LOG_EMERG, log, ngx_socket_errno,
                              ngx_socket_n " %s failed", ls[i].addr_text.data);
                return NGX_ERROR;
            }

            // ...

            // 设置socket属性
            if (setsockopt(s, SOL_SOCKET, SO_REUSEADDR,
                           (const void *) &reuseaddr, sizeof(int)) == -1) {
                ngx_log_error(NGX_LOG_EMERG, log, ngx_socket_errno,
                              "setsockopt(SO_REUSEADDR) %s failed",
                              ls[i].addr_text.data);
                return NGX_ERROR;
            }

            /* TODO: close on exit */

            // 设置不阻塞
            if (!(ngx_event_flags & NGX_USE_AIO_EVENT)) {
                if (ngx_nonblocking(s) == -1) {
                    ngx_log_error(NGX_LOG_EMERG, log, ngx_socket_errno,
                                  ngx_nonblocking_n " %s failed",
                                  ls[i].addr_text.data);
                    return NGX_ERROR;
                }
            }

            // ...

            // 绑定地址
            if (bind(s, ls[i].sockaddr, ls[i].socklen) == -1) {
                err = ngx_socket_errno;
                ngx_log_error(NGX_LOG_EMERG, log, err,
                              "bind() to %s failed", ls[i].addr_text.data);

                if (err != NGX_EADDRINUSE)
                    return NGX_ERROR;

                if (ngx_close_socket(s) == -1)
                    ngx_log_error(NGX_LOG_EMERG, log, ngx_socket_errno,
                                  ngx_close_socket_n " %s failed",
                                  ls[i].addr_text.data);

                failed = 1;
                continue;
            }

            // 执行 listen() 调用，监听端口
            if (listen(s, ls[i].backlog) == -1) {
                ngx_log_error(NGX_LOG_EMERG, log, ngx_socket_errno,
                              "listen() to %s failed", ls[i].addr_text.data);
                return NGX_ERROR;
            }

            /* TODO: deferred accept */

            ls[i].fd = s;
        }
    }
}
```

这里就执行了常规的TCP服务必须执行的两个函数，`bind` 和 `listen`，有了这两个，才能
在监听的端口上执行 `accept`。可是，是谁再执行 `ngx_open_listening_sockets` 呢？
搜索之后，发现原来是 `init_cycle`，这个函数，在 `master` 初始化的时候就调用了。

```c
ngx_cycle_t *ngx_init_cycle(ngx_cycle_t *old_cycle)
{
    // ...

    // 解析配置文件
    if (ngx_conf_parse(&conf, &cycle->conf_file) != NGX_CONF_OK) {
        ngx_destroy_pool(pool);
        return NULL;
    }

    // ...
    // 初始化配置
    for (i = 0; ngx_modules[i]; i++) {
        if (ngx_modules[i]->type != NGX_CORE_MODULE) {
            continue;
        }

        module = ngx_modules[i]->ctx;

        if (module->init_conf) {
            if (module->init_conf(cycle, cycle->conf_ctx[ngx_modules[i]->index])
                                                              == NGX_CONF_ERROR)
            {
                ngx_destroy_pool(pool);
                return NULL;
            }
        }
    }

    // ...

    if (!ngx_test_config && !failed) {
        // 监听每一个端口。worker fork之后，会继承父进程的监听端口
        if (ngx_open_listening_sockets(cycle) == NGX_ERROR) {
            failed = 1;
        }
    }

    // ...
}
```

`ngx_open_listening_sockets` 我们看到逻辑就是对于 `cycle->listening.nelts` 里的
每一个地址，都创建一个 `socket` 然后 `bind` 并且 `listen`。`cycle->listening.nelts`
就是Nginx要监听的端口，Nginx配置文件里，什么指令会指定监听端口呢？ `listen` 指令，
这就引导我们要去看一看 `ngx_conf_parse` 了。

```c
// 解析配置文件
char *ngx_conf_parse(ngx_conf_t *cf, ngx_str_t *filename)
{
    // ...
    // 调用该directive的handler函数，比如listen就在这里读取出端口，然后保存配置
    rv = cmd->set(cf, cmd, conf);
    // ...
}
```

我们看下 `listen` 指令的定义之处：

```c
static ngx_command_t  ngx_http_core_commands[] = {
    // ...
    { ngx_string("listen"),
      NGX_HTTP_SRV_CONF|NGX_CONF_TAKE1,
      ngx_set_listen,
      NGX_HTTP_SRV_CONF_OFFSET,
      0,
      NULL },
    // ...
}

static char *ngx_set_listen(ngx_conf_t *cf, ngx_command_t *cmd, void *conf)
{
    // ...
    if (!(ls = ngx_array_push(&scf->listen))) {
        return NGX_CONF_ERROR;
    }
    // ...
    port = ngx_atoi(&addr[p], args[1].len - p);
    // ...
    if (port == NGX_ERROR && p == 0) {

        /* "listen host" */
        ls->port = 80;
    else if {
        // ...
    } else if (p == 0) {
        ls->addr = INADDR_ANY;
        ls->port = (in_port_t) port;
        return NGX_CONF_OK;

    } else {
        ls->port = (in_port_t) port;
    }

    ls->addr = inet_addr((const char *) addr);
    if (ls->addr == INADDR_NONE) {
        h = gethostbyname((const char *) addr);

        if (h == NULL || h->h_addr_list[0] == NULL) {
            ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                              "can not resolve host \"%s\" "
                              "in \"%s\" directive", addr, cmd->name.data);
            return NGX_CONF_ERROR;
        }

        ls->addr = *(in_addr_t *)(h->h_addr_list[0]);
    }
}
```

所以现在我们知道了，Nginx 在 `init_cycle` 中解析配置文件时，提取出端口号，并且保存起来，然后执行
`ngx_open_listening_sockets` 打开监听的端口，并且设置好socket相关属性。等等，解析
配置文件，似乎是在 `master` 中就完成了？没错：

```c
int main(int argc, char *const *argv)
{
    // ...
    // 初始化 cycle。里面会初始化各个模块
    cycle = ngx_init_cycle(&init_cycle);
    // ...
    if (ngx_process == NGX_PROCESS_MASTER) {
        ngx_master_process_cycle(cycle, &ctx);

    } else {
        ngx_single_process_cycle(cycle, &ctx);
    }
}
```

非常好，现在我们明白了大概的流程：

- Nginx 从 `main` 开始执行，`main` 中调用 `ngx_init_cycle`
- `ngx_init_cycle` 调用 `ngx_conf_parse` 解析配置，解析出监听的地址
- `ngx_init_cycle` 调用 `ngx_open_listening_sockets` 进行监听
- `ngx_open_listening_sockets` 依次调用 `bind` 和 `listen` 进行监听
- `main` 函数调用 `ngx_master_process_cycle` 启动master
- `ngx_master_process_cycle` 调用 `ngx_start_worker_processes` 启动 `worker`，之后进入master的循环

那么，`accept` 在哪里调用？我们去看一下：

```c
void ngx_event_accept(ngx_event_t *ev)
{
    // ...
    // 调用accept函数，接受连接
    s = accept(ls->fd, sa, &len);
    // ...
    // 初始化c的各种属性，c是connection
    c->pool = pool;
    // ...
    if (ngx_add_conn && (ngx_event_flags & NGX_USE_EPOLL_EVENT) == 0) {
        // 又加到epoll里
        if (ngx_add_conn(c) == NGX_ERROR) {
            ngx_close_accepted_socket(s, log);
            ngx_destroy_pool(pool);
            return;
        }
    }
    // ...
    ls->listening->handler(c);
    // ...
}

static int ngx_epoll_add_connection(ngx_connection_t *c)
{
    struct epoll_event  ee;

    ee.events = EPOLLIN|EPOLLOUT|EPOLLET;
    ee.data.ptr = (void *) ((uintptr_t) c | c->read->instance);

    ngx_log_debug2(NGX_LOG_DEBUG_EVENT, c->log, 0,
                   "epoll add connection: fd:%d ev:%08X", c->fd, ee.events);

    if (epoll_ctl(ep, EPOLL_CTL_ADD, c->fd, &ee) == -1) {
        ngx_log_error(NGX_LOG_ALERT, c->log, ngx_errno,
                      "epoll_ctl(EPOLL_CTL_ADD, %d) failed", c->fd);
        return NGX_ERROR;
    }

    c->read->active = 1;
    c->write->active = 1;

    return NGX_OK;
}
```

可以看到，`ngx_event_accept` 做的事情就是，先 `accept`，然后再次把fd加到epoll里，
`ngx_event_accept` 是在哪里被调用的呢？搜了一下，在 `ngx_event_process_init` 里，
继续往上追调用：

```c
ngx_module_t  ngx_event_core_module = {
    NGX_MODULE,
    &ngx_event_core_module_ctx,            /* module context */
    ngx_event_core_commands,               /* module directives */
    NGX_EVENT_MODULE,                      /* module type */
    ngx_event_module_init,                 /* init module */
    ngx_event_process_init                 /* init process */
};
```

原来是 `ngx_event_process_init` 调用的，这不是在 `ngx_worker_process_cycle` 中调用的：

```c
// worker 处理循环在这里
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
}
```

搜索了一下，还真的只有 event 模块定义了这个函数：

```bash
$ ag 'init process'
src/event/ngx_event.c
116:    NULL                                   /* init process */
188:    ngx_event_process_init                 /* init process */

src/event/modules/ngx_epoll_module.c
129:    NULL                                 /* init process */

src/event/modules/ngx_select_module.c
71:    NULL                                   /* init process */

// ...
```

其余的都是NULL。

到现在，我们对Nginx又更加了解了，Nginx的worker启动之后，首先初始化模块，调用各个
模块的 `init_process`，不过只有event模块定义了这个函数。

`ngx_event_process_init` 中，把 `rev->event_handler` 设置为 `&ngx_event_accept`，
因此，当worker收到连接时，就会因为是可读事件，而执行 `&ngx_event_accept`，在
`&ngx_event_accept` 中，执行 `accept` 之后，设置socket的一些属性，然后分配一个
`ngx_connection_t` 代表这个连接，并且把socket再次加入到epoll中进行监听。

到现在，我们就了解到了TCP层的整个链路：

- Nginx 从 `main` 开始执行，`main` 中调用 `ngx_init_cycle`
- `ngx_init_cycle` 调用 `ngx_conf_parse` 解析配置，解析出监听的地址
- `ngx_init_cycle` 调用 `ngx_open_listening_sockets` 进行监听
- `ngx_open_listening_sockets` 依次调用 `bind` 和 `listen` 进行监听
- `main` 函数调用 `ngx_master_process_cycle` 启动master
- `ngx_master_process_cycle` 调用 `ngx_start_worker_processes` 启动 `worker`，之后进入master的循环
- `ngx_worker_process_cycle` 初始化模块，调用了事件模块的 `ngx_event_process_init` 函数
- `ngx_event_process_init` 设置读事件的处理函数为 `ngx_event_accept`
- 随后，worker开始进入自身的循环，找到最小的超时时间，然后执行 `epoll_wait`
- 当socket可读时，就会从 `epoll_wait` 返回，并且对监听的socket执行 `ngx_event_accept`
- `ngx_event_accept` 执行 `accept`，然后再次把socket加到epoll中监听

可是，HTTP相关的处理函数是怎么被调用的呢？

## HTTP 处理流程

翻了翻资料，最后我发现原来HTTP相关的请求，就在 `ngx_event_accept` 中被调用：

```c
void ngx_event_accept(ngx_event_t *ev)
{
    // ...
    // 调用accept函数，接受连接
    s = accept(ls->fd, sa, &len);
    // ...
    // 初始化c的各种属性，c是connection
    c->pool = pool;
    // ...
    ls->listening->handler(c); // 这里是 ngx_http_init_connection
    // ...
}
```

但是，是在哪里设置的呢？我搜索了一下代码：

```c
// 解析HTTP模块配置
static char *ngx_http_block(ngx_conf_t *cf, ngx_command_t *cmd, void *conf)
{
    // ...
    ls->handler = ngx_http_init_connection; // http handler在这里设置的
    // ...
}
```

整个流程终于可以串起来了，原来早在 `worker` 解析配置文件的时候，就已经设置好了
HTTP回调函数，当accept处理完之后，接下来就轮到 HTTP 回调函数登场。

接下来，我们读一读 `ngx_http_init_connection`。

```c
void ngx_http_init_connection(ngx_connection_t *c)
{
    // ...
    rev->event_handler = ngx_http_init_request; // http 处理请求的函数
    // ...
}
```

原来这里转手就把读事件的处理函数改成了 `ngx_http_init_request`：

```c
static void ngx_http_init_request(ngx_event_t *rev)
{
    // ...
    rev->event_handler = ngx_http_process_request_line; // 处理请求内容的函数
    // ...
}

static void ngx_http_process_request_line(ngx_event_t *rev)
{
    // ...
    for ( ;; ) {
        // ...
        rc = ngx_http_parse_request_line(r, r->header_in);
        // ...
    }
    // ...
        } else if (rc == NGX_HTTP_PARSE_HEADER_DONE) {
            rev->event_handler = ngx_http_block_read;
            ngx_http_handler(r);
    // ...
    }
}
```

从这里开始，Nginx真正开始处理请求了，上述几个函数，开始处理头部了。我们继续看
`ngx_http_handler`：

```c
void ngx_http_handler(ngx_http_request_t *r)
{
    ngx_http_log_ctx_t  *lcx;

    r->connection->unexpected_eof = 0;

    lcx = r->connection->log->data;
    lcx->action = NULL;

    switch (r->headers_in.connection_type) {
    case 0:
        if (r->http_version > NGX_HTTP_VERSION_10) {
            r->keepalive = 1;
        } else {
            r->keepalive = 0;
        }
        break;

    case NGX_HTTP_CONNECTION_CLOSE:
        r->keepalive = 0;
        break;

    case NGX_HTTP_CONNECTION_KEEP_ALIVE:
        r->keepalive = 1;
        break;
    }

    if (r->keepalive && r->headers_in.msie && r->method == NGX_HTTP_POST) {

        /*
         * MSIE may wait for some time if the response for the POST request
         * is sent over the keepalive connection
         */

        r->keepalive = 0;
    }

    if (r->headers_in.content_length_n > 0) {
        r->lingering_close = 1;

    } else {
        r->lingering_close = 0;
    }

    r->connection->write->event_handler = ngx_http_phase_event_handler;

    ngx_http_run_phases(r);

    return;
}

static void ngx_http_run_phases(ngx_http_request_t *r) // 传说中的11个阶段
{
    char                       *path;
    ngx_int_t                   rc;
    ngx_http_handler_pt        *h;
    ngx_http_core_loc_conf_t   *clcf;
    ngx_http_core_main_conf_t  *cmcf;

    cmcf = ngx_http_get_module_main_conf(r, ngx_http_core_module);

    for (/* void */; r->phase < NGX_HTTP_LAST_PHASE; r->phase++) {

        if (r->phase == NGX_HTTP_CONTENT_PHASE && r->content_handler) {
            r->connection->write->event_handler = ngx_http_empty_handler;
            rc = r->content_handler(r);
            ngx_http_finalize_request(r, rc);
            return;
        }

        h = cmcf->phases[r->phase].handlers.elts;
        for (r->phase_handler = cmcf->phases[r->phase].handlers.nelts - 1;
             r->phase_handler >= 0;
             r->phase_handler--)
        {
            rc = h[r->phase_handler](r);

            if (rc == NGX_DONE) {

                /*
                 * we should never use r here because 
                 * it could point to already freed data
                 */

                return;
            }

            if (rc == NGX_DECLINED) {
                continue;
            }

            if (rc >= NGX_HTTP_SPECIAL_RESPONSE || rc == NGX_ERROR) {
                ngx_http_finalize_request(r, rc);
                return;
            }

            if (r->phase == NGX_HTTP_CONTENT_PHASE) {
                ngx_http_finalize_request(r, rc);
                return;
            }

            if (rc == NGX_AGAIN) {
                return;
            }

            if (rc == NGX_OK && cmcf->phases[r->phase].type == NGX_OK) {
                break;
            }
        }
    }


    // ...

    ngx_http_finalize_request(r, NGX_HTTP_NOT_FOUND);
    return;
}
```

到了这里，我们终于抵达了传说中Nginx的11个处理阶段。

- NGX_HTTP_POST_READ_PHASE — First phase. The ngx_http_realip_module registers its handler at this phase to enable
substitution of client addresses before any other module is invoked.
- NGX_HTTP_SERVER_REWRITE_PHASE — Phase where rewrite directives defined in a server block (but outside a
location block) are processed. The ngx_http_rewrite_module installs its handler at this phase.
- NGX_HTTP_FIND_CONFIG_PHASE — Special phase where a location is chosen based on the request URI. Before this phase,
the default location for the relevant virtual server is assigned to the request, and any module requesting a location
configuration receives the configuration for the default server location. This phase assigns a new location to the
request. No additional handlers can be registered at this phase.
- NGX_HTTP_REWRITE_PHASE — Same as NGX_HTTP_SERVER_REWRITE_PHASE, but for rewrite rules defined in the location,
chosen in the previous phase.
- NGX_HTTP_POST_REWRITE_PHASE — Special phase where the request is redirected to a new location if its URI
changed during a rewrite. This is implemented by the request going through the NGX_HTTP_FIND_CONFIG_PHASE again.
No additional handlers can be registered at this phase.
- NGX_HTTP_PREACCESS_PHASE — A common phase for different types of handlers, not associated with access control.
The standard nginx modules ngx_http_limit_conn_module and ngx_http_limit_req_module register their handlers at this phase.
- NGX_HTTP_ACCESS_PHASE — Phase where it is verified that the client is authorized to make the request. Standard
nginx modules such as ngx_http_access_module and ngx_http_auth_basic_module register their handlers at this phase.
By default the client must pass the authorization check of all handlers registered at this phase for the request to
continue to the next phase. The satisfy directive, can be used to permit processing to continue if any of the phase
handlers authorizes the client.
- NGX_HTTP_POST_ACCESS_PHASE — Special phase where the satisfy any directive is processed. If some access phase
handlers denied access and none explicitly allowed it, the request is finalized. No additional handlers can be
registered at this phase.
- NGX_HTTP_PRECONTENT_PHASE — Phase for handlers to be called prior to generating content. Standard modules such
as ngx_http_try_files_module and ngx_http_mirror_module register their handlers at this phase.
- NGX_HTTP_CONTENT_PHASE — Phase where the response is normally generated. Multiple nginx standard modules
register their handlers at this phase, including ngx_http_index_module or ngx_http_static_module. They are called
sequentially until one of them produces the output. It's also possible to set content handlers on a per-location
basis. If the ngx_http_core_module's location configuration has handler set, it is called as the content handler
and the handlers installed at this phase are ignored.
- NGX_HTTP_LOG_PHASE — Phase where request logging is performed. Currently, only the ngx_http_log_module registers
its handler at this stage for access logging. Log phase handlers are called at the very end of request processing,
right before freeing the request.

这里我就不翻译了。之所以划分这么多个阶段，就是因为Web其实场景很复杂，而Nginx又要实现良好的扩展性。所以把一个请求
划分为多个阶段，每一个阶段都可以注册handler去处理，有点像web框架中的中间件的意思。

## 总结

这一篇文章中，我们从worker看起，然后分别去探索到底在哪里进行 `bind` 和 `listen` 的。随后跟随代码，我们发现worker
初始化模块的时候，就已经预留好了 `accept` 和 HTTP的处理函数。最后，我们终于理顺了整个流程，了解了master和worker是
怎么分工的，然后worker又是怎么一步一步设置，并且在运行过程中更换事件处理handler，最后终于可以开始接收请求的。
