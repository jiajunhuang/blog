# Golang的short variable declaration

Go中，常规声明方式为

```golang
var i, j int
```

也有一种短的方式

```golang
i, j := 1, 2
```

这种方式相当于

```golang
var i, j int
i, j = 1, 2
```

但是短的方式允许重复声明，条件是必须有一个以上重复

```golang
i, j := 1, 2
z, j := 3, 4
i, j := 5, 6 // 报错！
```

我们来看一段代码：

```golang
package main

import (
	"fmt"
	"os"
)

var cwd string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("error = %s\n", err)
	}

	fmt.Printf("init: cwd = %s\n", cwd)
}

func main() {
	fmt.Printf("main: cwd = %s\n", cwd)
}
```

运行结果：

```bash
jiajun@debian $ go run fun.go
init: cwd = /home/jiajun/test
main: cwd =
```

为什么cwd明明已经声明成了全局变量却没有被没改变呢？

> https://golang.org/ref/spec#Short_variable_declarations

Unlike regular variable declarations, a short variable declaration may
redeclare variables provided they were originally declared earlier in
the same block (or the parameter lists if the block is the function body)
with the same type, and at least one of the non-blank variables is new.

所以上面的代码想要能正常运行就得：

```golang
package main

import (
	"fmt"
	"os"
)

var cwd string

func init() {
    var err error
	cwd, err = os.Getwd()
	if err != nil {
		fmt.Printf("error = %s\n", err)
	}

	fmt.Printf("init: cwd = %s\n", cwd)
}

func main() {
	fmt.Printf("main: cwd = %s\n", cwd)
}
```
