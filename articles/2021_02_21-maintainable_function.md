# 编写可维护的函数

之所以写这篇，是因为看到 Dave Cheney 的一篇博客，深有同感：
https://dave.cheney.net/2019/09/24/be-wary-of-functions-which-take-several-parameters-of-the-same-type 。

日常开发中很有可能会遇到这样的傻bug：参数写反了。Go/Kotlin等强类型语言中，当连续几个参数的类型是一样时，
类型检查就没法帮我们避坑，而Python中则更是如此，无论传什么类型进去都可以，type hint在实际项目中的使用还不够广泛，
再加上Python和Kotlin都有默认参数。

我们先来回顾一下这三中语言中的此类问题。

- Go:

```go
package main

func foo(a, b int) {
	println(a, b)
}

func main() {
	foo(1, 2)
}
```

> Go语言中无法辨别两个同类型的参数传递的先后顺序。但是Go不支持keyword arguments，所以少了一些问题。

- Python:

```python
def foo(first, second, third=3, fourth=4):
    print(first, second, third, fourth)


foo(5, 4, 5)
```

> Python过于灵活，keyword arguments也可以用普通参数的形式传进去。

- Kotlin:

```kotlin
fun foo(first: Int, second: Int, third: Int=3, fourth: Int=4) {
    println("$first, $second, $third, $fourth")
}

fun main(args: Array<String>) {
    foo(1, 2, 4)
}
```

> Kotlin这点与Python一致。

这还只是一个简单的例子，当我们遇到一些函数比如 `copy`，`memmove` 这样的参数，则必须要打开manual来看看函数的签名，
Dave给了一种解决方案：

```go
type Source string

func (src Source) CopyTo(dest string) error {
	return CopyFile(dest, string(src))
}

func main() {
	var from Source = "presentation.md"
	from.CopyTo("/tmp/backup")
}
```

这是一种方案，但是如果都这么写的话，似乎又有点麻烦。我个人的习惯是在设计此类函数时，一定要保证签名的统一性，例如
在Go Web项目中，我们经常会需要返回成功或者失败，通常来说成功我们需要定制的内容只有返回结果，而失败时，我们则希望
定制返回状态码，原因和结果。如果是这样，那么就容易造成 `Success` 和 `Fail` 两个函数的签名不一致，导致每次使用时，
都可能要去查一查怎么用。所以我通常会为了可维护性而牺牲一下，多写一点代码，定义如下：

```go
package handlers

import (
	"github.com/gin-gonic/gin"
)

func Fail(c *gin.Context, code int, msg string, result interface{}) {
	c.JSON(code, gin.H{"code": code, "msg": msg, "result": result})
}

func Success(c *gin.Context, code int, msg string, result interface{}) {
	c.JSON(code, gin.H{"code": code, "msg": msg, "result": result})
}
```

这样子无论是成功时，还是失败时，签名顺序都是一样的，当然，如果在Python或者Kotlin这种支持keyword arguments的语言中，
可以给 `code` 和 `msg` 一个默认值，这样子就只有在需要定制的时候，才需要单独传参数。例如：

```python
def success(result, code=200, msg=""):
    pass


def fail(result, code=400, msg=""):
    pass
```

最后，不犯此类错误的一个帮手就是代码补全提供的类型，一定要看清楚然后再写(如果项目中统一了我所说的签名统一性，那么就
可以在这一点上提升效率了)。
