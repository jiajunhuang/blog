# Golang中实现禁止拷贝

Go中没有原生的禁止拷贝的方式，所以如果有的结构体，你希望使用者无法拷贝，只能指针传递保证全局唯一的话，可以这么干，定义
一个结构体叫 `noCopy`，要实现 `sync.Locker` 这个接口

```go
// noCopy may be embedded into structs which must not be copied
// after the first use.
//
// See https://golang.org/issues/8005#issuecomment-190753527
// for details.
type noCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*noCopy) Lock() {}
func (*noCopy) Unlock() {}
```

然后把 `noCopy` 嵌到你自定义的结构体里，然后 `go vet` 就可以帮我们进行检查了。举个例子：

```go
package main

import (
	"fmt"
)

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

type Demo struct {
	noCopy noCopy
}

func Copy(d Demo) {
	CopyTwice(d)
}
func CopyTwice(d Demo) {}

func main() {
	d := Demo{}
	fmt.Printf("%+v", d)

	Copy(d)

	fmt.Printf("%+v", d)
}
```

执行一下：

```bash
$ go vet main.go
# command-line-arguments
./main.go:16: Copy passes lock by value: main.Demo contains main.noCopy
./main.go:17: call of CopyTwice copies lock value: main.Demo contains main.noCopy
./main.go:19: CopyTwice passes lock by value: main.Demo contains main.noCopy
./main.go:23: call of fmt.Printf copies lock value: main.Demo contains main.noCopy
./main.go:25: call of Copy copies lock value: main.Demo contains main.noCopy
./main.go:27: call of fmt.Printf copies lock value: main.Demo contains main.noCopy
```

---

- https://golang.org/issues/8005#issuecomment-190753527
- https://github.com/jiajunhuang/go/blob/annotated/src/sync/cond.go#L94:6
- https://stackoverflow.com/questions/52494458/nocopy-minimal-example
