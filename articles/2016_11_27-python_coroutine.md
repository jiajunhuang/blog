# 借助coroutine用同步的语法写异步

首先我们构造一个耗时足够久的服务器：

```python
import tornado.gen
import tornado.ioloop
import tornado.web

class MainHandler(tornado.web.RequestHandler):
    @tornado.gen.coroutine
    def get(self):
        yield tornado.gen.sleep(1)
        self.write("Hello, world\n")

if __name__ == "__main__":
    application = tornado.web.Application([
        (r"/", MainHandler),
    ])
    application.listen(8888)
    tornado.ioloop.IOLoop.current().start()
```

每次请求都耗时一秒钟：

```bash
root@arch tests: nohup python test.py > /dev/null &
[1] 15597
nohup: ignoring input and redirecting stderr to stdout
root@arch tests:
root@arch tests: ls
sama  test.py
root@arch tests:
root@arch tests: time curl localhost:8888
Hello, world

real    0m1.018s
user    0m0.000s
sys 0m0.007s
root@arch tests: time curl localhost:8888
Hello, world

real    0m1.016s
user    0m0.003s
sys 0m0.003s
```

## 阻塞型请求

我们先来看代码:

```python
import socket
import time

PORT = 8888
CHUNK_SIZE = 4096


def request():
    sock = socket.socket()
    sock.connect(("", PORT))
    sock.send(b"GET / HTTP/1.1\r\n\r\n")
    data = sock.recv(CHUNK_SIZE)
    print(data.decode())


start = time.time()
request()
request()
end = time.time()
print("use time: %.2f second(s)" % (end - start))
```

这样子的话，请求一次就需要花费一秒，请求是一个接着一个来的，在这中间的时间 进程被投入睡眠。

## I/O多路复用

