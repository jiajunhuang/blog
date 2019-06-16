# fr转发p 源码阅读与分析(二)：TCP穿透实现

在 [上一篇](https://jiajunhuang.com/articles/2019_06_11-frpc_scurce_code_part1.md.html) 文章中，我们介绍了frp中的一些
概念和基础知识，这一篇中，我们在此前的基础之上，来看看frp是怎么实现TCP内网穿透的。

我们知道，要使用frp，必须有个服务端，然后要有个客户端。因此，我们从这里开始入手。

> 可以参考 《如何阅读源代码》：https://jiajunhuang.com/articles/2018_08_04-how_to_read_source_code.md.html

## frps

`cmd/frps/main.go` 是frps的入口处，我们从这里开始，`main` 函数的主体：

```go
func main() {
	crypto.DefaultSalt = "frp"

	Execute()
}
```

因此我们需要看到 `Execute()` 函数的内容，其实它是使用了 `cobra` 这个库，所以实际的入口在

```go
var rootCmd = &cobra.Command{
        Use:   "frps",
        Short: "frps is the server of frp (https://github.com/fatedier/frp)",
        RunE: func(cmd *cobra.Command, args []string) error {
                if showVersion {
                        fmt.Println(version.Full())
                        return nil
                }

                var err error
                if cfgFile != "" {
                        var content string
                        content, err = config.GetRenderedConfFromFile(cfgFile)
                        if err != nil {
                                return err
                        }
                        g.GlbServerCfg.CfgFile = cfgFile
                        err = parseServerCommonCfg(CfgFileTypeIni, content)
                } else {
                        err = parseServerCommonCfg(CfgFileTypeCmd, "")
                }
                if err != nil {
                        return err
                }

                err = runServer()
                if err != nil {
                        fmt.Println(err)
                        os.Exit(1)
                }
                return nil
        },
}
```

最终，也就是 `runServer()` 这个函数：

```go
func runServer() (err error) {
        log.InitLog(g.GlbServerCfg.LogWay, g.GlbServerCfg.LogFile, g.GlbServerCfg.LogLevel,
                g.GlbServerCfg.LogMaxDays)
        svr, err := server.NewService()
        if err != nil {
                return err
        }
        log.Info("Start frps success")
        server.ServerService = svr
        svr.Run()
        return
}
```

`svr.Run()`，其中的 `svr` 是来自 `server.NewService()`，仔细看一下，`server.NewService()` 其实就是初始化了一大堆东西。
我们直接看 `svr.Run()` 做了什么：

```go
func (svr *Service) Run() {
        if svr.rc.NatHoleController != nil {
                go svr.rc.NatHoleController.Run()
        }
        if g.GlbServerCfg.KcpBindPort > 0 {
                go svr.HandleListener(svr.kcpListener)
        }

        go svr.HandleListener(svr.websocketListener)
        go svr.HandleListener(svr.tlsListener)

        svr.HandleListener(svr.listener)
}
```

可以看到，最后frps会执行到 `svr.HandleListener(svr.listener)`，前面的都是什么 nat hole punching, kcp, websocket, tls等等，我们不看。直接看tcp。

```go
func (svr *Service) HandleListener(l frpNet.Listener) {
	// Listen for incoming connections from client.
	for {
		c, err := l.Accept()
		if err != nil {
			log.Warn("Listener for incoming connections from client closed")
			return
		}
		c = frpNet.CheckAndEnableTLSServerConn(c, svr.tlsConfig)

		// Start a new goroutine for dealing connections.
		go func(frpConn frpNet.Conn) {
        ...
        }
    }
}
```

这里就是监听之后，每来一个新的连接，就起一个goroutine去处理，也就是 `go func()...` 这一段，然后我们看看内容：

```go
switch m := rawMsg.(type) {
case *msg.Login:
    err = svr.RegisterControl(conn, m)
    // If login failed, send error message there.
    // Otherwise send success message in control's work goroutine.
    if err != nil {
        conn.Warn("%v", err)
        msg.WriteMsg(conn, &msg.LoginResp{
            Version: version.Full(),
            Error:   err.Error(),
        })
        conn.Close()
    }
case *msg.NewWorkConn:
    svr.RegisterWorkConn(conn, m)
case *msg.NewVisitorConn:
    if err = svr.RegisterVisitorConn(conn, m); err != nil {
        conn.Warn("%v", err)
        msg.WriteMsg(conn, &msg.NewVisitorConnResp{
            ProxyName: m.ProxyName,
            Error:     err.Error(),
        })
        conn.Close()
    } else {
        msg.WriteMsg(conn, &msg.NewVisitorConnResp{
            ProxyName: m.ProxyName,
            Error:     "",
        })
    }
default:
    log.Warn("Error message type for the new connection [%s]", conn.RemoteAddr().String())
    conn.Close()
}
```

这就是服务端启动之后，卡住的地方了。客户端建立连接之后，会发送一个消息，它的类型可能是 `msg.Login`, `msg.NewWorkConn`, `msg.NewVisitorConn`。上一篇我们说了，visitor
是用于stcp也就是端对端加密通信的，我们不看。workConn就是用于转发流量的，Login就是新的客户端连上去之后进行启动。

## frpc

同样，我们从 `cmd/frpc/main.go` 看起：

```go
func main() {
        crypto.DefaultSalt = "frp"

        sub.Execute()
}
```

跳转到 `sub.Execute()`：

```go
var rootCmd = &cobra.Command{
        Use:   "frpc",
        Short: "frpc is the client of frp (https://github.com/fatedier/frp)",
        RunE: func(cmd *cobra.Command, args []string) error {
                if showVersion {
                        fmt.Println(version.Full())
                        return nil
                }

                // Do not show command usage here.
                err := runClient(cfgFile)
                if err != nil {
                        fmt.Println(err)
                        os.Exit(1)
                }
                return nil
        },
}

func Execute() {
        if err := rootCmd.Execute(); err != nil {
                os.Exit(1)
        }
}
```

然后我们看 `runClient` 函数：

```go
func runClient(cfgFilePath string) (err error) {
        var content string
        content, err = config.GetRenderedConfFromFile(cfgFilePath)
        if err != nil {
                return
        }
        g.GlbClientCfg.CfgFile = cfgFilePath

        err = parseClientCommonCfg(CfgFileTypeIni, content)
        if err != nil {
                return
        }

        pxyCfgs, visitorCfgs, err := config.LoadAllConfFromIni(g.GlbClientCfg.User, content, g.GlbClientCfg.Start)
        if err != nil {
                return err
        }

        err = startService(pxyCfgs, visitorCfgs)
        return
}
```

基本上就是解析配置文件（因为frpc启动的时候要一个配置文件），然后执行 `startService`：

```go
func startService(pxyCfgs map[string]config.ProxyConf, visitorCfgs map[string]config.VisitorConf) (err error) {
        log.InitLog(g.GlbClientCfg.LogWay, g.GlbClientCfg.LogFile, g.GlbClientCfg.LogLevel, g.GlbClientCfg.LogMaxDays)
        if g.GlbClientCfg.DnsServer != "" {
                s := g.GlbClientCfg.DnsServer
                if !strings.Contains(s, ":") {
                        s += ":53"
                }
                // Change default dns server for frpc
                net.DefaultResolver = &net.Resolver{
                        PreferGo: true,
                        Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
                                return net.Dial("udp", s)
                        },
                }
        }
        svr, errRet := client.NewService(pxyCfgs, visitorCfgs)
        if errRet != nil {
                err = errRet
                return
        }

        // Capture the exit signal if we use kcp.
        if g.GlbClientCfg.Protocol == "kcp" {
                go handleSignal(svr)
        }

        err = svr.Run()
        if g.GlbClientCfg.Protocol == "kcp" {
                <-kcpDoneCh
        }
        return
}
```

同样的，执行 `client.NewService` 之后执行 `svr.Run()`，我们看看 `svr.Run()` 是什么：

```go
func (svr *Service) Run() error {
	// first login
	for {
		conn, session, err := svr.login()
		if err != nil {
			log.Warn("login to server failed: %v", err)

			// if login_fail_exit is true, just exit this program
			// otherwise sleep a while and try again to connect to server
			if g.GlbClientCfg.LoginFailExit {
				return err
			} else {
				time.Sleep(10 * time.Second)
			}
		} else {
			// login success
			ctl := NewControl(svr.runId, conn, session, svr.pxyCfgs, svr.visitorCfgs)
			ctl.Run()
			svr.ctlMu.Lock()
			svr.ctl = ctl
			svr.ctlMu.Unlock()
			break
		}
	}

	go svr.keepControllerWorking()

	if g.GlbClientCfg.AdminPort != 0 {
		err := svr.RunAdminServer(g.GlbClientCfg.AdminAddr, g.GlbClientCfg.AdminPort)
		if err != nil {
			log.Warn("run admin server error: %v", err)
		}
		log.Info("admin server listen on %s:%d", g.GlbClientCfg.AdminAddr, g.GlbClientCfg.AdminPort)
	}

	<-svr.closedCh
	return nil
}
```

可以看到，客户端启动之后，就是一个 `for` 循环，写入 `login` 信息，也就是刚才 `frps` 里的 `msg.loginMsg`，
然后起一个 `goroutine` 执行 `keepControllerWorking()`，之后主 `goroutine` 就阻塞在 `<-svr.closedCh`。
看看 `keepControllerWorking()` 的内容：

```go
func (svr *Service) keepControllerWorking() {
	maxDelayTime := 20 * time.Second
	delayTime := time.Second

	for {
		<-svr.ctl.ClosedDoneCh()
		if atomic.LoadUint32(&svr.exit) != 0 {
			return
		}

		for {
			log.Info("try to reconnect to server...")
			conn, session, err := svr.login()
			if err != nil {
				log.Warn("reconnect to server error: %v", err)
				time.Sleep(delayTime)
				delayTime = delayTime * 2
				if delayTime > maxDelayTime {
					delayTime = maxDelayTime
				}
				continue
			}
			// reconnect success, init delayTime
			delayTime = time.Second

			ctl := NewControl(svr.runId, conn, session, svr.pxyCfgs, svr.visitorCfgs)
			ctl.Run()
			svr.ctlMu.Lock()
			svr.ctl = ctl
			svr.ctlMu.Unlock()
			break
		}
	}
}
```

基本上就是一个循环，里面最终是为了成功连接然后执行 `ctl.Run()`：

```go
func (ctl *Control) Run() {
	go ctl.worker()

	// start all proxies
	ctl.pm.Reload(ctl.pxyCfgs)

	// start all visitors
	go ctl.vm.Run()
	return
}

// If controler is notified by closedCh, reader and writer and handler will exit
func (ctl *Control) worker() {
	go ctl.msgHandler()
	go ctl.reader()
	go ctl.writer()

	select {
	case <-ctl.closedCh:
		// close related channels and wait until other goroutines done
		close(ctl.readCh)
		ctl.readerShutdown.WaitDone()
		ctl.msgHandlerShutdown.WaitDone()

		close(ctl.sendCh)
		ctl.writerShutdown.WaitDone()

		ctl.pm.Close()
		ctl.vm.Close()

		close(ctl.closedDoneCh)
		if ctl.session != nil {
			ctl.session.Close()
		}
		return
	}
}

// msgHandler handles all channel events and do corresponding operations.
func (ctl *Control) msgHandler() {
	defer func() {
		if err := recover(); err != nil {
			ctl.Error("panic error: %v", err)
			ctl.Error(string(debug.Stack()))
		}
	}()
	defer ctl.msgHandlerShutdown.Done()

	hbSend := time.NewTicker(time.Duration(g.GlbClientCfg.HeartBeatInterval) * time.Second)
	defer hbSend.Stop()
	hbCheck := time.NewTicker(time.Second)
	defer hbCheck.Stop()

	ctl.lastPong = time.Now()

	for {
		select {
		case <-hbSend.C:
			// send heartbeat to server
			ctl.Debug("send heartbeat to server")
			ctl.sendCh <- &msg.Ping{}
		case <-hbCheck.C:
			if time.Since(ctl.lastPong) > time.Duration(g.GlbClientCfg.HeartBeatTimeout)*time.Second {
				ctl.Warn("heartbeat timeout")
				// let reader() stop
				ctl.conn.Close()
				return
			}
		case rawMsg, ok := <-ctl.readCh:
			if !ok {
				return
			}

			switch m := rawMsg.(type) {
			case *msg.ReqWorkConn:
				go ctl.HandleReqWorkConn(m)
			case *msg.NewProxyResp:
				ctl.HandleNewProxyResp(m)
			case *msg.Pong:
				ctl.lastPong = time.Now()
				ctl.Debug("receive heartbeat from server")
			}
		}
	}
}

// reader read all messages from frps and send to readCh
func (ctl *Control) reader() {
	defer func() {
		if err := recover(); err != nil {
			ctl.Error("panic error: %v", err)
			ctl.Error(string(debug.Stack()))
		}
	}()
	defer ctl.readerShutdown.Done()
	defer close(ctl.closedCh)

	encReader := crypto.NewReader(ctl.conn, []byte(g.GlbClientCfg.Token))
	for {
		if m, err := msg.ReadMsg(encReader); err != nil {
			if err == io.EOF {
				ctl.Debug("read from control connection EOF")
				return
			} else {
				ctl.Warn("read error: %v", err)
				ctl.conn.Close()
				return
			}
		} else {
			ctl.readCh <- m
		}
	}
}

// writer writes messages got from sendCh to frps
func (ctl *Control) writer() {
	defer ctl.writerShutdown.Done()
	encWriter, err := crypto.NewWriter(ctl.conn, []byte(g.GlbClientCfg.Token))
	if err != nil {
		ctl.conn.Error("crypto new writer error: %v", err)
		ctl.conn.Close()
		return
	}
	for {
		if m, ok := <-ctl.sendCh; !ok {
			ctl.Info("control writer is closing")
			return
		} else {
			if err := msg.WriteMsg(encWriter, m); err != nil {
				ctl.Warn("write message to control connection error: %v", err)
				return
			}
		}
	}
}
```

`reader` 从frps收信息，然后写到 `ctl.readCh` 这个 channel 里，`writer` 则相反，
从 `ctl.sendCh` 收信息，写到 frps，而 `msgHandler` 则从frpc里读取信息，放到 `ctl.sendCh`，
从 `ctl.readCh` 读取信息，处理之。

之所以这样设计，是为了能够异步处理所有消息。看看 `msgHandler` 的关键部分：

```go
switch m := rawMsg.(type) {
case *msg.ReqWorkConn:
    go ctl.HandleReqWorkConn(m)
case *msg.NewProxyResp:
    ctl.HandleNewProxyResp(m)
case *msg.Pong:
    ctl.lastPong = time.Now()
    ctl.Debug("receive heartbeat from server")
}
```

我们说过了，frps 每次收到一个请求之后，然后下发一个指令给frpc，要求frpc建立连接，
然后frps再把新来的连接与请求所在的连接串起来，完成代理，`msg.ReqWorkConn` 就是这个指令。

那么 `frps` 是在哪里下发指令的呢？

## frps 下发指令

每当公网来一个新的请求的时候，frps就会下发一个指令给frpc，要求建立一个新的连接，代码如下：

```go
func (pxy *BaseProxy) GetWorkConnFromPool(src, dst net.Addr) (workConn frpNet.Conn, err error) {
        // try all connections from the pool
        for i := 0; i < pxy.poolCount+1; i++ {
                if workConn, err = pxy.getWorkConnFn(); err != nil {
                        pxy.Warn("failed to get work connection: %v", err)
                        return
                }
                pxy.Info("get a new work connection: [%s]", workConn.RemoteAddr().String())
                workConn.AddLogPrefix(pxy.GetName())

                var (
                        srcAddr    string
                        dstAddr    string
                        srcPortStr string
                        dstPortStr string
                        srcPort    int
                        dstPort    int
                )

                if src != nil {
                        srcAddr, srcPortStr, _ = net.SplitHostPort(src.String())
                        srcPort, _ = strconv.Atoi(srcPortStr)
                }
                if dst != nil {
                        dstAddr, dstPortStr, _ = net.SplitHostPort(dst.String())
                        dstPort, _ = strconv.Atoi(dstPortStr)
                }
                message := msg.StartWorkConn{
                                ProxyName: pxy.GetName(),
                                SrcAddr:   srcAddr,
                                SrcPort:   uint16(srcPort),
                                DstAddr:   dstAddr,
                                DstPort:   uint16(dstPort),
                        }
                err := msg.WriteMsg(workConn, &message)
                workConn.Warn("===== HERE! HERE! HERE!, message is: %+v", message)
                if err != nil {
                        workConn.Warn("failed to send message to work connection from pool: %v, times: %d", err, i)
                        workConn.Close()
                } else {
                        break
                }
        }

        if err != nil {
                pxy.Error("try to get work connection failed in the end")
                return
        }
        return
}
```

那个 `===== HERE! HERE! HERE!, message is: %+v` 是我加上去的，为了方便看每次服务端下发什么信息。我们看的是TCP的代理，
它继承于 `server/proxy/proxy.go` 里的 `BaseProxy`，实现是在 `server/proxy/tcp.go` 里的 `type TcpProxy struct`。

那么什么时候会初始化 `TcpProxy` 呢？我们注意到，frps 接收的消息里，就有一种是消息类型是 `msg.NewProxy`，这是客户端和
服务端都启动，并且客户端成功login之后，客户端发送给服务端的消息。也就是 `frpc` 的配置文件里具体的代理，例如：

```ini
[common]
server_addr = xxxxx
server_port = 12345
tls_enable = true

[ssh]
type = tcp
local_ip = 127.0.0.1
local_port = 22
remote_port = 12346
```

中的 `ssh` 就是一个TCP代理。收到 `msg.NewProxy` 之后，服务端会起一个新的监听器监听在对应的端口，然后开始处理请求。

## 总结

这一篇中，我们在第一篇的基础之上看了frpc和frps的交互流程，了解了frp是如何进行TCP代理的。
