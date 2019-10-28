# 栈

日常业务中的确很少用到栈这个数据结构。但是实际上，代码中无处不是栈，为什么？因为函数调用就是通过栈来实现的。

首先我们来看看栈是怎样一种结构。维基百科上这样定义：

```
In computer science, a stack is an abstract data type that serves as a collection of elements, with two principal operations:

    push, which adds an element to the collection, and
    pop, which removes the most recently added element that was not yet removed.

The order in which elements come off a stack gives rise to its alternative name, LIFO (last in, first out).
```

也就是说，栈有这样的特性：

- 栈有两个操作，一个是push，即把数据压入该数据结构；另一个是pop，即从该数据结构中弹出一个数据
- 每次执行pop操作时，得到的总是该数据结构中最后进入的一个数据。

举个例子，如果我们按照下述步骤进行操作，那么会是这样的：

- `push 1`。栈中的数据是 `1`，没有数据返回。
- `push 2`。栈中的数据是 `1, 2`，没有数据返回。
- `push 3`。栈中的数据是 `1, 2, 3`，没有数据返回。
- `pop`。栈中的数据是 `1, 2`，返回的数据是3。
- `pop`。栈中的数据是 `1`，返回的数据是2。

## 栈的实际使用

上面我们说到，栈的一个典型应用就是函数调用。我们来看看函数调用是怎么通过栈来实现的。有以下代码：

```python
def inner(foo):
    print("进入inner函数")
    print("离开inner函数")


def outter(bar):
    print("进入outter函数")
    inner(bar)
    print("离开outter函数")

if __name__ == "__main__":
    outter("hey")
```

我们来看看调用结果：

```bash
$ python stack.py
进入outter函数
进入inner函数
离开inner函数
离开outter函数
```

一个典型的函数调用过程，就是把函数调用之后要执行的指令的地址压入栈中（例如此处outter中调用inner之后的地址），然后将要
调用的函数压入栈中，随后将(部分)参数压入栈中，然后把要执行的指令的地址改为所调用的函数的地址，当这个函数执行完成之后，
便会找到此前所保存的返回地址，再接着执行原本函数中的代码。

## 总结

这一篇中我们首先介绍了栈的基本属性，然后模拟了一个栈的执行，接着我们了解了栈这个数据结构在实际编程中的应用，那就是函数调用。

---

参考资料：

- [维基百科上的Stack词条](https://en.wikipedia.org/wiki/Stack_(abstract_data_type))
