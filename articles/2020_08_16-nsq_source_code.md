# NSQ源码分析

简单的翻了一下NSQ的源码，看看它是怎么实现的。我首先是从nsqtail开始看的，先从简单的入手。之后我看了nsqlookupd和nsqd。本文
只讲nsqd。

首先从nsqd的入口文件看起 `apps/nsqd/main.go`：

```go
func main() {
	prg := &program{}
	if err := svc.Run(prg, syscall.SIGINT, syscall.SIGTERM); err != nil {
		logFatal("%s", err)
	}
}

func (p *program) Start() error {
    // ...
	go func() {
		err := p.nsqd.Main()
		if err != nil {
			p.Stop()
			os.Exit(1)
		}
	}()

	return nil
}

func (n *NSQD) Main() error {
    // ...
	n.tcpServer.ctx = ctx
	n.waitGroup.Wrap(func() {
		exitFunc(protocol.TCPServer(n.tcpListener, n.tcpServer, n.logf))
	})

	httpServer := newHTTPServer(ctx, false, n.getOpts().TLSRequired == TLSRequired)
	n.waitGroup.Wrap(func() {
		exitFunc(http_api.Serve(n.httpListener, httpServer, "HTTP", n.logf))
	})
	n.waitGroup.Wrap(n.queueScanLoop)
	n.waitGroup.Wrap(n.lookupLoop)
    // ...
}
```

其中，`httpServer` 在 `nsq/nsqd/http.go` 中的 `newHTTPServer`，`tcpServer` 在 `nsq/nsqd/tcp.go` 的 `tcpServer`。

其中 tcpServer 的 Handle函数，核心在于 `err = prot.IOLoop(clientConn)`，`prot` 的定义是：

```go
// Protocol describes the basic behavior of any protocol in the system
type Protocol interface {
	IOLoop(conn net.Conn) error
}
```

实现在 `nsqd/protocol_v2.go` 中。这里是每次有一个新的连接时，代表着有一个新的client来了，于是就使用
`tcpServer.Handle` 来处理，Handle函数会有一个for循环：

```go
for {
    // 检查心跳
    if client.HeartbeatInterval > 0 {
        client.SetReadDeadline(time.Now().Add(client.HeartbeatInterval * 2))
    } else {
        client.SetReadDeadline(zeroTime)
    }

    // ReadSlice does not allocate new space for the data each request
    // ie. the returned slice is only valid until the next call to it
    // 读取内容
    line, err = client.Reader.ReadSlice('\n')
    if err != nil {
        if err == io.EOF {
            err = nil
        } else {
            err = fmt.Errorf("failed to read command - %s", err)
        }
        break
    }

    // trim the '\n'
    line = line[:len(line)-1]
    // optionally trim the '\r'
    if len(line) > 0 && line[len(line)-1] == '\r' {
        line = line[:len(line)-1]
    }
    // 解析命令
    params := bytes.Split(line, separatorBytes)

    p.ctx.nsqd.logf(LOG_DEBUG, "PROTOCOL(V2): [%s] %s", client, params)

    var response []byte
    // 执行命令
    response, err = p.Exec(client, params)
    if err != nil {
        ctx := ""
        if parentErr := err.(protocol.ChildErr).Parent(); parentErr != nil {
            ctx = " - " + parentErr.Error()
        }
        p.ctx.nsqd.logf(LOG_ERROR, "[%s] - %s%s", client, err, ctx)

        sendErr := p.Send(client, frameTypeError, []byte(err.Error()))
        if sendErr != nil {
            p.ctx.nsqd.logf(LOG_ERROR, "[%s] - %s%s", client, sendErr, ctx)
            break
        }

        // errors of type FatalClientErr should forceably close the connection
        if _, ok := err.(*protocol.FatalClientErr); ok {
            break
        }
        continue
    }

    // 返回结果
    if response != nil {
        err = p.Send(client, frameTypeResponse, response)
        if err != nil {
            err = fmt.Errorf("failed to send response - %s", err)
            break
        }
    }
}
```

这是server要做的事情。接下来我们看看，一条消息是如何发布的，也就是，我们跳到 `p.Exec` 里，看看 `PUB` 命令
是如何执行的。

