# Web开发系列(二)：HTTP协议

在上一次我们介绍TCP协议的最后([点这里](https://jiajunhuang.com/articles/2017_08_12-tcp_ip.md.html))，我们简单的
看了一下HTTP协议大概长什么样：当我们输入 `GET / HTTP/1.1` 回车之后，终端里便回显出服务器所返回的内容。今天我们
更详细一点的看一看HTTP协议。

基于 [从0构建TCP协议](https://jiajunhuang.com/articles/2017_08_12-tcp_ip.md.html) 这篇文章，我们了解到，HTTP协议
是明文协议，也就是我们可以直接肉眼看出来协议的内容，要详细了解一个协议，最精确地办法是读他的标准，于是我们打开Google
搜索 `HTTP RFC`，看到了 [RFC 2616](https://tools.ietf.org/html/rfc2616) 我们将从中抽取一部分内容来描述HTTP协议到底是什么。

## 什么是协议？

首先我们要搞清楚，协议这个词，意味着什么？协议意味着，约定。就像普通话，普通话是一种协议，是一种约定，大家都通过普通话
这种方式，于是全国人民相互沟通起来就会变得很方便，而不是A说广东话，B说闽南语，这样他们是无法交流的。英文也是如此，通过
英文，全世界人民都可以相互交流。

## HTTP协议长什么样？

HTTP协议中有两个概念，一个叫请求，一个叫响应。

在此之前我们需要了解计算机和计算机中常见的两种交流方式，一种是有一台相对计算能力比较强的计算机用来服务众多其他计算机，
这台计算能力比较强的计算机叫做(中心)服务器(Server)，而众多其他计算机叫做客户端(Client)。为什么我故意强调服务器的计算能力会
比较强呢？计算能力弱的计算机也能做服务器，但是一般会导致响应非常的慢。什么叫做响应呢？就是服务器发给客户端的内容，与之相对，
客户端发送给服务器的内容，就叫请求。

可以这样理解，方便记忆，客户端是没有这个资源，所以 **请求** 服务器把资源给他，而服务器便回答了客户端，也就是 **响应** 了客户端。

计算机中还有一种交流方式是没有中心服务器的，所有的机器互联，大家都可能是服务器，大家也都可能是客户端。我们常见的BT种子便是
使用这种方式在计算机之间流传，但这不是我们这次的重点，所以就此打住。

HTTP协议里，请求和响应的格式稍有不同，但大体相当，我们先来看一个请求的示例：

```
$ telnet www.baidu.com 80
Trying 14.215.177.38...
Connected to www.baidu.com.
Escape character is '^]'.
GET / HTTP/1.1
Host: www.baidu.com

HTTP/1.1 200 OK
Date: Sat, 14 Oct 2017 12:08:02 GMT
Content-Type: text/html
Content-Length: 14613
Last-Modified: Mon, 25 Sep 2017 03:07:00 GMT
Connection: Keep-Alive
Vary: Accept-Encoding
Set-Cookie: BAIDUID=0302E3C81EE6F3CFB80F4017871566F8:FG=1; expires=Thu, 31-Dec-37 23:55:55 GMT; max-age=2147483647; path=/; domain=.baidu.com
Set-Cookie: BIDUPSID=0302E3C81EE6F3CFB80F4017871566F8; expires=Thu, 31-Dec-37 23:55:55 GMT; max-age=2147483647; path=/; domain=.baidu.com
Set-Cookie: PSTM=1507982882; expires=Thu, 31-Dec-37 23:55:55 GMT; max-age=2147483647; path=/; domain=.baidu.com
P3P: CP=" OTI DSP COR IVA OUR IND COM "
Server: BWS/1.1
X-UA-Compatible: IE=Edge,chrome=1
Pragma: no-cache
Cache-control: no-cache
Accept-Ranges: bytes

<!DOCTYPE html><!--STATUS OK-->
<html>
...
```

看 `telnet www.baidu.com 80` 之后，我们输入了

```
GET / HTTP/1.1
Host: www.baidu.com

```

首先有一个动词，我们叫 `HTTP Method`，常见的有 `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`, `HEAD`，不常见的有 `TRACE`, `CONNECT`。
然后一个空格，后面接了一个 `/`，表示我们请求这个站点的根目录，或者根网页。再之后是 `HTTP/1.1`，这是HTTP协议的版本号，
根据RFC说明，HTTP协议的版本号要支持多位数字的比较，而不能直接用ASCII比较，例如 `HTTP/11.22` 要比 `HTTP/2.1`更大，
且版本号是 `HTTP/<major>.<minor>`的方式，major是大版本，minor是小版本，例如: `HTTP/0.9`, `HTTP/1.1`, `HTTP/2.0`。

然后是换行符。

在此之后，是 `Host: www.baidu.com`，这叫virtual host，是HTTP协议为了支持同一个IP上服务多个网站而来的，服务器通过判断
这个字段里的内容来将一个请求打到不同的内容服务器上。

在之后是两个换行符。

请注意，HTTP协议中的换行，是 `\r\n`，所以上面这个请求，把换行符和空格打出来，实际上是这样的：

```
GET<space>/<space>HTTP/1.1\r\nHost:<space>www.baidu.com\r\n\r\n
```

两个\r\n表示HTTP协议内容头部的结束，正文的开始，当然也可以不接正文。正文也是各种各样的字符串，那么问题来了，服务器怎么知道
客户端请求里带的是什么内容呢？所以有一个头部叫做 `Content-Type`，他就是用来表明此次携带的内容类型是什么，例如：

```
GET / HTTP/1.1
Host: www.baidu.com
Content-Type: text/html

```

其中 `Content-Type` 后面接什么，需要参考 [MIME](https://tools.ietf.org/html/rfc2045) 里所定义的内容。并且，这里的值只能
是当做参考，如果想的话，是可以完全忽略的，例如 `Content-Type: text/html`，但是却把内容当做json来解析，是一样可以的。

还有很多的头部可以写进去，这需要参考 [这里](https://en.wikipedia.org/wiki/List_of_HTTP_header_fields#Request_fields)

接下来我们讲响应，响应和请求的格式长得差不多，我们直接来看一个例子：

```
$ http http://jiajunhuang.com
HTTP/1.1 301 Moved Permanently
CF-RAY: 3ada8267001c2228-LAX
Connection: keep-alive
Content-Type: text/html
Date: Sat, 14 Oct 2017 12:23:21 GMT
Location: https://jiajunhuang.com/
Server: cloudflare-nginx
Set-Cookie: __cfduid=df544a54aa60644f072c6f3c237d5d7f61507983801; expires=Sun, 14-Oct-18 12:23:21 GMT; path=/; domain=.jiajunhuang.com; HttpOnly
Transfer-Encoding: chunked

<html>
<head><title>301 Moved Permanently</title></head>
<body bgcolor="white">
<center><h1>301 Moved Permanently</h1></center>
<hr><center>nginx/1.10.3 (Ubuntu)</center>
</body>
</html>

```

在上述响应中，首先是协议版本，然后是状态码，然后是状态码的内容，或者说状态码的意思是什么，然后是换行，之后便是各种头部。
再往后便是两个换行，紧接着相应的真正内容。看起来跟请求是不是很像？的确如此。

但是对于响应，我们需要多讲一个东西，那便是状态码：HTTP协议的状态码目前主要有这么几类：

- 1xx: Infomation
- 2xx: Successful
- 3xx: Redirection
- 4xx: Client Error
- 5xx: Server Error

具体可以参考这里：https://en.wikipedia.org/wiki/List_of_HTTP_status_codes

我们可以这样记忆(站在服务器的视角)：

- 1xx: 来，告诉你一点事情
- 2xx: 喏，成功了，这是你要的
- 3xx: 滚你的，去别的地方找
- 4xx: 你出错了！
- 5xx: 我出错了！

具体的状态码还得靠平时，用多了自然就记住了。

> 此外，RFC还说，HTTP协议的头部，请求URL中的域名部分都是不区分大小写的，但是URI确是却分大小写的等等，还有很多小细节在RFC里，强烈建议略读一下，为什么是略读呢？因为太长了。。。

## 解析HTTP

我们自己用Python写一个HTTP解析器，目标很简单，解析出版本号，请求的Host，请求的内容类型，请求的内容，首先我们需要打开socket，
bind，然后listen对应端口：

```python
import socket


server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server.bind(("0.0.0.0", 8088))
server.listen()
```

然后我们要做的事情就是，每来一个请求，我们就处理一个：

```python
import socket


server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server.bind(("0.0.0.0", 8088))
server.listen()


while True:
    client, address = server.accept()
    print(client.recv(4096))
```

我们新建一个终端执行 `http localhost:8088` 试试，可以看到服务器有这样的输出：

```
$ python t.py 
b'GET / HTTP/1.1\r\nHost: localhost:8088\r\nUser-Agent: HTTPie/0.9.9\r\nAccept-Encoding: gzip, deflate\r\nAccept: */*\r\nConnection: keep-alive\r\n\r\n'

```

喏，这里我们看到了刚才我们学习的内容，各种该有的字段都在这，`\r\n`是分隔符。再来看一个POST请求，我们POST一个JSON上去：

`$ http POST localhost:8088 hello=world`

可以看到服务器这样输出：

```
$ python t.py 
b'POST / HTTP/1.1\r\nHost: localhost:8088\r\nUser-Agent: HTTPie/0.9.9\r\nAccept-Encoding: gzip, deflate\r\nAccept: application/json, */*\r\nConnection: keep-alive\r\nContent-Type: application/json\r\nContent-Length: 18\r\n\r\n{"hello": "world"}'
```

仔细观察一下，是不是和我们之前所学到的内容相匹配呢？

接下来我们要写一个函数来处理请求里带上来的内容，然后返回给客户端success这个字段，并且关掉socket连接：

```python
import socket


server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
server.bind(("0.0.0.0", 8088))
server.listen()


def handle_client(content):
    lines = content.split("\r\n")
    method_line = lines[0]
    data = lines[-1]

    print("HTTP Method is: {}".format(method_line.split(" ")[0]))
    print("Data is: {}".format(data))


while True:
    client, address = server.accept()
    handle_client(client.recv(4096).decode("utf-8"))
    client.send(b"HTTP/1.1 200 OK\r\n\r\nsuccessful")
    client.close()
```

新建一个终端，执行：

```
$ http localhost:8088
$ http POST localhost:8088 hello=world
```

可以看到有如下响应：

```
$ http localhost:8088
HTTP/1.1 200 OK

successful

$ http POST localhost:8088 hello=world
HTTP/1.1 200 OK

successful

```

而服务器端则有输出：

```python
$ python t.py 
HTTP Method is: GET
Data is: 
HTTP Method is: POST
Data is: {"hello": "world"}
```

讲完，收工 :)
