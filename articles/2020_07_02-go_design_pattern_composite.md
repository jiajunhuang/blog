# Go设计模式：composite模式

典型的composite模式，是这样的，对于多个对象，由于我们只需要其中一部分共同的操作，因此我们可以通过定义一个父类，来规定
我们所需要的操作，却并不管具体每个子类到底是什么样的。看下维基百科的定义：

```
The composite pattern describes a group of objects that are treated the same way as a single instance
of the same type of object. The intent of a composite is to "compose" objects into tree structures to
represent part-whole hierarchies. Implementing the composite pattern lets clients treat individual
objects and compositions uniformly.
```

这里说明了几个问题：

- 我们会把一组对象当作同样地类型，也就是说，我们并不在乎它是什么类的实例，我们只在乎有什么操作
- 通常会用树状结构来表示，对应到编程语言，其实就是使用继承的方式

我们来看Python的例子，仍然以发短信为例子：

```python
import abc


class SMSSender:
    @abc.abstractmethod
    def send(self, user, message):
        raise NotImplementedError()


class AliyunSender(SMSSender):
    def send(self, user, message):
        print("使用aliyun向{}发送信息{}")

    def report(self):
        print("blabla")


class TencentSender(SMSSender):
    def send(self, user, message):
        print("使用tencent向{}发送信息{}")

    def shutdown(self):
        print("shutdown")
```

瞧，AliyunSender和TencentSender分别继承了SMSSender，他们都实现了 `send` 方法，而且他们还各自都有自己的不同的子类方法，
但是我们使用的时候并不在乎，因为我们只关心是否实现了 `send` 方法。

诶？等等，我们这不是讲的Go的设计模式吗？我们来看下Go语言里面如何实现。通常我们有两种方式，一种其实就是上面的翻版：

```go
package main

import (
	"fmt"
)

type Sender struct{}

func (s *Sender) Send(user, message string) {
	panic("not implemented")
}

type AliyunSender struct {
	Sender
}

func (a *AliyunSender) Send(user, message string) {
	fmt.Printf("使用aliyun向%s发送信息%s\n", user, message)
}

type TencentSender struct {
	Sender
}

func (t *TencentSender) Send(user, message string) {
	fmt.Printf("使用aliyun向%s发送信息%s\n", user, message)
}
```

但是这在Go语言里，并不是最优解，更多的，我们会使用接口。

```go
package main

import (
	"fmt"
)

type Sender interface {
	Send(user, message string)
}

type AliyunSender struct {
	Sender
}

func (a *AliyunSender) Send(user, message string) {
	fmt.Printf("使用aliyun向%s发送信息%s\n", user, message)
}

type TencentSender struct {
	Sender
}

func (t *TencentSender) Send(user, message string) {
	fmt.Printf("使用aliyun向%s发送信息%s\n", user, message)
}

var (
	_ Sender = &AliyunSender{}
	_ Sender = &TencentSender{}
)
```

composite 的核心并不是一定要用树状模式(也就是对应编程语言继承)来表示，而是说我们只关心是否实现了接口，并不关心它具体是啥。
这不就是Go里面接口的用法么？

这就是Composite模式在Go语言里的应用。实际项目中，composite模式可以用于递归的表示某些东西的情况下，比如文件系统、窗口系统
等大量共同属性、操作的情况下。

---

ref:

- https://en.wikipedia.org/wiki/Composite_pattern
