# 使用Redis的Stream模块实现群聊功能

Redis 5.0 加入了一个新的模块：[Stream](https://redis.io/topics/streams-intro)，在这篇文章中，我们使用它来实现IM中的群聊。

首先我们来看看我们的IM有哪些功能，回忆一下我们使用的最多的IM---微信，聊天的形式有两种：

- 单聊
- 群聊

单聊在数据上不算难，比如最简单的，我们可以使用一个关系型数据库来存储每一条聊天记录，或者每一个一对一的关系，我们使用一个
AOF文件来存储，诸如此类，今天我们的重点是群聊。我们继续分析：

- 微信的群聊是不保存历史记录的，意味着从一个手机切换到另外一个手机之后，历史记录就不存在了
- 群聊是多对多的，意味着每一个人都可以发送和接收消息

而Redis的Stream模块，完美的提供了我们所需要的功能。在使用Stream之前，我们得了解什么是Stream，下面是简单的描述，如果你想
更详细的了解它，可以参考 [这里](https://jiajunhuang.com/articles/2018_12_27-redis_stream.md.html)：

TODO

## MVP：阻塞版本

```python
from redis import Redis

r = Redis()

name = input("what's your name? ")
chat_stream = "hello"

while True:
    get = input("what you wanna say? ")

    print(r.xadd(chat_stream, {name: get}))

    print(r.xread({chat_stream: "$"}, None, 0))
```

## 改进：同时收发

## 改进：命令行界面调整

## 总结

---

参考资料：

- https://redis.io/topics/streams-intro
- https://redis-py.readthedocs.io/en/latest/#redis.Redis.xread
- https://jiajunhuang.com/articles/2018_12_27-redis_stream.md.html
