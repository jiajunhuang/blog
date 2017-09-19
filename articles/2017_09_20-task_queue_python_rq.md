# 任务队列怎么写？python rq源码分析

rq的意思是 Redis Queue。这个项目和redis是强结合的，此外还有一个重度依赖是pickle。
这是这个项目很简单的原因之一。

## 拷贝源码

首先我们需要fork一份源代码例如：https://github.com/jiajunhuang/rq ，然后拷贝
到本地。进入源码文件夹 `$ cd rq/rq`，我们可以看到目录结构：

```bash
$ tree
.
├── cli
│   ├── cli.py
│   ├── helpers.py
│   └── __init__.py
├── compat
│   ├── connections.py
│   ├── dictconfig.py
│   └── __init__.py
├── connections.py
├── contrib
│   ├── __init__.py
│   ├── legacy.py
│   └── sentry.py
├── decorators.py
├── defaults.py
├── dummy.py
├── exceptions.py
├── handlers.py
├── __init__.py
├── job.py
├── local.py
├── logutils.py
├── queue.py
├── registry.py
├── suspension.py
├── timeouts.py
├── utils.py
├── version.py
└── worker.py

3 directories, 26 files
```

## 入口

那么我们应该怎么读源码呢？入口点是什么？编写代码的时候我们的入口点是main函数，
那么读源代码的时候入口点应该是什么呢？我们是怎么启动rq的worker呢？

```bash
$ rq worker
```

所以我们看看rq这个命令是怎么来的：

```bash
cat `which rq`
#!/home/jiajun/.py3k/bin/python3

# -*- coding: utf-8 -*-
import re
import sys

from rq.cli import main

if __name__ == '__main__':
    sys.argv[0] = re.sub(r'(-script\.pyw?|\.exe)?$', '', sys.argv[0])
    sys.exit(main())
```

说明入口点在 `rq.cli` 的main函数里。接下来我们看看 `rq.cli` 从何而来。

```bash
cat cli/__init__.py 
# flake8: noqa
from .cli import main

# TODO: the following imports can be removed when we drop the `rqinfo` and
# `rqworkers` commands in favor of just shipping the `rq` command.
from .cli import info, worker
```

接下来我们看看 `cli/cli.py` 这个文件，里面可以看到 `def worker`，这就是我们要找
的入口点。可以看到真正开始工作的地方是 `worker.work(burst=burst)`

## 默认值

```python
DEFAULT_JOB_CLASS = 'rq.job.Job'
DEFAULT_QUEUE_CLASS = 'rq.Queue'
DEFAULT_WORKER_CLASS = 'rq.Worker'
DEFAULT_CONNECTION_CLASS = 'redis.StrictRedis'
DEFAULT_WORKER_TTL = 420
DEFAULT_RESULT_TTL = 500
```

一路追查worker初始化的地方的来源：

```python
queues = [cli_config.queue_class(queue,
                                    connection=cli_config.connection,
                                    job_class=cli_config.job_class)
            for queue in queues]
worker = cli_config.worker_class(queues,
                                    name=name,
                                    connection=cli_config.connection,
                                    default_worker_ttl=worker_ttl,
                                    default_result_ttl=results_ttl,
                                    job_class=cli_config.job_class,
                                    queue_class=cli_config.queue_class,
                                    exception_handlers=exception_handlers or None)

worker.work(burst=burst)
```

就可以追查到上述默认值，这些值我们之后还会看到。

## 探究 `work`

打开 `worker.py`，找到 `def work`：

```python
def work(self, burst=False, logging_level="INFO"):
    """Starts the work loop.

    Pops and performs all jobs on the current list of queues.  When all
    queues are empty, block and wait for new jobs to arrive on any of the
    queues, unless `burst` mode is enabled.

    The return value indicates whether any jobs were processed.
    """
    setup_loghandlers(logging_level)
    self._install_signal_handlers()

    did_perform_work = False
    self.register_birth()
    self.log.info("RQ worker {0!r} started, version {1}".format(self.key, VERSION))
    self.set_state(WorkerStatus.STARTED)

    try:
        while True:
            try:
                self.check_for_suspension(burst)

                if self.should_run_maintenance_tasks:
                    self.clean_registries()

                if self._stop_requested:
                    self.log.info('Stopping on request')
                    break

                timeout = None if burst else max(1, self.default_worker_ttl - 60)

                result = self.dequeue_job_and_maintain_ttl(timeout)
                if result is None:
                    if burst:
                        self.log.info("RQ worker {0!r} done, quitting".format(self.key))
                    break

                job, queue = result
                self.execute_job(job, queue)
                self.heartbeat()

                did_perform_work = True

            except StopRequested:
                break
    finally:
        if not self.is_horse:
            self.register_death()
    return did_perform_work
```

