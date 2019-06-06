# 并发编程

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

如果我们没有并发，计算10次 `x++`，最终 `x` 的值会是10。

```go
package main

import (
	"fmt"
)

func main() {
    var x = 0

	for i := 0; i < 10; i++ {
		x++
	}

	fmt.Println(x)
}
```

那么什么是并发呢？我推荐大家先阅读一下：https://blog.golang.org/concurrency-is-not-parallelism

在Go中，我们可以轻松地通过使用 [Goroutine](./goroutine.md) 来实现并发，但是一旦有并发，就意味着会有数据冲突，试想，如果
有10个Goroutine同时执行 `x++` 这一行代码，最后x会等于几呢？

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	var x = 0

	for i := 0; i < 10; i++ {
		go func() {
			x++
		}()
	}

	time.Sleep(time.Second * time.Duration(1))
	fmt.Println(x)
}
```

没有人知道。这是因为虽然我们写的代码只有一行，但是我们使用的是高级语言，他最后会被编译成机器码，而CPU层面的操作实际上
是：

- 读取 `x` 的值
- 计算 `x + 1`
- 讲结果重写写入 `x`

因此，当多个人同时执行这段代码的时候，有可能出现这么一种情况，大家都算好了 `x+1`，然后大家都一起写入 `x`，如果 `x` 最开始等于１，
那么最终 `x` 的值会是2．

所以我们需要锁。锁的作用就是，把数据锁定，持有锁的人才能执行代码，其他人都是等待。

## sync

`sync` 包提供锁，是这样用的：

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var x = 0
	var l sync.Mutex

	for i := 0; i < 10; i++ {
		go func() {
			l.Lock()
			defer l.Unlock()

			x++
		}()
	}

	time.Sleep(time.Second * time.Duration(1))
	fmt.Println(x)
}
```

> 你可能注意到了 `defer` 这个陌生的关键字，它的作用就是延迟执行，和 `go` 关键字一样，他的后面要接一个函数，而和 `go` 关键字
> 不一样的是，`defer` 并不会并发执行，而是延迟执行，并且是按照调用顺序，先调用的最后执行。比如 `defer func1()`, `defer func2()`，
> `func2()` 会比 `func1()` 先执行。

上面的代码里，我们并发10个Goroutine来执行函数，函数里我们先加锁，然后才执行代码。这样就可以保证x的并发安全。

除了 `sync.Mutex`，`sync` 包还提供一种类型的锁： `RWMutex`，它是读写锁，读写锁使用起来和 `Mutex` 差不多，但是它多了两个方法：

- `RLock()`
- `RUnlock()`

写操作之间会互相阻塞，写操作和读操作之间会互相阻塞，但是读操作之间不会互相阻塞。

## sync/atomic

`atomic` 这个库提供了CPU层级的原子操作实现，主要是这么几类函数：

- Add 这类函数把目标值值与参数相加然后存储下来
- CompareAndSwap 这类函数把目标值与参数比较，如果目标值与参数中的old相等，就把目标值替换为new
- Load 读取目标值
- Store 更新目标值
- Swap 将目标值替换为new，并且同时返回old值

参考：https://golang.org/pkg/sync/atomic/#pkg-index

---

- 上一篇：[Channel](./channel.md)
- 下一篇：[go tools](./go_tool.md)
