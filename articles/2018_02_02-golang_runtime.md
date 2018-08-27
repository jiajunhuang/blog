# Golang runtime 源码阅读

> 只想简单地写写，就不写的太复杂了。注释都在: https://github.com/jiajunhuang/go

- Go的编译方式是静态编译，把runtime直接编译到最终的可执行文件里。首先我们把代码考过来，然后编译出 `go` 这个可执行文件
出来。

- 编写以下代码，然后用我们自己编译出来的go来编译出一个二进制文件。注意要带调试信息并且禁止优化的，要不然不方便看。

```go
package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello world!")
}
```

```bash
$ ../bin/go build -gcflags "-N -l" -o test_demo1 demo1.go
$ gdb test_demo1
(gdb) source /home/jiajun/Code/go/src/runtime/runtime-gdb.py
Loading Go Runtime support.
(gdb) info files
Symbols from "/home/jiajun/Code/go/analysis/test_demo1".
Local exec file:
	`/home/jiajun/Code/go/analysis/test_demo1', file type elf64-x86-64.
	Entry point: 0x44fa90
	0x0000000000401000 - 0x0000000000482608 is .text
	0x0000000000483000 - 0x00000000004c4e3f is .rodata
	0x00000000004c4f60 - 0x00000000004c5ac0 is .typelink
	0x00000000004c5ac0 - 0x00000000004c5b00 is .itablink
	0x00000000004c5b00 - 0x00000000004c5b00 is .gosymtab
	0x00000000004c5b00 - 0x0000000000514042 is .gopclntab
	0x0000000000515000 - 0x0000000000521bdc is .noptrdata
	0x0000000000521be0 - 0x00000000005286f0 is .data
	0x0000000000528700 - 0x0000000000544d88 is .bss
	0x0000000000544da0 - 0x00000000005474b8 is .noptrbss
	0x0000000000400f9c - 0x0000000000401000 is .note.go.buildid
(gdb) b *0x44fa90
Breakpoint 1 at 0x44fa90: file /home/jiajun/Code/go/src/runtime/rt0_linux_amd64.s, line 8.
```

然后就跳到 `rt0_linux_amd64.s` 看。虽然没有系统的学汇编，但是边看边猜还是可以继续看下去的。

```asm
TEXT _rt0_amd64_linux(SB),NOSPLIT,$-8
	JMP	_rt0_amd64(SB)
```

然后发现 `_rt0_amd64` 继续不下去了。于是：

```bash
(gdb) b _rt0_amd64
Breakpoint 2 at 0x44c2b0: file /home/jiajun/Code/go/src/runtime/asm_amd64.s, line 15.
```

发现跳到了 `rt0_go`，不过gdb直接打断点发现打不出来，于是就在同一个文件里尝试搜索。
发现在： https://github.com/jiajunhuang/go/blob/67a58c5a2401e89fd4f688e8f70fd3be9506cea5/src/runtime/asm_amd64.s#L87

- 继续跟踪，发现有标签，最后到了 `ok` 这个标签。

```bash
(gdb) b runtime.g0
Function "runtime.g0" not defined.
Make breakpoint pending on future shared library load? (y or [n]) n
(gdb) b runtime.m0
Function "runtime.m0" not defined.
Make breakpoint pending on future shared library load? (y or [n]) n
(gdb) b runtime.check
Breakpoint 3 at 0x434890: file /home/jiajun/Code/go/src/runtime/runtime1.go, line 141.
(gdb) b runtime.args
Breakpoint 4 at 0x434340: file /home/jiajun/Code/go/src/runtime/runtime1.go, line 65.
(gdb) b runtime.osinit
Breakpoint 5 at 0x424750: file /home/jiajun/Code/go/src/runtime/os_linux.go, line 274.
(gdb) b runtime.schedinit
Breakpoint 6 at 0x428b30: file /home/jiajun/Code/go/src/runtime/proc.go, line 508.
(gdb) b runtime.mainPC
Function "runtime.mainPC" not defined.
Make breakpoint pending on future shared library load? (y or [n]) n
(gdb) b runtime.main
Breakpoint 7 at 0x427980: file /home/jiajun/Code/go/src/runtime/proc.go, line 131.
(gdb) b runtime.newproc
Breakpoint 8 at 0x42f540: file /home/jiajun/Code/go/src/runtime/proc.go, line 3321.
(gdb) b runtime.mstart
Breakpoint 9 at 0x42a920: file /home/jiajun/Code/go/src/runtime/proc.go, line 1208.
```

因此函数调用链是：

`runtime.check` -> `runtime.args` -> `runtime.osinit` -> `runtime.schedinit` -> `runtime.newproc`

最后一步里的 `runtime.newproc` 之前有把 `runtime.mainPC` 压栈。

```asm
MOVQ    $runtime·mainPC(SB), AX     // entry

