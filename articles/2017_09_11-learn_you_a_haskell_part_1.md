# Haskell简明教程（一）：从递归说起

> 这一系列是我学习 `Learn You a Haskell For Great Good` 之后，总结，编写的学习笔记。

这个系列主要分为五个部分：

- [从递归说起](#)
- [从命令式语言进行抽象](./2017_09_17-learn_you_a_haskell_part_2.md.html)
- [Haskell初步：语法](./2017_09_18-learn_you_a_haskell_part_3.md.html)
- [Haskell进阶：Monoid, Applicative, Monad](./2017_09_25-learn_you_a_haskell_part_4.md.html)
- [实战：Haskell和JSON](./2017_09_26-learn_you_a_haskell_part_5.md.html)

虽然我们最终的目标是初窥 `Haskell` ，但是就我本人的学习经历来说，日常是学习命令式
编程为主，相信绝大部分同学也是一样，而不是先学会了函数式编程之后才开始学习命令式编程。

所以这一系列教程最开始会从命令式语言进行切入，逐渐过度到 `Haskell` 这一函数式语言。
最开始讲述例子所用的语言包括但不限于 `Python`，`Golang`，使用这两门语言的原因很简单，
因为我目前比较熟悉的语言是这两门 :)

此外讲解的时候会用一些基本的，简单地数据结构，为什么不一如往常，从现实业务切入
慢慢慢慢抽象到这些呢？因为：如上所说，现实业务一步步抽象之后，就能转化成数据结构，
此外，如果我们再加上这样一步的话，这个系列怕是要再长不少，我更愿意把这些独立成一个
新的系列来讲，如果大家有兴趣的话，可以在最下面留言，或者发邮件给我，或者其他方式
告知我 :)

## 先不用递归

我们来看看列表，或者数组，或者切片(slice)，在命令式语言里我们要怎么遍历。

```go
package main

import (
	"fmt"
)

func main() {
	var simpleArray = [3]int{1, 2, 3}

	for i, v := range simpleArray {
		fmt.Printf("index: %d, value: %d\n", i, v)
	}
}
```

输出：

```bash
$ go run t.go 
index: 0, value: 1
index: 1, value: 2
index: 2, value: 3
```

我们的遍历是，从左往右，一个一个来，数组里的元素就像是一个一个连着的多米诺骨牌。

好，我们按下不表，再看一个例子，在二叉查找树上找子节点。

```go
package main

import (
	"fmt"
)

type Node struct {
	v      int
	lchild *Node
	rchild *Node
}

type Tree *Node

func build(array []int) Tree {
	if len(array) < 1 {
		return nil
	}
	var t Tree = &Node{array[0], nil, nil}

	for _, v := range array[1:len(array)] {
		insert(t, v)
	}

	return t
}

func insert(t *Node, v int) {
	var cursor = t
	for {
		if v <= cursor.v {
			if cursor.lchild == nil {
				cursor.lchild = &Node{v, nil, nil}
				return
			}
			cursor = cursor.lchild
		} else {
			if cursor.rchild == nil {
				cursor.rchild = &Node{v, nil, nil}
				return
			}
			cursor = cursor.rchild
		}
	}
}

func query(t *Node, v int) (found bool, n *Node) {
	var cursor = t

	for cursor.lchild != nil && cursor.rchild != nil {
		if cursor.v == v {
			return true, cursor
		} else if v < cursor.v {
			cursor = cursor.lchild
		} else {
			cursor = cursor.rchild
		}
	}

	return false, nil
}

func main() {
	var t = build([]int{7, 3, 6, 8, 1})

	found, addr := query(t, 3)
	fmt.Printf("found? %t, addr: %v\n", found, addr)
	found, addr = query(t, 10)
	fmt.Printf("found? %t, addr: %v\n", found, addr)
}
```

我们是如何查找子节点的呢？如果发现当前值和要查找的值相同，那么就返回，如果
要查找的值比当前值更小，就往左走，否则往右走。

平时我们编程就是这样，仔细的检查每一个边界情况，然后控制下一步怎么走。这就是所谓
的命令式编程，我们描述的是每一步该怎么走。

## 初探递归

