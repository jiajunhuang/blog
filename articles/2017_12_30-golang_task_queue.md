# Golang 分布式异步任务队列 Machinery 教程

Golang的分布式任务队列还不算多，目前比较成熟的应该就只有 [Machinery](https://github.com/RichardKnop/machinery) 了。

这篇文章里我们简略的看一下Machinery怎么用。但是我们首先简单介绍一下异步任务这个概念。

如果你熟悉Python中的异步任务框架的话，想必一定听过Celery。异步任务框架是什么呢？异步任务的主要作用是将需要长时间执行
的代码放到一个单独的程序中，例如调用第三方邮件接口，但是这个接口可能非常慢才响应，而你又想确保自己的API及时响应。这个
时候就可以采用异步任务来进行解耦。

一般来说，异步任务都由这么几部分组成：

    - broker：broker是用来传递信息的，我们可以想象成“信使”，“外卖配送员”，它的作用是暂时保存产生的任务以便于消费
    - 生产者：它负责产生任务
    - 消费者：它负责消费任务
    - result backend：这个不是必需，但是如果有保存结果的需要，那么就需要它。

而流程则是：

```
生产者发布任务 -> broker -> 消费者竞争一个任务，然后进行消费 -> (可选：消费后向broker确认已经消费，然后broker删除此任务，
否则将超时重发任务) -> result backend保存结果
```

## Machinery

首先我们来把 `Machinery` 代码拉下来：

```bash
$ go get -u github.com/RichardKnop/machinery/v1
```

Machinery 对消息的定义是：

```go
// Signature represents a single task invocation
type Signature struct {
	UUID           string
	Name           string
	RoutingKey     string
	ETA            *time.Time
	GroupUUID      string
	GroupTaskCount int
	Args           []Arg
	Headers        Headers
	Immutable      bool
	RetryCount     int
	RetryTimeout   int
	OnSuccess      []*Signature
	OnError        []*Signature
	ChordCallback  *Signature
}
```

就如同自己写任务队列可能用json一样。

一般生产者先调用 `signature := tasks.NewSignature` 定义好任务，然后 `machineryServer.SendTask` 就完成了任务的产生。

Machinery 的异步任务长这样：

```go
func Add(args ...int64) (int64, error) {
  sum := int64(0)
  for _, arg := range args {
    sum += arg
  }
  return sum, nil
}
```

要注意一点，函数的最后一个参数必需是 error。然后这样注册任务。

```go
server.RegisterTasks(map[string]interface{}{
  "add":      Add,
})
```

消费者先调用 `worker := machineryServer.NewWorker("send_sms", 10)` 然后 `worker.Launch()` 开始监听broker并且消费任务。
当你产生一个任务，名字是 `add` 时，这个函数就会被调用。

一般你可以把生产者和消费者放到两个文件里，分别定义main函数，然后自己写Makefile，这样就可以直接make然后产生两个可执行
文件，不过我个人更喜欢用 `flag` 来标识到底是什么身份：

```go
func main() {
	// parse cmd args
	flag.Parse()

	// init config
	initConfig()

	// init machinery worker
	initMachinery()

	// register tasks
	machineryServer.RegisterTask("sendSMS", sendSMS)

	if *worker {
		startWorker()
	} else {
		startWebServer()
	}
}
```
