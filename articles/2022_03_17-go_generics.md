# Go 泛型

Go 1.18 发布了，其中一大特性就是泛型，说到泛型， 曾经我也是不怎么喜欢泛型的，
因为泛型会增加代码的复杂度，但是后来随着经验的增加，我还是正视了泛型，泛型
有其存在的必要性，例如当实现一些容器库、算法库的时候，有泛型就会简单方便很多，
Go里，`sort` 库就是一个反例。

## 没有泛型的时候

为了理解泛型，我们还是要先从原始代码看起，也就是没有泛型的时候。举个例子，
如果我们想要实现一个 `sum` 库，提供各种类型累加的函数。我们知道，Go里面有
以下基本类型：

```go
bool

string

int  int8  int16  int32  int64
uint uint8 uint16 uint32 uint64 uintptr

byte // alias for uint8

rune // alias for int32
     // represents a Unicode code point

float32 float64

complex64 complex128
```

其中数字、complex和string都可以实现累加。由于没有泛型，那么只能我们自己累一点，
为每一个基础类型都封装一个包装函数：

```go
package summer

func Ints(elems []int) int {
	var sum int
	for _, elem := range elems {
		sum += elem
	}
	return sum
}

func Strings(elems []string) string {
	var sum string
	for _, elem := range elems {
		sum += elem
	}
	return sum
}

func Float64s(elems []float64) float64 {
	var sum float64
	for _, elem := range elems {
		sum += elem
	}
	return sum
}

func Ints64(elems []int64) int64 {
	var sum int64
	for _, elem := range elems {
		sum += elem
	}
	return sum
}

// ...
```

对此你有什么感受呢？其实实现泛型不难，就是苦了库的作者，而且，无法扩展，比如
我们如果想要 `decimal` 支持，就得在代码库里自己实现一遍。

## 泛型出场

这个时候，我们就可以请泛型出场了。简单来说，泛型就是把上面我们手工一个一个类型
实现累加函数，变成有编译器来实现。泛型，就是指函数实现的时候，我们还不知道是
一个什么类型，但是运行的时候，它一定是一个具体的类型。我们实现的时候暂且不管
是啥类型，只要求这个类型符合一定的约束，比如，可以累加，可以比较。

用实际例子来说，上面的代码，我们可以改写成：

```go
type Summer interface {
	~int | ~float32 | ~float64 | ~string | ~int64
}

func sum[V Summer](vs ...V) V {
	var total V

	for _, v := range vs {
		total += v
	}

	return total
}
```

这样，代码就从上面的近白行，变成了十几行。计算机世界里，有一句话叫做，没有什么
是加一个中间层解决不了的，如果有，那就再加一层。泛型就是这么一个意思，由于
我们不想针对每一个类型都写一遍重复的逻辑，而且其实我们只需要这些类型符合一定
的特征就行，所以我们加了一层抽象，告诉编译器，我们这里将来会是一个具体的类型，
但是现在我还不知道具体是啥类型，只要他们符合一定的约束即可，符合这些约束的参数，
就可以执行这些逻辑。

以上面的例子来说：

- `func sum[V Summer](vs ...V) V` 声明了我们有一个参数类型为V，V符合约束Summer，函数 `sum` 会传入变长参数，每一个参数的类型都是V，返回值类型也是V
- `Summer` 这个约束的意思，看上面的类型定义，意思是只要求 `Summer` 的底层类型是 `int`, `float32`, `float64`, `string` 或者 `int64`
- 接下来就是具体的逻辑，`var total V` 声明变量 `total`，类型为 `V`
- 然后是 `for` 循环，累加
- 最后就是返回 `total`

其实看到这里，不难发现泛型代码和具体类型的代码相差不大，只不过从一个具体的
类型变成了一个不确定的类型。如果一开始觉得难以理解泛型的话，可以先写具体类型
的代码，然后把具体类型代码改造成泛型代码：

- 先把具体类型，改成 `T` 或者 `V`，其实叫啥都行，只不过我们一般都用一个大写字母
- 然后往函数名和参数的圆括号之间，加一个中括号，声明 `T` 或者 `V` 都符合什么约束