这个时候我们的老前辈们就有新办法了，好，我们翻开《UNIX环境高级编程》，里面有 专门讲select的，看完这个之后，我们来看看 `Python 3` 提供的 [selectors 模块](https://docs.python.org/3/library/selectors.html#selectors.BaseSelector.register)

我们把上面的代码改改：

```python
import selectors
import socket
import time

PORT = 8888
CHUNK_SIZE = 4096
COUNT = 0

selector = selectors.DefaultSelector()


def request(selector):
    global COUNT
    sock = socket.socket()
    sock.connect(("", PORT))
    selector.register(sock.fileno(), selectors.EVENT_WRITE, data=lambda: writable(selector, sock))
    COUNT += 1


def writable(selector, sock):
    selector.unregister(sock.fileno())
    sock.send(b"GET / HTTP/1.1\r\n\r\n")
    selector.register(sock.fileno(), selectors.EVENT_READ, data=lambda: readable(selector, sock))


def readable(selector, sock):
    global COUNT
    selector.unregister(sock.fileno())
    COUNT -= 1
    data = sock.recv(CHUNK_SIZE)
    print(data.decode())


start = time.time()
request(selector)
request(selector)

while COUNT:
    for key, _ in selector.select():
        callback = key.data
        callback()

end = time.time()
print("use time: %.1f second(s)" % (end - start))
```

```bash
root@arch tests: python client.py
HTTP/1.1 200 OK
Content-Type: text/html; charset=UTF-8
Etag: "7b4758d4baa20873585b9597c7cb9ace2d690ab8"
Server: TornadoServer/4.4.2
Content-Length: 13
Date: Sun, 27 Nov 2016 14:02:38 GMT

Hello, world

HTTP/1.1 200 OK
Content-Type: text/html; charset=UTF-8
Etag: "7b4758d4baa20873585b9597c7cb9ace2d690ab8"
Server: TornadoServer/4.4.2
Content-Length: 13
Date: Sun, 27 Nov 2016 14:02:38 GMT

Hello, world

use time: 1.0 second(s)
```

再运行一下发现两次请求也只花一秒钟时间。这就是I/O多路复用模型的作用～ 但是呢，大把大把的callback把函数拆的四分五散，很不利于阅读。所以接下来我们 就要介绍主角出场: `coroutine`

## coroutine

```python
In [6]: def use_yield():
...:     print("enter the func")
...:     value = yield "hello"
...:     print("got: ", value)
...:     return value
...:

In [7]: gen = use_yield()

In [8]: gen.send(None)
enter the func
Out[8]: 'hello'

In [9]: gen.send("world")
got:  world
---------------------------------------------------------------------------
StopIteration                             Traceback (most recent call last)
<ipython-input-9-ffdc45971c0a> in <module>()
----> 1 gen.send("world")

StopIteration: world

In [10]: type(gen)
Out[10]: generator
```

说好的coroutine呢？怎么最后输出的是generator？别着急，且听我慢慢说来。

首先我们先下个定义，包含了yield关键字的函数就叫generator。来我们先 默念三遍，包含了yield关键字的函数就叫generator； 包含了yield关键字的函数就叫generator； 包含了yield关键字的函数就叫generator。

什么叫generator呢？就是这个函数可以执行到中间某句话的时候，把控制权转让给别人。 并且在未来，别人可以让这个函数从那句话处继续执行。我们通过next让generator执行 到下一个yield处，如果之后没有了yield就会执行到函数结尾，然后抛一个 `StopIteration` 异常。而且我们还可以通过 `.send` 给generator发送数据，恢复它的执行。

个人的理解就是，在python的世界里，coroutine是建立在generator的语法基础上的产物。 并没有具体的形式，coroutine就是用户来控制程序切换。具体在python里就是用户通过 yield把控制权丢出去，通过 `.send` 或者 `next` 来切回那个函数里继续执行。

> 注：接下来所有说用 `next` 的地方，实际代码上我都是用的 `.send`

## coroutine based I/O

我想在等待I/O的时候，把cpu控制权丢出去，让别人继续执行，等到I/O准备完成的时候， 再来执行我。这句话有点熟悉，就跟我们站在第一人称描述I/O多路复用的时候一样： 我想在等待I/O的时候把我挂起，让别人执行，我给你一个回调函数，等到I/O准备完成的 时候，你去执行这个回调函数。

那如果我们想通过yield来抹平回调函数把原本一个函数切分成两个函数的缝隙呢？ 函数执行的一个缺点就是执行完之后，函数中的变量状态就丢失了。

注：我们简单说一下Python的VM，Python是有自己的指令的，就跟x86的cpu有 自己的指令一样。我们来简单看一下：

```python
In [17]: def foo():
    ...:     bar()
    ...:

In [18]: def bar():
    ...:     pass
    ...:

In [19]: import dis

In [20]: dis.dis(foo)
2           0 LOAD_GLOBAL              0 (bar)
            3 CALL_FUNCTION            0 (0 positional, 0 keyword pair)
            6 POP_TOP
            7 LOAD_CONST               0 (None)
            10 RETURN_VALUE

In [21]: dis.dis(bar)
2           0 LOAD_CONST               0 (None)
            3 RETURN_VALUE
```

首先执行foo函数的时候，会由其它函数把环境准备好，把回退指针准备好，然后 调用。

- `LOAD_GLOBAL` 首先从global()里加载bar函数
- `CALL_FUNCTION` 会调用该函数
- `POP_TOP` 会把该函数的栈清掉
- `LOAD_CONST` 把None加载到栈顶，因为这是foo函数的默认返回值
- `RETURN_VALUE` 把None返回

其实我们可以直接把一系列的函数存到 `selector.register` 的data里，但是我们 把它抽出来，就跟ES6里的 `Promise` 一样，我们管它叫 `Future` 。就是一个 普通的类，用来保存回调函数和执行结果的。

```python
class Future:
    def __init__(self):
        self._reuslt = None
        self._callbacks = []

    def set_result(self, result):
        self._result = result
        for callback in self._callbacks:
            callback()

    def add_done_callback(self, callback):
        self._callbacks.append(callback)
```

所以我们把 `register` 改成：

```python
selector.register(sock.fileno(), selectors.EVENT_READ, data=fut)
```

然后在下面的 `select` 处改成：

```python
for key, _ in selector.select():
    fut = key.data
    fut.set_result(None)
```

因为在这里，key.data 已经不是回调函数，而是我们的Future了。

但是我们希望的结果是能够切回我们的函数继续执行，这时候就靠 `next` 了，那我们 要想个办法，让future执行完之后调用 `next(coro)` 。首先我们要找个地方保存住 对coro的引用，所以和Future一样，我们用一个类或者函数来保存都行。为了以后更方便 理解asyncio和tornado，我们用一个类，名字叫 `Task` ：

```python
class Task:
    def __init__(self, coro):
        self.coro = coro

    def step(self):
        try:
            fut = self.coro.send(None)
        except StopIteration:
            return
        fut.add_done_callback(self.step)
```

这样我们调用的时候就是 `task = Task(request())` 然后 `task.step()` 了， 首先 `task = Task(request())` 会执行 `Task.__init__` 会把request()这个 generator保存下来，为啥参数里叫做coro呢？因为我们把它用作coroutine，好以后 我们统称coroutine吧。

接下来通过 `task.step()` 启动coroutine，然后增加一个回调函数，一直执行 到 `selctor.register` ，然后yield。接着执行第二个 `Task(request(selector)).step()` 同样yield。接着执行 `while COUNT` 循环，然后执行 `selctor.select` 并且阻塞 于此，当socket可读时，就会执行 `fut.set_result(None)` 然后就会执行里面的 callback函数，其中有一个callback就是执行上面的 `step` ，借此执行了 `self.coro.send(None)` 从而恢复了coroutine的执行。

如果使用函数的形式，可以通过闭包达到这一点。

```python
def task(coro):
    try:
        fut = coro.send(None)
    except StopIteration:
        return
    fut.add_done_callback(lambda: task(coro))
```

结合上面所说，代码应该是这样的：

```python
import selectors
import socket
import time

PORT = 8888
CHUNK_SIZE = 4096
COUNT = 0

selector = selectors.DefaultSelector()


class Future:
    def __init__(self):
        self._result = None
        self._callbacks = []

    def set_result(self, result):
        self._result = result
        for callback in self._callbacks:
            callback()

    def add_done_callback(self, callback):
        self._callbacks.append(callback)


class Task():
    def __init__(self, coro):
        self.coro = coro

    def step(self):
        try:
            fut = self.coro.send(None)
        except StopIteration:
            return
        fut.add_done_callback(self.step)


def request(selector):
    global COUNT
    fut = Future()

    sock = socket.socket()
    sock.connect(("", PORT))
    selector.register(sock.fileno(), selectors.EVENT_WRITE, data=fut)
    COUNT += 1

    yield fut

    selector.unregister(sock.fileno())
    sock.send(b"GET / HTTP/1.1\r\n\r\n")

    fut = Future()  # 原来的fut已经用完了，我们要来个新的
    selector.register(sock.fileno(), selectors.EVENT_READ, data=fut)

    yield fut

    selector.unregister(sock.fileno())
    COUNT -= 1
    data = sock.recv(CHUNK_SIZE)
    print(data.decode())


start = time.time()
Task(request(selector)).step()
Task(request(selector)).step()

while COUNT:
    for key, _ in selector.select():
        fut = key.data
        fut.set_result(None)

end = time.time()
print("use time: %.1f second(s)" % (end - start))
```

另外， `sock.connect` 是阻塞的，这个时候我们需要把socket设置 成非阻塞的。 `socket.setblocking(False)` 可以把它设置成非阻塞的。

```python
import selectors
import socket
import time

PORT = 8888
CHUNK_SIZE = 4096
COUNT = 0

selector = selectors.DefaultSelector()


class Future:
    def __init__(self):
        self._result = None
        self._callbacks = []

    def set_result(self, result):
        self._result = result
        for callback in self._callbacks:
            callback()

    def add_done_callback(self, callback):
        self._callbacks.append(callback)


class Task():
    def __init__(self, coro):
        self.coro = coro

    def step(self):
        try:
            fut = self.coro.send(None)
        except StopIteration:
            return
        fut.add_done_callback(self.step)


def request(selector):
    global COUNT
    COUNT += 1

    fut = Future()

    sock = socket.socket()
    sock.setblocking(False)

    try:
        sock.connect(("", PORT))
    except BlockingIOError:
        pass

    selector.register(sock.fileno(), selectors.EVENT_WRITE, data=fut)
    yield fut
    selector.unregister(sock.fileno())

    sock.send(b"GET / HTTP/1.1\r\n\r\n")

    fut = Future()  # 原来的fut已经用完了，我们要来个新的

    selector.register(sock.fileno(), selectors.EVENT_READ, data=fut)
    yield fut
    selector.unregister(sock.fileno())

    data = sock.recv(CHUNK_SIZE)
    print(data.decode())
    COUNT -= 1


start = time.time()
Task(request(selector)).step()
Task(request(selector)).step()

while COUNT:
    for key, _ in selector.select():
        fut = key.data
        fut.set_result(None)

end = time.time()
print("use time: %.1f second(s)" % (end - start))
```

这份代码对比起一开始的阻塞型代码，结构上就很类似了，不会因为回调而把一个 函数拆的四分五裂。好了，今天就写到这里吧，下一篇我准备讲讲 `yield` `yield from` `await` `async` -------- `yield` 的前世今生。
