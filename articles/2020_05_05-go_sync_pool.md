# Golang sync.Pool源码阅读与分析

Go的很多地方都有用到 sync.Pool，这是作为一个内存池来使用的。例如 `fmt.Printf`：

```go
// These routines end in 'f' and take a format string.

// Fprintf formats according to a format specifier and writes to w.
// It returns the number of bytes written and any write error encountered.
func Fprintf(w io.Writer, format string, a ...interface{}) (n int, err error) {
	p := newPrinter()
	p.doPrintf(format, a)
	n, err = w.Write(p.buf)
	p.free()
	return
}
```

其中的 `p := newPrinter()` 就是这样的：

```go
func newPrinter() *pp {
	p := ppFree.Get().(*pp)
	p.panicking = false
	p.erroring = false
	p.wrapErrs = false
	p.fmt.init(&p.buf)
	return p
}
```

我们来看看 `sync.Pool` 的基本用法：

```go
package main

import (
	"bytes"
	"fmt"
	"sync"
)

var (
	// 声明一个全局变量（或者局部变量也可以）用于存储内存池
	bytesPool = sync.Pool{
		New: func() interface{} { return bytes.Buffer{} },
	}
)

// NewBufferFromPool new bytes.Buffer from sync.Pool
func NewBufferFromPool() bytes.Buffer {
	return bytesPool.Get().(bytes.Buffer) // 通过Get来获得一个
}

// NewBuffer return new bytes.Buffer
func NewBuffer() bytes.Buffer {
	return bytes.Buffer{}
}

func main() {
	a := NewBuffer()
	b := NewBufferFromPool()
	fmt.Printf("%b, %b\n", a, b)
	bytesPool.Put(b)
}
```

由此可见 sync.Pool 的基本用法。

## 源码分析

我们来看看 sync.Pool 是怎么实现的：

```go
type Pool struct {
	noCopy noCopy

	local     unsafe.Pointer // local fixed-size per-P pool, actual type is [P]poolLocal
	localSize uintptr        // size of the local array

	victim     unsafe.Pointer // local from previous cycle
	victimSize uintptr        // size of victims array

	// New optionally specifies a function to generate
	// a value when Get would otherwise return nil.
	// It may not be changed concurrently with calls to Get.
	New func() interface{}
}

// 看看 poolLocal 的结构
type poolLocal struct {
	poolLocalInternal

	// Prevents false sharing on widespread platforms with
	// 128 mod (cache line size) = 0 .
	pad [128 - unsafe.Sizeof(poolLocalInternal{})%128]byte
}

// 看看 poolLocalInternal 的结构
// Local per-P Pool appendix.
type poolLocalInternal struct {
	private interface{} // Can be used only by the respective P.
	shared  poolChain   // Local P can pushHead/popHead; any P can popTail.
}
```

然后我们来看看 `Get` 是怎么工作的：

```go
// Get selects an arbitrary item from the Pool, removes it from the
// Pool, and returns it to the caller.
// Get may choose to ignore the pool and treat it as empty.
// Callers should not assume any relation between values passed to Put and
// the values returned by Get.
//
// If Get would otherwise return nil and p.New is non-nil, Get returns
// the result of calling p.New.
func (p *Pool) Get() interface{} {
	if race.Enabled {
		race.Disable()
	}
	l, pid := p.pin()
	x := l.private
	l.private = nil
	if x == nil {
		// Try to pop the head of the local shard. We prefer
		// the head over the tail for temporal locality of
		// reuse.
		x, _ = l.shared.popHead()
		if x == nil {
			x = p.getSlow(pid)
		}
	}
	runtime_procUnpin()
	if race.Enabled {
		race.Enable()
		if x != nil {
			race.Acquire(poolRaceAddr(x))
		}
	}
	if x == nil && p.New != nil {
		x = p.New()
	}
	return x
}
```

这一段的主要作用就是，优先从当前执行的Processor里
获取[参考Golang的GMP](https://jiajunhuang.com/articles/2018_02_02-golang_runtime.md.html)，如果没有的话，
就从共享池子里拿，如果还是没有的话，就调用 `getSlow` 里拿，再不行的话，就调用 `New` 函数了。

我们来看看上述例子的跑分：

```go
$ cat main_test.go 
package main

import (
	"testing"
)

func BenchmarkNewBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewBuffer()
		_ = p
	}
}

func BenchmarkNewBufferFromPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewBufferFromPool()
		bytesPool.Put(p)
	}
}
$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/jiajunhuang/test
BenchmarkNewBuffer-8           	1000000000	         1.14 ns/op
BenchmarkNewBufferFromPool-8   	10978279	       207 ns/op
PASS
ok  	github.com/jiajunhuang/test	3.645s
```

你会惊讶的发现，用了Pool比不用还要慢。为啥呢？经过我的测试发现主要是在类型转换上比较费时，如果去掉这个，就会快很多，但是
去掉类型转换之后，用了Pool还是比不用更慢，这又是为啥呢？因为bytes所使用的内存比较小，使用内存池的效果并不好：

```go
$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/jiajunhuang/test
BenchmarkNewBuffer-8           	1000000000	         0.951 ns/op
BenchmarkNewBufferFromPool-8   	43858023	        25.7 ns/op
PASS
ok  	github.com/jiajunhuang/test	2.989s
```

把代码改成使用这种大块内存的就好很多：

```go
type MyStruct struct {
	http.Request
	http.Response
	a http.Request
	b http.Request
	c http.Request
	d http.Request
}
```

效果如下：

```go
$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/jiajunhuang/test
BenchmarkNewBuffer-8           	14430404	        83.5 ns/op
BenchmarkNewBufferFromPool-8   	45602824	        32.1 ns/op
PASS
ok  	github.com/jiajunhuang/test	3.490s
```

所以，对于 sync.Pool 的使用要注意两点：

- 类型转换(type casting)很费CPU
- 对于大块的内存，使用内存池才有意义

---

参考资料：

- https://golang.org/pkg/sync/#Pool
- https://medium.com/a-journey-with-go/go-understand-the-design-of-sync-pool-2dde3024e277
