# 避免全局变量

> 睡觉前好像确实不应该想问题，大半夜的思维活跃睡不着了。当然，那只臭蚊子也有功劳，要是被我发现了我要灭了它。
> 不过，既然睡不着，那就起来写篇博客。

最近突然想到以前的一个项目，一个用来做token认证的微服务，当时为了快速实现，没有严格遵守MVC，很多controller里就有类似的
代码：

```go
func XXXHandler() {
    db.Where("xxx = ?", 123).Find(&User)
}
```

后来同事接入opentracing的时候，就很痛苦。当然了，来新公司之后也写过类似的代码，主要是之前没有想到特别好的解决方案，以及
严格遵守MVC的必要性。Go的ORM实在是太难用了，以至于无法完全的将对象和数据库表解耦，如你所见，代码里还到处都是SQL的影子。
如果是SQLAlchemy还真的很难看出这样做有什么不好。像上面的代码，至少有这么几个坏处：

- 暴露了具体SQL实现给外界，此处的外界是Controller。因为Go的ORM特别难用，里面嵌入了大量的SQL语句，所以其实是和具体数据库
强相关的。也就是说，这样以来，假设啥时候要改个数据库，那就完蛋了，因为到处充斥着这样的代码，手都能改断。

- 没有重复利用代码。举个例子，根据 `user_id` 拿 `User` 信息的代码，肯定到处都需要。如果所有的地方都是直接 `db.Where(xxx)`
这样的用法，就会造成和上面一条说到的一样的问题。

- 无法对数据库操作进行一些特定的，统一的操作。举个例子，加tracing。如果我们把提取数据的函数封装在M里，那么我们在每个方法
里加一行 `defer BlablaTracing()` 就可以达到我们的目的，但是像上面那样，就不好办了。

所以正确的方法应该是，遵循MVC。把数据库操作封装到M里，例如，model层这样写：

```go
var db sql.DB // 不暴露db出去，把db限制在model这个包里

type User struct{}

func GetUserByID(id uint32) (*User, error) {
	user := User{}
	if err := db.Where("id = ?", id).Find(&user).Error; err != nil {
		return nil, err
	} else {
		return user, nil
	}
}
```

然后，controller里这样写：

```go
func XXXHandler() {
    user, err := GetUserByID(user_id)
    xxxxxx
}
```

所以说，有些懒，偷不得。为了不让同事想开车从你身上碾几遍，还是好好设计，好好想好少挖坑吧😁
