# 容器类型

## 目录

- [安装](./installation_linux.md)
    - [Windows](./installation_windows.md)
    - [Linux](./installation_linux.md)
    - [macOS](./installation_mac_os.md)
- [Hello, World](./hello_world.md)
- [Go语言简介](./intro.md)
- [基本类型](./basic_types.md)
- [容器类型](./composite_types.md)
- [函数](./function.md)
- [流程控制](./flow.md)
- [错误处理](./errors.md)
- [面向对象编程](./oo.md)
- [面向接口编程](./interface.md)
- [指针](./pointers.md)
- [Goroutine](./goroutine.md)
- [Channel](./channel.md)
- [并发编程](./concurrency.md)
- [go tools](./go_tool.md)

## array

数组是一种数据结构，它是同一种类型的多个值拼在一起，而且，数组是有固定长度的，同时，数组里的每个元素都有它的下标，例如：

```
数组：1 2 3 4 5
下标：0 1 2 3 4
```

使用Go语言声明时是这样写：

```go
package main

import (
	"fmt"
)

func main() {
	var x = [5]int{1, 2, 3, 4, 5}

	fmt.Printf("%+v\n", x)
}
```

在Go语言里，数组下标(也叫index)是从0开始的。不过通常在Go语言里，我们极少使用数组，而是使用切片(slice)。

## slice

slice的底层实现就是数组，而类型声明上也是与数组极其相似，看例子：

```go
package main

import (
	"fmt"
)

func main() {
	var x = []int{1, 2, 3, 4, 5}

	fmt.Printf("%+v\n", x)
}
```

对比一下可以发现，唯一的区别就在于，数组是使用 `[5]int{xxx}` 来进行初始化，而 slice 是使用 `[]int{xxx}`。

在Go中，我们经常使用slice。slice的特点是，长度是可以改变的，也就是说，我们可以无限追加元素到slice中。其他特点slice与数组并无区别。

## map

map是哈希表，Go语言中，声明一个map是这样用：

```go
var x map[string]string
```

但是注意，上面只是说明x的类型是 `map[string]string`，但是x的值却是nil。

## struct

struct是用来把基本类型组合在一起的，举个例子，我们有个struct叫 `Person`，我们把名字，年龄组合在一起：

```go
type Person struct {
    Name string
    age int
}
```

这样我们就可以代表一个人。不知道你是否注意到了，上面的例子中，
`Name` 是大写的，而 `age` 是小写的，这有什么区别呢？

在Go语言中，大写开头的变量名是包外可以访问的，而小写的则是不可以的。还记得 `fmt.Printf` 吗？正是因为 `Printf` 是大写开头，所以我们
才能调用这个函数，如果是 `fmt.printf`，那么我们是不可以调用的。

---

- 上一篇：[基本类型](./basic_types.md)
- 下一篇：[函数](./function.md)
