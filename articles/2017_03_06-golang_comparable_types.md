# 什么是 Golang Comparable Types

Golang 中有很多时候要用到comparable types，例如比较struct的时候，例如map里的
key。弄清楚有哪些类型是comparable的非常重要。

> https://golang.org/ref/spec#Comparison_operators

- Boolean values
- Integer values
- Floating point values
- Complex values
- String values
- Pointer values
- Channel values
- Interface values
- Struct values are comparable if all their fields are comparable
- Array values are comparable if values of the array element type are comparable
- A value x of non-interface type X and a value t of interface type T
are comparable when values of type X are comparable and X implements T
