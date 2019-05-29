# Golang实现平滑重启(优雅重启)

> 平滑重启，也叫优雅重启，热重启，平滑升级，热升级，热更新等等。

最近在看traefik的源代码, 看到其中有一个功能是平滑重启, 不过他是通过一个叫做 [graceful](https://github.com/tylerb/graceful)
的库来做到的, 是在HTTP Server的层级. 于是我探索了一下方案, 在TCP层级做了一个demo出来.

## 先看traefik的实现方案

上面说的graceful这个库,如它在Github的简介所说: "Graceful is a Go package enabling graceful shutdown of an http.Handler server."
只提供了优雅关闭, 不提供优雅重启. 那么什么叫做优雅关闭呢? 意思就是服务器要关闭了, 会拒绝新的连接,但是老的连接不会被强制关闭,而是
会等待一定时间, 等待客户端主动关闭, 除非客户端一直没有关闭, 到了预设的超时时间才进行服务器端关闭.

我们来看看traefik是怎么做的:

```go
func main() { // goroutine 0
	goAway := false
	go func() {  // goroutine 1
		sig := <-sigs
		fmt.Println("I have to go...", sig)
		goAway = true
		srv.Stop(10 * time.Second)
	}()

	for{
		if (goAway){
			break
		}
		fmt.Println("Started")
		srv = &graceful.Server{
			Timeout: 10 * time.Second,
			NoSignalHandling: true,

			ConnState: func(conn net.Conn, state http.ConnState) {
				fmt.Println( "Connection ", state)
			},

			Server: &http.Server{
				Addr: ":8001",
				Handler: userRouter,
			},
		}

		go srv.ListenAndServe()  // goroutine 2
		<- srv.StopChan() // goroutine 0
		fmt.Println("Stopped")
	}
}
```

可以看到, 我们用 `goroutine + 数字` 来表示所注释的代码会在哪个goroutine里执行, main函数我们假设是在goroutine 0里执行.

- 在 goroutine 1处, 支出一个goroutine, 监听 `sigs` 这个channel, 如果有, 就把闭包变量(来自main函数)设置为true, 从而会导致
下面的for循环不会进行下一次循环.
- 在for循环中,我们首先新建graceful的server, 然后支出一个goroutine 2去开始提供服务
- 在下一行, goroutine 0中, 这个goroutine会被阻塞在这个channel上, 直到channel中有值可以消费

而追踪进去就会发现, `srv.StopChan()` 是一个确保 graceful Server正确初始化 `srv.stopChan` 的函数. 搜索stopChan, 我们可以看到

```go
func (srv *Server) shutdown(shutdown chan chan struct{}, kill chan struct{}) {
	// Request done notification
	done := make(chan struct{})
	shutdown <- done

	srv.stopLock.Lock()
	defer srv.stopLock.Unlock()
	if srv.Timeout > 0 {
		select {
		case <-done:
		case <-time.After(srv.Timeout):
			close(kill)
		}
	} else {
		<-done
	}
	// Close the stopChan to wake up any blocked goroutines.
	srv.chanLock.Lock()
	if srv.stopChan != nil {
		close(srv.stopChan)
	}
	srv.chanLock.Unlock()
}
```

在调用shutdown之后, 就会关闭这个channel, 然后上面所说的for循环就会重新初始化. 于是似乎就实现了一次 "平滑重启". 为什么打
引号呢? 因为在关闭服务器端的监听和下一次for循环重新执行到 `srv.ListenAndServe()` 之间的这一段时间间隙, 很有可能会有新的
连接到来却因为服务器端没有监听而连接失败. 所以这个实现和我们直接执行 `sudo systemctl restart nginx` 是类似的.

更详细的traefik源码阅读与分析我会另外再写一篇博客来分析, 这里就此打住. 接下来来看一下简单的在TCP层面实现平滑重启的服务器.

## TCP平滑重启

首先我们来看看怎么起一个TCP服务器:

```go
package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	conn.Write([]byte("hello"))
	conn.Close()
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(ln.Addr())
	}
	for {
		if conn, err := ln.Accept(); err == nil {
			fmt.Println("new conn...")
			go handleConnection(conn)
		}
	}
}
```

为了测试,我们写一个Python脚本(Python果然还是更加简洁):

```python
import socket


def foo():
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.connect(("127.0.0.1", 8080))
    s.close()


if __name__ == "__main__":
    while True:
        foo()
```

执行之后就可以看到输出.

我们知道多个进程是不可以监听在同一个(IP地址,端口号)对上的, 即,不能对同一对(IP地址,端口号)
执行多次listen函数,我们可以做个实验,把ListenAndServe抽出来,起另外一个goroutine去执行,为了方便
区分,我们加入一个参数,就是goroutine的名字:

```go
package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	conn.Write([]byte("hello"))
	conn.Close()
}

func ListenAndServe(name string) {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(ln.Addr())
	}

	fmt.Println(name)
	for {
		if conn, err := ln.Accept(); err == nil {
			go handleConnection(conn)
		}
	}
}

func main() {
	go ListenAndServe("server1")
	ListenAndServe("server2")
}
```

执行一下:

```bash
$ go run t.go
[::]:8080
server2
listen tcp :8080: bind: address already in use
server1
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x20 pc=0x4dc9a4]
```

但是我们可以在同一个socket对上, 共享同一个监听套接字地址, 然后在多个goroutine中执行accept函数:

```go
package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	conn.Write([]byte("hello"))
	conn.Close()
}

func ListenAndServe(ln net.Listener, name string) {
	for {
		if conn, err := ln.Accept(); err == nil {
			fmt.Println(name)
			go handleConnection(conn)
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(ln.Addr())
	}

	go ListenAndServe(ln, "server1")
	ListenAndServe(ln, "server2")
}
```

但是这还远远不是平滑重启, 只是证明了平滑重启是可行的, 毕竟平滑重启的前提就是在父子进程中能够
共享同一个套接字, 而且在不同的地方可以进行 `accept` 操作. 接下来我们来看一下怎么fork, 然后带上
socket套接字的文件描述符, 然后再在子进程中重新把套接字描述符还原成 `tcp.Listener`:

- 先来看怎么把套接字转换成文件描述符, 传递给另外一个goroutine, 然后这个goroutine还原成listener:

```go
package main

import (
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) {
	conn.Write([]byte("hello"))
	conn.Close()
}

func listenAndServe(ln net.Listener, name string) {
	for {
		if conn, err := ln.Accept(); err == nil {
			fmt.Println(name)
			go handleConnection(conn)
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(ln.Addr())
	}
	l := ln.(*net.TCPListener)
	newFile, _ := l.File()
	fmt.Println(newFile.Fd())

	anotherListener, _ := net.FileListener(newFile)

	go listenAndServe(anotherListener, "listener 1")
	listenAndServe(ln, "listener 2")
}
```

接下来我们要在go中进行fork并且传递文件描述符, 查看了文档, 可以通过 `exec.Cmd` 里的 `ExtraFiles []*os.File` 来传递:

```go
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
)

var (
	graceful = flag.Bool("graceful", false, "-graceful")
)

func handleConnection(conn net.Conn) {
	conn.Write([]byte("hello"))
	conn.Close()
}

func listenAndServe(ln net.Listener, name string) {
	for {
		if conn, err := ln.Accept(); err == nil {
			fmt.Println(name)
			go handleConnection(conn)
		}
	}
}

func gracefulRestart() {
	ln, err := net.FileListener(os.NewFile(3, "graceful server"))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(ln)
	}

	listenAndServe(ln, "graceful server")
}

func main() {
	flag.Parse()
	fmt.Printf("given args: %t\n", *graceful)

	if *graceful {
		gracefulRestart()
	} else {
		ln, err := net.Listen("tcp", ":8080")
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(ln.Addr())
		}
		l := ln.(*net.TCPListener)
		newFD, _ := l.File()
		fmt.Println(newFD.Fd())

		cmd := exec.Command(os.Args[0], "-graceful")
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		cmd.ExtraFiles = []*os.File{newFD}
		cmd.Run()
	}
}
```

当然我们还可以做的更好, 例如让graceful server支持再次graceful restart, 于是代码变成了这样:

```go

ckage main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var (
	graceful = flag.Bool("graceful", false, "-graceful")
)

// Accepted accepted connection
type Accepted struct {
	conn net.Conn
	err  error
}

func handleConnection(conn net.Conn) {
	conn.Write([]byte("hello"))
	conn.Close()
}

func listenAndServe(ln net.Listener, sig chan os.Signal) {
	accepted := make(chan Accepted, 1)
	go func() {
		for {
			conn, err := ln.Accept()
			accepted <- Accepted{conn, err}
		}
	}()

	for {
		select {
		case a := <-accepted:
			if a.err == nil {
				fmt.Println("handle connection")
				go handleConnection(a.conn)
			}
		case _ = <-sig:
			fmt.Println("gonna fork and run")
			forkAndRun(ln)
			break
		}
	}
}

func gracefulListener() net.Listener {
	ln, err := net.FileListener(os.NewFile(3, "graceful server"))
	if err != nil {
		fmt.Println(err)
	}

	return ln
}

func firstBootListener() net.Listener {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
	}

	return ln
}

func forkAndRun(ln net.Listener) {
	l := ln.(*net.TCPListener)
	newFile, _ := l.File()
	fmt.Println(newFile.Fd())

	cmd := exec.Command(os.Args[0], "-graceful")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.ExtraFiles = []*os.File{newFile}
	cmd.Run()
}

func main() {
	flag.Parse()
	fmt.Printf("given args: %t, pid: %d\n", *graceful, os.Getpid())
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)

	var ln net.Listener
	if *graceful {
		ln = gracefulListener()
	} else {
		ln = firstBootListener()
	}

	listenAndServe(ln, c)
}
```
