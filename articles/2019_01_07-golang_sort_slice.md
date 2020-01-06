# Golang不那么蛋疼的sort

以前Go里写排序，如果不能用 `sort.Ints`, `sort.Strings`, `sort.Float64s` 等等快捷函数，就只能实现 `sort.Interface` 这个
接口了：

```go
type Interface interface {
    // Len is the number of elements in the collection.
    Len() int
    // Less reports whether the element with
    // index i should sort before the element with index j.
    Less(i, j int) bool
    // Swap swaps the elements with indexes i and j.
    Swap(i, j int)
}
```

很蛋疼对不对？经群友提醒，Go 1.8以后，可以使用 `sort.Slice` 这个快捷函数快速实现排序而不用实现上面那个接口了，看例子：

```go
package main

import (
	"fmt"
	"sort"
)

func main() {
	people := []struct {
		Name string
		Age  int
	}{
		{"Gopher", 7},
		{"Alice", 55},
		{"Vera", 24},
		{"Bob", 75},
	}
	sort.Slice(people, func(i, j int) bool { return people[i].Name < people[j].Name })
	fmt.Println("By name:", people)

	sort.Slice(people, func(i, j int) bool { return people[i].Age < people[j].Age })
	fmt.Println("By age:", people)
}
```

完美！比以前简单多了对不对。

---

参考资料：

- [官方文档](https://golang.org/pkg/sort/#Slice)
