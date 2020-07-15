# Go设计模式：装饰器模式和访问者模式

今天我们来介绍两个设计模式，一个是老朋友，decorator，第二个就是访问者模式。如果使用过python的话，我想decorator都
不用我介绍了，我们来看个例子：

```python
import functools


def foo():
    print("=== foo ===")


if __name__ == "__main__":
    foo()
```

如果我们想在执行foo函数之前或者之后做点什么事情，就可以用上decorator（比如在web开发中，我们就经常有这种需求，例如在一个
请求的开始，我们初始化一个事务，在请求结束之后，我们尝试提交或者回滚事务）：

```python
import functools


def with_tx(func):
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        print("=== start tx ===")
        result = func(*args, **kwargs)
        print("=== commit tx ===")
        return result

    return wrapper


@with_tx
def foo():
    print("=== foo ===")


if __name__ == "__main__":
    foo()
```

执行一下：

```bash
$ python main.py 
=== start tx ===
=== foo ===
=== commit tx ===
```

我们先来介绍Python中decorator的语法糖，为什么在使用 `with_tx` 之后，我们仍然能以 `foo` 的名义来调用被装饰过的函数呢？
我们来把 `with_tx` 这个函数拆城几部分来看：

```python
def with_tx(func):  # 定义 with_tx 函数，这个函数接收一个函数作为参数
    @functools.wraps(func)  # functools.wraps 的作用，是把传入的func的文档等资料，放到wrapper函数里，它也是一个decorator
    def wrapper(*args, **kwargs):  # 定义一个闭包函数，闭包函数可以使用外层的变量，因此也就可以使用func。
        print("=== start tx ===")  # 闭包函数决定何时调用被包装的函数，比如我们这里先print，再调用
        result = func(*args, **kwargs)
        print("=== commit tx ===")
        return result

    return wrapper  # 把闭包函数返回
```

所以，我们可以得出这么几个结论：

- `with_tx` 接受一个函数作为参数，同时它返回一个函数
- `with_tx` 内的闭包函数最后是被返回的，它的实现决定了何时调用被包装的函数
- 使用 `@with_tx` 之后里面返回的函数，最后却是以 `foo` 的函数名调用，其实是因为这相当于，把返回的wrapper函数直接
重新赋值给 foo 函数，相当于 `foo = with_tx(foo)`

这就是装饰器模式，一种不改变原有代码，却能增加点功能的设计模式。

## Go语言的装饰器模式

在了解Python中的decorator模式之后，我们再来看Go语言如何实现装饰器模式，就很简单了：

```go
package main

import (
	"fmt"
)

type Decoer func(i int, s string) bool

func foo(i int, s string) bool {
	fmt.Printf("=== foo ===\n")
	return true
}

func withTx(fn Decoer) Decoer {
	return func(i int, s string) bool {
		fmt.Printf("=== start tx ===\n")
		result := fn(i, s)
		fmt.Printf("=== commit tx ===\n")

		return result
	}
}

func main() {
	foo := withTx(foo)
	foo(1, "hello")
}
```

由于Go没有Python中那样的语法糖，因此只能手动重新赋值给同名的变量。

这就是装饰器模式。