在[Thinking Recursively](https://jiajunhuang.com/articles/2015_09_05-thinking_recursively.rst.html)
中，简略的介绍了一下递归。

我们现在换一种角度来看上面的问题。什么是递归呢，我们来看看维基百科的解释，
https://en.wikipedia.org/wiki/Recursion_(computer_science)。

`Recursion in computer science is a method where the solution to a problem
depends on solutions to smaller instances of the same problem (as opposed to iteration).`

将问题分解成相同的子问题。相同，也就是说，我们有一个大的箱子，现在我们把它变成
无数个小箱子。

我们先来看看列表该怎么抽象。遵循上面的规则，我们把列表拆成更小的相同的子问题。
也就是说我们可以把列表拆成1个节点和剩下的列表，也可以把列表拆成n个节点。没错，这
都是抽象，我们先来看看后面这种，当我们把列表拆成n个节点，怎么进行遍历呢？没错，
其实就是从左往右一个一个来，好像回到了上一节。那如果我们把粒度放大呢？两个两个？
试想一下。其实还是一样，得进行遍历。但是有一种特殊情况我们不用遍历，试想，我们
把列表拆成一个子节点，和一个列表。会是怎样？例如 `[1, 2, 3, 4, 5]` 我们拆成
`1` 和 `[2, 3, 4, 5]`。首先我们打印 `1`，然后我们处理后面的这个列表。

似乎可行，用 `Golang` 试试。

```bash
package main

import (
	"fmt"
)

func printArray(array []int) {
	fmt.Printf("%d", array[0])
	printArray(array[1:len(array)])
}

func main() {
	var array = []int{6, 5, 4, 7, 8, 9}
	printArray(array)
}
```

```
$ go run t.go 
654789panic: runtime error: index out of range

goroutine 1 [running]:
main.printArray(0xc420043f68, 0x0, 0x0)
	/home/jiajun/tests/t.go:8 +0xec
main.printArray(0xc420043f68, 0x1, 0x1)
	/home/jiajun/tests/t.go:9 +0xdd
main.printArray(0xc420043f60, 0x2, 0x2)
	/home/jiajun/tests/t.go:9 +0xdd
main.printArray(0xc420043f58, 0x3, 0x3)
	/home/jiajun/tests/t.go:9 +0xdd
main.printArray(0xc420043f50, 0x4, 0x4)
	/home/jiajun/tests/t.go:9 +0xdd
main.printArray(0xc420043f48, 0x5, 0x5)
	/home/jiajun/tests/t.go:9 +0xdd
main.printArray(0xc420043f40, 0x6, 0x6)
	/home/jiajun/tests/t.go:9 +0xdd
main.main()
	/home/jiajun/tests/t.go:14 +0x5c
exit status 2
```

为啥会这样呢？我们来模拟一下程序运行时的情况，首先 `[6, 5, 4, 7, 8, 9]` 执行
printArray时，传入 `[6, 5, 4, 7, 8, 9]`，打印 6，然后调用 printArray，传入的
是 `[5, 4, 7, 8, 9]` ... 一直到 最后只剩下 `[9]` 的时候，接下来调用 printArray
传入 `[]`，然后调用 `array[0]` 结果就panic了。

所以递归我们需要判断一下特殊条件。如果传入的是空数组，就啥也不干，退出。

```go
package main

import (
	"fmt"
)

func printArray(array []int) {
	if len(array) == 0 {
		return
	}

	fmt.Printf("%d", array[0])
	printArray(array[1:len(array)])
}

func main() {
	var array = []int{6, 5, 4, 7, 8, 9}
	printArray(array)
}
```

这样就好了。

```bash
$ go run t.go
654789
```

那二叉查找树该怎么抽象呢？同样，就是本节点和左边的树，右边的树。这就留作思考吧 :)

## 总结

递归是什么呢？分拆成相同的子问题，把剔出来的那一部分解决之后，再去解决子问题。
通过这一篇，我们从实际代码看命令式编程是怎样一步一步操作，然后跳过来从另一个
角度看，了解了什么是递归。下一篇，我们继续看命令式编程，看我们如何从实际业务
代码脱身，进行抽象。下一篇我们讲述一个比较简单的问题，就是移动端推送(虽然我
已经讲过好几遍了，但这个还真是一个用来讲抽象的好例子)。
