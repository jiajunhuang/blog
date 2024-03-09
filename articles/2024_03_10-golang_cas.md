# Golang CAS 操作是怎么实现的

在Go语言中，CAS(Compare and Swap) 操作一般都是通过 `atomic` 操作来实现的，我们来探究一下底层是怎么实现的。

我们以 `CompareAndSwapInt32` 为例，首先找到源码，位于 `doc.go`：

```go
// CompareAndSwapInt32 executes the compare-and-swap operation for an int32 value.
// Consider using the more ergonomic and less error-prone [Int32.CompareAndSwap] instead.
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool)
```

在Go语言中，这种只有函数声明，没有函数实现的，通常意味着函数的实现在Go汇编中。在 `asm.s` 中可以找到定义：

```go
TEXT ·CompareAndSwapInt32(SB),NOSPLIT,$0
	JMP	runtime∕internal∕atomic·Cas(SB)
```

然后跟着路径找 `internal/atomic/atomic_amd64.s`：

```go
// bool Cas(int32 *val, int32 old, int32 new)
// Atomically:
//	if(*val == old){
//		*val = new;
//		return 1;
//	} else
//		return 0;
TEXT ·Cas(SB),NOSPLIT,$0-17
	MOVQ	ptr+0(FP), BX
	MOVL	old+8(FP), AX
	MOVL	new+12(FP), CX
	LOCK
	CMPXCHGL	CX, 0(BX)
	SETEQ	ret+16(FP)
	RET
```

我们来看下这段汇编代码的意思，以下解释来自ChatGPT：

```go
TEXT ·Cas(SB),NOSPLIT,$0-17
```
这行代码定义了一个名为 `Cas` 的函数，`SB` 是一个汇编器符号，表示当前包的起始地址。
`NOSPLIT` 表示该函数不会发生栈的分裂，`$0-17` 表示函数没有输入参数，但有 17 个字节的输出参数。

```go
MOVQ	ptr+0(FP), BX
```
这行代码将函数的第一个输入参数 `ptr` 的值加载到寄存器 `BX` 中。

```go
MOVL	old+8(FP), AX
```
这行代码将函数的第二个输入参数 `old` 的值加载到寄存器 `AX` 中。

```go
MOVL	new+12(FP), CX
```
这行代码将函数的第三个输入参数 `new` 的值加载到寄存器 `CX` 中。

```go
LOCK
```
这行代码是一个前缀指令，用于告诉处理器后面的指令是原子操作，需要获取总线锁。

```go
CMPXCHGL	CX, 0(BX)
```
这行代码使用 CMPXCHG 指令进行比较和交换操作。它比较内存地址 `0(BX)` 处的值与寄存器 `CX` 的值是否相等，如果相等，则将寄存器 `CX` 的值写入内存地址 `0(BX)` 中。

```go
SETEQ	ret+16(FP)
```
这行代码根据 CMPXCHG 指令的结果设置标志位，如果比较和交换成功，则将标志位设置为 1，否则设置为 0。

```go
RET
```
这行代码表示函数的返回。

总的来说，这段汇编代码实现了一个 CAS 操作，它比较内存地址中的值与期望值是否相等，如果相等，则将新的值写入内存地址中，并返回操作是否成功。这个 CAS 操作使用了 CMPXCHG 指令来实现原子的比较和交换操作。

其实说到底，Golang的CAS操作，就是依靠 `LOCK` + `CMPXCHGL` 来实现的，其他的语言也是一样的，需要依赖CPU提供这种底层
能力，才能够真正的做到 CAS。
