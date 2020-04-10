# Golang validator使用教程

validator应该是Golang里进行表单校验的事实标准了，比如在Web框架GIN中，就是默认使用它。表单校验的作用，就是对输入的数据
进行合法判断，如果是不合法的，那么就输出错误。比如：

```go
package main

import (
        "log"

        "github.com/go-playground/validator/v10"
)

type MyStruct struct {
        FirstName string `json:"firstname" validate:"required"`
}

func main() {
        v := validator.New()

        s := MyStruct{"blabla"}
        err := v.Struct(s)
        log.Printf("%+v", err)

        s2 := MyStruct{}
        err = v.Struct(s2)
        log.Printf("%+v", err)
}
```

执行结果为：

```bash
$ go run main.go
2020/04/10 21:28:41 <nil>
2020/04/10 21:28:41 Key: 'MyStruct.FirstName' Error:Field validation for 'FirstName' failed on the 'required' tag
```

而validator基本的用法，其实也就和上面的类似，只不过，除了 `v.Struct` 之外，还有好几个：

- `func (v *Validate) Struct(s interface{}) error` 接收的参数为一个struct
- `func (v *Validate) StructExcept(s interface{}, fields ...string) error` 校验struct中的选项，不过除了fields里所给的字段
- `func (v *Validate) StructFiltered(s interface{}, fn FilterFunc) error` 接收一个struct和一个函数，这个函数的返回值为bool，决定是否跳过该选项
- `func (v *Validate) StructPartial(s interface{}, fields ...string) error` 接收一个struct和fields，仅校验在fields里的值
- `func (v *Validate) Var(field interface{}, tag string) error` 接收一个变量和一个tag的值，比如 `validate.Var(i, "gt=1,lt=10")`
- `func (v *Validate) VarWithValue(field interface{}, other interface{}, tag string) error` 将两个变量进行对比，比如 `validate.VarWithValue(s1, s2, "eqcsfield")`

> 上述方法均有另外一种形式，就是带上context的那种，函数名字就是上述函数名，最后加一个 `Ctx`。

了解了这些之后，我们还需要了解的一个东西就是，在Golang中，我们使用 `struct tag` 来定义表单合法的值。比如如果我们希望某个
字段是邮件，那么可以这样定义：

```go
type MyStruct struct {
    Email string `validate:"email"`
}
```

如果有多个校验条件，可以用英文逗号 `,` 来进行连接，他们是 `AND` 的关系，如果想要 `OR` 的关系，那么使用 `|`，比如：

```go
package main

import (
        "log"

        "github.com/go-playground/validator/v10"
)

type MyStruct struct {
        Age int `json:"age" validate:"lt=10|gt=20"`
}

func main() {
        v := validator.New()

        s := MyStruct{15}
        err := v.Struct(s)
        log.Printf("%+v", err)

        s2 := MyStruct{9}
        err = v.Struct(s2)
        log.Printf("%+v", err)
}
```

接着我们来看看常见的校验写法：

- `email` 邮件
- `url` 链接
- `json` JSON
- `file` 文件路径
- `base64` Base64
- `containsany=!@#?`, `contains=@`, `containsrune=@` 都表示包含
- `excludes=@`, `excludesall=!@#?`, `excludesrune=@` 都表示不包含
- `startswith=hello` 和 `endswith=goodbye` 表示字符串的起始和结束是否等于值
- `latitude` 和 `longitude` 分别表示是否是纬度和经度
- `ip`, `ipv4`, `ipv6` 分别表示对应的IP地址类型
- `datetime=2006-01-02` 表示是否是这种格式的日期字符串
- `uuid`, `uuid3`, `uuid4`, `uuid5` 表示是否是对应的UUID类型
- `lowercase` 和 `uppercase` 表示是否是对应的小写或者大写
- `gt`, `lt` 分别是大于和小于，`eq` 表示等于，如果是大于等于，那么是 `gte`，小于等于则是 `lte`
- `oneof='red green' 'blue yellow'`, `oneof=5 7 9` 表示是其中一个值
- `min`, `max` 表示其值是否满足此表达式
- `required` 表示此项是必填的，不能为0值
- `-` 忽略此属性，即不校验此属性

这就是常见的标签的用法和意思，快用起来吧！

---

参考资料：

- https://godoc.org/gopkg.in/go-playground/validator.v10