```go
func (p *protocolV2) Exec(client *clientV2, params [][]byte) ([]byte, error) {
	if bytes.Equal(params[0], []byte("IDENTIFY")) {
		return p.IDENTIFY(client, params)
	}
	err := enforceTLSPolicy(client, p, params[0])
	if err != nil {
		return nil, err
	}
	switch {
	case bytes.Equal(params[0], []byte("FIN")):
		return p.FIN(client, params)
	case bytes.Equal(params[0], []byte("RDY")):
		return p.RDY(client, params)
	case bytes.Equal(params[0], []byte("REQ")):
		return p.REQ(client, params)
	case bytes.Equal(params[0], []byte("PUB")):
		return p.PUB(client, params)
	case bytes.Equal(params[0], []byte("MPUB")):
		return p.MPUB(client, params)
	case bytes.Equal(params[0], []byte("DPUB")):
		return p.DPUB(client, params)
	case bytes.Equal(params[0], []byte("NOP")):
		return p.NOP(client, params)
	case bytes.Equal(params[0], []byte("TOUCH")):
		return p.TOUCH(client, params)
	case bytes.Equal(params[0], []byte("SUB")):
		return p.SUB(client, params)
	case bytes.Equal(params[0], []byte("CLS")):
		return p.CLS(client, params)
	case bytes.Equal(params[0], []byte("AUTH")):
		return p.AUTH(client, params)
	}
	return nil, protocol.NewFatalClientErr(nil, "E_INVALID", fmt.Sprintf("invalid command %s", params[0]))
}


func (p *protocolV2) PUB(client *clientV2, params [][]byte) ([]byte, error) {
    // ...
	msg := NewMessage(topic.GenerateID(), messageBody)
	err = topic.PutMessage(msg)
    // ...
}
```

接下来我们看看 `topic.PutMessage`：

```go
// PutMessage writes a Message to the queue
func (t *Topic) PutMessage(m *Message) error {
	t.RLock()
	defer t.RUnlock()
	if atomic.LoadInt32(&t.exitFlag) == 1 {
		return errors.New("exiting")
	}
	err := t.put(m)
	if err != nil {
		return err
	}
	atomic.AddUint64(&t.messageCount, 1)
	atomic.AddUint64(&t.messageBytes, uint64(len(m.Body)))
	return nil
}


func (t *Topic) put(m *Message) error {
	select {
	case t.memoryMsgChan <- m:
	default:
		b := bufferPoolGet()
		err := writeMessageToBackend(b, m, t.backend)
		bufferPoolPut(b)
		t.ctx.nsqd.SetHealth(err)
		if err != nil {
			t.ctx.nsqd.logf(LOG_ERROR,
				"TOPIC(%s) ERROR: failed to write message to backend - %s",
				t.name, err)
			return err
		}
	}
	return nil
}
```

这里就是NSQ放消息的时候，优先放内存，如果超过一定量大小时，就会放磁盘的逻辑。我们看到放内存的时候，是放到
`t.memoryMsgChan` 这个channel里的，那接下来如何追踪呢？自然就是搜索一下，看哪里从channel取出，搜索之后，我发现
是在 `messagePump` 函数里：

```go
// messagePump selects over the in-memory and backend queue and
// writes messages to every channel for this topic
func (t *Topic) messagePump() {
    // ...
	// main message loop
	for {
		select {
		case msg = <-memoryMsgChan:
		case buf = <-backendChan:
			msg, err = decodeMessage(buf)
			if err != nil {
				t.ctx.nsqd.logf(LOG_ERROR, "failed to decode message - %s", err)
				continue
			}
		case <-t.channelUpdateChan:
			chans = chans[:0]
			t.RLock()
			for _, c := range t.channelMap {
				chans = append(chans, c)
			}
			t.RUnlock()
			if len(chans) == 0 || t.IsPaused() {
				memoryMsgChan = nil
				backendChan = nil
			} else {
				memoryMsgChan = t.memoryMsgChan
				backendChan = t.backend.ReadChan()
			}
			continue
		case <-t.pauseChan:
			if len(chans) == 0 || t.IsPaused() {
				memoryMsgChan = nil
				backendChan = nil
			} else {
				memoryMsgChan = t.memoryMsgChan
				backendChan = t.backend.ReadChan()
			}
			continue
		case <-t.exitChan:
			goto exit
		}

		for i, channel := range chans {
			chanMsg := msg
			// copy the message because each channel
			// needs a unique instance but...
			// fastpath to avoid copy if its the first channel
			// (the topic already created the first copy)
			if i > 0 {
				chanMsg = NewMessage(msg.ID, msg.Body)
				chanMsg.Timestamp = msg.Timestamp
				chanMsg.deferred = msg.deferred
			}
			if chanMsg.deferred != 0 {
				channel.PutMessageDeferred(chanMsg, chanMsg.deferred)
				continue
			}
			err := channel.PutMessage(chanMsg)
			if err != nil {
				t.ctx.nsqd.logf(LOG_ERROR,
					"TOPIC(%s) ERROR: failed to put msg(%s) to channel(%s) - %s",
					t.name, msg.ID, channel.name, err)
			}
		}
	}
    // ...
}
```

