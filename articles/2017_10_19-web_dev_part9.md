# Web开发系列(九)：消息队列，异步任务

有这样一个需求，第三方请求向我们的用户发送一个推送消息。我们必须尽快响应第三方：你的请求我们收到了，但是第三方又想知道
结果。一种办法是等，第三方等我们的系统处理好了，然后返回结果给他。这样做有个优点，代码逻辑简单，但是缺点似乎更大，因为
用户要等待结果，所以这个TCP连接是不会断掉的，也就意味着，如果请求的并发量比较大，那么对我们的系统负载是非常高的，因为要
维护很多个TCP连接。此外对第三方的系统来说也是如此，假设这个请求是从移动端发来的，那影响则更甚。

所以我们需要另外一种方法，异步任务。

Python中，异步任务的首选似乎是Celery，不过我在生产环境中遇到过Celery的问题是无故假死，一直卡在futex锁上。后来切换到rq就
没有再出现过这个问题了，但是Python-rq的问题是使用的并发模型是来一个任务fork一次，对系统性能消耗特别大，所以我改了一下Worker，
加入了Gevent：

```python
class GeventWorker(Worker):
    def execute_job(self, job, queue):
        self.set_state(WorkerStatus.BUSY)
        self.log.debug("gonna spawn a greenlet to execute the given job.")
        gevent.spawn(self.perform_job, job, queue).join()
        self.log.debug("greenlet executed.")
        self.set_state(WorkerStatus.IDLE)


def gevent_worker(queues):
    client = Client(config.SENTRY_DSN, transport=HTTPTransport)
    worker = GeventWorker(
        queues=queues,
        connection=StrictRedis.from_url(config.WORKER_BROKER)
    )
    register_sentry(client, worker)
    worker.work()
```

然后再pre-fork多个进程，每个进程中使用协程处理任务，本来还可以改成并发处理多个任务的，但是因为没有这么高的并发要求，
所以就没有进一步改的更复杂。

而 `Python-rq` 的原理我也在 [这一篇](https://jiajunhuang.com/articles/2017_09_20-task_queue_python_rq.md.html) 文章中
有说过，即生产者进行enqueue，而消费者监听对应的queue，一有任务到来便开始进行消费。中间的queue，也叫broker，是用来存储
在生产者和消费者传递的消息用的，Celery可以选rabbit-mq, redis等好几种作为broker，而rq则只支持redis一种。

> 此外，从严格意义上来说，Redis并不能算上是正统的消息队列
