# Go设计模式：单例模式、原型模式和Builder模式

这篇文章记录三个设计模式，因为他们都比较简单，因此较短的篇幅就可以描述完，就把这三个放在一起。

## 单例模式

单例模式，就是为了确保全局唯一。在Go语言里实现单例模式，好像也没啥好办法，一般就是：

- 全局变量
- sync.Once

然而我之前写过 [Python中实现单例模式的四种方式](https://jiajunhuang.com/articles/2018_08_24-python_singleton.md.html)，
究其本质，其实也是通过全局变量来实现的，要么就是全局变量，要么就是类变量，而类本身也是全局唯一的。

## 原型模式

在Go语言里，我倒是很少见到使用原型模式，原型模式是这样一种情况：通常来说我们新建一个对象都是直接实例化比如：

- `new(SomeStruct)`
- `make(SomeStruct)`
- `SomeStruct{}`

但是原型模式并不直接通过类或者结构体来实例化，而是通过一个实例对自身进行clone来得到一个新的实例(其实一般情况也就是clone
方法自己偷偷的实例化了一个对象然后把属性copy过去)，原型模式和直接实例化的最大区别就是通过原型模式，可以直接把实例clone时
自身的状态也一起copy过去。

我从来没直接用过原型模式，不过GORM里有，我们来看看他是怎么实现的：

```go
func (stmt *Statement) clone() *Statement {
	newStmt := &Statement{
		Table:                stmt.Table,
		Model:                stmt.Model,
		Dest:                 stmt.Dest,
		ReflectValue:         stmt.ReflectValue,
		Clauses:              map[string]clause.Clause{},
		Distinct:             stmt.Distinct,
		Selects:              stmt.Selects,
		Omits:                stmt.Omits,
		Joins:                map[string][]interface{}{},
		Preloads:             map[string][]interface{}{},
		ConnPool:             stmt.ConnPool,
		Schema:               stmt.Schema,
		Context:              stmt.Context,
		RaiseErrorOnNotFound: stmt.RaiseErrorOnNotFound,
	}

	for k, c := range stmt.Clauses {
		newStmt.Clauses[k] = c
	}

	for k, p := range stmt.Preloads {
		newStmt.Preloads[k] = p
	}

	for k, j := range stmt.Joins {
		newStmt.Joins[k] = j
	}

	return newStmt
}
```

瞧，他就是新建一个，然后把属性copy过去。

## Builder模式

Builder模式适用于这么一种情况：无法或不想一次性把实例的所有属性都给出，而是要分批次、分条件构造，举个例子，不是这样实例化：

```go
a := SomeStruct{1, 2, "hello"}
```

而是这样：

```go
a := SomeStruct{}
a.setAge(1)
a.setMonth(2)
if (blabla) {
    a.setSlogan("hello")
}
```

这种模式的一个用处就是，上古时期的动态网站就是靠Builder模式来生成HTML的，大家都这么玩：

```go
a := emptyPage{}
a.addTag("p", "balblabla")
a.addTag("br")
```

可能最后就会生成这么一个HTML：

```html
<p>balblabla</p>
<br />
```

可能有人要问了，为啥不直接初始化实例的时候，把属性放进去呢？

- 可能初始化的时候不知道有啥，例如上面的例子，生成HTML，可能会需要根据某种条件，生成不同的HTML插进去
- 如果直接通过初始化实例属性，而不是各种 `setXXX` 方法的话，那就等于暴露了结构体内部构造，也就使得其他代码与结构体成员耦合（实际上一般都没啥问题）

Builder模式除了上面例子中的形态，还有一种变种，那就是链式：

```go
a := SomeStruct{}
a = a.setAge(1).setMonth(2).setSlogan("hello")
```

那这是怎么实现的呢？其实就是在每一个函数的最后，把实例自身返回。那Builder模式在哪里有用到呢？比如 [go-resty](https://github.com/go-resty/resty):

```go
// Create a Resty Client
client := resty.New()

resp, err := client.R().
      SetQueryParams(map[string]string{
          "page_no": "1",
          "limit": "20",
          "sort":"name",
          "order": "asc",
          "random":strconv.FormatInt(time.Now().Unix(), 10),
      }).
      SetHeader("Accept", "application/json").
      SetAuthToken("BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F").
      Get("/search_result")


// Sample of using Request.SetQueryString method
resp, err := client.R().
      SetQueryString("productId=232&template=fresh-sample&cat=resty&source=google&kw=buy a lot more").
      SetHeader("Accept", "application/json").
      SetAuthToken("BC594900518B4F7EAC75BD37F019E08FBC594900518B4F7EAC75BD37F019E08F").
      Get("/show_product")
```

就是通过Builder模式来构造请求的。