可以看到这一段代码做的事情：

- 配置好日志
- 安装好信号处理器
- 注册worker
- 把状态设置成开始工作
- 然后开始进入循环
    - 检查当前worker是否被暂停了
    - 弹出一个job来
    - 开始执行job
    - 执行完成之后发送心跳

job?这应该就是我们的任务了，那么，它是从何而来呢？我们的worker是怎么知道哪个任务从何而来呢？

## job从何而来

我们可以看到，job是从 `job, queue = result` 来的，而 result 是从
`result = self.dequeue_job_and_maintain_ttl(timeout)`来的。我们看看后者：

```python
def dequeue_job_and_maintain_ttl(self, timeout):
    result = None
    qnames = self.queue_names()

    self.set_state(WorkerStatus.IDLE)
    self.procline('Listening on {0}'.format(','.join(qnames)))
    self.log.info('')
    self.log.info('*** Listening on {0}...'.format(green(', '.join(qnames))))

    while True:
        self.heartbeat()

        try:
            result = self.queue_class.dequeue_any(self.queues, timeout,
                                                    connection=self.connection,
                                                    job_class=self.job_class)
            if result is not None:
                job, queue = result
                self.log.info('{0}: {1} ({2})'.format(green(queue.name),
                                                        blue(job.description), job.id))

            break
        except DequeueTimeout:
            pass

    self.heartbeat()
    return result
```

继续追查 `self.queue_class.dequeue_any` 就是 `queue.py` 里的 `Queue` 的函数：

```python
@classmethod
def dequeue_any(cls, queues, timeout, connection=None, job_class=None):
    """Class method returning the job_class instance at the front of the given
    set of Queues, where the order of the queues is important.

    When all of the Queues are empty, depending on the `timeout` argument,
    either blocks execution of this function for the duration of the
    timeout or until new messages arrive on any of the queues, or returns
    None.

    See the documentation of cls.lpop for the interpretation of timeout.
    """
    job_class = backend_class(cls, 'job_class', override=job_class)

    while True:
        queue_keys = [q.key for q in queues]
        result = cls.lpop(queue_keys, timeout, connection=connection)
        if result is None:
            return None
        queue_key, job_id = map(as_text, result)
        queue = cls.from_queue_key(queue_key,
                                    connection=connection,
                                    job_class=job_class)
        try:
            job = job_class.fetch(job_id, connection=connection)
        except NoSuchJobError:
            # Silently pass on jobs that don't exist (anymore),
            # and continue in the look
            continue
        except UnpickleError as e:
            # Attach queue information on the exception for improved error
            # reporting
            e.job_id = job_id
            e.queue = queue
            raise e
        return job, queue
    return None, None
```

看到了 `result = cls.lpop`，继续追查下去：

```python
@classmethod
def lpop(cls, queue_keys, timeout, connection=None):
    """Helper method.  Intermediate method to abstract away from some
    Redis API details, where LPOP accepts only a single key, whereas BLPOP
    accepts multiple.  So if we want the non-blocking LPOP, we need to
    iterate over all queues, do individual LPOPs, and return the result.

    Until Redis receives a specific method for this, we'll have to wrap it
    this way.

    The timeout parameter is interpreted as follows:
        None - non-blocking (return immediately)
            > 0 - maximum number of seconds to block
    """
    connection = resolve_connection(connection)
    if timeout is not None:  # blocking variant
        if timeout == 0:
            raise ValueError('RQ does not support indefinite timeouts. Please pick a timeout value > 0')
        result = connection.blpop(queue_keys, timeout)
        if result is None:
            raise DequeueTimeout(timeout, queue_keys)
        queue_key, job_id = result
        return queue_key, job_id
    else:  # non-blocking variant
        for queue_key in queue_keys:
            blob = connection.lpop(queue_key)
            if blob is not None:
                return queue_key, blob
        return None
```

