# Channel

## 目录

- [安装](./installation_linux.md)
    - [Windows](./installation_windows.md)
    - [Linux](./installation_linux.md)
    - [macOS](./installation_mac_os.md)
- [Hello, World](./hello_world.md)
- [Go语言简介](./intro.md)
- [基本类型](./basic_types.md)
- [容器类型](./composite_types.md)
- [函数](./function.md)
- [流程控制](./flow.md)
- [错误处理](./errors.md)
- [面向对象编程](./oo.md)
- [面向接口编程](./interface.md)
- [指针](./pointers.md)
- [Goroutine](./goroutine.md)
- [Channel](./channel.md)
- [并发编程](./concurrency.md)
- [go tools](./go_tool.md)

## channel

channel 是Go语言中用于Goroutine之间通信的工具，它相当于一个管道，有往channel里发送数据的发送方，也有从channel里读取数据的
消费方。发送和消费channel中的数据都是使用 `<-` 这个符号，但是它的位置放在左边还是右边就代表着不同的意思，箭头指向谁，数据
就往那个方向流动：

- `v := <-myChan` 就说明数据是从 `myChan` 里读取出来，流动到 `v` 这个变量里
- `myChan <- v` 则说明数据是从v流动到 `myChan` 里

初始化channel要使用 `make` 关键字，例如：

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	x := make(chan int)

	go func() {
		fmt.Println(<-x)
	}()

	time.Sleep(time.Second * time.Duration(1))
	x <- 1
	time.Sleep(time.Second * time.Duration(1))
}
```

`channel` 有三种形式，双向、只读、只写:

```go
package main

import (
	"fmt"
)

func main() {
	x := make(chan int)
	y := make(chan<- int)
	z := make(<-chan int)

	fmt.Printf("%T, %T, %T\n", x, y, z)
}
```

注意上面的例子里，`make(chan<- int)` 创建出来的channel只能写入数据，`make(<-chan int)` 创建出来的channel只能读取数据，他们都是
单向channel。

## select

读取channel中的数据是会被阻塞住的，那么我们要怎么才能实现如果channel中没有数据就继续执行呢？答案是使用 `select`：

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	x := make(chan int)
	go func() {
		// 这里要睡10s，但是主Goroutine没有睡10s，因此这里是执行不到的，从而主Goroutine会在select里跳到default里
		time.Sleep(time.Second * time.Duration(10))
		x <- 1
	}()

	select {
	case i := <-x:
		fmt.Printf("从x中接收到%d\n", i)
	default:
		fmt.Printf("其他分支都阻塞了，所以轮到我执行\n")
	}

	fmt.Println("退出")
}
```

`select` 语句在形式上与 `switch...case...` 是一样的，不过 `select` 用于从多个阻塞的 `goroutine` 里监听，哪个分支
最先唤醒，就执行那个分支，当然，也有 `default` 分支，用于当所有 `case` 都阻塞时，跳过这个 `select`。

---

- 上一篇：[Goroutine](./goroutine.md)
- 下一篇：[并发编程](./concurrency.md)
