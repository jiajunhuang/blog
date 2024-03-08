# PostgreSQL 当MQ来使用

一般我们都是用 Redis/RabbitMQ等等 来做MQ，那么，能不能使用关系型数据库来做这件事情呢？显然，可以。

## 设计存储表

首先我们要创建一张表，来存储消息：

```sql
create table pg_queue (
    id serial primary key,
    queue_name text not null,
    payload jsonb not null,
    created_at timestamp with time zone default now(),
    updated_at timestamp with time zone default now(),
    executed_at timestamp with time zone default now(),
    status text default 'pending'
);

-- add index
create index pg_queue_queue_name_status_idx on pg_queue (queue_name, status, executed_at);
```

请注意，上面的字段中，`executed_at` 是执行时间，这样我们可以实现定时执行，这在常见的MQ里可是一个相对高级的属性了，
`status` 为该任务的状态，我设计的MQ里，有 `pending`, `processing`, `done`, `failed` 等几种状态，如果想要实现ACK，
那完全可以再加一种 `ack` 的状态进去。`queue_name` 是队列名，有这一个字段就可以实现多个队列，`payload` 是消息内容。

然后创建一个联合索引，这可不就妥妥的一个高级队列出来了么。

### 生产任务

```sql
-- add task to queue
INSERT INTO pg_queue (queue_name, payload) VALUES ('test', '1');
INSERT INTO pg_queue (queue_name, payload) VALUES ('test', '2');
INSERT INTO pg_queue (queue_name, payload) VALUES ('test', '3');
INSERT INTO pg_queue (queue_name, payload) VALUES ('test', '4');
INSERT INTO pg_queue (queue_name, payload) VALUES ('test', '5');
INSERT INTO pg_queue (queue_name, payload) VALUES ('test', '6');
INSERT INTO pg_queue (queue_name, payload) VALUES ('test', '7');
INSERT INTO pg_queue (queue_name, payload) VALUES ('test', '8');
INSERT INTO pg_queue (queue_name, payload) VALUES ('test', '9');
INSERT INTO pg_queue (queue_name, payload) VALUES ('test', '10');
```

上述SQL会插入任务，默认的执行时间就是立即执行，如果需要做延时的话，就可以指定 `executed_at` 字段。

## 取任务消费

```sql
-- get task from queue
UPDATE pg_queue SET status = 'processing', updated_at = NOW() WHERE id = (
    SELECT id FROM pg_queue WHERE queue_name = 'test' AND status = 'pending' AND executed_at <= NOW() ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED
) RETURNING id, queue_name, payload, created_at, updated_at, executed_at, status;
```

这里的SQL，意思是取 `test` 队列中，状态为 `pending` 且到了执行时间的任务，一次取一个（可以调整为多个），跳过已经被锁定的任务。

## 更新任务状态

```sql
-- mark task as done
UPDATE pg_queue SET status = 'done', updated_at = NOW() WHERE id = 1;
```

```sql
-- mark task as failed
UPDATE pg_queue SET status = 'failed', updated_at = NOW() WHERE id = 1;
```

## 任务重试

```sql
-- auto retry task if status is failed
UPDATE pg_queue SET status = 'pending', updated_at = NOW(), executed_at = NOW() WHERE queue_name = 'test' AND status = 'failed' AND executed_at <= NOW();
```

## 定期清理过期任务

```sql
-- auto truncate pg_queue
DELETE FROM pg_queue WHERE created_at < NOW() - INTERVAL '1 day';
```

## 总结

不愧是PG，通过这么几个简单的SQL，就可以实现一个高级队列，支持：

- 持久化、事务支持
- 多端写入
- 多端消费
- 延时任务、定时任务
- 历史消息记录
- ACK(如果想要的话完全可以用 `status` 字段实现)
- 多队列(框架内可以实现优先队列)支持，延时队列，死信队列(加一个状态即可)
- 批量发送和接收
- 重试、重入队列

其余的高级特性，都可以通过一个辅助性框架来实现。
