# Go语言简介

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

编程语言中，我们常常会见到以下概念：

- 变量
- 常量
- 声明
- 类型声明
- 赋值
- 包
- 作用域

接下来，我们以Go语言为例，依次说明以上的概念。

## 变量

在一门编程语言中，变量是最长用的一个东西，它是用来存储一些可以发生变动的值的，因此叫做变量。

我们用一个简单地例子来讲解：

```go
package main // package 是关键字，声明一个包。main是所声明的包的名字

import ( // import 是关键字，告诉Go编译器要导入一个或多个包
    "fmt" // "fmt" 是字符串。也是我们要导入的包的名字
)

func main() { // func 是关键字，用于声明一个函数，而main则是我们声明的函数的名字
    words := "Hello World!"  // words 是一个变量，它的内容是 "Hello World!" 这个字符串

    fmt.Println(words) // fmt.Println 是调用 fmt 这个包里的 Println 函数，而括号里的就是给这个函数的参数。
}
```

在上面的例子里，每个单词都有它的含义。编程语言就是这样，通过定义一些特定的单词，来表达一些事情。就像普通话里的 "你", "我", "吃" 这些词语一样，它们都代表着一些特殊的含义。

而Go语言中，有特定含义的单词，我们叫做 "关键字"，Go语言的关键字有25个，它们分别是：

```
break default func interface select case defer go map struct
chan else goto package switch const fallthrough if range type
continue for import return var
```

正是因为这些关键字是有特殊含义的，所以我们不能滥用。除此之外，还有三类单词我们不能滥用，它们分别是：

- Go语言内置的常量：

    ```
    true false iota nil
    ```

- Go语言内置的类型：

    ```
    int int8 int16 int32 int64 uint uint8 uint16 uint32 uint64
    uintptr float32 float64 complex64 complex128 bool byte
    rune string error
    ```

- Go语言内置的函数：

    ```
    make len cap new append copy close delete complect real
    imag panic recover
    ```

这三类虽然不是关键字，但是也不应该随意使用。

## 常量

变量是可以变化的，常量就是不可以变化的。在Go语言中，我们使用 `const` 关键字声明一个常量，例如：

```go
package main

import (
    "fmt"
)

const hello = "Hello, World"

func main() {
    fmt.Println(hello)
}
```

如果尝试去修改常量，那么就会得到一个报错。这样的代码是运行不了的。

## 声明

声明是指在代码里宣布有这样一个变量名，它的类型是xxx。例如：

```go
var x int
```

对于Go而言，由于Go有类型推断，也就是说，我们可以不写是什么类型，而Go语言能够根据代码自动推断出来变量是什么类型，
所以上面的代码我们也可以这样写：

```go
var x = 1
// 或者使用Go的短声明 x := 1，注意，使用了 := 就不需要 var 关键字了。而且 := 表达式只能在函数里使用。
```

Go里面能够进行声明的关键字有4个：

- `var` 用于声明变量
- `const` 用于声明常量
- `type` 用于声明一个新的类型
- `func` 用于声明一个函数

举个例子就是这样：

```go
package main

import "fmt" // 如果只导入一个包，也可以不加括号，使用这种写法，其实多个包也可以写多行。但是一般还是使用括号，毕竟可以少些很多次 import

const boilingF = 212.0

func main() {
    var f = boilingF
    var c = (f - 32) * 5 / 9
    fmt.Printf("boiling point = %gF or %g C\n", f, c)
}
```

> 注意上面的代码中, // 开头的是注释。/* 和 */ 中间的所有字符、单词也都是注释。

## 类型声明

从上一节我们可以知道，使用 `var` 可以声明变量，`const`　声明常量，`func`　声明函数。那么 `type`　是做什么用的呢？

`type`　是用来声明我们自己的类型用的。举个例子：

```go
type MyInt int
```

这里我们就声明了一个我们自己的类型叫做 `MyInt`，只不过它的底层类型就是 `int` 罢了。因此相当于我们给 `int`　取了个别名，叫做 `MyInt`。

## 赋值

什么叫赋值呢？其实就是等号。看到等号，我们从右边往左边读，也就是，把右边的值　**赋值** 给左边的变量。

例如：

```go
var x = 1
```

这里就是，我们声明了一个变量，并且把这个变量的初始值设置成１，也就是，把１赋值给ｘ。

## 包

包，就是把一堆Go文件打包在一起，它们作为一个整体对外提供函数。

这就是我们上面所看到的 `fmt` 包和 `main` 包。虽然我们的 `main` 包只有一个函数。

此外需要注意的是，能运行的 Go程序，必须要有一个包叫 `main`，而且 `main` 包里，必须有一个函数叫做 `main`。也就是说，Go的程序，
是从 `main` 包的  `main` 函数开始执行的。

---

- 上一篇：[Hello, World](./hello_world.md)
- 下一篇：[基本类型](./basic_types.md)
