# socketserver 源码阅读

首先看到文件上半部分的注释，讲清楚了这个module里的类继承关系。

```
+------------+
| BaseServer |
+------------+
        |
        v
+-----------+        +------------------+
| TCPServer |------->| UnixStreamServer |
+-----------+        +------------------+
        |
        v
+-----------+        +--------------------+
| UDPServer |------->| UnixDatagramServer |
+-----------+        +--------------------+
```

不在上面描述的有线程和进程模型，它们是通过Mixin实现的。一个同步的服务器基本上
就是以下这样的套路：

```python
import socket

sock = socket.socket()
sock.bind(('localhost', 8080))
sock.listen()


def handle(conn, client_addr):
    conn.send(b"some data")
    conn.close()


while True:
    conn, client_addr = sock.accept()
    handle(conn, client_addr)
```

socketserver 也不例外，虽然里面用到了 `selectors.PollSelector`，这个文件里最精彩
的部分在于实现 `ThreadingMixIn` 和 `ForkingMixIn`。我也来写一个更简单的版本：

```python
import socket
import threading
import selectors


class BaseServer:
    def __init__(self):
        self.selector = selectors.PollSelector()

    def fileno(self):
        raise NotImplemented()

    def handle_request(self):
        raise NotImplemented()

    def serve_forever(self):
        self.selector.register(self, selectors.EVENT_READ)

        while True:
            if self.selector.select():
                self.handle_request()


class TCPServer(BaseServer):
    def __init__(self):
        super().__init__()
        self.socket = socket.socket()
        self.socket.bind(('localhost', 8080))
        self.socket.listen()

    def fileno(self):
        return self.socket.fileno()

    def handle_request(self):
        return self.__handle_request()

    def _handle_request(self):
        conn, addr = self.socket.accept()
        conn.send(b'hello world\n')
        conn.close()


class ThreadingMixIn:
    def handle_request(self):
        t = threading.Thread(
            target=self._handle_request,
        )
        t.start()


class ThreadingTCPServer(ThreadingMixIn, TCPServer):
    pass


if __name__ == "__main__":
    # TCPServer().serve_forever()
    ThreadingTCPServer().serve_forever()
```
