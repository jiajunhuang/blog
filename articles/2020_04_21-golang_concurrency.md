# Golang中的并发控制

之前我写过一篇博客介绍 [Python中的并发控制](https://jiajunhuang.com/articles/2020_02_12-python_concurrency.md.html)。

Python的并发控制可以说很优雅，Java的也是类似的，那么，Go语言咋办？如果我也想实现类似的需求，例如：同时不超过8个goroutine
执行任务，那该咋办呢？我在写 [gotasks 这个异步任务框架](https://github.com/jiajunhuang/gotasks) 就有这种需求，因此我
把Go的并发控制抽象成如下代码，以后就可以直接使用了：

```go
package pool

type GoPool struct {
	MaxLimit int

	tokenChan chan struct{}
}

type GoPoolOption func(*GoPool)

func WithMaxLimit(max int) GoPoolOption {
	return func(gp *GoPool) {
		gp.MaxLimit = max
		gp.tokenChan = make(chan struct{}, gp.MaxLimit)

		for i := 0; i < gp.MaxLimit; i++ {
			gp.tokenChan <- struct{}{}
		}
	}
}

func NewGoPool(options ...GoPoolOption) *GoPool {
	p := &GoPool{}
	for _, o := range options {
		o(p)
	}

	return p
}

// Submit will wait a token, and then execute fn
func (gp *GoPool) Submit(fn func()) {
	token := <-gp.tokenChan // if there are no tokens, we'll block here

	go func() {
		fn()
		gp.tokenChan <- token
	}()
}

// Wait will wait all the tasks executed, and then return
func (gp *GoPool) Wait() {
	for i := 0; i < gp.MaxLimit; i++ {
		<-gp.tokenChan
	}

	close(gp.tokenChan)
}

func (gp *GoPool) size() int {
	return len(gp.tokenChan)
}
```

来看看用法：

```go
gopool := pool.NewGoPool(pool.WithMaxLimit(3))
defer gopool.Wait()

gopool.Submit(func() {//你的代码})
```

是不是也挺优美的？这里注意两点：

- `gopool.Submit` 在令牌不足时，会阻塞当前调用(因此Go runtime会执行其他不阻塞的代码)
- `gopool.Wait()` 会等到回收所有令牌之后，才返回

这样就可以实现我们的需求了，例如并发3个Goroutine执行任务。

---

参考资料：

- https://github.com/jiajunhuang/gotasks/blob/master/pool/pool.go
