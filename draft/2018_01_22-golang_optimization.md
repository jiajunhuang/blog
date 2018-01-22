# Go性能优化

- 尽量避免GC，所以要避免创建过多的对象
- 尽量的复用已经创建的对象，其中就包括如果可以的话，预先创建好对象
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