// 然后再下面就有：
DATA    runtime·mainPC+0(SB)/8,$runtime·main(SB)
```

所以应该是 `runtime.mainPC` 作为入口点，由 `runtime.newproc` 来执行进入。

- 需要提前把我读到的知识剧透，方便读者理解。

    - Go的runtime中，M是 `Machine`，代表操作系统的线程
    - P是 `Processor`，意思是逻辑处理器，在最初的版本里是没有P的
    - G是 `Goroutine`。是Go中执行任务的单元，也是coroutine中的最小个体。
    - 在最初的版本里没有P，所以`M`和`G`是 M:N。历史原因不是特别了解，不过我猜测引入P的原因是，当M执行系统调用或者cgo代码
    而阻塞时，如果没有P的存在，那么该M上的所有G就无法执行。而引入P之后，可以把该M上的P摘掉，放到别的M上执行。

- coroutine又称微线程，协程，纤程等。原因是，线程是操作系统调度的最小单位，而coroutine则是用户态的 "线程"。线程的
创建，销毁，切换代价非常的高。通过线程池无法解决c10k问题，而 I/O多路复用+回调的方式写起来又比较反人类，所以有协程
这么一个东西。在用户态，以同步的方式写异步。通过某些关键字主动让出执行权限，而后等到 I/O 事件准备好时，再切换回来。

例如Python中的 `Tornado` 就通过yield+I/O多路复用回调实现了协程。gevnet则更黑一点，直接利用Python导入的机制把标准库的
代码patch。`AsyncIO` 其实差不多。不过 `Tornado` 和 `AsyncIO` 的异步代码都具有传染性，说的是例如 `@gen.coroutine` 这种
玩意儿。

所以有了Go这种，在语言层面实现异步的方式(gevent其实与此十分类似)。

- `newproc` 执行 `systemstack` 函数。这个函数的作用是在系统栈中调用给定的函数 fn。看他的注释：

```go
// systemstack runs fn on a system stack.
// If systemstack is called from the per-OS-thread (g0) stack, or
// if systemstack is called from the signal handling (gsignal) stack,
// systemstack calls fn directly and returns.
// Otherwise, systemstack is being called from the limited stack
// of an ordinary goroutine. In this case, systemstack switches
// to the per-OS-thread stack, calls fn, and switches back.
// It is common to use a func literal as the argument, in order
// to share inputs and outputs with the code around the call
// to system stack:
//
//	... set up y ...
//	systemstack(func() {
//		x = bigcall(y)
//	})
//	... use x ...
//
// systemstack如果是由g0调用，或者收到信号而调用，就会调用fn然后返回。
// 否则，systemstack切换到 per-OS-thread栈执行完fn之后，又切回去
//go:noescape
func systemstack(fn func())
```

而 `newproc` 给 `systemstack` 传的参数便是 `newproc1`。而 `newproc1` 做的事情就是新建一个Goroutine丢到队列里。而执行的fn
就是 `runtime.mainPC`，就是 `runtime.main`

- 接下来读到 `runtime.main`，没多远就执行了一个

```go
151     systemstack(func() {                                                                         
152         newm(sysmon, nil)                                                                        
153     })
```

这个sysmon是 `system monitor`，负责抢占，检查网络事件等。

然后执行

    - `runtime_init`。这个是动态生成的。
    - `gcenable` 启动gc
    - `main_init` 动态生成的。
    - `main_main` 就是我们 `main` 包里的main函数了。
    - 通过for循环确保 `&runningPanicDefers` 为0才退出。
    - `exit` 退出。

> `_init` 的函数都是动态生成的，顺序与 `import` 顺序不一定一致，但是被依赖的包的init完了才会init当前文件。

--------------------------------------------------------------------------------------------------------------

似乎到这里，整个启动流程就看完了。接下来我们跳到其中的细节去看。

--------------------------------------------------------------------------------------------------------------

- tcmalloc. tcmalloc是 `thread cache malloc` 的简写，看完之后暗自觉得，大神就是大神。。。设计非常漂亮。

    - https://github.com/gperftools/gperftools
    - https://en.wikipedia.org/wiki/C_dynamic_memory_allocation#Thread-caching_malloc_(tcmalloc)

- `schedinit` 里有整个调度器初始化的代码：

https://github.com/jiajunhuang/go/blob/67a58c5a2401e89fd4f688e8f70fd3be9506cea5/src/runtime/proc.go#L508

- `findrunnable` 中实现了work stealing:

    - 检查是否处于GC态
    - 检查本地有没有可执行的G
    - 检查全局队列有没有可执行的G
    - 检查网络I/O有没有可以恢复执行的G
    - 去别的队列里偷
    - 还是没有，就把自己挂起

- runtime里最重要的几个文件：

    - `runtime1.go` 初始化时的检测
    - `runtime2.go` 初始化等
    - `proc.go` 调度，work stealing等
    - `mheap.go` 和 `malloc.go` 内存分配相关实现

- 如何实现协程？可以看看我的这个项目：

https://github.com/jiajunhuang/storm

虽然看起来是Cython写的，其实是Python，当时为了练习一下Cython就全部加上了类型改了后缀名，没差。Python中实现协程
毕竟简单多了，方便理解。
