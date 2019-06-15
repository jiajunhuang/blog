# NatProxy

NatProxy是一个方便、快捷的内网穿透工具，借助natproxy，你可以远程访问在家里的电脑。例如，通过NatProxy你可以做到：

- 远程访问在家里的笔记本电脑
- 将本地开发的网页发给朋友看
- 将局域网内的其他服务共享给别人
- 等等

## 如何使用NatProxy

### 下载

首先，我们需要下载NatProxy客户端，在 [这个页面](https://github.com/jiajunhuang/natproxy-client/releases/)
点击最新版本下载，下载完成之后，把它放到你想要的目录，如果是Linux/macOS用户，记得添加可执行权限：

```bash
$ sudo chmod +x ./natproxy
```

为了方便在命令行里执行，你还可以把它添加到 `/usr/local/bin/` 下面：

```bash
$ sudo mv ./natproxy /usr/local/bin/
```

### 注册和登录

为了natproxy提供的免费服务器不被滥用，我们需要先注册一个帐号，通过命令行便可以做到，Linux和macOS用户打开终端，Windows用户
打开cmd，如果已经把natproxy添加到了 `PATH` 环境变量里(Linux/macOS直接把文件放到 `/usr/local/bin` 下即是，Windows用户需要
更改环境变量)，那么可以直接输入 `natproxy`，如果没有的话，需要切换目录到 `natproxy` 或 `natproxy.exe` 所在的目录下执行。

> 如果没有添加到 PATH 里，那么下面所有的命令，例如 `natproxy -register` 都应该该写成 `./natproxy -register`，windows用户
> 应该写成 `.\natproxy -register`。

首先我们注册：

```bash
jiajun@idea  ~ $ natproxy -register -email='herosim@qq.com' -password='xxxxxxxx'
2019/06/15 10:59:46 注册成功
```

然后登录一下获取token：

```bash
jiajun@idea  ~ $ natproxy -login -email='herosim@qq.com' -password='xxxxxxxx'
2019/06/15 10:59:59 登录成功，token是 b1c6abcdabcdabcd92666f980fcaabcd
```

接着我们就可以确认一下token是否有效：

```bash
jiajun@idea  ~ $ natproxy -token=b1c6abcdabcdabcd92666f980fcaabcd
2019/06/15 11:00:17 准备连接到服务器(natproxy.laizuoceshi.com:8443)...
2019/06/15 11:00:17 成功连接到服务器(natproxy.laizuoceshi.com:8443)
2019/06/15 11:00:18 检查当前服务端是否已经把本账号设置成断开连接: false
2019/06/15 11:00:19 服务器分配的公网地址是nats-cn1.laizuoceshi.com:25861
^C
```

如果出现 `服务器分配的公网地址是....` 的字符，那就说明已经注册成功了

### 使用

我们简单地看一下帮助文档

```bash
jiajun@idea  ~ $ natproxy --help
Usage of natproxy:
  -email string
    	注册邮箱
  -local string
    	-local=<你本地需要转发的地址> (default "127.0.0.1:8080")
  -login
    	是否登录
  -password string
    	注册密码
  -register
    	是否注册
  -server string
    	-server=<你的服务器地址> (default "natproxy.laizuoceshi.com:8443")
  -socketBufferSize int
    	连接缓冲区大小，越大越快，但是也更吃内存 (default 32768)
  -tls
    	-tls=true 默认使用TLS加密 (default true)
  -token string
    	-token=<你的token> (default "balalaxiaomoxian")
  -toolsAPI string
    	tools API (default "https://tools.jiajunhuang.com")
```

可以知道，如果我们要使用natproxy进行内网穿透，那么首先要有token，token我们已经有了，然后就是告诉natproxy，我们想把什么
暴露出去。例如，我的例子里，是把 `127.0.0.1:80` 这个地址暴露出去。

首先我们要确认地址是可以连通的，我们使用telnet来确认：

```bash
jiajun@idea  ~ $ telnet 127.0.0.1 80
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.
^]
telnet> 
Connection closed.
```

这样说明能连通，如果卡着半天没有动静，那么就说明无法连通。转发一个无法连通的地址，是无效的。
我们来把 `127.0.0.1:80` 进行转发：

```bash
jiajun@idea  ~ $ natproxy -token=b1c69eba0770434192666f980fcafa1e -local='127.0.0.1:80'
2019/06/15 11:00:44 准备连接到服务器(natproxy.laizuoceshi.com:8443)...
2019/06/15 11:00:45 成功连接到服务器(natproxy.laizuoceshi.com:8443)
2019/06/15 11:00:45 服务器分配的公网地址是nats-cn1.laizuoceshi.com:25861
2019/06/15 11:00:48 检查当前服务端是否已经把本账号设置成断开连接: false
```

这个时候我们自己开一个新的终端来试试是不是已经成功了：

```bash
jiajun@idea  ~ $ http HEAD nats-cn1.laizuoceshi.com:25861
HTTP/1.1 200 OK
Connection: keep-alive
Content-Encoding: gzip
Content-Type: text/html; charset=utf-8
Date: Sat, 15 Jun 2019 03:01:19 GMT
Server: nginx/1.16.0
```

大功告成！通过这个公网地址，我们已经访问了在本地的HTTP服务。

### 使用systemd开机自启

编辑 `/etc/systemd/system/natproxy.service`:

```systemd
[Unit]
Description=NatProxy Client Service
After=network.target

[Service]
Type=simple
User=nobody
Restart=on-failure
RestartSec=5s
ExecStart=/usr/local/bin/natproxy -token='<你的token>' -local='<你的目标地址>'

[Install]
WantedBy=multi-user.target
```

然后执行：

```bash
$ sudo systemctl daemon-reload
$ sudo systemctl enable natproxy
$ sudo systemctl start natproxy
$ sudo systemctl status natproxy  # 查看是否成功启动
```

## NatProxy 适合做什么？

- ssh远程访问内网Linux
- rdp远程访问内网Windows
- 将局域网内的HTTP服务临时暴露出来
- ...等等类似的

## NatProxy 不适合做什么？

- 暴露游戏服务器：游戏服务器很多都是使用udp协议进行通信，natproxy暂时还没有支持udp
- 高并发服务：NatProxy目前没有支持连接池，每次公网请求到来，服务端才下发指令要求客户端发起连接进行转发，因此很难应对高并发
的需求，且服务端为了安全，对每个用户进行了连接数限制。

## 注意事项

- 由于会把服务暴露在公网，因此请务必注意不能转发带有私密信息的服务
- 如果要发送带私密信息的服务，我建议：使用安全的通信协议，例如SSH，或者如果是HTTP服务，在你本地的服务上加TLS，这样可以
保证无人能进行中间人攻击；如果是HTTP服务，加上basic auth
- 数据安全的问题所有的内网穿透工具都会有，NatProxy客户端和服务端之间的通信是加密的，也就是说，中间人无法知道你会分配在公网的
哪一台服务器。并且安全始终是NatProxy所重视的，后续会推出新的安全相关的功能来进行应对。

## 反馈问题

你可以加我微信 `gansteed` 注明“natproxy” 进natproxy用户群，也可以在 https://github.com/jiajunhuang/natproxy-client/issues
提issue。

## 为什么不用frp？

因为frp需要自己有一台服务器，而NatProxy可以免费获得。
