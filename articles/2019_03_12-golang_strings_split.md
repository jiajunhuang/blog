# 一个想当然的bug

许久不写Python，`"".split(",")` 的结果我已经不能脑内运算了：

```python
In [1]: "".split(",")
Out[1]: ['']
```

我记成了 `"".split(",")` 的运算结果是 `[]`，因此，昨天在debug这段代码(Golang)的时候花了我好一会儿：

```go
$ cat main.go
package main

import (
	"fmt"
	"strings"
)

func main() {
	l := strings.Split("", ",")
	if len(l) > 0 {
		fmt.Printf("l: %+v\n", l)
	}
}
```

执行一下：

```bash
$ go run main.go
l: []
```

问题主要在于：

- 我想当然的以为 `"".split(",")` 的结果会是一个空list
- go的 `fmt.Printf` 真是一个大坑，空字符串你为啥不给我打出来？`[""]` 才是正确的打印方式呀！
