# 队列

今天我们来简单介绍一下队列这种数据结构，根据维基百科的定义，队列是这样一种数据结构：

In computer science, a queue is a collection of entities that are maintained in a sequence and
can be modified by the addition of entities at one end of the sequence and removal from the other end of the sequence.

所以说，队列的特点就是，数据往一个方向流动，就像是一根水管，我们从一端注水，水从另一端流出。

## 队列的实际使用

那么日常项目中什么地方会用到队列呢？我们想想带队列的名词：消息队列、任务队列。没错，他们就是队列。而实现一个最简单的
消息队列可以用Redis中的list，我们需要两个命令：RPUSH和BLPOP。前者的作用是将任务塞入队列，后者的作用是从队列中弹出一个
任务。如[Python-RQ源码分析](https://jiajunhuang.com/articles/2017_09_20-task_queue_python_rq.md.html) 中所说，借助BLPOP
命令：

```python
@classmethod
def lpop(cls, queue_keys, timeout, connection=None):
    connection = resolve_connection(connection)
    if timeout is not None:  # blocking variant
        if timeout == 0:
            raise ValueError('RQ does not support indefinite timeouts. Please pick a timeout value > 0')
        result = connection.blpop(queue_keys, timeout)
    ...
```

## 总结

这一篇中我们介绍了队列的特性，以及队列在实际项目中的应用。

---

参考资料：

- [维基百科中对queue的定义](https://en.wikipedia.org/wiki/Queue_(abstract_data_type))
