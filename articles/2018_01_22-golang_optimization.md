# Go语言性能优化实战

**过早优化是万恶之源，这里都是黑魔法，不是性能瓶颈慎用**

- 根据pprof数据优化
- 尽量避免GC，所以要避免创建过多的对象，也可以通过设置 `GOGC` 环境变量来增加触发GC的阈值，缺点是费内存。
- 尽量的复用已经创建的对象，其中就包括如果可以的话，预先创建好对象。参考： https://golang.org/pkg/sync/#Pool
- 避免锁，可以考虑 CAS。https://golang.org/pkg/sync/atomic/
- 如果可以的话，用struct代替map。一个简单的例子就可以说明：

```go
package main

type structDemoStruct struct {
	First int
}

var (
	mapDemo    = make(map[int]int, 1)
	structDemo = structDemoStruct{}
)

func mapIncr() {
	mapDemo[1]++
}

func structIncr() {
	structDemo.First++
}
```

```go
package main

import (
	"testing"
)

func BenchmarkMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mapIncr()
	}
}

func BenchmarkStruct(b *testing.B) {
	for i := 0; i < b.N; i++ {
		structIncr()
	}
}
```

测试结果：

```bash
$ go test -bench .
goos: linux
goarch: amd64
BenchmarkMap-4      	100000000	        20.5 ns/op
BenchmarkStruct-4   	1000000000	         2.02 ns/op
PASS
ok  	_/home/jiajun/tests	4.299s
```

- `defer` 虽然好用，但是也会带来性能损伤。如果是高并发的服务，可能要注意一下
- `time.Now` 比你想象中的慢，如果时间精确度不高，可以自己实现一个粗略的时钟
- `[]byte`和string转换比你想象中的慢，而且不能愉快的重复利用
- 减少锁的使用，缩小临界区
- 如果是CPU Bound的话，可以考虑设置 `GOGC` 来用内存换CPU时间
