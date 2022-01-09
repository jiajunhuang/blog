# 使用反射(reflect)对结构体赋值

[上一篇](./2022_01_09-gin_binding_args.md) 中，我们看了GIN是如何绑定参数并且校验的，本着知道如何使用也要知道底层原理的
探索精神，这一篇中，我们自己来使用 `reflect` 实现一个轻量版的参数绑定。

不过在此之前，我们需要先了解和熟悉 `reflect` 库。

## 学习 reflect

对于代码中的一个变量来说，它有两个信息：1. 类型；2. 值。类型是指，这个变量具体是什么类型，比如是否是 `string`, `int`,
`bool`, `*ptr`；对于值，就是具体的赋值，比如 `x := 1`，`x` 的类型是 `int`, 值是 `1`。这就引入了 `reflect` 库中最重要的
两个函数：

- `reflect.TypeOf`。类型是 `func TypeOf(i interface{}) Type`。
- `reflect.ValueOf`。类型是 `func ValueOf(i interface{}) Value`。

我们来看看两者分别是什么：

```go
// TypeOf returns the reflection Type that represents the dynamic type of i.
// If i is a nil interface value, TypeOf returns nil.
func TypeOf(i interface{}) Type {
	eface := *(*emptyInterface)(unsafe.Pointer(&i))
	return toType(eface.typ)
}

// Type is the representation of a Go type.
//
// Not all methods apply to all kinds of types. Restrictions,
// if any, are noted in the documentation for each method.
// Use the Kind method to find out the kind of type before
// calling kind-specific methods. Calling a method
// inappropriate to the kind of type causes a run-time panic.
//
// Type values are comparable, such as with the == operator,
// so they can be used as map keys.
// Two Type values are equal if they represent identical types.
type Type interface {
	// Methods applicable to all types.

	// Align returns the alignment in bytes of a value of
	// this type when allocated in memory.
	Align() int

	// FieldAlign returns the alignment in bytes of a value of
	// this type when used as a field in a struct.
	FieldAlign() int

	// Method returns the i'th method in the type's method set.
	// It panics if i is not in the range [0, NumMethod()).
	//
	// For a non-interface type T or *T, the returned Method's Type and Func
	// fields describe a function whose first argument is the receiver,
	// and only exported methods are accessible.
	//
	// For an interface type, the returned Method's Type field gives the
	// method signature, without a receiver, and the Func field is nil.
	//
	// Methods are sorted in lexicographic order.
	Method(int) Method

	// MethodByName returns the method with that name in the type's
	// method set and a boolean indicating if the method was found.
	//
	// For a non-interface type T or *T, the returned Method's Type and Func
	// fields describe a function whose first argument is the receiver.
	//
	// For an interface type, the returned Method's Type field gives the
	// method signature, without a receiver, and the Func field is nil.
	MethodByName(string) (Method, bool)

	// NumMethod returns the number of methods accessible using Method.
	//
	// Note that NumMethod counts unexported methods only for interface types.
	NumMethod() int

	// Name returns the type's name within its package for a defined type.
	// For other (non-defined) types it returns the empty string.
	Name() string

	// PkgPath returns a defined type's package path, that is, the import path
	// that uniquely identifies the package, such as "encoding/base64".
	// If the type was predeclared (string, error) or not defined (*T, struct{},
	// []int, or A where A is an alias for a non-defined type), the package path
	// will be the empty string.
	PkgPath() string

	// Size returns the number of bytes needed to store
	// a value of the given type; it is analogous to unsafe.Sizeof.
	Size() uintptr

	// String returns a string representation of the type.
	// The string representation may use shortened package names
	// (e.g., base64 instead of "encoding/base64") and is not
	// guaranteed to be unique among types. To test for type identity,
	// compare the Types directly.
	String() string

	// Kind returns the specific kind of this type.
	Kind() Kind

	// Implements reports whether the type implements the interface type u.
	Implements(u Type) bool

	// AssignableTo reports whether a value of the type is assignable to type u.
	AssignableTo(u Type) bool

	// ConvertibleTo reports whether a value of the type is convertible to type u.
	// Even if ConvertibleTo returns true, the conversion may still panic.
	// For example, a slice of type []T is convertible to *[N]T,
	// but the conversion will panic if its length is less than N.
	ConvertibleTo(u Type) bool

	// Comparable reports whether values of this type are comparable.
	// Even if Comparable returns true, the comparison may still panic.
	// For example, values of interface type are comparable,
	// but the comparison will panic if their dynamic type is not comparable.
	Comparable() bool

	// Methods applicable only to some types, depending on Kind.
	// The methods allowed for each kind are:
	//
	//	Int*, Uint*, Float*, Complex*: Bits
	//	Array: Elem, Len
	//	Chan: ChanDir, Elem
	//	Func: In, NumIn, Out, NumOut, IsVariadic.
	//	Map: Key, Elem
	//	Ptr: Elem
	//	Slice: Elem
	//	Struct: Field, FieldByIndex, FieldByName, FieldByNameFunc, NumField

	// Bits returns the size of the type in bits.
	// It panics if the type's Kind is not one of the
	// sized or unsized Int, Uint, Float, or Complex kinds.
	Bits() int

	// ChanDir returns a channel type's direction.
	// It panics if the type's Kind is not Chan.
	ChanDir() ChanDir

	// IsVariadic reports whether a function type's final input parameter
	// is a "..." parameter. If so, t.In(t.NumIn() - 1) returns the parameter's
	// implicit actual type []T.
	//
	// For concreteness, if t represents func(x int, y ... float64), then
	//
	//	t.NumIn() == 2
	//	t.In(0) is the reflect.Type for "int"
	//	t.In(1) is the reflect.Type for "[]float64"
	//	t.IsVariadic() == true
	//
	// IsVariadic panics if the type's Kind is not Func.
	IsVariadic() bool

	// Elem returns a type's element type.
	// It panics if the type's Kind is not Array, Chan, Map, Ptr, or Slice.
	Elem() Type

	// Field returns a struct type's i'th field.
	// It panics if the type's Kind is not Struct.
	// It panics if i is not in the range [0, NumField()).
	Field(i int) StructField

	// FieldByIndex returns the nested field corresponding
	// to the index sequence. It is equivalent to calling Field
	// successively for each index i.
	// It panics if the type's Kind is not Struct.
	FieldByIndex(index []int) StructField

	// FieldByName returns the struct field with the given name
	// and a boolean indicating if the field was found.
	FieldByName(name string) (StructField, bool)

	// FieldByNameFunc returns the struct field with a name
	// that satisfies the match function and a boolean indicating if
	// the field was found.
	//
	// FieldByNameFunc considers the fields in the struct itself
	// and then the fields in any embedded structs, in breadth first order,
	// stopping at the shallowest nesting depth containing one or more
	// fields satisfying the match function. If multiple fields at that depth
	// satisfy the match function, they cancel each other
	// and FieldByNameFunc returns no match.
	// This behavior mirrors Go's handling of name lookup in
	// structs containing embedded fields.
	FieldByNameFunc(match func(string) bool) (StructField, bool)

	// In returns the type of a function type's i'th input parameter.
	// It panics if the type's Kind is not Func.
	// It panics if i is not in the range [0, NumIn()).
	In(i int) Type

	// Key returns a map type's key type.
	// It panics if the type's Kind is not Map.
	Key() Type

	// Len returns an array type's length.
	// It panics if the type's Kind is not Array.
	Len() int

	// NumField returns a struct type's field count.
	// It panics if the type's Kind is not Struct.
	NumField() int

	// NumIn returns a function type's input parameter count.
	// It panics if the type's Kind is not Func.
	NumIn() int

	// NumOut returns a function type's output parameter count.
	// It panics if the type's Kind is not Func.
	NumOut() int

	// Out returns the type of a function type's i'th output parameter.
	// It panics if the type's Kind is not Func.
	// It panics if i is not in the range [0, NumOut()).
	Out(i int) Type

	common() *rtype
	uncommon() *uncommonType
}
```

