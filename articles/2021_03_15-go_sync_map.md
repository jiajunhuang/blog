# Golang sync.Map源码分析

今天看了一下 `sync.Map` 的实现，首先我们从一个demo入手：

```go
package main

import (
	"sync"
)

func main() {
	a := sync.Map{}
	a.Load()
	a.Delete()
	a.Store()
}
```

然后我们就可以跳到源码去看看到底是怎么实现的。首先看 `Map` 的定义：

```go
type Map struct {
	mu Mutex

	// read contains the portion of the map's contents that are safe for
	// concurrent access (with or without mu held).
	//
	// The read field itself is always safe to load, but must only be stored with
	// mu held.
	//
	// Entries stored in read may be updated concurrently without mu, but updating
	// a previously-expunged entry requires that the entry be copied to the dirty
	// map and unexpunged with mu held.
	read atomic.Value // readOnly

	// dirty contains the portion of the map's contents that require mu to be
	// held. To ensure that the dirty map can be promoted to the read map quickly,
	// it also includes all of the non-expunged entries in the read map.
	//
	// Expunged entries are not stored in the dirty map. An expunged entry in the
	// clean map must be unexpunged and added to the dirty map before a new value
	// can be stored to it.
	//
	// If the dirty map is nil, the next write to the map will initialize it by
	// making a shallow copy of the clean map, omitting stale entries.
	dirty map[interface{}]*entry

	// misses counts the number of loads since the read map was last updated that
	// needed to lock mu to determine whether the key was present.
	//
	// Once enough misses have occurred to cover the cost of copying the dirty
	// map, the dirty map will be promoted to the read map (in the unamended
	// state) and the next store to the map will make a new dirty copy.
	misses int
}
```

在Map的注释里，写明了，`sync.Map` 在两种情况下比较好使：

```go
// The Map type is optimized for two common use cases: (1) when the entry for a given
// key is only ever written once but read many times, as in caches that only grow,
// or (2) when multiple goroutines read, write, and overwrite entries for disjoint
// sets of keys. In these two cases, use of a Map may significantly reduce lock
// contention compared to a Go map paired with a separate Mutex or RWMutex.
```

一种是读多写少的情况下，一种是当多个goroutine并发读或写不同的key时。

可以看到 `Map` 有4个属性，`mu` 就是锁，`read` 是一个只读的值，可以看到，既然用了 `atomic.Value` 那么大概率是用 CAS 来进行操作的，
`dirty` 是一个map，`misses` 是用来统计未命中的次数的，注释里说，当
`misses` 达到一定值时，就把 `dirty` 里的值放到 `read` 里。

`read` 的类型其实是 `readOnly`，定义如下：

```go
// readOnly is an immutable struct stored atomically in the Map.read field.
type readOnly struct {
	m       map[interface{}]*entry
	amended bool // true if the dirty map contains some key not in m.
}
```

接下来我们就可以看看具体操作来进一步理解为啥要这样设计了。

```go
// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *Map) Load(key interface{}) (value interface{}, ok bool) {
	read, _ := m.read.Load().(readOnly)
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		// Avoid reporting a spurious miss if m.dirty got promoted while we were
		// blocked on m.mu. (If further loads of the same key will not miss, it's
		// not worth copying the dirty map for this key.)
		read, _ = m.read.Load().(readOnly)
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			// Regardless of whether the entry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if !ok {
		return nil, false
	}
	return e.load()
}
```

首先加载 `read`，如果没有取到值，就去 `dirty` 里拿，并且记录下来未命中：

```go
func (m *Map) missLocked() {
	m.misses++
	if m.misses < len(m.dirty) {
		return
	}
	m.read.Store(readOnly{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}
```

可以看到，如果 `m.misses < len(m.dirty)` 那么就啥也不干，否则就把`m.dirty` 的值存储到 `m.read` 里。

这里我有一个疑问，就是直接把 `dirty` 替换上去的话，`read` 里的值咋办呢？除非是 `dirty` 里有所有 `read` 的值。

