# 当JS遇到uint64：JS整数溢出问题

最近遇到一个问题，就是传了一个超级大的uint64，结果前端发现数字对不上，精度丢失了！原因就在于，JS的Number是
"64-bit floating point IEEE 754 number"。最大能表示的数是 `Number.MAX_SAFE_INTEGER`，一般来说，就是：

- `2 ** 53 - 1`, 或者
- `+/- 9,007,199,254,740,991`, 或者
- nine quadrillion seven trillion one hundred ninety-nine billion two hundred
fifty-four million seven hundred forty thousand nine hundred ninety-one

总之，就是如果表示整数，最多能表示到 `2 ** 53 - 1`。解决方案，就是用string来表示，对于GIN来说，就是在定义的struct
中，uint64的tag里，加一个 `string`，比如如果原来是 `json:"id"`，现在就是 `json:"id,string"`：

```go
package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

type Demo struct {
	ID uint64 `json:"id,string"`
}

func bigNumHandler(c *gin.Context) {
	d := Demo{}
	if err := c.BindJSON(&d); err != nil {
		log.Printf("error: %s", err.Error())
		c.JSON(400, d)
		return
	}

	d.ID -= 2
	c.JSON(200, d)
}

func main() {
	route := gin.Default()
	route.POST("/big", bigNumHandler)
	route.Run(":8085")
}
```

```bash
$ http :8085/big id=18446744073709551615
HTTP/1.1 200 OK
Content-Length: 29
Content-Type: application/json; charset=utf-8
Date: Tue, 22 Jun 2021 02:51:10 GMT

{
    "id": "18446744073709551613"
}


$ http :8085/big id='18446744073709551615'
HTTP/1.1 200 OK
Content-Length: 29
Content-Type: application/json; charset=utf-8
Date: Tue, 22 Jun 2021 02:51:13 GMT

{
    "id": "18446744073709551613"
}
```

这个坑，比较难遇上，你得同时遇到：

- 刚好你用了 uint64/int64
- 刚好你的值大于 `2 ** 53`

以上就是这个坑的总结。

---

- https://stackoverflow.com/questions/307179/what-is-javascripts-highest-integer-value-that-a-number-can-go-to-without-losin
- https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Number/MAX_SAFE_INTEGER