`Type` 代表Go语言中的类型，它是一个接口，但是有一大堆方法。我们需要仔细读一下注释，注释中有说明，其中很多方法都是
特定类型才能使用的，否则会panic。

接下来我们来看 `Value`:

```go
// ValueOf returns a new Value initialized to the concrete value
// stored in the interface i. ValueOf(nil) returns the zero Value.
func ValueOf(i interface{}) Value {
	if i == nil {
		return Value{}
	}

	// TODO: Maybe allow contents of a Value to live on the stack.
	// For now we make the contents always escape to the heap. It
	// makes life easier in a few places (see chanrecv/mapassign
	// comment below).
	escapes(i)

	return unpackEface(i)
}

// Value is the reflection interface to a Go value.
//
// Not all methods apply to all kinds of values. Restrictions,
// if any, are noted in the documentation for each method.
// Use the Kind method to find out the kind of value before
// calling kind-specific methods. Calling a method
// inappropriate to the kind of type causes a run time panic.
//
// The zero Value represents no value.
// Its IsValid method returns false, its Kind method returns Invalid,
// its String method returns "<invalid Value>", and all other methods panic.
// Most functions and methods never return an invalid value.
// If one does, its documentation states the conditions explicitly.
//
// A Value can be used concurrently by multiple goroutines provided that
// the underlying Go value can be used concurrently for the equivalent
// direct operations.
//
// To compare two Values, compare the results of the Interface method.
// Using == on two Values does not compare the underlying values
// they represent.
type Value struct {
	// typ holds the type of the value represented by a Value.
	typ *rtype

	// Pointer-valued data or, if flagIndir is set, pointer to data.
	// Valid when either flagIndir is set or typ.pointers() is true.
	ptr unsafe.Pointer

	// flag holds metadata about the value.
	// The lowest bits are flag bits:
	//	- flagStickyRO: obtained via unexported not embedded field, so read-only
	//	- flagEmbedRO: obtained via unexported embedded field, so read-only
	//	- flagIndir: val holds a pointer to the data
	//	- flagAddr: v.CanAddr is true (implies flagIndir)
	//	- flagMethod: v is a method value.
	// The next five bits give the Kind of the value.
	// This repeats typ.Kind() except for method values.
	// The remaining 23+ bits give a method number for method values.
	// If flag.kind() != Func, code can assume that flagMethod is unset.
	// If ifaceIndir(typ), code can assume that flagIndir is set.
	flag

	// A method value represents a curried method invocation
	// like r.Read for some receiver r. The typ+val+flag bits describe
	// the receiver r, but the flag's Kind bits say Func (methods are
	// functions), and the top bits of the flag give the method number
	// in r's type's method table.
}
```

