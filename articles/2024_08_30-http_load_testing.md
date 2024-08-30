# HTTP 压力测试中的 Coordinated Omission

HTTP负载测试工具有很多，常见的有 ApacheBench，也就是 ab，还有 wrk、vegata等。在了解这些工具时，我了解到一个新的概念：

Coordinated Omission，来看 GPT 的解释：

> "Coordinated Omission"是性能测试中一个重要的概念。在性能测试中，我们常常需要测量系统的延迟，例如一个请求从发送到收到响应所需的时间。然而，如果在系统延迟的情况下，我们暂停发送新的请求，那么我们测量的延迟就会低于实际值。这就是所谓的"Coordinated Omission"（协同遗漏）。简单来说，"Coordinated Omission"是一种测试偏差，它发生在以下情况：当系统延迟时，我们暂停发送新的请求，等待系统恢复后再发送。这样，我们就忽略了在系统延迟期间应该发送但没有发送的请求。因此，我们测量的延迟结果会低于实际值。例如，假设我们每秒发送10个请求，每个请求的预期延迟是100毫秒。然而，在某一时刻，由于系统延迟，一个请求花了200毫秒才得到响应。如果我们在这200毫秒内暂停发送新的请求，那么我们就会忽略掉这200毫秒内应该发送但没有发送的1个请求。因此，我们测量的平均延迟会低于实际值。"Coordinated Omission"的问题在于，它会导致我们低估系统在高负载情况下的真实性能。因此，在进行性能测试时，我们需要采取措施避免"Coordinated Omission"，以获得更准确的测试结果。

这个问题，会导致我们低估系统在高负载情况下的真实性能。举个简单的例子来理解，如果我们实现一个压测工具，采用的是
for循环来实现，其中一个请求耗时较长导致后续请求被阻塞，那么这个工具就会有 Coordinated Omission 问题。

wrk 就有 Coordinated Omission 的问题，因此有人提出了 wrk2 来解决这个问题。不过由于要自己编译，还挺麻烦的，我选择
使用 vegata 来进行压测(vegata可以携带 payload 进行请求)。

## vegata 使用示例

首先我们用Go语言写一个简单的HTTP服务：

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.POST("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
```

然后我们用 vegata 来进行压测：

```bash
$ echo "POST http://localhost:8080/" | vegeta attack -body post.json -rate=10000 -duration=30s | vegeta report
Requests      [total, rate, throughput]         300000, 10000.39, 10000.35
Duration      [total, attack, wait]             29.999s, 29.999s, 118.062µs
Latencies     [min, mean, 50, 90, 95, 99, max]  49.042µs, 152.644µs, 137.14µs, 181.872µs, 250.288µs, 466.057µs, 3.108ms
Bytes In      [total, mean]                     5400000, 18.00
Bytes Out     [total, mean]                     7500000, 25.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:300000
Error Set:
```

这样就可以看到持续30s，每秒10000个请求的情况了。
