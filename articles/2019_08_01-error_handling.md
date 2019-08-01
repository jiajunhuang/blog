# C, Go, Python的错误处理和异常机制杂谈

很多语言都混用了错误和异常，本文中，我们对错误和异常进行定义：

- 错误(error) 是不可恢复的，只能通过修复程序或者输入输出来修正的
- 异常(exception)是可以恢复的，可以通过捕捉异常来重新让运行中的程序继续运行

但是，很多语言由于各种原因，混用了上述两个概念。

- C语言中没有异常，所有的异常，都是通过错误来展示(errno)，根据判断errno之后决定如何处理
- Go语言与C一脉相承，虽然有异常(panic+recover)，但是一般不会在业务代码中使用，而是使用错误(例如 `io.EOF`)
- Python有错误和异常两个概念，但是一般错误只用于例如语法错误等这类，而异常则是代码中常用的处理方式

我们来看看三种语言是如何处理系统调用的错误的。

### C语言

对于系统调用出错来说，是这样的流程(直接引用《FreeBSD设计与实现》上的原文)：

Eventually, the system call returns to the calling process, either successfully or unsuccessfully.
On the PC architecture, success or failure is returned as the carry bit in the user process’s
program status longword: If it is zero, the return was successful; otherwise, it was unsuccessful.
On many machines, return values of C functions are passed back through a general-purpose
register (for the PC, data register EAX). The routines in the kernel that implement system calls
return the values that are normally associated with the global variable errno. After a system call,
the kernel system-call handler leaves this value in the register. If the system call failed, a C
library routine moves that value into errno, and sets the return register to -1. The calling process
is expected to notice the value of the return register, and then to examine errno. The mechanism
involving the carry bit and the global variable errno exists for historical reasons derived from
the PDP-11.

简单来说，就是系统调用返回时，会放一个错误码在寄存器上，进入到用户空间的代码之后(通常来说也就是glibc或者
其它C标准库里的系统调用包装代码里)，会对错误码进行检查。如果是0，那么没有报错，否则，就根据错误码，
设置errno，然后这个系统调用的包装函数返回-1。

### Go

对于Go来说，情况与C类似，不过是替换成Go内置的错误，例如 `io.EOF`。

### Python

Python的官方解释器是使用C写的，因此，错误返回时，已经能够获取errno来判断错误，而Python的系统调用错误异常便是
建立在这个基础之上，检查errno，如果有错误，那么则抛出对应的异常。

### 错误判断和抛出异常的区别

对于C和Go的开发者来说，每一次调用都要判断一下返回是否出错，这就是Go语言饱受诟病的 `if err != nil {}` 的来源。
这种方式需要逐层处理，例如如果想把错误抛到上一层，就得加一行 `if err != nil {return err}`，而Python的异常
则可以自动跨越多个函数调用抛出。

### 错误日志和处理

什么时候应该记录日志？如果只使用Go标准库中的 `errors` 库，那么很麻烦的一个地方在于，`errors` 无法包含错误发生时
的信息，因此古老的方式是发生错误的地方打一行日志：

```go
if err != nil {
    log.Errorf("blablabla, error blablabla")
    return err
}
```

这样层级多了之后就容易发生错误日志多次重复，非常容易被错误日志埋没。

我们可以参考Python中的处理方式：

```python
if err:
    raise SomeException("exception context")
```

在异常中包含错误发生时的错误信息(或者说上下文)，然后在上层统一进行处理。

当然，Go语言中，Go 2草案对此进行了改进，社区中也有Go 1.X 的实现：https://github.com/pkg/errors:

```go
_, err := ioutil.ReadAll(r)
if err != nil {
        return errors.Wrap(err, "read failed")
}
```

---

参考资料：

- FreeBSD设计与实现：https://book.douban.com/subject/25856073/
- https://docs.python.org/3/tutorial/errors.html
- https://stackoverflow.com/questions/912334/differences-between-exception-and-error
- https://wiki.haskell.org/Error_vs._Exception