除了 `Type` 和 `Value`，我们还需要知道一个东西，那就是 `Kind`。我们可以通过一个简单的例子来看看他们的区别：

```go
package main

import (
	"fmt"
	"reflect"
)

type MyStruct struct {
	i int
}

func main() {
	m := MyStruct{1}

	t := reflect.TypeOf(m)
	v := reflect.ValueOf(m)

	fmt.Printf("t: %s, kind: %s, v: %s, kind: %s\n", t, t.Kind(), v, v.Kind())
}
```

运行一下：

```go
$ go run main.go
t: main.MyStruct, kind: struct, v: {%!s(int=1)}, kind: struct
```

可以看到，`Type` 保存的是变量的类型，而 kind 是变量最终在Go里存在时的原生类型。并且从 `Type` 和 `Value` 都能拿到这个信息。
`Kind` 的种类有：

```go
// A Kind represents the specific kind of type that a Type represents.
// The zero Kind is not a valid kind.
type Kind uint

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
)
```

## 自己实现参数绑定

有了上述知识，我们就可以自己实现一个参数绑定甚至是JSON序列化和反序列化的库了。首先我们确定好函数的签名：

```go
func mapping(dst interface{}, m map[string][]string)
```

入参是一个结构体的指针，因为Go里所有的函数调用传参都是 pass by value，也就是值拷贝，如果想要修改一个变量，必须把指针
传进去，而第二个参数的类型则是 `map[string][]string`，这是因为我们在上一篇文章中已经了解到了，`url.Values`，`headers`
的底层表示，都是这个类型，所以我们直接使用这个类型。

