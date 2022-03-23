# Nginx 源码阅读（三）: 连接池、内存池

这一篇文章中，我们主要来看看Nginx怎么实现连接池和内存池。通过这两个池，Nginx就可以避免频繁申请内存，从而提高性能。

- [Nginx 源码阅读（一）: 启动流程](https://jiajunhuang.com/articles/2022_03_21-nginx_source_code_1.md.html)
- [Nginx 源码阅读（二）: 请求处理](https://jiajunhuang.com/articles/2022_03_22-nginx_source_code_2.md.html)
- [Nginx 源码阅读（三）: 连接池、内存池、buf](https://jiajunhuang.com/articles/2022_03_23-nginx_source_code_3.md.html)

## 连接池

Nginx 配置中，有这么一个选项：`worker_connections`，默认值512。Sets the maximum number of simultaneous
connections that can be opened by a worker process.

It should be kept in mind that this number includes all connections (e.g. connections with proxied servers,
among others), not only connections with clients. Another consideration is that the actual number of
simultaneous connections cannot exceed the current limit on the maximum number of open files, which can be
changed by worker_rlimit_nofile.

这个选项可以控制一个worker最多可以打开多少个连接，那么它存储在哪里呢？这就简单了，我们来看看在哪里配置的便知。
在 Nginx v0.1.0 中，这个配置叫做 "connections"，所以我们要搜索：

```bash
$ ag -Q '"connections"'
src/event/ngx_event.c
127:    { ngx_string("connections"),
```

来看代码：

```c
static ngx_command_t  ngx_event_core_commands[] = {
    // ...
    { ngx_string("connections"),
      NGX_EVENT_CONF|NGX_CONF_TAKE1,
      ngx_event_connections,
      0,
      0,
      NULL },
    // ...
}

static char *ngx_event_connections(ngx_conf_t *cf, ngx_command_t *cmd,
                                   void *conf)
{
    ngx_event_conf_t  *ecf = conf;

    ngx_str_t  *value;

    if (ecf->connections != NGX_CONF_UNSET_UINT) {
        return "is duplicate" ;
    }

    value = cf->args->elts;
    ecf->connections = ngx_atoi(value[1].data, value[1].len);
    if (ecf->connections == (ngx_uint_t) NGX_ERROR) {
        ngx_conf_log_error(NGX_LOG_EMERG, cf, 0,
                           "invalid number \"%s\"", value[1].data);

        return NGX_CONF_ERROR;
    }

    cf->cycle->connection_n = ecf->connections;

    return NGX_CONF_OK;
}
```

可以看到，这里就是解析出后面的数值，然后赋值给 `cf->cycle->connection_n`：

```c
struct ngx_cycle_s {
    void           ****conf_ctx;
    ngx_pool_t        *pool;

    ngx_log_t         *log;
    ngx_log_t         *new_log;

    ngx_array_t        listening;
    ngx_array_t        pathes;
    ngx_list_t         open_files;

    ngx_uint_t         connection_n; // 此处
    ngx_connection_t  *connections;
    ngx_event_t       *read_events;
    ngx_event_t       *write_events;

    ngx_cycle_t       *old_cycle;

    ngx_str_t          conf_file;
    ngx_str_t          root;
};
```

这可不就是我们的老朋友，`cycle`嘛。我们搜索一下 `connection_n` 看看是怎么使用的，在 `ngx_event_process_init` 里找到了：

```c
static ngx_int_t ngx_event_process_init(ngx_cycle_t *cycle)
{
    // ...
    ecf = ngx_event_get_conf(cycle->conf_ctx, ngx_event_core_module);
    // ...
    cycle->connection_n = ecf->connections;
    // ...
    cycle->connections = ngx_alloc(sizeof(ngx_connection_t) * ecf->connections, cycle->log);
    // ...
    c = cycle->connections;
    for (i = 0; i < cycle->connection_n; i++) {
        c[i].fd = (ngx_socket_t) -1;
        c[i].data = NULL;
    }
    // ...
    cycle->read_events = ngx_alloc(sizeof(ngx_event_t) * ecf->connections, cycle->log);
    if (cycle->read_events == NULL) {
        return NGX_ERROR;
    }

    rev = cycle->read_events;
    for (i = 0; i < cycle->connection_n; i++) {
        rev[i].closed = 1;
    }
    cycle->write_events = ngx_alloc(sizeof(ngx_event_t) * ecf->connections, cycle->log);
    if (cycle->write_events == NULL) {
        return NGX_ERROR;
    }

    wev = cycle->write_events;
    for (i = 0; i < cycle->connection_n; i++) {
        wev[i].closed = 1;
    }

    // ...
}

// 事件的各种标记
struct ngx_event_s {
    void            *data;

    unsigned         write:1;

    unsigned         accept:1;

    unsigned         oneshot:1;

    /* used to detect the stale events in kqueue, rt signals and epoll */
    unsigned         instance:1;

    /*
     * the event was passed or would be passed to a kernel;
     * in aio mode - operation was posted.
     */
    unsigned         active:1;

    unsigned         disabled:1;

    /* the ready event; in aio mode 0 means that no operation can be posted */
    unsigned         ready:1;

    /* aio operation is complete */
    unsigned         complete:1;

    unsigned         eof:1;
    unsigned         error:1;

    unsigned         timedout:1;
    unsigned         timer_set:1;

    unsigned         delayed:1;

    unsigned         read_discarded:1;

    unsigned         unexpected_eof:1;

    unsigned         deferred_accept:1;

    /* the pending eof reported by kqueue or in aio chain operation */
    unsigned         pending_eof:1;


    u_int            index;

    ngx_log_t       *log;

    /* TODO: threads: padding to cache line */

    /*
     * STUB: The inline of "ngx_rbtree_t  rbtree;"
     */

    ngx_int_t        rbtree_key;
    void            *rbtree_left;
    void            *rbtree_right;
    void            *rbtree_parent;
    char             rbtree_color;


    unsigned         closed:1;
};

struct ngx_connection_s {
    void               *data;
    ngx_event_t        *read;
    ngx_event_t        *write;

    ngx_socket_t        fd;

    ngx_recv_pt         recv;
    ngx_send_chain_pt   send_chain;

    ngx_listening_t    *listening;

    off_t               sent;

    void               *ctx;
    void               *servers;


    ngx_log_t          *log;

    ngx_pool_t         *pool;

    struct sockaddr    *sockaddr;
    socklen_t           socklen;
    ngx_str_t           addr_text;

    ngx_buf_t          *buffer;

    ngx_uint_t          number;

    unsigned            log_error:2;  /* ngx_connection_log_error_e */

    unsigned            buffered:1;
    unsigned            single_connection:1;
    unsigned            unexpected_eof:1;
    unsigned            timedout:1;
    signed              tcp_nopush:2;
};
```

可以看到，Nginx的worker初始化的时候，会调用各个模块的 `init_process` 函数，而事件模块的这个函数 `ngx_event_process_init`
被调用时，就会申请好 `connection_n` 大小的内存，用来放连接，同时还会申请同样大小的 `read_events` 和 `write_events` 两个
数组，用来存放对应fd上的读和写的事件，这些事件都是 `ngx_event_t` 类型，里面存放了事件的各种信息和状态，以及当前处理函数。 执行完这些，我们就已经创建好了对应大小的撒个数组，分别存放连接、读事件、写事件，这就是Nginx的连接池。

Nginx通过这三个数组缓存住一个连接的信息，把文件描述符(上文的fd)作为数组的下标，我们来看看Nginx把监听端口的fd放进去的逻辑：

```c
static ngx_int_t ngx_event_process_init(ngx_cycle_t *cycle)
{
    /* for each listening socket */
    // 每一个监听的socket，在这里加入epoll
    s = cycle->listening.elts;
    for (i = 0; i < cycle->listening.nelts; i++) {
        fd = s[i].fd;
        c = &cycle->connections[fd];
        rev = &cycle->read_events[fd];
        wev = &cycle->write_events[fd];

        ngx_memzero(c, sizeof(ngx_connection_t));
        ngx_memzero(rev, sizeof(ngx_event_t));
        // ... 各种赋值和设置状态标记
    }
}
```

逻辑就是以fd为下标，取出数组里对应的 `ngx_connection_t` 结构体和 `ngx_event_t` 结构体，然后进行赋值，之后在要用的时候，
就可以直接用了。那么我们怎么确保不会更改到错误的数组元素呢？如何确保不会超出数组的范围呢？对于第一个问题，fd由操作系统
保证唯一，所以只要你从操作系统那里拿到fd，就不会改到别人的元素。对于第二个问题，`listen` 出来的fd不会有这个问题，因为
操作系统分配fd都是以进程为限制范围，从可用的最小的整数开始分配，但是 `accept` 可能会有这个问题，比如当连接数特别多的时候，
我们来看下Nginx如何处理 `accept`：

```c
void ngx_event_accept(ngx_event_t *ev)
{
    // 调用accept函数，接受连接
    s = accept(ls->fd, sa, &len);
    if (s == -1) {
        // 处理异常
    }

    // 判断是否超出最大可用fd，是的话，关闭连接
    if ((ngx_uint_t) s >= ecf->connections) {

        ngx_log_error(NGX_LOG_ALERT, ev->log, 0,
                        "accept() on %s returned socket #%d while "
                        "only %d connections was configured, "
                        "closing the connection",
                        ls->listening->addr_text.data, s, ecf->connections);

        ngx_close_accepted_socket(s, log);
        ngx_destroy_pool(pool);
        return;
    }

    // 开始使用
    c = &ngx_cycle->connections[s];
    rev = &ngx_cycle->read_events[s];
    wev = &ngx_cycle->write_events[s];
}
```

你看，逻辑很简单，加一个判断便是。好了，Nginx连接池我们就看到这里。接下来我们来看看内存池。

## 内存池

在读Nginx源码的时候，我们看到了很多调用 `ngx_create_pool`，`ngx_destroy_pool`。这里就是对内存池的操作，包括
`ngx_connection_t` 里，也有对内存池的引用，因为Nginx有一个新的连接来时，就会为连接分配一个 `pool`，当关闭连接
时，就销毁。这样避免了到处都是 `malloc` 和 `free`。

我们先来看看 `pool` 的定义和创建销毁：

```c
struct ngx_pool_s {
    char              *last; // 最后分配出去的地址
    char              *end; // 本次内存块结尾的地址
    ngx_pool_t        *next; // 下一个块的地址
    ngx_pool_large_t  *large; // 大内存块的引用
    ngx_log_t         *log; // 日志
};

ngx_pool_t *ngx_create_pool(size_t size, ngx_log_t *log)
{
    ngx_pool_t  *p;

    if (!(p = ngx_alloc(size, log))) { // 申请大小为 size 的内存，ngx_alloc 底层就是 malloc
       return NULL;
    }

    p->last = (char *) p + sizeof(ngx_pool_t); // 申请内存的最前面就是 ngx_pool_t 的各种属性，所以要从 p + sizeof(ngx_pool_t) 开始使用
    p->end = (char *) p + size; // 总大小为 size
    p->next = NULL; // 初始化
    p->large = NULL; // 初始化
    p->log = log; // 设置log

    return p;
}


void ngx_destroy_pool(ngx_pool_t *pool)
{
    ngx_pool_t        *p, *n;
    ngx_pool_large_t  *l;

    // 从大块开始销毁，一个一个free
    for (l = pool->large; l; l = l->next) {

        ngx_log_debug1(NGX_LOG_DEBUG_ALLOC, pool->log, 0,
                       "free: " PTR_FMT, l->alloc);

        if (l->alloc) {
            free(l->alloc);
        }
    }

    // 然后销毁其余块
    for (p = pool, n = pool->next; /* void */; p = n, n = n->next) {
        free(p);

        if (n == NULL) {
            break;
        }
    }
}
```

那么，如何使用呢？我们继续看代码：

```c
void *ngx_palloc(ngx_pool_t *pool, size_t size)
{
    char              *m;
    ngx_pool_t        *p, *n;
    ngx_pool_large_t  *large, *last;

    // 如果大小符合小块，并且该还有足够的内存
    if (size <= (size_t) NGX_MAX_ALLOC_FROM_POOL
        && size <= (size_t) (pool->end - (char *) pool) - sizeof(ngx_pool_t))
    {
        // 迭代
        for (p = pool, n = pool->next; /* void */; p = n, n = n->next) {
            m = ngx_align(p->last);

            // 如果找到了可以存放所要求大小的内存块，增加 p->last 然后返回地址
            if ((size_t) (p->end - m) >= size) {
                p->last = m + size ;

                return m;
            }

            if (n == NULL) {
                break;
            }
        }

        /* allocate a new pool block */

        // 没有的话，就只能申请一个新的同样大小的块追加在当前块的后面了
        if (!(n = ngx_create_pool((size_t) (p->end - (char *) p), p->log))) {
            return NULL;
        }

        p->next = n;
        m = n->last;
        n->last += size;

        return m;
    }

    /* allocate a large block */
    /*
    否则的话，就申请一个大块了，看了一下 NGX_MAX_ALLOC_FROM_POOL 的值是 
    * NGX_MAX_ALLOC_FROM_POOL should be (ngx_page_size - 1), i.e. 4095 on x86.
    * On FreeBSD 5.x it allows to use the zero copy sending.
    * On Windows NT it decreases a number of locked pages in a kernel.
    * #define NGX_MAX_ALLOC_FROM_POOL  (ngx_pagesize - 1)
    */

    large = NULL;
    last = NULL;

    if (pool->large) {
        for (last = pool->large; /* void */ ; last = last->next) {
            if (last->alloc == NULL) {
                large = last;
                last = NULL;
                break;
            }

            if (last->next == NULL) {
                break;
            }
        }
    }

    if (large == NULL) {
        if (!(large = ngx_palloc(pool, sizeof(ngx_pool_large_t)))) {
            return NULL;
        }

        large->next = NULL;
    }

#if 0
    if (!(p = ngx_memalign(ngx_pagesize, size, pool->log))) {
        return NULL;
    }
#else
    if (!(p = ngx_alloc(size, pool->log))) {
        return NULL;
    }
#endif

    if (pool->large == NULL) {
        pool->large = large;

    } else if (last) {
        last->next = large;
    }

    large->alloc = p;

    return p;
}


ngx_int_t ngx_pfree(ngx_pool_t *pool, void *p)
{
    ngx_pool_large_t  *l;

    // 看来释放的话，只能释放大块
    for (l = pool->large; l; l = l->next) {
        if (p == l->alloc) {
            ngx_log_debug1(NGX_LOG_DEBUG_ALLOC, pool->log, 0,
                           "free: " PTR_FMT, l->alloc);
            free(l->alloc);
            l->alloc = NULL;

            return NGX_OK;
        }
    }

    return NGX_DECLINED;
}
```

读完这些，我们就知道Nginx的内存池是怎么实现的了。

## 总结

这一篇文章中，我们看了Nginx是如何实现连接池和内存池的，通过使用这两个池，Nginx可以复用很多内存，从而避免了频繁申请
内存，尤其是在处理高并发情况的时候，频繁申请销毁内存更是会让系统性能急剧下降，更别提到处申请内存，只要有一个地方忘记
释放，就会造成严重的内存泄漏。相信阅读这两个池的源码，会给我们在未来类似的设计带来帮助。
