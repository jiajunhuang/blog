# Go设计模式：Iterator

最常见的迭代模式莫过于循环：

```go
package main

func main() {
	for i := 0; i < 10; i++ {
		println(i)
	}
}
```

迭代器会依次把元素呈现给我们，当没有元素时，迭代终止。

那么Go语言里还有哪些常见的迭代器呢？比如 `sync.Map` 里，就有：

```go
package main

import (
	"fmt"
	"sync"
)

func main() {
	aMap := sync.Map{}

	aMap.Store("hello", "world")
	aMap.Store("world", "hello")

	aMap.Range(func(k, v interface{}) bool {
		fmt.Printf("key: %s, value %s\n", k, v)
		return true
	})
}
```

我们来看看这个 Range 是怎么实现的：

```go
func (m *Map) Range(f func(key, value interface{}) bool) {
	// We need to be able to iterate over all of the keys that were already
	// present at the start of the call to Range.
	// If read.amended is false, then read.m satisfies that property without
	// requiring us to hold m.mu for a long time.
	read, _ := m.read.Load().(readOnly)
	if read.amended {
		// m.dirty contains keys not in read.m. Fortunately, Range is already O(N)
		// (assuming the caller does not break out early), so a call to Range
		// amortizes an entire copy of the map: we can promote the dirty copy
		// immediately!
		m.mu.Lock()
		read, _ = m.read.Load().(readOnly)
		if read.amended {
			read = readOnly{m: m.dirty}
			m.read.Store(read)
			m.dirty = nil
			m.misses = 0
		}
		m.mu.Unlock()
	}

	for k, e := range read.m {
		v, ok := e.load()
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}

// readOnly is an immutable struct stored atomically in the Map.read field.
type readOnly struct {
	m       map[interface{}]*entry
	amended bool // true if the dirty map contains some key not in m.
}
```

其中 `read.m` 是一个map，所以还是使用Go内置的map来做的，那如果我们想自己做一个该咋办呢？我们来参考
[golang-set](https://github.com/deckarep/golang-set) 的实现方式：

先看用法：

```go
package main

import (
	"fmt"

	mapset "github.com/deckarep/golang-set"
)

func main() {
	aSet := mapset.NewSet()
	aSet.Add("hello")
	aSet.Add("world")
	aSet.Add("hello")
	aSet.Add("world")

	for v := range aSet.Iterator().C {
		fmt.Printf("v: %s\n", v.(string))
	}
}
```

来看看实现：

```go
type Iterator struct {
	C    <-chan interface{}
	stop chan struct{}
}

func (set *threadSafeSet) Iterator() *Iterator {
	iterator, ch, stopCh := newIterator()

	go func() {
		set.RLock()
	L:
		for elem := range set.s {
			select {
			case <-stopCh:
				break L
			case ch <- elem:
			}
		}
		close(ch)
		set.RUnlock()
	}()

	return iterator
}
```

可以看出来，其实它也是通过底层的for循环，加channel来实现的，我们自己来实现一个：

```go
package main

import (
	"fmt"
)

// Iterator 声明接口
type Iterator interface {
	Iterator(m iterSet) Iter
}

// Iter 迭代器的实现
type Iter struct {
	C chan interface{}
}

func newIter(i *iterSet) Iter {
	iter := Iter{make(chan interface{})}

	go func() {
		for k := range i.m {
			iter.C <- k
		}
		close(iter.C)
	}()

	return iter
}

// 我们自己的set
type iterSet struct {
	m map[string]bool
}

// Add 添加元素
func (i *iterSet) Add(k string) {
	i.m[k] = true
}

// Iterator 返回一个迭代器
func (i *iterSet) Iterator() Iter {
	return newIter(i)
}

func main() {
	aSet := iterSet{map[string]bool{}}
	aSet.Add("hello")
	aSet.Add("hello")
	aSet.Add("world")
	aSet.Add("world")

	iter := aSet.Iterator()

	for v := range iter.C {
		fmt.Printf("key: %s\n", v.(string))
	}
}
```

duang，搞定！

---

参考资料：

- https://en.wikipedia.org/wiki/Iterator_pattern
