# Golang context源码阅读与分析

Golang中使用context作为goroutine之间的控制器，例如：

```go
package main

import (
	"context"
	"log"
	"time"
)

func UseContext(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("context is done with error %s", ctx.Err())
			return
		default:
			log.Printf("nothing just loop...")
			time.Sleep(time.Second * time.Duration(1))
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go UseContext(ctx)

	time.Sleep(time.Second * time.Duration(1))
	cancel()
	time.Sleep(time.Second * time.Duration(2))
}
```

这样就可以在 `main` 函数里告知 `UseContext` 所在的goroutine，主函数已经想要退出了。当然，并不是什么很神奇的方法，
也就是相当于main函数里传递一个变量，而UseContext里不断的去检查变量而已，只不过在Golang里，我们使用的是chanel来实现。

来看看 `context` 是啥，原来是个接口：

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}
```

既然是接口，那么我们就得来找一个具体的实现来看看，比如 `context.Background()`：

```go
func Background() Context {
	return background
}

// but what is background? it's:
var (
	background = new(emptyCtx)
	todo       = new(emptyCtx)
)
```

原来是一个 `emptyCtx` 实例。那 `WithCancel()` 是咋实现的呢？

```go
// WithCancel returns a copy of parent with a new Done channel. The returned
// context's Done channel is closed when the returned cancel function is called
// or when the parent context's Done channel is closed, whichever happens first.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func WithCancel(parent Context) (ctx Context, cancel CancelFunc) {
	c := newCancelCtx(parent)
	propagateCancel(parent, &c)
	return &c, func() { c.cancel(true, Canceled) }
}

// A cancelCtx can be canceled. When canceled, it also cancels any children
// that implement canceler.
type cancelCtx struct {
	Context

	mu       sync.Mutex            // protects following fields
	done     chan struct{}         // created lazily, closed by first cancel call
	children map[canceler]struct{} // set to nil by the first cancel call
	err      error                 // set to non-nil by the first cancel call
}
```

可以看出来，done是一个channel，我们在 `UseContext` 里，通过select来检测ctx是否已经关闭，其实就是看done这个channel
是否关闭了。不信可以看看 `cancel` 的实现：

```go
// cancel closes c.done, cancels each of c's children, and, if
// removeFromParent is true, removes c from its parent's children.
func (c *cancelCtx) cancel(removeFromParent bool, err error) {
	if err == nil {
		panic("context: internal error: missing cancel error")
	}
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return // already canceled
	}
	c.err = err
	if c.done == nil {
		c.done = closedchan
	} else {
		close(c.done) // NOTE(jiajun): here it is :)
	}
	for child := range c.children {
		// NOTE: acquiring the child's lock while holding parent's lock.
		child.cancel(false, err)
	}
	c.children = nil
	c.mu.Unlock()

	if removeFromParent {
		removeChild(c.Context, c)
	}
}
```

瞧，调用cancel的时候就是调用了 `close(c.done)`，然后下一次循环的时候，`UseContext` 就能检测到ctx已经关闭了。

## 总结

这篇博客里我们看到了context是怎么实现的，读者如果有兴趣的话，可以看看Context中的传值和取值是怎么实现的，以及思考一下
是否有更好的实现方案。