这里就是处理消息的逻辑，看到 `for i, channel := range chans` 这里，就是NSQ把topic下每一个消息，分别复制分发到多个
`channel` 的逻辑。那么，channel只选一个client把消息放进去的逻辑，在哪里呢？

```go
// PutMessage writes a Message to the queue
func (c *Channel) PutMessage(m *Message) error {
	c.RLock()
	defer c.RUnlock()
	if c.Exiting() {
		return errors.New("exiting")
	}
	err := c.put(m)
	if err != nil {
		return err
	}
	atomic.AddUint64(&c.messageCount, 1)
	return nil
}

func (c *Channel) put(m *Message) error {
	select {
	case c.memoryMsgChan <- m:
	default:
		b := bufferPoolGet()
		err := writeMessageToBackend(b, m, c.backend)
		bufferPoolPut(b)
		c.ctx.nsqd.SetHealth(err)
		if err != nil {
			c.ctx.nsqd.logf(LOG_ERROR, "CHANNEL(%s): failed to write message to backend - %s",
				c.name, err)
			return err
		}
	}
	return nil
}
```

我们得看看在哪里消费了 `c.memoryMsgChan` 里的消息，搜了一圈，发现在 `nsqd/protocol_v2.go` 的 `func (p *protocolV2) messagePump`
里：

```go
func (p *protocolV2) messagePump(client *clientV2, startedChan chan bool) {
	var err error
	var memoryMsgChan chan *Message
	var backendMsgChan <-chan []byte
	var subChannel *Channel
	// NOTE: `flusherChan` is used to bound message latency for
	// the pathological case of a channel on a low volume topic
	// with >1 clients having >1 RDY counts
	var flusherChan <-chan time.Time
	var sampleRate int32

	subEventChan := client.SubEventChan
	identifyEventChan := client.IdentifyEventChan
	outputBufferTicker := time.NewTicker(client.OutputBufferTimeout)
	heartbeatTicker := time.NewTicker(client.HeartbeatInterval)
	heartbeatChan := heartbeatTicker.C
	msgTimeout := client.MsgTimeout

	// v2 opportunistically buffers data to clients to reduce write system calls
	// we force flush in two cases:
	//    1. when the client is not ready to receive messages
	//    2. we're buffered and the channel has nothing left to send us
	//       (ie. we would block in this loop anyway)
	//
	flushed := true

	// signal to the goroutine that started the messagePump
	// that we've started up
	close(startedChan)

	for {
		if subChannel == nil || !client.IsReadyForMessages() {
			// the client is not ready to receive messages...
			memoryMsgChan = nil
			backendMsgChan = nil
			flusherChan = nil
			// force flush
			client.writeLock.Lock()
			err = client.Flush()
			client.writeLock.Unlock()
			if err != nil {
				goto exit
			}
			flushed = true
		} else if flushed {
			// last iteration we flushed...
			// do not select on the flusher ticker channel
			memoryMsgChan = subChannel.memoryMsgChan
			backendMsgChan = subChannel.backend.ReadChan()
			flusherChan = nil
		} else {
			// we're buffered (if there isn't any more data we should flush)...
			// select on the flusher ticker channel, too
			memoryMsgChan = subChannel.memoryMsgChan
			backendMsgChan = subChannel.backend.ReadChan()
			flusherChan = outputBufferTicker.C
		}

		select {
		case <-flusherChan:
			// if this case wins, we're either starved
			// or we won the race between other channels...
			// in either case, force flush
			client.writeLock.Lock()
			err = client.Flush()
			client.writeLock.Unlock()
			if err != nil {
				goto exit
			}
			flushed = true
		case <-client.ReadyStateChan:
		case subChannel = <-subEventChan:
			// you can't SUB anymore
			subEventChan = nil
		case identifyData := <-identifyEventChan:
			// you can't IDENTIFY anymore
			identifyEventChan = nil

			outputBufferTicker.Stop()
			if identifyData.OutputBufferTimeout > 0 {
				outputBufferTicker = time.NewTicker(identifyData.OutputBufferTimeout)
			}

			heartbeatTicker.Stop()
			heartbeatChan = nil
			if identifyData.HeartbeatInterval > 0 {
				heartbeatTicker = time.NewTicker(identifyData.HeartbeatInterval)
				heartbeatChan = heartbeatTicker.C
			}

			if identifyData.SampleRate > 0 {
				sampleRate = identifyData.SampleRate
			}

			msgTimeout = identifyData.MsgTimeout
		case <-heartbeatChan:
			err = p.Send(client, frameTypeResponse, heartbeatBytes)
			if err != nil {
				goto exit
			}
		case b := <-backendMsgChan:
			if sampleRate > 0 && rand.Int31n(100) > sampleRate {
				continue
			}

			msg, err := decodeMessage(b)
			if err != nil {
				p.ctx.nsqd.logf(LOG_ERROR, "failed to decode message - %s", err)
				continue
			}
			msg.Attempts++

			subChannel.StartInFlightTimeout(msg, client.ID, msgTimeout)
			client.SendingMessage()
			err = p.SendMessage(client, msg)
			if err != nil {
				goto exit
			}
			flushed = false
		case msg := <-memoryMsgChan:
			if sampleRate > 0 && rand.Int31n(100) > sampleRate {
				continue
			}
			msg.Attempts++

			subChannel.StartInFlightTimeout(msg, client.ID, msgTimeout)
			client.SendingMessage()
			err = p.SendMessage(client, msg)
			if err != nil {
				goto exit
			}
			flushed = false
		case <-client.ExitChan:
			goto exit
		}
	}

exit:
	p.ctx.nsqd.logf(LOG_INFO, "PROTOCOL(V2): [%s] exiting messagePump", client)
	heartbeatTicker.Stop()
	outputBufferTicker.Stop()
	if err != nil {
		p.ctx.nsqd.logf(LOG_ERROR, "PROTOCOL(V2): [%s] messagePump error - %s", client, err)
	}
}
```

