# Python RQ(Redis Queue)添加gevent支持

python-rq简单好用，但缺点是，默认的实现是使用fork的模式，关于这点可以看
[python-rq源码分析](https://jiajunhuang.com/articles/2017_09_20-task_queue_python_rq.md.html) 。

所以我们要对他进行改造，每次执行任务，我们就使用一个coroutine。gevent的文档中这样写道：

> Patching should be done as early as possible in the lifecycle of the program.

因此，我们在最上面就开始进行 monkey patch。此外，我把queue定义在了 `jobs/queue.py` 里。直接上代码：

```python
# worker.py
import gevent.monkey
gevent.monkey.patch_all()  # noqa

import logging
from rq.worker import (
    Worker,
    WorkerStatus,
)
import redis

from config import config
from jobs import (
    money_q,
    message_q,
)


class GeventWorker(Worker):
    def execute_job(self, job, queue):
        self.set_state(WorkerStatus.BUSY)
        self.log.debug("gonna spawn a greenlet to execute job %s from queue", job, queue)
        gevent.spawn(self.perform_job, job, queue).join()
        self.log.debug("job %s from queue %s executed", job, queue)
        self.set_state(WorkerStatus.IDLE)


def gevent_worker(queues):
    worker = GeventWorker(
        queues=queues,
        connection=redis.StrictRedis.from_url(config.WORKER_BROKER)
    )
    worker.work()


if __name__ == "__main__":
    gevent_worker([money_q, message_q])
```

```python
# queue.py
from rq import Queue
import redis

from config import config

__conn = redis.StrictRedis.from_url(config.WORKER_BROKER)

money_q = Queue("money", connection=__conn)
message_q = Queue("message", connection=__conn)
```

解释一下实现原理：

首先阅读 rq 默认的worker实现，就会发现，所有的worker都有 `execute_job` 这个方法，因此我们继承 `Worker` 并且
重写这个方法，在我们的实现里，新起一个coroutine来执行相关代码。

---

- http://www.gevent.org/api/gevent.monkey.html
