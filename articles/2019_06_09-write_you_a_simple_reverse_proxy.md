# 自己动手写一个反向代理

[此前谈到网络编程的重要性](https://jiajunhuang.com/articles/2019_06_06-ip_ban.md.html)，放假在家做了一个反向代理。

目前来说，比较好用的反向代理是 [frp](https://github.com/fatedier/frp)。但是用归用，造轮子归造轮子。明白了底层原理，才心安。

先来看看frp的架构图，基本上反向代理都是这样的架构。

![frp 架构](./img/frp_architecture.png)

注意两点：

- 我们需要一个客户端，用于与服务端保持长连接
- 我们在服务端需要单独监听一个端口，当有新的连接时，就把请求内容转发到客户端与服务端所建立的长连接中

因此，我的 [natrp](https://github.com/jiajunhuang/natrp) 的流量示意图是这样的：

```
                            /---->---\      /--->-----\
Internet(互联网，公网客户端)        公网服务器        局域网的机器
                           \----<---/       \-----<----/
```

上代码，客户端：

```go
package main

import (
	"context"
	"flag"
	"net"
	"time"

	"github.com/jiajunhuang/natrp/dial"
	"github.com/jiajunhuang/natrp/pb/serverpb"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	logger, _ = zap.NewProduction()

	localAddr  = flag.String("local", "127.0.0.1:80", "-local=<你本地需要转发的地址>")
	serverAddr = flag.String("server", "natrp.jiajunhuang.com:10022", "-server=<你的服务器地址>")
	token      = flag.String("token", "balalaxiaomoxian", "-token=<你的token>")
)

func main() {
	defer logger.Sync()

	flag.Parse()
	retryCount := 0

	for {
		func() {
			md := metadata.Pairs("natrp-token", *token)
			ctx := metadata.NewOutgoingContext(context.Background(), md)

			client, conn, err := dial.WithServer(ctx, *serverAddr, false)
			if err != nil {
				logger.Error("failed to connect to server server", zap.Error(err))
				return
			}
			defer conn.Close()

			logger.Info("try to connect to server", zap.String("addr", *serverAddr))

			stream, err := client.Msg(ctx)
			if err != nil {
				logger.Error("failed to communicate with server", zap.Error(err))
				return
			}

			logger.Info("success to connect to server", zap.String("addr", *serverAddr))
			retryCount = 0

			localConn, err := net.Dial("tcp", *localAddr)
			if err != nil {
				logger.Error("failed to communicate with local port", zap.Error(err))
				return
			}
			defer localConn.Close()

			logger.Info("start to build a brige between local and server", zap.String("server", *serverAddr), zap.String("local", *localAddr))

			go func() {
				defer localConn.Close()

				for {
					req, err := stream.Recv()
					if err != nil {
						logger.Error("failed to read", zap.Error(err))
						return
					}

					if _, err := localConn.Write(req.Data); err != nil {
						logger.Error("failed to write", zap.Error(err))
						return
					}
				}
			}()

			data := make([]byte, 1024)
			for {
				n, err := localConn.Read(data)
				if err != nil {
					logger.Error("failed to read", zap.Error(err))
					return
				}

				if err := stream.Send(&serverpb.MsgRequest{Data: data[:n]}); err != nil {
					logger.Error("failed to write", zap.Error(err))
					return
				}
			}
		}()

		if retryCount < 10 {
			time.Sleep(time.Microsecond * time.Duration(100*retryCount))
		} else if retryCount < 60 {
			time.Sleep(time.Second * time.Duration(1))
		} else if retryCount > 60 {
			time.Sleep(time.Second * time.Duration(10))
		}
		logger.Info("trying to reconnect", zap.String("addr", *serverAddr))
		retryCount++
	}
}
```

服务端：

```go
func (s *service) Msg(stream serverpb.ServerService_MsgServer) error {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.ErrBadMetadata
	}
	logger.Info("metadata", zap.Any("metadata", md))
	token := md.Get("natrp-token")
	if len(token) != 1 {
		return errors.ErrBadMetadata
	}

	listenAddr := getListenAddrByToken(token[0])

	listener, err := reuse.Listen("tcp", listenAddr)
	if err != nil {
		logger.Error("failed to listen", zap.Error(err))
		return err
	}
	defer listener.Close()
	addrList := strings.Split(listener.Addr().String(), ":")
	addr := fmt.Sprintf("%s:%s", wanIP, addrList[len(addrList)-1])
	logger.Info("server listen at", zap.String("addr", addr))

	conn, err := listener.Accept()
	if err != nil {
		logger.Error("failed to accept", zap.Error(err))
		return err
	}
	defer conn.Close()

	go func() {
		defer conn.Close()

		for {
			req, err := stream.Recv()
			if err != nil {
				logger.Error("failed to read", zap.Error(err))
				return
			}

			if _, err := conn.Write(req.Data); err != nil {
				logger.Error("failed to write", zap.Error(err))
				return
			}
		}
	}()

	data := make([]byte, 1024)
	for {
		n, err := conn.Read(data)
		if err != nil {
			logger.Error("failed to read", zap.Error(err))
			return err
		}

		if err := stream.Send(&serverpb.MsgResponse{Data: data[:n]}); err != nil {
			logger.Error("failed to write", zap.Error(err))
			return err
		}
	}
}

func getListenAddrByToken(token string) string {
	return "0.0.0.0:10033"
}
```

服务端与客户端之间通信使用gRPC，服务端与公网请求通信使用裸的TCP。

写这个玩意儿发现几个问题：

- proxy要妥善的处理两边的 `net.Conn` 的异常情况，一个关闭之后，能够迅速的关闭另一端
- frp的客户端重试机制应该做的不错，我应该要去阅读一下源码学习一下

---

参考资料：

- https://zh.wikipedia.org/zh-hans/%E5%8F%8D%E5%90%91%E4%BB%A3%E7%90%86
- https://github.com/fatedier/frp
- https://github.com/jiajunhuang/natrp