看

```go
		case msg := <-memoryMsgChan:
			if sampleRate > 0 && rand.Int31n(100) > sampleRate {
				continue
			}
			msg.Attempts++

			subChannel.StartInFlightTimeout(msg, client.ID, msgTimeout)
			client.SendingMessage()
			err = p.SendMessage(client, msg)
			if err != nil {
				goto exit
			}
			flushed = false
```

这一段，channel里每次只有一个client消费，就是这一段完成的，为啥呢？因为Go的Channel就是这样，多个goroutine监听
一个Channel时，一次只有一个Channel会消费该消息，我们看个demo：

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	var c = make(chan int, 1)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			num := <-c
			fmt.Printf("got %d from channel, I'm channel %d\n", num, i)
			wg.Done()
		}(i)
	}

	for i := 0; i < 10; i++ {
		c <- i
	}

	wg.Wait()
}
```

执行结果：

```bash
$ go run test.go 
got 0 from channel, I'm channel 9
got 2 from channel, I'm channel 4
got 3 from channel, I'm channel 5
got 4 from channel, I'm channel 6
got 5 from channel, I'm channel 7
got 1 from channel, I'm channel 3
got 6 from channel, I'm channel 8
got 9 from channel, I'm channel 2
got 8 from channel, I'm channel 1
got 7 from channel, I'm channel 0
```

## 我的疑惑

不过，我也有一个地方还有疑惑，那就是上面代码的这一段：

```go
		if subChannel == nil || !client.IsReadyForMessages() {
			// the client is not ready to receive messages...
			memoryMsgChan = nil
			backendMsgChan = nil
			flusherChan = nil
			// force flush
			client.writeLock.Lock()
			err = client.Flush()
			client.writeLock.Unlock()
			if err != nil {
				goto exit
			}
			flushed = true
```

由于刚初始化，这段代码似乎一定会被执行？如果也有在看NSQ，并且知道这段代码逻辑的话，麻烦告诉我一声。

## 总结

这篇文章，记录了一下NSQ源码分析流程，了解了NSQ是如何实现一个topic下的消息会被复制到多个channel，
一个channel下的消息只会被多个client中的一个消费，以及消息是如何实现在内存与磁盘中存储的。
