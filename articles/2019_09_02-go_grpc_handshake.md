# 一个Golang gRPC握手错误的坑

在 [这个issue](https://github.com/grpc/grpc-go/issues/2406) 里所提到的feature实现之前，Go的gRPC实现里，
客户端和服务端握手过程中，客户端并不会等待HTTP/2协议握手完成之后才开始交互，因此Go的gRPC v1.18之后开始
改变这种行为，实现前面所说的这个feature。然而，这就引入了一个不兼容问题，也引入了一大堆bug。很不幸，我就
踩中了。

> 这个feature可以通过设置 `GRPC_GO_REQUIRE_HANDSHAKE=on` 这个环境变量来开启，也可以通过设置
> `GRPC_GO_REQUIRE_HANDSHAKE=off` 来关闭。

开发一个gRPC应用的时候，客户端不断的报错：

```go
error: rpc error: code = Unavailable desc = all SubConns are in TransientFailure, latest connection error: timed out waiting for server handshake
```

这个错误我以前遇到过，然后就想着设置 `GRPC_GO_REQUIRE_HANDSHAKE=off` 来解决，然后发现并没有用，经过一番查阅之后，发现
原来是在所使用的版本里，即便设置这个环境变量，也没有用：

```go
google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55
google.golang.org/grpc v1.23.0
```

解决方案就是降级版本：

```go
google.golang.org/genproto v0.0.0-20190404172233-64821d5d2107
google.golang.org/grpc v1.19.0
```

over...

---

- [https://github.com/grpc/grpc-go/issues/2406](https://github.com/grpc/grpc-go/issues/2406)
- [https://github.com/grpc/grpc-go/issues/2636](https://github.com/grpc/grpc-go/issues/2636)
- [https://github.com/grpc/grpc-go/issues/2663](https://github.com/grpc/grpc-go/issues/2663)
