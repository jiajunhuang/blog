# 错误处理实践

一个接口，或者一个系统，总是要处理很多异常流程，通常我们都需要以某种形式表达出来，例如返回错误、抛异常。这篇文章我想
讲讲怎么去处理错误。

## 接口错误

通常对于RESTful接口，我会建议这样返回错误:

```js
{
    "code": 200000,
    "msg": "success",
    "result": {
        "users": []
    }
}
```

所有的真实payload，都会返回在 `result` 字段中。为什么要这样设计呢？我们对接的编程语言中，有动态语言，有静态语言。
动态语言处理响应结果非常方便，直接反序列化用就行，对于静态语言，我们往往需要先定义好一个结构体或者类，然后才能
把结果反序列化进对象。所以我们总是有这几个字段，对于定义结构体就会非常方便。

那么为什么 `code` 是一个比较大的数字呢？大部分场景下，其实我们都不需要使用错误码来指示错误，但是有一些场景我们又需要，
错误码通常是用来指示一个具体的业务错误，我们需要靠查表去看到底是什么错误。另外一个方面，我们其实对HTTP协议中定义的
状态码比较熟悉，比如我们大家都知道200系列是成功，400系列是客户端错误，500系列是服务端错误。因此，我在设计的时候，都会
把错误码，分配在不同的系列里，比如我们想表示用户名不对，首先我们知道，这属于客户端请求错误，因此他要分配在400系列中，
然后在这里找到第一个可用的错误码，例如 400001，依此类推。当然，HTTP状态码会和 `code` 保持一致，例如当 `code` 为
400001 时，HTTP状态码返回400。

## 程序内错误

对于程序内的错误，一般会给我们传递两个信息：1. 属于哪一类错误；2. 具体的错误信息。对于Go语言来说，我一般推荐使用标准库。
我们首先也要对错误进行分类，例如：

- ErrBadArgument
- ErrUnauthorized
- ErrNotFound
- ErrInternalError

等等，这在Python中很常见，Python自带的异常就在标准库中已经进行过分类，然后业务异常再基于他们开始派生。在Go标准库支持
error继承之后，我们也可以使用类似的方式。

此外，我们还需要将一些错误替换成上述基类错误，例如 `gorm.ErrNotFound` 替换成 `ErrNotFound`，然后我们在中间件中进行
统一的处理，将 `400` 系列的错误信息返回，将 `500` 系列的错误打印，但是替换掉敏感信息。

## 不同语言错误处理方式

编程语言都会内置错误处理，这是一门编程语言中非常重要的一部分。

- 对于C语言来说，最常见的处理方式就是用返回值来指示错误，0表示正常，-1表示出错。有时还会借助errno来表示具体错误。
- 对于Go语言来说，按照约定，如果想要返回错误，使用返回值中最后一个来返回错误。Go语言中还有 panic/recover 来处理不符合预期的错误。
- 对于Python来说，我们会使用Exception来表示错误，用 `try...except...` 来表达。Java、JS与此类似。
- 对于Haskell，使用 Maybe 来表示错误。Rust 使用类似的方式。

## 总结

这篇文章中，简单的描述了我个人对于错误处理的一些经验和总结，以及我为什么会这样设计。希望对大家有所帮助。
