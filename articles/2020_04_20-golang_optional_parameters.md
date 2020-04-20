# Golang的可选参数实践

Golang处理可选参数+默认参数的时候，常见的代码是这样的：

```go
type Queue struct {
	Name string
}

func NewQueue(name string) *Queue {
	return &Queue{name}
}
```

然而随着项目的发展，这个参数可能会变多：

```go
type Queue struct {
	Name     string
	MaxLimit int
}
```

那么这个时候要如何处理呢？有几种方案：

- 破坏兼容性：使用 `func NewQueue(name string, maxLimit int) *Queue` 的方式
- 不破坏兼容性，增加 `func NewQueueWithLimit(name string, maxLimit int) *Queue`，从而保留原来的方式
- 破坏兼容性，使用一个 `Config` 来保存配置，就变成了 `func NewQueue(config *QueueConfig) *Queue`，但是这种方式有一个
很大的缺点，就是不好处理零值，零值到底是零值，还是填的刚好是零值？我该不该用默认参数替代之？

因此就演变出下面这种方式，在我看来是一个很好的解决方案：

```go
type Queue struct {
	Name     string
	MaxLimit int

	// monitor
	MonitorInterval int
}

type QueueOption func(*Queue)

func WithMaxLimit(max int) QueueOption {
	return func(q *Queue) {
		q.MaxLimit = max
	}
}

func WithMonitorInterval(seconds int) QueueOption {
	return func(q *Queue) {
		q.MonitorInterval = seconds
	}
}

func NewQueue(name string, options ...QueueOption) *Queue {
	queue := &Queue{name, 10, 5}

	for _, o := range options {
		o(queue)
	}

	return queue
}
```

使用起来，就可以这样：`NewQueue("abcd", WithMaxLimit(10))`，然而后面的 With 函数是可选的，如果不选，就会使用我们的默认值，
这样做的好处是不会破坏兼容性，以后如果增加了更多选项，那么增加一个 With函数即可。

不过天下没有免费的午餐，这种方式的弊端也很明显，就是比较费键盘 --- 每次新增一个选项时都要新增加一个函数。

---

参考资料：

- https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
- https://www.sohamkamani.com/golang/options-pattern/
