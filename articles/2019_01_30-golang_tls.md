# 记一次Golang TLS编程踩坑

最近在写一个HTTP/2代理，一开始使用h2c玩的好好的，结果往测试环境发布，因为跨了公网，因此要加证书，踩了一个坑。

发起连接的客户端代码：

```go
var backendConn net.Conn
var err error
if useTLS {
    backendConn, err = tls.Dial("tcp", backendAddr, nil)
} else {
    backendConn, err = net.Dial("tcp", backendAddr)
}
```

就会发现，没有TLS的情况下，是OK的，但是一旦连上了TLS（我用的Nginx来起TLS服务），Nginx直接返回400。400，那就是客户端错误。对TLS并不熟悉，
最后在 [@jingwei大神](https://jingwei.link/) 的指点下，发现原来是 [ALPN](https://en.wikipedia.org/wiki/Application-Layer_Protocol_Negotiation) 的问题，ALPN是TLS用来给服务端和客户端协商上层协议用的扩展，而上面的代码并没有提供客户端支持什么协议，因此Nginx直接报400。加入即可：

```go
var backendConn net.Conn
var err error
if useTLS {
    backendConn, err = tls.Dial("tcp", backendAddr, &tls.Config{NextProtos: []string{"h2"}})
} else {
    backendConn, err = net.Dial("tcp", backendAddr)
}
```
