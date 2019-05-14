# 密码技术简明教程(三)：证书和TLS

在 [第一篇](https://jiajunhuang.com/articles/2019_05_12-crypto.md.html) 和 [第二篇](https://jiajunhuang.com/articles/2019_05_14-crypto_part2.md.html) 中我们学到了
使用对称加密加密信息，非对称加密配送密钥，使用散列确认文件没有被篡改，使用消息认证码确保知晓密码的才能发送消息，使用数字签名来证明消息的发送者。

那么，如果我们在互联网上访问一个网站进行通信的时候，使用非对称加密进行密钥配送时，怎么确保对方就是对方呢？也就是说，怎么确认我们收到的所谓的 `Alice` 的公钥就真的是 `Alice` 的公钥呢？

答案是，对公钥使用数字签名。

## 证书

因为大家都有公钥，于是世界上各大商家就形成了一个组织，这个组织负责对公钥进行签名。只要是这个组织签名过的公钥我们都信任。这个组织就叫 `CA(Certification Authority)`。

举个例子，如果 `Alice` 要向 `Bob` 发送密文，那么首先 `Bob` 将它的证书发送给 `Alice`。

首先证书有两部分组成：

- `Bob` 的公钥
- `CA` 使用 `CA` 的私钥生成的数字签名

`Alice` 接收到证书之后，首先使用 `CA` 的公钥对数字签名进行确认，如果发现没有问题，确实
是 `CA` 签发的证书，那么接下来，就是用 `Bob` 的公钥对消息进行加密，然后发送给 `Bob`。
`Bob` 再使用自己的私钥进行解密，就可以得到 `Alice` 发出的明文。

如果全世界只有一个 `CA`，由它来签发所有的证书，那么它肯定忙不过来，而且也会由于处于垄断地位而搞各种幺蛾子，因此，`CA` 组织一般都是树状结构，是分层的，而最顶层的 `CA` 就叫 根证书(Root CA)。

## TLS

TLS 就是我们之前学到的几种密码工具的组合，举个例子，我们日常所使用的HTTP协议，他是明文协议，HTTP请求大概长这样：

```
GET / HTTP/1.1
Host: www.example.com
```

而响应大概长这样：

```
HTTP/1.1 200 OK
Date: Mon, 23 May 2005 22:38:34 GMT
Content-Type: text/html; charset=UTF-8
Content-Length: 138
Last-Modified: Wed, 08 Jan 2003 23:11:55 GMT
Server: Apache/1.3.3.7 (Unix) (Red-Hat/Linux)
ETag: "3f80f-1b6-3e1cb03b"
Accept-Ranges: bytes
Connection: close

<html>
  <head>
    <title>An Example Page</title>
  </head>
  <body>
    <p>Hello World, this is a very simple HTML document.</p>
  </body>
</html>
```

可以看到，他们其实就是人类可读的字符，所以对HTTP协议进行中间人攻击，就可以获取所传输的一切，因此我们
需要一种能够兼容HTTP协议(HTTP应用这么广，不可能一下子把它废掉)，但是又能进行加密的技术，它就是TLS。

TLS详细可以参考这里：https://en.wikipedia.org/wiki/Transport_Layer_Security

TLS的工作流程是：

- 客户端首先发送 `ClientHello`，表明客户端可以理解的对称加密算法列表
- 服务端回复 `ServerHello`，从客户端所发送的对称加密算法列表中挑选一个自己也能支持的发送回去，并且附带上自己的证书
- 客户端对证书进行校验，确保这是经过验证的真实的服务器发送来的消息
- 客户端发送对称加密所用的密码
- 到此握手完成
- 之后协议切换到对称加密，然后使用刚才协商的对称加密的密码进行通信

可以发现，TLS其实就是把我们之前所学到的密码技术进行了一个封装。

---

至此我们的密码技术简明教程就结束了，希望读者能够理解日常生活和工作中的种种加密技术，最后，请记住，一定要使用公开算法的，已经被证明安全的加密算法。
