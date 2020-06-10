# Go设计模式：模板模式

模板模式，大家猜猜是什么？其实我们天天用，离我们最近的就是面向对象里的继承。模板的作用是什么呢？是框定一个框架，但是
不填充具体细节，细节由具体的子类来完成，举个例子，如果我们想写一个短信系统，他要对接各种第三方短信，但是其实大家
都知道，发短信，大家接口不一样而已，实际上他们可以抽象成相同的流程：

```go
package main

import (
	"fmt"
)

type SMSSender struct{}

func (s *SMSSender) Send(content string, receivers []string) {
	panic("not implement yet")
}

type AliyunSMS struct {
	SMSSender
}

func (a *AliyunSMS) Send(content string, receivers []string) {
	fmt.Printf("阿里云发短信的各种API调用")
}

type TencentSMS struct {
	SMSSender
}

func (t *TencentSMS) Send(content string, receivers []string) {
	fmt.Printf("腾讯云发短信的各种API调用")
}
```

瞧，他们就像一颗倒过来的树：

```
                SMSSender

            /                \
        AliyunSMS          TencentSMS
```

`SMSSender` 规定了有这么一个方法，而 `AliyunSMS` 和 `TencentSMS` 才是细节实现者，这就是传统的面向对象范式里的继承。

所以模板模式并不是什么新奇玩意儿，原来只是我们的老朋友(上述例子中，Go语言里，更好的实现是使用interface)。