原来就是从给定的queue里 `lpop` 出来，然后，查一下 blpop 的返回值，是返回的值所在
的list名和值。

> Once new data is present on one of the lists, the client returns with the name of the key unblocking it and the popped value.

然后我们跳回到上一个函数。发现接下来的步骤是根据所得的job_id和queue_key实例化
Queue和Job。

那么我们看看其中调用的 `Job.fetch`：

```python
@classmethod
def fetch(cls, id, connection=None):
    """Fetches a persisted job from its corresponding Redis key and
    instantiates it.
    """
    job = cls(id, connection=connection)
    job.refresh()
    return job
```

`job.refresh()` 很可疑，因为到这一步之前，我们的 job的信息都还只是字符串。
在worker端worker是怎么知道要去调用哪里的函数呢？

我仔细看了看，差点就放过了 `self.data = obj['data']` 这一步，跟进去一看，结果发现不是，
其他地方也没有看到。

尴尬。

那就很奇怪了哈，肯定有个地方从字符串转回python对象的吧。于是我去翻了翻 [文档](http://python-rq.org/docs/)，
发现文档上写了，它是用pickle的，那肯定有地方用了 `pickle.loads`，于是就搜到了
`loads = pickle.loads`。继续搜看哪里用到了loads。

```bash
$ ack loads
job.py
25:loads = pickle.loads
46:    This is a helper method to not have to deal with the fact that `loads()`
51:        obj = loads(pickled_string)
394:                self._result = loads(rv)
```

分别看了一下51行和394行，感觉51行更像，于是就搜 `unpickle`，于是找到了 Job 里的

```python
def _unpickle_data(self):
    self._func_name, self._instance, self._args, self._kwargs = unpickle(self.data)
```

继续搜 `_unpickle_data`，我们发现有四个地方用了它：

- `func_name`
- `instance`
- `args`
- `kwargs`

原来在引用这些 property 的时候，如果还没有反序列化，就会先反序列化一下，算了，我们
先放下这个，看看接下来是如何执行job的好了。

## Job是如何执行的

继续看看 `Worker.work` 的代码，拿到job之后，就开始执行 `self.execute_job(job, queue)`，
跟进去看，看到了 `self.fork_work_horse(job, queue)`，继续跟进去看，看到了：

```python
def fork_work_horse(self, job, queue):
    """Spawns a work horse to perform the actual work and passes it a job.
    """
    child_pid = os.fork()
    os.environ['RQ_WORKER_ID'] = self.name
    os.environ['RQ_JOB_ID'] = job.id
    if child_pid == 0:
        self.main_work_horse(job, queue)
    else:
        self._horse_pid = child_pid
        self.procline('Forked {0} at {1}'.format(child_pid, time.time()))
```

fork之后返回0的是子进程，我们继续看 `self.main_work_horse`：

```python
def main_work_horse(self, job, queue):
    """This is the entry point of the newly spawned work horse."""
    # After fork()'ing, always assure we are generating random sequences
    # that are different from the worker.
    random.seed()

    self.setup_work_horse_signals()

    self._is_horse = True
    self.log = logger

    success = self.perform_job(job, queue)

    # os._exit() is the way to exit from childs after a fork(), in
    # constrast to the regular sys.exit()
    os._exit(int(not success))
```

继续看 `self.perform_job`，发现中间执行了 `job.perform`，然后发现调用了
`self._execute`，然后发现使用了 `self.func` 这个属性，进去一看，发现使用了
`self.func_name`！恍然大悟！这个时候终于发序列化了：

```python
@property
def func(self):
    func_name = self.func_name
    if func_name is None:
        return None

    if self.instance:
        return getattr(self.instance, func_name)

    return import_attribute(self.func_name)
```

找到了 func_name，然后导入，然后把 args和kwargs塞进去执行。就是这样！

## 那enqueue呢？

猜测一下，enqueue是怎么执行的？肯定是把函数的func_name，args，kwargs全部dump
然后塞到对应的queue里啊，答案就在 `queue.py` 的 `enqueue_call`里。我就不继续写
了，就当作是练习吧 :)
