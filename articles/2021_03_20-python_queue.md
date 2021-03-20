# Python Queue源码分析

今天读源码时发现了 [Queue](https://docs.python.org/3/library/queue.html) 这个标准库，是Python标准库里对队列的实现，翻了
一下源码，发现还挺有意思的。

如果实现一个简单的队列，那其实不难，关键在于我看到 `Queue` 支持阻塞，我很好奇是怎么实现的，于是翻了一下。首先我们来看
Queue提供的方法，其实不多：

```python
class Queue:
    '''Create a queue object with a given maximum size.

    If maxsize is <= 0, the queue size is infinite.
    '''

    def __init__(self, maxsize=0):
        pass

    def task_done(self):
    def join(self):
    def qsize(self):
    def empty(self):
    def full(self):
    def put(self, item, block=True, timeout=None):
    def get(self, block=True, timeout=None):
    def put_nowait(self, item):
    def get_nowait(self):
```

方法不算多，首先我们来看看，它是咋存储队列里的内容的，我们知道，既然是队列，那么肯定是先进先出，比如LPUSH，RPOP，如果
我们用list来实现，那么问题就在于，list的 `insert(0, item)` 效率很差，其实Python那么强大的标准库里，有一个东西叫做 `deque`，
就是一个双向队列的实现，这个实现很经典，很有意思，大家可以去看看。而Queue这个标准库就是基于它来实现的：

```python
    def __init__(self, maxsize=0):
        self.maxsize = maxsize
        self._init(maxsize)

    def _init(self, maxsize):
        self.queue = deque()
```

接着我们来看看 `put` 和 `get`：

```python
    def put(self, item, block=True, timeout=None):
        '''Put an item into the queue.

        If optional args 'block' is true and 'timeout' is None (the default),
        block if necessary until a free slot is available. If 'timeout' is
        a non-negative number, it blocks at most 'timeout' seconds and raises
        the Full exception if no free slot was available within that time.
        Otherwise ('block' is false), put an item on the queue if a free slot
        is immediately available, else raise the Full exception ('timeout'
        is ignored in that case).
        '''
        with self.not_full:
            if self.maxsize > 0:
                if not block:
                    if self._qsize() >= self.maxsize:
                        raise Full
                elif timeout is None:
                    while self._qsize() >= self.maxsize:
                        self.not_full.wait()
                elif timeout < 0:
                    raise ValueError("'timeout' must be a non-negative number")
                else:
                    endtime = time() + timeout
                    while self._qsize() >= self.maxsize:
                        remaining = endtime - time()
                        if remaining <= 0.0:
                            raise Full
                        self.not_full.wait(remaining)
            self._put(item)
            self.unfinished_tasks += 1
            self.not_empty.notify()

    def get(self, block=True, timeout=None):
        '''Remove and return an item from the queue.

        If optional args 'block' is true and 'timeout' is None (the default),
        block if necessary until an item is available. If 'timeout' is
        a non-negative number, it blocks at most 'timeout' seconds and raises
        the Empty exception if no item was available within that time.
        Otherwise ('block' is false), return an item if one is immediately
        available, else raise the Empty exception ('timeout' is ignored
        in that case).
        '''
        with self.not_empty:
            if not block:
                if not self._qsize():
                    raise Empty
            elif timeout is None:
                while not self._qsize():
                    self.not_empty.wait()
            elif timeout < 0:
                raise ValueError("'timeout' must be a non-negative number")
            else:
                endtime = time() + timeout
                while not self._qsize():
                    remaining = endtime - time()
                    if remaining <= 0.0:
                        raise Empty
                    self.not_empty.wait(remaining)
            item = self._get()
            self.not_full.notify()
            return item
```

这里就解释清楚了我之前好奇的地方，阻塞是咋做到的，原来是使用了 `threading.Condition`：

```python
    def __init__(self, maxsize=0):
        self.maxsize = maxsize
        self._init(maxsize)

        # mutex must be held whenever the queue is mutating.  All methods
        # that acquire mutex must release it before returning.  mutex
        # is shared between the three conditions, so acquiring and
        # releasing the conditions also acquires and releases mutex.
        self.mutex = threading.Lock()

        # Notify not_empty whenever an item is added to the queue; a
        # thread waiting to get is notified then.
        self.not_empty = threading.Condition(self.mutex)

        # Notify not_full whenever an item is removed from the queue;
        # a thread waiting to put is notified then.
        self.not_full = threading.Condition(self.mutex)

        # Notify all_tasks_done whenever the number of unfinished tasks
        # drops to zero; thread waiting to join() is notified to resume
        self.all_tasks_done = threading.Condition(self.mutex)
        self.unfinished_tasks = 0
```

要阻塞的话，就调用某个 `Condition` 的 `wait` 方法，这个方法还可以带 `timeout`，原来最终还是操作系统提供的功能。

对于一个 `condition`，调用 `wait` 则开始阻塞，调用 `notify` 就会唤醒其中一个等待的线程，调用 `notify_all` 就会唤醒所有
等待的线程，原来如此。

看完这里，对Queue的实现也就恍然大悟，其余的操作其实就是在拿住 `self.mutex` 的情况下去对 `self.queue` 进行操作了，这里就
不分析了。

全文完。

---

Refs:

- https://docs.python.org/3/library/queue.html
- https://docs.python.org/3/library/threading.html#condition-objects
