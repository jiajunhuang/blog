# Go语言的defer, panic和recover

> 本文源自 https://blog.golang.org/defer-panic-and-recover

## defer

defer语句在Go的函数中属于一个栈的形式，即在函数运行完成时，根据该语句出现的顺序
后进先出进行调用。从流程上来讲，类似于如下python代码：

```python
def defer(func):
    def wrapper(*args, **kwargs):
        defers = []
        kwargs['defers'] = defers
        func(*args, **kwargs)
        for f in reversed(defers):
            f()
    return wrapper


@defer
def foo(defers=None):
    defers.append(lambda: print("1st defered function in foo"))
    print("hello")
    defers.append(lambda: print("2st defered function in foo"))
    print("world")
    defers.append(lambda: print("3st defered function in foo"))


if __name__ == "__main__":
    foo()
```

运行结果：

```bash
root@arch tests: python defer.py
hello
world
3st defered function in foo
2st defered function in foo
1st defered function in foo
```

但是实际上Go中的defer和上面的代码有很大不同。Go的defer有三个特点：

- 被defer的函数的参数值在defer产生的那一行代码处就确定了
- defer的函数的执行顺序为后进先出----最后derfer的函数最先执行，它们都在函数
返回了值之后开始
- defer的函数可以修改函数返回值

我们来看一个例子：

```go
package main

import "fmt"


func test_defer() {
    var i int = 1
    defer fmt.Println("1st in test_defer, i = ", i)
    fmt.Println("hello")
    i += 1
    defer fmt.Println("1st in test_defer, i = ", i)
    fmt.Println("world")
    i += 1
    defer fmt.Println("1st in test_defer, i = ", i)
}


func main() {
    test_defer()
}
```

运行结果：

```bash
root@arch tests: go run defer.go
hello
world
1st in test_defer, i =  3
1st in test_defer, i =  2
1st in test_defer, i =  1
```

由此可证第一第二点。

接下来我们再看一个例子：

```go
package main

import "fmt"


func test_defer() (i int){
    defer func () {
        fmt.Println(i)
        i++
        fmt.Println(i)
    } ()
    return 1
}


func main() {
    i := test_defer()
    fmt.Println(i)
}
```

运行结果：

```bash
root@arch tests: go run defer.go
1
2
2
```

由此可以看出，函数的执行过程是，执行函数体，然后到return语句，后进先出执行defer
语句，最后返回函数值。

## panic, recover

他们俩都是内置函数，和Python中的raise和 `try...except...` 类似。值得注意的是，
即便函数在运行过程中发生了panic，也会执行完被defer的函数。

```go
package main

import "fmt"


func caller() {
    defer func () {
        fmt.Println("defered anonymous function in caller")
    } ()
    callee()

    if r := recover(); r != nil {
        fmt.Println("recovered from callee")
    }
}


func callee() {
    panic(1111)
}


func main() {
    caller()
}
```

运行结果：

```bash
root@arch tests: go run panic.go
defered anonymous function in caller
panic: 1111

goroutine 1 [running]:
panic(0x48a0a0, 0xc420056190)
        /usr/lib/go/src/runtime/panic.go:500 +0x1a1
main.callee()
        /root/tests/panic.go:19 +0x61
main.caller()
        /root/tests/panic.go:10 +0x4e
main.main()
        /root/tests/panic.go:24 +0x14
exit status 2
```

可以看出，在panic那一行，该函数立即退出，整个函数也立即退出，但是利用defer的
特性，我们可以做到异常处理：

```go
package main

import "fmt"


func caller() {
    defer func () {
        fmt.Println("defered anonymous function in caller")
        if r := recover(); r != nil {
            fmt.Println("recovered from callee")
        }
    } ()
    callee()
}


func callee() {
    panic(1111)
}


func main() {
    caller()
}
```

运行结果：

```bash
root@arch tests: go run panic.go
defered anonymous function in caller
recovered from callee
```
