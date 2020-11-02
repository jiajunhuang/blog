# 为啥Redis使用pipelining会更快？

这是一个很考究细节的问题，大部分人都会说：因为减少了网络开销，那么，看如下例子：

```python
import time

import redis

client = redis.Redis(decode_responses=True)
count = 10000


def no_pipelining():
    for i in range(count):
        client.set("test:nopp:{}".format(i), i, ex=100)


def with_pipelining():
    pp = client.pipeline()

    for i in range(count):
        pp.set("test:withpp:{}".format(i), i, ex=100)

    pp.execute()


if __name__ == "__main__":
    start = time.time()
    no_pipelining()
    mid = time.time()
    with_pipelining()
    end = time.time()

    print("no_pipelining: {} seconds; with_pipelining: {} seconds".format(mid - start, end - mid))
```

为什么执行结果相差如此之大呢？

```python
$ python test.py
no_pipelining: 2.3809118270874023 seconds; with_pipelining: 0.4370129108428955 seconds
```

因为这是连接本地的redis，所以网络开销非常小，当然，这里仍然有一部分是网络开销影响，可是除此之外是否还有其它影响因素呢？
答案是有，比如OS进程调度，当不使用管道时，Redis处理每个命令之间是有时间空隙的，因此OS很有可能会将Redis进程转换为sleep状态，
然后运行其它程序，而使用pipelining时，可以提高CPU利用率，Redis空闲的时间没有那么多，因此，这也是pipelining速度会更快的
重要原因之一。

---

ref:

- https://redis.io/topics/pipelining
