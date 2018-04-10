# Golang和Thrift

Thrift是一款RPC协议+工具。我们团队选择了Thrift的主要原因是之前gRPC对gevent的支持不够好。目前虽然有支持，但是合并也
还没有多久。而Thrift有饿了么搞的一套，相对来说好用一些。

## 翻滚吧，RESTful

RESTful这些年来可谓是大红大紫，因为跨平台，human-readable等等。但是实际上我们接RESTful接口的时候，就很蛋疼了。一般
我们都这样干：

- 准备好请求对应的接口的参数，也许要加一堆的头部
- 请求对应的接口，设置超时
- 判断返回的状态码，是否200，400，500等等
- 如果是200，解析json
- 一般返回的json都不会是只有一级的，所以我们还要拿json里的某一层。举个例子，返回的是：

```json
{
    "code": 200,
    "message": "success",
    "result": {
        "name": "someone like you"
    }
}
```

- 如果是Python这种动态语言，取name可能是这样：

```python
>>> name = json_dict.get("result", {}).get("name")
>>> if name:
        print(name)
```

- 如果是Golang，Java等静态语言，还要先定义好结构体或者类，然后unmarshal，并且判断是否marshal出错。。。

所以，RESTful写一个两个还算简单，但是接多了真的是要疯。有了RPC，它会自动帮你生成native的代码，远程调用就像是调用
一个函数一样简单。不过说到底，RESTful只是一种表现形式，通过调用RESTful接口其实也是一种RPC，不过是一种蛋疼得RPC。
我们还是用Thrfit或者gRPC吧。

## Thrift

Thrift有如下几个概念：

- Protocol: 协议，可以类比为HTTP协议

    - TBinaryProtocol 二进制数据
    - TCompactProtocol 紧凑的数据
    - TDenseProtocol 类似于TCompactProtocol不过传输的时候会省略meta infomation
    - TJSONProtocol 使用JSON来传输
    - TSimpleJSONProtocol write-only protocol using JSON
    - TDebugProtocol human-readable text format 方便debug

- Transport: 如何传输，可以类比为TCP

    - TSocket 阻塞I/O
    - TFramedTransport 用frame来发送数据，用非阻塞server时就要用这个
    - TFileTransport 使用文件来传输数据
    - TMemoryTransport 使用内存来传输数据
    - TZlibTransport 传输数据时会使用zlib压缩

- Server: 一个组合上述东西的抽象概念，可以类比为web server

    - TSimpleServer 单线程阻塞IO的server
    - TThreadPoolServer 多线程阻塞IO的server
    - TNonblockingServer 多线程，使用非阻塞IO的server

## Thrift数据类型

- bool
- byte
- i16
- i32
- i64
- double
- string
- binary
- list
- set
- map
- struct 类似于Go的struct
- exception 异常
- service 类似于Go和Java的接口

没有unsigned的类型。论文里说原因是很多编程语言没有这玩意儿，另外据观察用的也少(其实我用的不少啊啊啊啊啊)。

## Go和Thrift

Go的server类似于这样：

```go
func rpcServer() {
	nagatoHandler := &NagatoRPCHandler{}

	transportFactory := thrift.NewTBufferedTransportFactory(BufferSize)
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	transport, err := thrift.NewTServerSocket(config.rpcAddr)

	if err != nil {
		logrus.Fatalf("failed to start rpc socket: %s", err)
	}
	processor := CustomizedTProcessor{p: nagato.NewNagatoServiceProcessor(nagatoHandler)}
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)

	logrus.Infof("rpc server is on %s", config.rpcAddr)
	server.Serve()
}
```

其中最上面的nagatoHandler就是一个 `type NagatoRPCHandler struct{}` 然后给他实现service里定义的方法。因为最后
RPC生成代码里的Processor其实是一个接口。实现了那些方法就好了。

更详细的例子看：https://thrift.apache.org/tutorial/go

## 坑

- thrift-compiler 不能太旧，否则编译出的代码都是用不了的，还会给你报各种奇奇怪怪的错误。一开始我就被debian里的0.9坑了。
自己编译一个0.11才是正道。。。
- RPC的服务端和客户端使用的protocol要对的上
- 没有中间件支持不能加监控。我自己的解决方案是实现 TProcessor 这个接口然后传入到 `thrift.NewTSimpleServer4` 里去。

```go
// CustomizedTProcessor 是定制化的TProcessor，用来搞一些事情
type CustomizedTProcessor struct {
	p thrift.TProcessor
}

// Process 是为了搞事情。。。
func (c CustomizedTProcessor) Process(ctx context.Context, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
	start := time.Now()

	// 执行
	success, err = c.p.Process(ctx, iprot, oprot)

	// 统计，然后返回
	end := time.Now()
	latency := end.Sub(start)

	var status string
	if success {
		status = "200"
	} else {
		status = "400"
	}
	/*
	   endpoint暂时不好拿。可以参考生成的代码里有这么一行：
	   name, _, seqId, err := iprot.ReadMessageBegin()

	   但是目前我还没有看完所有的thrift代码，不敢断定是否所有的protocol实现都不会受影响。所以暂时不这么干。使用reflect
	   拿出一个可以做处标识的先。

	   -。-其实现在这里endpoint也标识不出啥。。。but。。。
	*/
	endpoint := reflect.TypeOf(c.p).String()

	entry := logrus.WithFields(logrus.Fields{
		"request-id": "UNKNOW",
		"status":     status,
		"method":     "rpc",
		"uri":        endpoint,
		"ip":         "UNKNOW",
		"latency":    latency,
		"user-agent": "ThriftRPC",
		"time":       end.Format(time.RFC3339),
	})
	if success {
		entry.Info()
	} else {
		entry.Error(err.Error())
	}

	histogramVec.With(
		prometheus.Labels{
			"method":   "rpc",
			"endpoint": endpoint,
			"service":  "nagato",
			"status":   status,
		},
	).Observe(latency.Seconds())

	return success, err
}
```

当然，目前这个实现还很粗糙。本来是可以拿到具体是哪个processor的。但是 `name, _, seqId, err := iprot.ReadMessageBegin()`
这一行，有点侵入到thrift的实现了。。。而且还没有读完thrift的代码，不敢乱动。。。

----------------------------

- http://jnb.ociweb.com/jnb/jnbJun2009.html
- https://thrift.apache.org/static/files/thrift-20070401.pdf
- https://godoc.org/github.com/apache/thrift/lib/go/thrift
