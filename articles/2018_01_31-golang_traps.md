# Golang的一些坑

- 传给 `signal.Notify` 的channel必须是一个buffered channel, 否则收不到信号

- channel默认是unbuffered channel, 因此在没有消费者之前, 放入channel的动作都会被阻塞, 例如:

```go
func main() {
    c := make(chan int)

    for i := 0; i < 3; i++ {
        go func() {
            c <- 1
        }()
    }

    fmt.Println(<-c)
}
```

此函数退出时,会有两个goroutine被阻塞在channel上, 然而gc不会回收. 因此, 如果大量出现这种情况, 将会导致goroutine leak.

- `for...range` 语句会在执行之前执行一次。而且所有值都是拷贝，而非返回指针。

    - https://golang.org/ref/spec#For_statements
    - https://github.com/golang/go/wiki/Range
    - https://garbagecollected.org/2017/02/22/go-range-loop-internals/
    - https://github.com/golang/gofrontend/blob/master/go/statements.cc#L5343

- channel略慢，为何？

```go
type hchan struct {
	qcount   uint           // total data in the queue。总数
	dataqsiz uint           // size of the circular queue。make(chan, 3)中的3.大小。也是下面的buf的大小
	buf      unsafe.Pointer // points to an array of dataqsiz elements
	elemsize uint16         // 每个元素有多大
	closed   uint32
	elemtype *_type // element type 元素是啥类型
	sendx    uint   // send index 发送的序号
	recvx    uint   // receive index 接收的序号
	recvq    waitq  // list of recv waiters 等待接收的G。sudog链表。
	sendq    waitq  // list of send waiters 等待发送的G。sudog链表。

	// lock protects all fields in hchan, as well as several
	// fields in sudogs blocked on this channel.
	//
	// Do not change another G's status while holding this lock
	// (in particular, do not ready a G), as this can deadlock
	// with stack shrinking.
	// 加锁加锁。不用atomic的原因大概是chan的缓冲数量，等待发送数量和接受者数量都不定的吧。
    // 加锁简单好用，代价就是性能略差
	lock mutex
}
```

操作channel都要加锁。所以略慢。

- slice是共享底层数据的，为何？

```go
// slice的结构体，一个指针，一个长度，一个容量
type slice struct {
	array unsafe.Pointer
	len   int
	cap   int
}
```

因为结构体就是这么定义的 :)

从fasthttp里学到一招避免 `string` 转 `[]byte` 开销的方式：

```go
package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"unsafe"
)

// s2b converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func s2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")
	hello, _ := reader.ReadString('\n')
	fmt.Println(hello)
	helloBytes := s2b(hello)
	helloBytes[1] = 'w'
	fmt.Println(hello)
}
```

执行一下：

```bash
$ ./tests 
Enter text: hello
hello
hwllo
```

不过要注意不能直接写 `hello := "hello"`这样来改，因为如果显式写明字符串的话，
编译器会把它放在栈里，而不是像上面的代码一样在堆里搞。
