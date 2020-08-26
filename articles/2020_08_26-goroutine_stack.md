# Goroutine是如何处理栈的？

为什么Go的栈是无限大小的？为什么Go的栈策略从 segmented stacks 切换到 contiguous stacks？逃逸分析是什么？这篇文章是我的学习记录，分别解答了上述问题。

Go在1.3以前，是使用一种叫做 segmented stacks 的栈策略。而1.3时，换成了 contiguous stacks ，我们分别来了解一下这两种策略。

## segmented stacks

在1.2之前，每一个Goroutine创建的时候，都会申请一个8KB大小的空间作为该Goroutine的栈。当8KB用完时，Go会通过 `morestack` 函数为之
再申请一块内存，然后把两个栈连起来。

当函数返回时，把新申请的那个栈释放掉。

segmented stacks有这么一个缺点，就是当一个栈快要满时，会申请一个
新的栈来执行子函数，当子函数执行完成时，会把栈回收。
如果不断重复上面这个过程，那么就会出现这个Goroutine频繁的申请和释放栈，因此产生了 "hot split" 问题。

## contiguous stacks

为了解决上述问题，Go在1.13时，切换到了新的策略，叫做 "contiguous stacks"，它的策略如下：

- 当栈不够用时，创建一个更大的栈
- 把老栈的所有内容copy到新的栈
- 调整copy过去的内容中的指针地址(把老栈的地址，改成新栈的)
- 销毁老的栈

为什么能做到第三点呢？要想做到第三点，调整地址，必须有一个先决条件，
那就是栈里的地址，只被栈内使用，而堆里没有使用。
这是为什么呢？想要调整栈内的地址调用，不算难，把栈内的地址减去老栈的起始地址，就是
它们相对于栈的起始位置的偏移量，
然后再加上新栈的起始位置，就可以得到新的内存地址。

然而，在堆里的变量，我们无法知道哪个变量引用了栈内的地址，也就无法更改它的值，所幸，Go使用了一种叫做 "逃逸分析" 的技术，避免了这个
问题。

## 逃逸分析

在一个编程语言里，变量要么在栈中传递，要么在堆中共享，对于一些只读变量，可能会分配在代码段等其它地方。此处我们只讨论栈和堆。

我们通过一个例子看看Go的逃逸分析报告：

```bash
$ cat -n main.go 
     1	package main
     2	
     3	import (
     4		"fmt"
     5	)
     6	
     7	func add(a, b int) int {
     8		result := a + b
     9		fmt.Printf("result is %v\n", &result)
    10		return result
    11	}
    12	
    13	func main() {
    14		result := add(1, 2)
    15		fmt.Printf("result is %v\n", &result)
    16	}
$ go build -gcflags "-m -l" && ./test 
# _/home/jiajun/Code/test
./main.go:8:2: moved to heap: result
./main.go:9:12: ... argument does not escape
./main.go:14:2: moved to heap: result
./main.go:15:12: ... argument does not escape
result is 0xc000014108
result is 0xc000014100
```

可以看到，第8行和第14行的result变量，分别都逃逸到了堆。

对于逃逸分析的结果，可以用这么一句话简单概括：函数内，变量如果没有被传到函数外，就没有逃逸，否则，则逃逸。对于逃逸的变量，分配到堆上，否则，分配到栈上。

上面逃逸了，是因为传给 `fmt.Printf` 时，传递了 `result` 的地址，我们改掉再来看看：

```bash
$ cat -n main.go 
     1	package main
     2	
     3	import (
     4		"fmt"
     5	)
     6	
     7	func add(a, b int) int {
     8		result := a + b
     9		return result
    10	}
    11	
    12	func main() {
    13		result := add(1, 2)
    14		fmt.Printf("result is %v\n", result)
    15	}
$ go build -gcflags "-m -l" && ./test 
# _/home/jiajun/Code/test
./main.go:14:12: ... argument does not escape
./main.go:14:13: result escapes to heap
result is 3
```

可以看到 `add` 函数内就没有发生逃逸了。

那么逃逸分析有什么用呢？通过逃逸分析，我们可以对变量的位置分配，提前进行优化，没有逃逸的分配在栈上，逃逸的在堆上。这洋酒可以减小
GC的压力，因为堆上的东西少了，GC就快了。

同时，我们还可以获得一个加成效果，那就是 contiguous stacks 对栈内
变量的要求。

逃逸分析的规则还很复杂，以上只是一个简述(比如被Go的fmt.Printf函数引用了，那就一定会逃逸等等)。

## 继续讲contiguous stacks

我们现在知道，Go通过逃逸分析技术，为我们提供了一个重要的保证：

> the only pointers to data in a stack, are in that stack themselves (there are some exceptions though). If any pointer escapes (eg. the pointer is returned to the caller) it means that the pointed data are allocated on the heap instead.

因此，第三点可以行得通，整个计划就行得通。

这样带来的好处，就是不再有 "hot split" 问题，Goroutine的栈变大了之后，
不再收缩，而是一直保持这个大小，直到Goroutine被回收。

## 结论

这篇文章中我们学习了Go的栈管理策略，了解了他们的异同，了解了逃逸分析技术，以及它为Go切换栈管理策略奠定的基础。

---

Ref:

- https://agis.io/post/contiguous-stacks-golang/
- https://docs.google.com/document/d/1wAaf1rYoM4S4gtnPh0zOlGzWtrZFQ5suE8qr2sD8uWQ/pub
- https://groups.google.com/g/golang-dev/c/i7vORoJ3XIw?pli=1
- https://medium.com/a-journey-with-go/go-how-does-the-goroutine-stack-size-evolve-447fc02085e5
- https://dave.cheney.net/2013/06/02/why-is-a-goroutines-stack-infinite
- https://golang.org/src/runtime/stack.go
- https://www.ardanlabs.com/blog/2017/05/language-mechanics-on-escape-analysis.html
- https://zhuanlan.zhihu.com/p/91559562
