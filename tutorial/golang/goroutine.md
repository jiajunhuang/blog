# Goroutine

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

Goroutine 是Go语言中的一个特色，它是在代码层级实现并发的根基。我们举个简单的例子，
首先我们看如何打印0-9，如果没有并发，我们是这样写：

```go
package main

import (
	"fmt"
)

func main() {
	for i := 0; i < 10; i++ {
		fmt.Println(i)
	}
}
```

这样的输出效果一定是从0-9，每个数字一行的，因为这里不是并发执行，而是顺序执行，即for循环，从0，然后到1，然后一直到9。

而借助Goroutine我们可以这样并发打印0-9：

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	for i := 0; i < 10; i++ {
		go fmt.Println(i)
	}

	time.Sleep(time.Second * time.Duration(1))
}
```

只需要简单的通过 `go` 这个关键字，就可以实现10个Goroutine一起执行，为什么我们这里要加 `time.Sleep` 呢？因为Goroutine里，
main函数所在的Goroutine叫做 `main goroutine`，如果这个主Goroutine退出了，那么整个Go程序就会跟着一起退出，为了把上面的代码
执行完，我们需要等待一小会儿，可以看看输出：

```bash
$ go run main.go 
0
8
9
6
2
1
3
4
5
7
```

看到了吗，他们的输出顺序是不一定的，每一次执行，他们的输出结果都不会一样。

> 注意，`go` 关键字后只能接一个函数，可以是有名字的函数，也可以是匿名函数。

> 注意，Go语言的函数是支持闭包的，也就是说，函数里的函数，内层的函数是可以使用外层函数的变量的。

> 扩展阅读：如果对协程(coroutine)有兴趣的话，可以阅读 [协程(coroutine)简介 - 什么是协程？](https://jiajunhuang.com/articles/2018_04_03-coroutine.md.html)

---

- 上一篇：[指针](./pointers.md)
- 下一篇：[Channel](./channel.md)
