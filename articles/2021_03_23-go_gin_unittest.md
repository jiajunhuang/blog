# Golang GIN写单测时，愉快的使用返回值

我写的接口，基本长这样：

```js
{
    "code": 200,
    "msg": "原因",
    "result": {} // 或者空或者其它
}
```

所以在Go里，定义如下：

```go
type Resp struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result"`
}
```

然后，在返回值时，就可以定义好结构，比如：

```go
type UserResult struct {
	UserID int `json:"user_id"`
}

```

然后把值塞进去。这样就愉快的返回了结果，不过，单测的时候可咋办呢？因为我想判断 `result` 里的值。最开始我尝试了如下方案：

```go
func bodyToStruct(byteArray []byte, s interface{}) {
	err := json.Unmarshal(byteArray, &s)
	if err != nil {
		log.Printf("failed to unmarshal %s: %s", byteArray, err)
	}
}
```

这是一个把 []byte 转换成对应结构体的帮助函数，不过，result是不好转的，一开始我有尝试这样：

```go
resp := Resp{Result: UserResult{}}
```

但是没有用，拿出来以后，`Result` 的值是一个map，于是我采用了曲线救国的方式，加一个工具函数，先把 `Result` 的值 marshal，
然后unmarshal到对应的结构体：

```go
func resultToStruct(result, s interface{}) {
    byteArray, err := json.Marshal(result)
    if err != nil {
        logrus.Errorf("failed to marshal %s: %s", result, err)
        return
    }
 
    err = json.Unmarshal(byteArray, s)
    if err != nil {
        logrus.Errorf("failed to unmarshal %s: %s", byteArray, err)
        return
    }
}
```

很明显，有点low，转来转去。经过群友的提醒，是我最上面应该传指针，于是改成这样就可以了：

```go
resp := Resp{Result: &Result}
```

这样就可以愉快的把值一次性通过 `bodyToStruct` 放到对应的结构体里，但是，要怎么样才能愉快的取值呢？因为即便做到了刚才那样，
如果你直接 `resp.Result.UserID` 还是不行，因为 `Resp.Result` 的定义是一个 `interface{}`，解决方案如下：

```go
resp := Resp{}
result := Result{}
resp.Result = &result

bodyToStruct(bytes, &resp)
```

于是就可以了。看看demo代码：

```go
package main

import (
	"encoding/json"
	"log"
)

type Resp struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result"`
}

type UserResult struct {
	UserID int `json:"user_id"`
}

func bodyToStruct(byteArray []byte, s interface{}) {
	err := json.Unmarshal(byteArray, &s)
	if err != nil {
		log.Printf("failed to unmarshal %s: %s", byteArray, err)
	}
}

func main() {
	byteArray := []byte(`{"code":200,"msg":"","result":{"user_id": 1}}`)

	resp := Resp{}
	result := UserResult{}
	resp.Result = &result

	bodyToStruct(byteArray, &resp)
	log.Printf("user id is %d", result.UserID)
}
```

于是，就终于可以愉快的类型安全的使用返回值了。这样做有几个好处：

- 类型安全
- 可以复用定义接口时写的struct
- 可以补全

有同学可能认为，为啥不直接用interface，断言？嗯，其实我曾经年少轻狂的时候，经常说我是interface走天下，直到我的程序不断的
崩崩崩，我才老老实实重新做人，老老实实的类型安全。

这就是这次想要分享的技巧。
