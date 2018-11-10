# Golang的反射

反射是指，在程序运行的时候，可以动态的检查自身的类型，结构和其他信息。

要了解Go语言的反射，我们需要了解两个概念。

- `reflect.Type`
- `reflect.Value`

我们想要在运行时知道类型和值，怎么办呢？很明显就是编译器帮我们把这些信息存下来在程序的某个地方。
Type就是类型，Value就是值。reflect包里有几个重要的函数，需要了解一下：

- `reflect.TypeOf` 获取所传入值的类型，返回的是 `reflect.Type`。而Type又分为很多种Kind，Kind有这么几种取值：

```go
const (
        Invalid Kind = iota
        Bool
        Int
        Int8
        Int16
        Int32
        Int64
        Uint
        Uint8
        Uint16
        Uint32
        Uint64
        Uintptr
        Float32
        Float64
        Complex64
        Complex128
        Array
        Chan
        Func
        Interface
        Map
        Ptr
        Slice
        String
        Struct
        UnsafePointer
)
```

有点晕是不是？多看几遍就不晕了。

- `reflect.ValueOf` 返回所传入值的值，但是类型是 `reflect.Value`。根据 `reflect.Value` 可以拿到它的 Type，使用 `Type()`：

```go
package main

import (
	"fmt"
	"reflect"
)

func main() {
	v := reflect.ValueOf(3)

	fmt.Printf("TypeOf: %s\n", v.Type())
}
```

- `Value.CanSet` 返回这个Value是不是能被更新值。设想，如果我们想从环境变量里，把环境变量里设置的值刷到一个结构体里，就需要
用到这个函数了。注意，只有变量才能被更新，什么是变量呢？就是，我们可以拿到它在内存中的地址，我们就可以更改值的意思。看看下面的例子：

```
x := 2 // 2 是常量，x是变量，x存的值是2。x是可以被更新的。
a := reflect.ValueOf(2) // a 不是变量，因为a代表的值是2，是一个常量。
b := reflect.ValueOf(x) // b 不是变量，因为b代表的值是x包含的值，也是2，是一个常量。
c := reflect.ValueOf(&x) // c 不是变量，因为c代表的值是x的地址，例如0x11223344，也是一个常量。
d := c.Elem() // d是变量，因为c是x的地址的值，c.Elem() 是x地址所指向的那个值，我们知道了x的地址，自然就可以更新x了。
```

有点难理解对不对？难就对了，因为这个要对内存有一定的了解，如果实在看不懂，就还是去看看基础的书吧。我目前一句两句话也解释不清楚。
如果想理解的话，推荐：

- 《深入理解计算机系统》
- 《编程范式》

------

- https://blog.golang.org/laws-of-reflection
- 《The Go Programming Language》
- https://golang.org/pkg/reflect
