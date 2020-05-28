# Python中的并发控制

最近有个需求，要大批量写数据，项目是用Python写的。所以有个问题就是如何使用Python做并发控制，如果是Go语言，那么就可以使用
[WaitGroup](https://golang.org/pkg/sync/#WaitGroup)：

```go
package main

import (
	"sync"
)

type httpPkg struct{}

func (httpPkg) Get(url string) {}

var http httpPkg

func main() {
	var wg sync.WaitGroup
	var urls = []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}
	for _, url := range urls {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Launch a goroutine to fetch the URL.
		go func(url string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			http.Get(url)
		}(url)
	}
	// Wait for all HTTP fetches to complete.
	wg.Wait()
}
```

但是这样还只能做到等待所有任务完成后再退出，但是并发控制就得用别的方法，比如使用一个channel或者是加锁。但是Python总是有
非常成熟的方案：concurrent.futures.Executor。

Executor有两个具体实现：ThreadPoolExecutor 和 ProcessPoolExecutor，分别是线程池和进程池。用法都是一样的，都符合
`Executor` 的接口定义。我们来看实际例子：

```python
from concurrent.futures import ThreadPoolExecutor
import time


def thread_pool_executor_callback(worker):
    if worker.exception():
        logging.exception("worker %s got exception", worker)


def loooooong_task(i):
    print("task %s sleeping..." % i)
    time.sleep(10)
    print("task %s done..." % i)


with ThreadPoolExecutor(max_workers=2) as executor:
    for i in range(10):
        executor.submit(loooooong_task, i).add_done_callback(thread_pool_executor_callback)
```

是不是非常简单，这就是Python的魔力所在。其中 `add_done_callback` 是用来处理异常的一个回调函数，如果不弄这个的话，
发生异常以后，ThreadPoolExecutor 不会打印出异常，而是直接执行别的任务。

---

附：另外还可以参考参考 [gevent的pool](http://www.gevent.org/api/gevent.pool.html)

---

参考资料：

- [官方文档](https://docs.python.org/3.8/library/concurrent.futures.html#concurrent.futures.Executor)
- [Gevent Pool](http://www.gevent.org/api/gevent.pool.html)
- [Golang WaitGroup](https://golang.org/pkg/sync/#WaitGroup)
