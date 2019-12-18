# Go 1.13的errors挺香

前段时间Go发布了1.13,但是因为还没有进Arch的官方库，所以没去尝试。今天抽空试了一下，非常香。我们先来看个例子，然后
看看源码：

```go
package main

import (
	"errors"
	"fmt"
)

func doSomethingWrong(o error) error {
	return fmt.Errorf("%w wrapped error", o)
}

func main() {
	// o: original
	// n: new error
	o := errors.New("original error")

	fmt.Printf("error: %s\n", o)
	n := doSomethingWrong(o)
	fmt.Printf("error: %s\n", n)

	fmt.Printf("n is o: %t\n", errors.Is(n, o))
}
```

这次增加了四个函数：

- `errors.Is` 判断是否a错误是否是b错误的后代
- `errors.Unwrap` 将a错误的包装剔除一层
- `errors.As` 将a错误一直剔除到错误类型为 B 类型为止
- `fmt.Errrrf("%w", err)` 将err错误包装一层

我们来看看实现：

```go
func Errorf(format string, a ...interface{}) error {
	p := newPrinter()
	p.wrapErrs = true
	p.doPrintf(format, a)
	s := string(p.buf)
	var err error
	if p.wrappedErr == nil {
		err = errors.New(s)
	} else {
		err = &wrapError{s, p.wrappedErr}
	}
	p.free()
	return err
}

type wrapError struct {
	msg string
	err error
}

func (e *wrapError) Error() string {
	return e.msg
}

func (e *wrapError) Unwrap() error {
	return e.err
}
```

这是 `fmt.Errorf` 的实现。原理很简单，就是将 `err` 放到 `wrapError` 的 `err` 属性里，将错误信息放到 `msg` 里。

```go
func Unwrap(err error) error {
	u, ok := err.(interface {
		Unwrap() error
	})
	if !ok {
		return nil
	}
	return u.Unwrap()
}
```

就是调用 `wrapError.Unwrap` 方法。

```go
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}

	isComparable := reflectlite.TypeOf(target).Comparable()
	for {
		if isComparable && err == target {
			return true
		}
		if x, ok := err.(interface{ Is(error) bool }); ok && x.Is(target) {
			return true
		}
		// TODO: consider supporing target.Is(err). This would allow
		// user-definable predicates, but also may allow for coping with sloppy
		// APIs, thereby making it easier to get away with them.
		if err = Unwrap(err); err == nil {
			return false
		}
	}
}
```

就是一层一层往上检查看是否能匹配。

```go
func As(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	val := reflectlite.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflectlite.Ptr || val.IsNil() {
		panic("errors: target must be a non-nil pointer")
	}
	if e := typ.Elem(); e.Kind() != reflectlite.Interface && !e.Implements(errorType) {
		panic("errors: *target must be interface or implement error")
	}
	targetType := typ.Elem()
	for err != nil {
		if reflectlite.TypeOf(err).AssignableTo(targetType) {
			val.Elem().Set(reflectlite.ValueOf(err))
			return true
		}
		if x, ok := err.(interface{ As(interface{}) bool }); ok && x.As(target) {
			return true
		}
		err = Unwrap(err)
	}
	return false
}
```

`As` 就是一层一层网上剥离，然后看哪一层能够和给定的错误类型匹配，然后把值放进去。

真香。

---

参考资料：

- [Go官方博客对此的使用介绍](https://blog.golang.org/go1.13-errors)
