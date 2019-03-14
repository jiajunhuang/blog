# Golang 的槽点

Golang虽然说打着简洁的名号，但是很多设计实际上并不简洁，且由于Go1兼容性保证的原因，这些都不能被修正。

## 声明方式太多

- `var x int = 1`
- `var x = 1`
- `var x int; x = 1`
- `var x = int(1)`
- `x := 1`

## 初始化方式太多

- `x := MyStruct{}`
- `var x MyStruct; x = MyStruct{}`
- `x := new(MyStruct)`

对于map, slice, channel还要算上 `make`。

## 糟糕的错误处理

`if err != nil` 这绝对是Gopher最常见的代码之一。怀念大Python的 `try...except...`

## 多余的返回值变量声明

`func Bla() (i int, err error) {}` 有时候看一个变量到底是什么类型，结果一路往上看都没找到声明或者初始化的地方，结果最后发现
是在返回值那里进行声明的。这种声明方式加剧了Golang的不易读性。而这种方式唯一的作用就是，当你需要在defer里修改返回值时，就
需要这种奇怪的返回方式。

## 没有泛型

这就导致很多代码必须与interface搭上勾，这样做就失去了类型安全。如果想要类型安全，那么就得为每一种类型写上一份相同的代码。

## 没有好用的ORM

上述的种种底层原因，就导致了业务上的不好用之处，没有一个好用的ORM。不过没关系，sqlx够用。

## 建议

我一直认为，统一和简洁是非常重要的，因此个人遵循如下原则

- 不用 `new` ，除了 `channel` 之外，也不使用make。统一使用 `MyStruct{}`, `map[string]interface{}{}` 这样的方式来初始化
- 对于同一个block内多次使用的变量，在最顶部使用 `var x int` 进行声明
- 对于需要指定类型的变量，使用 `vat x uint = 1` 进行声明并初始化
- 等Go2的错误处理和泛型

---

- https://dave.cheney.net/practical-go/presentations/qcon-china.html
