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

Stream的中文意思是流，你可以把它想象成水流。水流中流动的是水，而Redis中的流里，流动的是信息。就像一根水管，我们可以
从一段注水，从另一端放水一样，在Redis中，我们可以从一端注入信息，我们称之为产生信息(产生信息的一端叫做生产者)，也可以
在另一端消费信息(消费信息的一端叫做消费者)。

Stream中产生消息，用的是 `XADD` 这个命令，而消费则是用 `XREAD`。

接下来，我们就要用这两个命令来实现一个简单的群聊对话。

## MVP：阻塞版本

```python
from redis import Redis  # 导入包

r = Redis()  # 初始化Redis对象实例，这里没有填参数，因此会连接本地的redis: 127.0.0.1:6379

name = input("what's your name? ")  # 首先要求输入一个名字，作为待会儿群聊时的身份认证标识
chat_stream = "my_chat_stream"  # 这是这个群聊的一个标识，相当于一个群的名字

while True:  # 进入死循环
    user_input = input("what you wanna say? ")  # 首先输出你想说啥？提示用户输入内容

    r.xadd(chat_stream, {name: user_input})  # 然后发送输入的内容到群聊内容里

    print(r.xread({chat_stream: "$"}, None, 0))  # 输出从群里读出来的内容
```

运行一下，起一个终端，执行 `python main.py`：

```python
$ python main.py 
what's your name? jhon
what you wanna say? hello world
[[b'hello', [(b'1586415762129-0', {b'marry': b'nothing just kidding'})]]]
what you wanna say? hello, marry
[[b'hello', [(b'1586415773558-0', {b'marry': b'well'})]]]

```

起另外一个终端，同样执行 `python main.py`：

```python
$ python main.py 
what's your name? marry
what you wanna say? nothing just kidding
[[b'hello', [(b'1586415768566-0', {b'jhon': b'hello, marry'})]]]
what you wanna say? well

```

可以看到，它们互相可以看到内容，但是有一个缺点，那就是，每次都要等读取完之后，才可以进行下一次输入，接下来我们着手
改进这一点。

## 改进：同时收发

由于我们收取消息的时候，超时时间填的是0,也就是说，没有收取到消息就一直等待，所以如果没有群聊消息来，我们就只能一直
等着，而不能输入新的消息。比较简单的一个改进，是改成没有收到消息的时候不等待，直接进入下一次循环，不过这样就有另外
一个缺点，那就是如果不输入内容，就一直看不到新的消息。

有没有办法让收消息和发消息互相不干扰呢？有办法，我们要使用线程：一个线程负责收消息，一个负责发消息：

```python
import threading

from redis import Redis  # 导入包

r = Redis()  # 初始化Redis对象实例，这里没有填参数，因此会连接本地的redis: 127.0.0.1:6379

name = input("what's your name? ")  # 首先要求输入一个名字，作为待会儿群聊时的身份认证标识
chat_stream = "my_chat_stream"  # 这是这个群聊的一个标识，相当于一个群的名字


def send_msgs():
    while True:  # 进入死循环
        user_input = input("what you wanna say? ")  # 首先输出你想说啥？提示用户输入内容
        r.xadd(chat_stream, {name: user_input})  # 然后发送输入的内容到群聊内容里


def recv_msgs():
    while True:
        print(r.xread({chat_stream: "$"}, None, 0))  # 输出从群里读出来的内容


if __name__ == "__main__":
    threading.Thread(target=recv_msgs).start()
    send_msgs()
```

如果你运行一下，就会发现现在已经可以了，但是仍然有缺点，那就是两端都会把自己的信息打印出来，这是因为我们对收到的信息
没有做处理，而是直接打印出来了。下一步，就是改进这些：

```python
import threading

from redis import Redis  # 导入包

r = Redis()  # 初始化Redis对象实例，这里没有填参数，因此会连接本地的redis: 127.0.0.1:6379

name = input("what's your name? ")  # 首先要求输入一个名字，作为待会儿群聊时的身份认证标识
chat_stream = "my_chat_stream"  # 这是这个群聊的一个标识，相当于一个群的名字


def send_msgs():
    while True:  # 进入死循环
        user_input = input("what you wanna say? ")  # 首先输出你想说啥？提示用户输入内容
        if user_input:
            r.xadd(chat_stream, {name: user_input})  # 然后发送输入的内容到群聊内容里


def handle_msgs(msgs):
    # msgs结构是：[[b'my_chat_stream', [(b'1586416610013-0', {b'jhon': b'nothing'})]]]
    for msg in msgs:  # 迭代，因此msg是 [b'my_chat_stream', [(b'1586416610013-0', {b'jhon': b'nothing'})]]
        _, msg_list = msg  # 解包，因此 msg_list 是 [(b'1586416610013-0', {b'jhon': b'nothing'})]
        for _, content in msg_list:  # 再次解包并迭代，因此 content是 {b'jhon': b'nothing'}
            for user_name, user_input in content.items():  # 迭代，因此user_name是 b'jhon' 而 user_input 是 b'nothing'
                decoded_user_name = user_name.decode()
                decoded_user_input = user_input.decode()
                if decoded_user_name == name:
                    continue

                print("[{}]: {}".format(decoded_user_name, decoded_user_input))


def recv_msgs():
    while True:
        msgs = r.xread({chat_stream: "$"}, None, 0)  # 获取从群里读出来的内容
        handle_msgs(msgs)  # 因为逻辑不算简单，为了这里看起来简单易懂，我们把处理消息的逻辑放在另外一个函数里


if __name__ == "__main__":
    threading.Thread(target=recv_msgs).start()
    send_msgs()
```

## 总结

这一篇文章中，我们通过一步一步的迭代，借助Redis的Stream功能，实现了一个简单的群聊功能，首先我们用一个简单的死循环，
输入后输入内容，接着我们进行改进，让输入和输出分别处理，互补干扰，最后我们对处理消息的逻辑进行进一步优化，使得
输出的内容看起来干净整洁。

---

参考资料：

- https://redis.io/topics/streams-intro
- https://redis-py.readthedocs.io/en/latest/#redis.Redis.xread
- https://jiajunhuang.com/articles/2018_12_27-redis_stream.md.html
