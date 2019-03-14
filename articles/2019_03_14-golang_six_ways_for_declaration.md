# Golang 的槽点

Golang虽然说打着简洁的名号，但是很多设计实际上并不简洁，且由于Go1兼容性保证的原因，这些都不能被修正

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

## 建议

我一直认为，统一和简洁是非常重要的，因此个人遵循如下原则

- 不用 `new` ，除了 `channel` 之外，也不使用make。统一使用 `MyStruct{}`, `map[string]interface{}{}` 这样的方式来初始化
- 对于同一个block内多次使用的变量，在最顶部使用 `var x int` 进行声明
- 对于需要指定类型的变量，使用 `vat x uint = 1` 进行声明并初始化

---

- https://dave.cheney.net/practical-go/presentations/qcon-china.html
