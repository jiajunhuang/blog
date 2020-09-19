# Linux常用命令(一)：netcat

netcat是常见的用于网络相关的工具，比如连接，监听，扫描端口等。和所有的UNIX命令一样，netcat也是一把瑞士军刀。接下来我们
来学习如何使用。

## 常见版本

使用之前我们要先了解一下netcat的几个版本：

- 原版netcat `sudo apt install netcat-traditional` 安装。这是最初的版本。
- GNU版netcat `sudo apt install netcat` 这是GNU为了符合GPL协议，重写的版本。
- OpenBSD版netcat `sudo apt install netcat-openbsd` 这是openbsd的版本。也是用的比较多的版本，我们也使用这个版本。

接下来我们了解一下netcat的命令行如何使用。它的参数有两种方式：

- 分开描述，比如 `nc -h`, `nc -w -v -z`
- 放在一起描述，比如 `nc -zu`

当然，上述的参数方式是可以组合的，比如：

```bash
$ nc -zv baidu.com -w 3 80
Connection to baidu.com 80 port [tcp/http] succeeded!
```

一般是，如果参数里还要带值的话，就用分开描述，否则就放在一起好了。

剩下的选项，我根据功能把它们划分开来：

### 网络模式和代理相关

- `-l` 监听，作服务器。不填时作客户端。
- `-u` UDP模式。不填时默认TCP模式。
- `-X` 和 `-x` 是代理相关的选项

### 其余常用选项

- `-v` verbose模式，打印更多日志
- `-z` 连接以后就断开，用于测试网络连接是否连通
- `-w` 超时时间，单位是秒
- `-s` 指定source addr
- `-p` 指定source port
- `-n` 不查询DNS
- `-k` 处理完一个请求之后，继续监听下一个

## 常见的例子及其释义

了解了常见的参数之后，我们就来用用它：

- 最简单的echo server和echo client：一个终端里执行 `nc -l 8080`，
另外一个终端里执行 `nc 127.0.0.1 8080`，他们就通过 TCP 连接了，解释一下上面的参数，`nc -l 8080`
就是监听在8080端口，本来应该是 `nc -l 127.0.0.1 8080`，这样的格式，但是如果默认不写监听的地址的话，
那么就是使用默认的 `127.0.0.1`。而第二个终端里的 `nc 127.0.0.1 8080` 就不能省略地址了，作为客户端的
时候，必须指明IP地址和端口。如果我们加上 `-u` 参数，那么就是UDP版本的echo server和client了。

- 测试某个地址是否能连通。一般我们会用telnet，不过telnet还需要通过 `Ctrl-]` 来退出。用nc更简单：

```bash
$ nc -zv baidu.com 80
Connection to baidu.com 80 port [tcp/http] succeeded!

```

解释一下参数，`-v` 就是打印日志的意思。如果不加的话，nc不会打印下面的这一行 `Connection ...`，而
`-z` 则是建立TCP连接之后就断开的意思。所以我们就可以通过这两个命令组合，来快速判断某个地址是否可以连通。

- 端口扫描。有时候我们想要范围扫描一下端口，那么就可以这样：

```bash
$ nc -vz 127.0.0.1 20-25
nc: connect to 127.0.0.1 port 20 (tcp) failed: Connection refused
nc: connect to 127.0.0.1 port 21 (tcp) failed: Connection refused
Connection to 127.0.0.1 22 port [tcp/ssh] succeeded!
nc: connect to 127.0.0.1 port 23 (tcp) failed: Connection refused
nc: connect to 127.0.0.1 port 24 (tcp) failed: Connection refused
Connection to 127.0.0.1 25 port [tcp/smtp] succeeded!

```

## 总结

这就是netcat的常见选项和使用方法了，通过这个工具，结合其它Linux命令还可以发挥更强大的功能。
