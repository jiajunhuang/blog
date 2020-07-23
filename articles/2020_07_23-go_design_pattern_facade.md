# Go设计模式：facade模式和观察者模式

## facade模式(外观模式)

初看这个名字，觉得很陌生。但是我说这个模式，其实就是抽象封装，那么我想它就不陌生了。facade模式的作用就是，原本随着项目
变得越来越大，代码之间可能会有一些顺序，如果把他们封装起来，那么对于调用者而言，就只需要调用一个函数，而并不需要知道
这个函数里面，具体做了什么，有什么顺序依赖。

## observer模式(观察者模式)

观察者模式的主要特征是大家都关注一个特定信息，当这个信息发生改变时，所有人都会收到通知。最明显的例子，就是Redis中的
pub/sub。我们来看看Go如何实现观察者模式：

```go
package main

import (
	"fmt"
)

type Observer interface {
	Update(event string)
}

type EventSource struct {
	observers []Observer
}

func (e *EventSource) AddObserver(o Observer) {
	e.observers = append(e.observers, o)
}

func (e *EventSource) Publish(event string) {
	for _, o := range e.observers {
		o.Update(event)
	}
}

type AObserver struct{}

func (a *AObserver) Update(event string) {
	fmt.Printf("A received event %s\n", event)
}

type BObserver struct{}

func (b *BObserver) Update(event string) {
	fmt.Printf("B received event %s\n", event)
}

type CObserver struct{}

func (c *CObserver) Update(event string) {
	fmt.Printf("C received event %s\n", event)
}

func main() {
	eventSource := EventSource{}
	eventSource.AddObserver(&AObserver{})
	eventSource.AddObserver(&BObserver{})
	eventSource.AddObserver(&CObserver{})

	eventSource.Publish("whoops")
}
```

执行一下：

```bash
$ go run main.go 
A received event whoops
B received event whoops
C received event whoops
```

## 总结

这一次我们介绍了两个设计模式，外观模式和观察者模式。外观模式其实就是日常编程中的抽象封装，观察者模式，则是大家都去订阅
一个消息的更新，当有更新发生时，会及时通知订阅者。