就这样简单两步，就可以改造成一个泛型函数了。

## 泛型语法

看完了具体的例子，我们来看看Go都支持什么地方用泛型，也就是泛型的语法：

- Functions can have an additional type parameter list that uses square brackets but otherwise looks like an ordinary parameter list: func F[T any](p T) { ... }. 函数里，用中括号声明函数将会用到的泛型，例如 `func sum[V Summer](vs ...V) {...}`。
- These type parameters can be used by the regular parameters and in the function body. 函数声明的泛型类型，函数体里也可以使用。
- Types can also have a type parameter list: type M[T any] []T. `type` 关键字也可以使用泛型，例如：`type Vector[T comparable] []T`。
- Each type parameter has a type constraint, just as each ordinary parameter has a type: func F[T Constraint](p T) { ... }. 每一个类型参数，都有约束，语法和变量声明一样，类型参数在前，约束在后，例如 `V Summer`。
- Type constraints are interface types. 类型参数的约束，其实底层是一个接口。
- The new predeclared name any is a type constraint that permits any type. 可以用 `any` 来表示任意类型，any的类型就是 `interface{}`。
- Interface types used as type constraints can embed additional elements to restrict the set of type arguments that satisfy the contraint: 类型约束的接口可以包含一些元素，来表示具体的约束
    - an arbitrary type T restricts to that type 比如一个具体的类型，表示必须符合这个约束，例如 `int` 表示必须是 int 类型。
    - an approximation element ~T restricts to all types whose underlying type is T 具体类型前加一个 `~` 表示底层类型符合这个就行，例如 `~int` 表示底层类型是 `int` 就行，因此 `type MyInt int` 的参数也可以用。
    - a union element T1 | T2 | ... restricts to any of the listed elements 可以用 `|` 来拼接多个类型，他们之间的关系是 `或`。
- Generic functions may only use operations supported by all the types permitted by the constraint. 泛型函数只能使用符合约束的操作。也就是说，约束里说明了支持啥操作，才能用啥操作，就跟接口一样。
- Using a generic function or type requires passing type arguments. 使用泛型的时候，要把具体参数传入，但是通常Go编译器可以推导出来，所以可以省掉，本来调用上面的 `sum` 函数得这样：`sum[int](ints...)` 但实际上通常都可以省略，写成 `sum(ints...)`
- Type inference permits omitting the type arguments of a function call in common cases. 通常编译器可以推导出类型，具体在上一条中说明了。

了解完了Go语言里，泛型的基本语法，剩下的就是愉快的使用了。如上面所说，正规的使用其实就是先声明泛型符合什么约束，不那么
正规的使用，其实可以直接把约束放到函数那里，比如 `sum` 函数可以改造成：

```go
func sum[V ~int | ~float64 | ~string](vs ...V) V {
	var total V

	for _, v := range vs {
		total += v
	}

	return total
}
```

## 总结

以上就是Go泛型你需要了解的东西，剩余的就是Go是如何实现泛型的，在提案中，有
`Go` 和 `Rust`、`Java`、`C++` 的对比，具体的实现还得看编译器的源码，据说是
由编译器来为每一种类型产生一份代码，然后再删掉无用的代码。但是具体还得去翻
一下编译器源码才知道，此处就不赘述了。

最后我想提一点的是，泛型的使用。毫无疑问，在容器类型(即各种数据结构)、以及实现某些算法的时候，其实他们的抽象层次是够高的，与具体类型关系不大，所以使用
泛型可以减少很多重复代码，比如 `sort` 库，就可以改造成泛型实现，就不用像现在
这样，提供各种 `sort.Ints`，`sort.Slice` 等等。

但是，这并不等同于可以滥用泛型，不得不说，使用泛型的代码，可读性会降低，因为
增加了读者的心智负担，尤其是当泛型类型很多的时候。业务代码中，应该视情况使用，而不能滥用。

最后，赶快下载 Go 1.18 一起体验泛型吧！

---

参考资料：

- https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md
- https://go.dev/doc/tutorial/generics
