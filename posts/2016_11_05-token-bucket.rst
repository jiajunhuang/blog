Token Bucket 算法
==================

最近在看celery代码，看到worker的代码，发现了tocket bucket算法，自己之前还一直
在想像有些API对调用次数有限制是怎么做到的，看完才发现，原来是这么朴素的算法。
真是无知，无知啊。嗯，以后要多读读源码。

celery外链里但是写上了注释的代码(稍微做了一些改动)：

.. code:: python

   # coding: utf-8
    import time

    class TokenBucket(object):
        def __init__(self, tokens, fill_rate):
            """tokens is the total tokens in the bucket. fill_rate is the
            rate in tokens/second that the bucket will be refilled."""
            self.capacity = tokens  # 桶的容量
            self._tokens = tokens  # 令牌们
            self.fill_rate = fill_rate  # 每秒放入的令牌数量
            self.timestamp = int(time.time())  # 上次请求令牌的时间

        def consume(self, tokens):
            """Consume tokens from the bucket. Returns True if there were
            sufficient tokens otherwise False."""
            if tokens <= self.__get_tokens():
                self._tokens -= tokens
            else:
                return False
            return True

        def __get_tokens(self):
            now = int(time.time())
            if self._tokens < self.capacity:
                delta = self.fill_rate * (now - self.timestamp)
                self._tokens = min(self.capacity, self._tokens + delta)
                print("delta: %s" % delta)
            self.timestamp = now
            return self._tokens

    if __name__ == "__main__":
        bucket = TokenBucket(80, 1)
        print("tokens = %s" % bucket._tokens)
        print("consume(10) = %s" % bucket.consume(10))
        print("consume(10) = %s" % bucket.consume(10))
        time.sleep(1)
        print("tokens = %s" % bucket._tokens)
        time.sleep(1)
        print("tokens = %s" % bucket._tokens)
        print("consume(90) = %s" % bucket.consume(90))
        print("tokens = %s" % bucket._tokens)
        print("consume(90) = %s" % bucket.consume(90))
        print("tokens = %s" % bucket._tokens)

我们看一下测试结果:

.. code:: bash

    $ python token_bucket.py
    tokens = 80
    consume(10) = True
    delta: 0
    consume(10) = True
    tokens = 60
    tokens = 60
    delta: 2
    consume(90) = False
    tokens = 62
    delta: 0
    consume(90) = False
    tokens = 62

最后我们用大白话来描述一下上面的代码：

1，初始化的时候，指定了桶的大小和每秒钟放入令牌的速率

2，每次消耗令牌的时候，都会计算，上次消耗到本次消耗之间产生了多少令牌，如果产生
令牌的数量超过了容量，则丢弃多余的令牌。

3，如果要消耗的令牌数量大于现有的令牌数量，则返回失败。


.. [#] `https://en.wikipedia.org/wiki/Token_bucket`_

.. [#] `http://code.activestate.com/recipes/511490/`_
