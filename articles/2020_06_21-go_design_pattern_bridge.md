# Go设计模式：桥接模式和策略模式

桥接模式在日常编码中还是经常用到的，这也是我比较喜欢的一个设计模式。从定义上来说，桥接模式比较晦涩难懂：

“桥接模式是软件设计模式中最复杂的模式之一，它把事物对象和其具体行为、具体特征分离开来，使它们可以各自独立的变化。”。

它的重要特征是，将一个实例(具体实现A)传入到另外一个类(类B)，供初始化，当类B的实例B1调用某些方法时，实际上调用的是
具体实现A的某些方法，听起来就很绕对不对，我们来看一个简单地例子(Python)：

```python
class EatApple:
    def eat(self):
        print("eating apple...")


class Eat:
    def __init__(self, sth):
        self.__sth = sth

    def eat(self):
        self.__sth.eat()


if __name__ == "__main__":
    Eat(EatApple()).eat()
```

看这个例子，这里就使用了桥接模式，分别是用Eat和EatApple，Eat抽象了吃这个动作，但是具体怎么吃，吃什么是由EatApple来实现
的，而要使用Eat类，你必须在初始化的时候传入一个EatApple，这样就把动作和实现分离了，这就是桥接模式。那为啥说在Go语言里，
经常可以看到桥接模式的身影呢？你是不是经常看到类似的代码？

```go
package main

import (
	"fmt"
)

type Set struct {
	impl map[string]bool
}

func NewSet() *Set {
	return &Set{make(map[string]bool)}
}

func (s *Set) Add(key string) {
	s.impl[key] = true
}

func (s *Set) Iter(f func(key string)) {
	for key := range s.impl {
		f(key)
	}
}

func main() {
	s := NewSet()
	s.Add("hello")
	s.Add("world")
	s.Iter(func(key string) {
		fmt.Printf("key: %s\n", key)
	})
}
```

我们在Set这个结构体里，封装了其他底层实现，Set规定了它所提供的方法，但是底层具体的实现，确是由 `map[string]bool` 真正
提供的，但是Set的使用者并不知道这个事实，因此对调用者而言，Set实现提供了功能，但是没有暴露底层实现。

将功能和实现隔离开来，有什么好处呢？好处就在于解耦，如果我们想要把Set改成并发版本的，那么我们将代码改为如下即可：

```go
package main

import (
	"fmt"
	"sync"
)

type Set struct {
	impl sync.Map
}

func NewSet() *Set {
	return &Set{sync.Map{}}
}

func (s *Set) Add(key string) {
	s.impl.Store(key, true)
}

func (s *Set) Iter(f func(key string)) {
	s.impl.Range(func(key, value interface{}) bool {
		f(key.(string))
		return true
	})
}

func main() {
	s := NewSet()
	s.Add("hello")
	s.Add("world")
	s.Iter(func(key string) {
		fmt.Printf("key: %s\n", key)
	})
}
```

瞧，我们更改了底层实现，但是 `main` 函数作为调用者，却不需要更改任何一行代码。这就是桥接模式的威力！

策略模式的代码结构和上面的例子其实很类似，当然，教科书上会说，策略模式是行为型模式，而桥接模式是结构型模式，在此我就不
咬文嚼字了。策略模式封装的，是算法实现，所以叫策略模式。