我们带着疑惑看看 `Store` 是怎么处理的：

```go
// Store sets the value for a key.
func (m *Map) Store(key, value interface{}) {
	read, _ := m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok && e.tryStore(&value) {
		return
	}

	m.mu.Lock()
	read, _ = m.read.Load().(readOnly)
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			// The entry was previously expunged, which implies that there is a
			// non-nil dirty map and this entry is not in it.
			m.dirty[key] = e
		}
		e.storeLocked(&value)
	} else if e, ok := m.dirty[key]; ok {
		e.storeLocked(&value)
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			m.dirtyLocked()
			m.read.Store(readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
	}
	m.mu.Unlock()
}
```

同样，首先先把 `read` 取出来，然后看看 `read` 里是否有这个值，如果有的话，就更新：

```go
func (e *entry) tryStore(i *interface{}) bool {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == expunged {
			return false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
			return true
		}
	}
}
```

类似于一个自旋锁，循环到存进去了为止。如果没有存进去，那么就继续走下面的逻辑，开始再次尝试读 `read`，如果有的话，更新值，没有的话就看 `dirty` 里有没有，
如果 `read` 和 `dirty` 都没有的话，并且此时 `read` 也没有修改，就会调用

`m.dirtyLocked()`：

```go
func (m *Map) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read, _ := m.read.Load().(readOnly)
	m.dirty = make(map[interface{}]*entry, len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
}
```

会把 `read` 里的值全部拷贝过来，然后赋值给 `read`，但是如果 `read` 的 `amended` 标记为true就不会执行。
然后最终还是写到了 `m.dirty`。

这上面的逻辑很重要的一个原因就是和读取时的顺序有关系，都是保证先读取 `read`，然后加锁之后还会二次确认是否在 `read` 里，然后才会考虑去
`dirty` 里看。

上一个疑惑倒是解开了，但是新的疑惑又来了，`read.amended` 是啥意思，从字面上来看，是说 `dirty` 里有数据，那么就会标记为 `true`，我们来看看相关操作：

```go
func (m *Map) Store(key, value interface{}) {
        ...
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			m.dirtyLocked()
			m.read.Store(readOnly{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
        ...
}

        ...


func (m *Map) missLocked() {
	m.misses++
	if m.misses < len(m.dirty) {
		return
	}
	m.read.Store(readOnly{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}
```

懂了，原来就是，当 `dirty` 里开始写入数据的时候，如果这个key在 `read` 和 `dirty` 里都没有，那就没法直接更新，所以就得往
`m.dirty` 写数据，只要 `m.dirty` 里有数据，那么 `read.amended` 一定是 `true`，而当上一次做完 `dirty` 替换为 `read` 时，
这个时候 `read.amended` 为 `false`，就会执行把read的值全部copy过来然后初始化 `dirty` 的操作。

挺绕的，可以说 `sync.Map` 设计的有点精妙，也可以说是挺复杂的，如果逻辑没有搞清楚，很容易就会出并发问题。

总结一下，`sync.Map` 为了实现高性能的并发Map，采用了类似读写分离的设计，读取的逻辑是优先读 `read`，其次读 `dirty`；写入
的逻辑是，优先更新 `read`，其次尝试更新 `dirty`。正是因为读取时，会先读取 `read` 然后才是 `dirty`，所以写入的时候，即使
`read` 没有，也可以写入到 `dirty`。

最后就是，我们来聊一下 `mutex` 和 `CAS` 的问题，一般来说，锁的性能比CAS略差，锁是基于 [test and set](https://en.wikipedia.org/wiki/Test-and-set)，
CAS是基于 [compare and swap](https://en.wikipedia.org/wiki/Compare-and-swap) 来实现的，两者都是CPU提供的指令，当CPU没有
提供test and set指令时，也可以用compare and swap指令来模拟锁，当然，性能就要差一些了。