接下来就是大体的逻辑：

- 首先我们要检验参数的类型
- 然后我们拿到结构体的类型信息，依次迭代结构体的每一个成员并且根据类型尝试解析 `m` 里的值，最后赋值

逻辑不难，最主要是要搞清楚 `reflect` 提供的能力，我们直接看代码，代码中有注释：

```go
package main

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

func mapping(dst interface{}, m map[string][]string) {
	typ := reflect.TypeOf(dst)

	// 首先判断传入参数的类型
	if !(typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct) {
		log.Printf("Should pass ptr to destination struct object. Usage: mapping(&someStruct, m)")
		return
	}

	// 拿到指针所指向的元素的类型
	typ = typ.Elem()
	// 拿到指针所指向的元素的值
	value := reflect.ValueOf(dst).Elem()

	// 遍历每一个字段
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// 忽略非导出字段
		if !field.IsExported() {
			log.Printf("field %s is not exported, ignore", field.Name)
			continue
		}

		// 判断是否设置了这个tag
		formTag := field.Tag.Get("form")
		if formTag == "" {
			log.Printf("tag `form` not exist in field, ignore")
			continue
		}

		// 查看是否有取值
		vs := m[formTag]
		if len(vs) == 0 {
			log.Printf("vs by formTag %s not found, ignore", formTag)
			continue
		}
		v := vs[len(vs)-1]

		// 根据类型来设置值
		switch fieldType := field.Type.Kind(); fieldType {
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
			typedV, _ := strconv.ParseInt(v, 10, 64)
			value.Field(i).SetInt(typedV)
		case reflect.String:
			value.Field(i).SetString(v)
		case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			typedV, _ := strconv.ParseUint(v, 10, 64)
			value.Field(i).SetUint(typedV)
		case reflect.Bool:
			value.Field(i).SetBool(v == "true")
		default:
			log.Printf("field type %s not support yet", fieldType)
		}
	}
}

func main() {
	m := map[string][]string{
		"name":  {"jhony"},
		"age":   {"1"},
		"money": {"10010010"},
	}

	type Person struct {
		Name     string `form:"name"`
		Age      uint   `form:"age"`
		Money    int64  `form:"money"`
		unexport string `form:"unexport"`
		NotFound bool   `form:"not_found"`
		NoTag    int8
	}

	i := 1
	mapping(i, m)
	mapping(&i, m)

	p := Person{}
	mapping(&p, m)

	fmt.Printf("%v\n", p)
}
```

运行一下：

```bash
$ go run main.go 
2022/01/09 17:21:55 Should pass ptr to destination struct object. Usage: mapping(&someStruct, m)
2022/01/09 17:21:55 Should pass ptr to destination struct object. Usage: mapping(&someStruct, m)
2022/01/09 17:21:55 field unexport is not exported, ignore
2022/01/09 17:21:55 vs by formTag not_found not found, ignore
2022/01/09 17:21:55 tag `form` not exist in field, ignore
{jhony 1 10010010  false 0}
```

搞定！

## 总结

上一篇文章中，我们看到GIN大概是如何绑定参数的，这一篇文章中，我们自己来实现一套轻量版的逻辑，通过这样实战一番，对
`reflect` 就会更加熟悉。
