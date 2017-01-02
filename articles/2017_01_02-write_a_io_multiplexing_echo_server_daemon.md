# 重读完APUE，自己写一个I/O多路复用的echo server daemon

重读完APUE，感觉又回忆起和理解了更多东西，动手写了一段代码：

```python
import logging
import os
import selectors
import socket
import sys


DATA_LEN = 4096
selector = selectors.DefaultSelector()


def daemonize():
    # 1, fork off the parent process
    if os.fork() != 0:
        exit()

    # 2, change file mode mask
    os.umask(0)

    # 3, open log
    logging.basicConfig(filename='/tmp/test_daemon.log', level=logging.DEBUG)

    # 4.1, create a unique session id
    os.setsid()

    # 4.2, fork again to prevent session leader get the terminal
    if os.fork() != 0:
        exit()

    # 5, change the current working directory to a safe place
    os.chdir('/')

    # 6, close the STDOUT/STDIN
    sys.stdout.close()
    sys.stdin.close()

    # 7, execute the daemon code
    return


def accept(sock, mask):
    conn, addr = sock.accept()
    logging.info("accept %s from %s" % (conn, addr))
    conn.setblocking(False)
    selector.register(conn, selectors.EVENT_READ, read)


def read(conn, mask):
    data = conn.recv(DATA_LEN)

    def write():
        conn.send(data)
        conn.close()
        selector.unregister(conn)

    selector.unregister(conn)
    logging.info("received data: %s" % data)
    selector.register(
        conn, selectors.EVENT_WRITE, lambda fileobj, mask: write()
    )


def main():
    daemonize()
    sock = socket.socket()
    sock.bind(('localhost', 1234))
    sock.listen(100)
    sock.setblocking(False)
    selector.register(sock, selectors.EVENT_READ, accept)

    while True:
        events = selector.select()
        for key, mask in events:
            callback = key.data
            callback(key.fileobj, mask)


if __name__ == "__main__":
    main()
```

验证一下：

```bash
root@arch ~: curl localhost:1234
GET / HTTP/1.1
Host: localhost:1234
User-Agent: curl/7.52.1
Accept: */*

root@arch ~: cat /tmp/test_daemon.log | wc
    8      84    1216
root@arch ~: ps aux | grep test_daemon
root      2046  0.0  0.4  37112  8860 ?        S    22:47   0:00 python test_daemon.py
root      2430  0.0  0.1  12492  2232 pts/1    S+   22:51   0:00 grep --color=auto test_daemon
```
