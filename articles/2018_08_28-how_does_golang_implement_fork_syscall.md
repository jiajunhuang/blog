# How to implement fork syscall in Golang?

ref: https://github.com/moby/moby/tree/master/pkg/reexec

We don't have a `fork` syscall in Golang, we have:

- [syscall.ForkExec](https://golang.org/pkg/syscall/#ForkExec)
- [os.StartProcess](https://golang.org/pkg/os/#StartProcess)
- [exec.Cmd](https://golang.org/pkg/os/exec/#Command)

All those three functions is like a combination of `fork + exec`, but there has no pure `fork` syscall just like in
C programming language(after syscall, which will return pid in caller). Reasons can be found in [here](https://stackoverflow.com/questions/28370646/how-do-i-fork-a-go-process/28371586#28371586):

It mainly says：

- `fork()` has been invented at the time when no threads were used at all, and a process had always had just a single thread of execution in it, and hence forking it was safe.
- In C, you control all the threads by your hand, but in Go, you cannot, so threads will be out of control after call `fork` without `exec`, so Go provides `fork + exec` only.


Let's have a look at how we use pure `fork` in C:

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

Run it：

```bash
$ gcc main.c && ./a.out
main process
main exit
child process
```

Let's look how can we implements this in Go:

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

Run it：

```bash
$ go run main.go
2018/03/08 19:52:39 init start, os.Args = [/tmp/go-build209640177/b001/exe/main]
2018/03/08 19:52:39 main start, os.Args = [/tmp/go-build209640177/b001/exe/main]
2018/03/08 19:52:39 init start, os.Args = [childProcess]
2018/03/08 19:52:39 childProcess
2018/03/08 19:52:39 main exit
```

Explanation:

`init` will be execute before `main` function. When you execute the binary executable file from command line,
`os.Args[0]` will be the name of binary executable file, but, `reexec.Command` will change `os.Args[0]`, so
child process will find function registed by `reexec.Register`, and execute it, return `true`, then call `os.Exit(0)`.
