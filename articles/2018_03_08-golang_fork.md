# Golang中实现典型的fork调用

[English Version](https://jiajunhuang.com/articles/2018_08_28-how_does_golang_implement_fork_syscall.md.html)

## 背景

在日常的系统编程中，我们经常有这样的需求：`fork` 多个子进程去执行任务。例如典型的 `master-slave` 架构便是这样的，`fork` 调用之后，父进程开始作为管理进程管理子进程，而子进程则开始执行处理请求的代码。但是在Go语言中，由于Go语言屏蔽了进程、线程的概念，而只提供了 `goroutine` 的概念，导致我们无法很好的直接操作 `fork` 调用。在Go语言中，此类解决方案只有以下三种：

- [syscall.ForkExec](https://golang.org/pkg/syscall/#ForkExec)
- [os.StartProcess](https://golang.org/pkg/os/#StartProcess)
- [exec.Cmd](https://golang.org/pkg/os/exec/#Command)

这三个都类似于 `fork + exec`，但是没有类似C中的fork调用可以达到在fork之后根据返回的pid然后进入不同的函数的方案。原因主要是：

- `fork` 早出现在只有进程，没有线程的年代
- C中是自行控制线程，这样fork之后才不会发生紊乱。一般都是单线程fork之后，才会开始多线程执行。
- Go中多线程是runtime自行决定的，所以Go中没有提供单纯的fork，而是fork之后立即就exec执行新的二进制文件

为了实现上述需求，我们需要利用Go语言的一些特性进行模拟。

## C语言的实现方式

在了解Go如何实现fork之前，先看一下C里的传统使用方式：

```c
#include <sys/types.h>
#include <unistd.h>
#include <stdio.h>
#include <sys/wait.h>

void child() {
    printf("child process\n");
}

int main() {
    printf("main process\n");
    pid_t pid = fork();
    int wstatus;

    if (pid == 0) {
        child();
    } else {
        printf("main exit\n");
        waitpid(pid, &wstatus, 0);
    }
}
```

运行一下：

```bash
$ gcc main.c && ./a.out 
main process
main exit
child process
```

## Go语言的实现方案

然后我们看看Docker提供的实现fork的方式：

```go
package main

import (
	"log"
	"os"

	"github.com/docker/docker/pkg/reexec"
)

func init() {
	log.Printf("init start, os.Args = %+v\n", os.Args)
	reexec.Register("childProcess", childProcess)
	if reexec.Init() {
		os.Exit(0)
	}
}

func childProcess() {
	log.Println("childProcess")
}

func main() {
	log.Printf("main start, os.Args = %+v\n", os.Args)
	cmd := reexec.Command("childProcess")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Panicf("failed to run command: %s", err)
	}
	if err := cmd.Wait(); err != nil {
		log.Panicf("failed to wait command: %s", err)
	}
	log.Println("main exit")
}
```

运行一下：

```bash
$ go run main.go
2018/03/08 19:52:39 init start, os.Args = [/tmp/go-build209640177/b001/exe/main]
2018/03/08 19:52:39 main start, os.Args = [/tmp/go-build209640177/b001/exe/main]
2018/03/08 19:52:39 init start, os.Args = [childProcess]
2018/03/08 19:52:39 childProcess
2018/03/08 19:52:39 main exit
```

这样我们就使用Go语言模拟了传统的 `fork` 调用。

其原理是在Go语言中，运行时会保证 `init` 一定会在 `main` 之前执行。而初次执行程序的时候 `os.Args[0]` 是可执行文件的名字，但是 `reexec.Command` 却可以修改子进程的 `os.Args[0]`，所以子进程会直接找到 `reexec.Init` 上`reexec.Register` 所注册的函数，然后执行，返回true，最后调用 `os.Exit(0)`，因此就能达到传统的 `fork` 之后，子进程执行另外一个函数的效果。

## 总结

通过利用Go语言里，`runtime` 总是会保证 `init` 函数优先于 `main` 函数执行的特点，我们模拟了传统的 `fork` 调用，使得Go语言也可以方便的像其他语言一样，调用 `fork` 之后执行代码中的其他代码段。

---

参考资料：

- https://github.com/moby/moby/tree/master/pkg/reexec 
- https://stackoverflow.com/questions/28370646/how-do-i-fork-a-go-process/28371586#28371586
