# Golang slice 源码阅读

翻Golang代码：

```go
type slice struct {
	array unsafe.Pointer
	len   int
	cap   int
}
```

slice 的array是一个指针，指向一块连续内存。

再看 `growslice` 函数，这是append调用的函数

```go
func growslice(et *_type, old slice, cap int) slice {
	if raceenabled {
		callerpc := getcallerpc(unsafe.Pointer(&et))
		racereadrangepc(old.array, uintptr(old.len*int(et.size)), callerpc, funcPC(growslice))
	}
	if msanenabled {
		msanread(old.array, uintptr(old.len*int(et.size)))
	}

	if et.size == 0 {
		if cap < old.cap {
			panic(errorString("growslice: cap out of range"))
		}
		// append should not create a slice with nil pointer but non-zero len.
		// We assume that append doesn't need to preserve old.array in this case.
		return slice{unsafe.Pointer(&zerobase), old.len, cap}
	}

	newcap := old.cap
	doublecap := newcap + newcap
	if cap > doublecap {
		newcap = cap
	} else {
		if old.len < 1024 {
			newcap = doublecap
		} else {
			for newcap < cap {
				newcap += newcap / 4
			}
		}
	}

	var lenmem, newlenmem, capmem uintptr
	const ptrSize = unsafe.Sizeof((*byte)(nil))
	switch et.size {
	case 1:
		lenmem = uintptr(old.len)
		newlenmem = uintptr(cap)
		capmem = roundupsize(uintptr(newcap))
		newcap = int(capmem)
	case ptrSize:
		lenmem = uintptr(old.len) * ptrSize
		newlenmem = uintptr(cap) * ptrSize
		capmem = roundupsize(uintptr(newcap) * ptrSize)
		newcap = int(capmem / ptrSize)
	default:
		lenmem = uintptr(old.len) * et.size
		newlenmem = uintptr(cap) * et.size
		capmem = roundupsize(uintptr(newcap) * et.size)
		newcap = int(capmem / et.size)
	}

	if cap < old.cap || uintptr(newcap) > maxSliceCap(et.size) {
		panic(errorString("growslice: cap out of range"))
	}

	var p unsafe.Pointer
	if et.kind&kindNoPointers != 0 {
		p = mallocgc(capmem, nil, false)
		memmove(p, old.array, lenmem)
		// The append() that calls growslice is going to overwrite from old.len to cap (which will be the new length).
		// Only clear the part that will not be overwritten.
		memclrNoHeapPointers(add(p, newlenmem), capmem-newlenmem)
	} else {
		// Note: can't use rawmem (which avoids zeroing of memory), because then GC can scan uninitialized memory.
		p = mallocgc(capmem, et, true)
		if !writeBarrier.enabled {
			memmove(p, old.array, lenmem)
		} else {
			for i := uintptr(0); i < lenmem; i += et.size {
				typedmemmove(et, add(p, i), add(old.array, i))
			}
		}
	}

	return slice{p, old.len, newcap}
}
```

可以看出slice每次扩容的实现是在小于1024的时候每次乘以2，之后在小于cap的时候每次
加1/4，一直到超过为止。

所以来看看同事发的一段代码：

```go
package main

import "fmt"

func main() {
	s := []int{5}
	s = append(s, 7)
	s = append(s, 9)
	x := append(s, 11)
	y := append(s, 12)
	fmt.Println(s, x, y)
}
```

最开始同事问我你猜这段代码会输出什么的时候我答错了，本以为Golang的语义不会
实现成这样的。不过翻了实现才发现，well, ahh...

```bash
root@arch test: go run test.go 
[5 7 9] [5 7 9 12] [5 7 9 12]
```

按照上面的内存翻倍策略，`s := []int{5}` 的时候，array容量是1，`s = append(s, 7)`
时为2，`s = append(s, 9)` 时为3，执行 `x := append(s, 11)` 时为4，但是
执行 `y := append(s, 12)` 时容量仍然为4，因为尚未执行 `x := append(s, 11)` 时，
`x.array` 指向了连续内存（数组），len为3，cap为4，s最后一个元素是9，`append`
之后会把9后面的元素填成11，然后返回这样一个 `slice` 对象给x, 同样，
`y := append(s, 12)` 时一样，执行之前，len为3，cap为4，s的最后一个元素是9，
所以执行之后，就把原来的11给覆盖了。

还是挺坑的，内存直接被盖掉了。
