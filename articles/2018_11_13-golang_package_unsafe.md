# Go的unsafe包

常规的Go代码是可以跨平台的，因为常规代码只关心使用哪种数据结构，而不关心数据结构的内部实现，内部实现细节由编译器处理。

而 `unsafe` 这个包就是用来访问内部实现细节的。所以，使用了unsafe这个包的代码，因为Go的编译器实现细节改变，很有可能会造成兼容性问题。

## 三个函数

`unsafe` 主要提供了三个函数：

- `Alignof`: 输出给定类型内存对齐的大小
- `Offsetof`: 输出给定结构体具体属性相对于结构体其实内存位置的偏移量
- `Sizeof`: 输出给定类型所占内存的大小

看示例：

```go
package main

import (
	"fmt"
	"unsafe"
)

type Demo struct {
	s  string
	i  int
	f  float64
	bs []byte
}

func main() {
	d := Demo{}

	fmt.Println("Alignof:")
	fmt.Println(unsafe.Alignof(d.s))
	fmt.Println(unsafe.Alignof(d.i))
	fmt.Println(unsafe.Alignof(d.f))
	fmt.Println(unsafe.Alignof(d.bs))

	fmt.Println("Offsetof:")
	fmt.Println(unsafe.Offsetof(d.s))
	fmt.Println(unsafe.Offsetof(d.i))
	fmt.Println(unsafe.Offsetof(d.f))
	fmt.Println(unsafe.Offsetof(d.bs))

	fmt.Println("Sizeof:")
	fmt.Println(unsafe.Sizeof(d.s))
	fmt.Println(unsafe.Sizeof(d.i))
	fmt.Println(unsafe.Sizeof(d.f))
	fmt.Println(unsafe.Sizeof(d.bs))
}
```

执行结果：

```bash
$ go run main.go
Alignof:
8
8
8
8
Offsetof:
0
16
24
32
Sizeof:
16
8
8
24
```

## `unsafe.Pointer`

这个类型相当于C语言里的 `void *`。举个例子，如果想把 `[]byte` 转换成 string 或者反之，该怎么做呢？看看fasthttp里的一段代码：

```go
// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// s2b converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func s2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}
```

------

- https://golang.org/pkg/unsafe/
- https://en.wikipedia.org/wiki/Data_structure_alignment
- https://github.com/valyala/fasthttp/blob/fcaab424cac756cafb79fb3c08b5a1bc6b7d63e7/bytesconv.go#L378:6
