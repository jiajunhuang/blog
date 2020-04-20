# 善用闭包(closure)让Go代码更优雅

通常来说我们降低一个函数的复杂度的方法是拆分。即大事化小，各个击破的原理。不过拆分成函数调用有一个问题，那就是修改
函数参数的时候很蛋疼。

比如，原本由于各种原因我们有一个巨长的函数，他的作用是发工资：

```go
func loooooooooooog(userID int) {
    // 比如1000行代码
}
```

因为我们要拆分成若干个小函数，这自然是一件好事情，降低了变量之间的依赖：

```go
func loooooooooooog(userID int) {
    checkUser(userID)
    checkSalary(userID)
    Send(userID, salary)
    Notify(userID)
}
```

但是现在我们决定引入 `Context`，用以随时取消Goroutine的执行，当然啦，在Go里没法自动取消，只能手动档：

```go
func checkUser(ctx context.Context, userID int) error {
    select {
        case <-ctx.Done():
            err := ctx.Err()
            log.Printf("ctx is done for %s", err)
            return err
        default:
    }

    // blablabla 业务代码
}
```

同样的代码还会出现在若干个子函数中，例如 `checkSalary`, `Send`, `Notify`等等。难道有谁愿意写这么多重复的代码吗？
至少我不愿意，所以捏，我决定创造一个函数，取缔这些重复的context检查，幸好Go支持闭包，要不然就不好弄了：

```go
// Execute execute fn until ctx.Done() is received
func Execute(ctx context.Context, fn func()) error {
    select {
    case <-ctx.Done():
        err := ctx.Err()
        log.Printf("ctx is done for %s", err)
        return err
    default:
        return fn()
    }
}
```

那么接下来我就可以这样用了：

```go
func loooooooooooog(ctx context.Context, userID int) {
    Execute(ctx, func() error {/*业务代码在这里*/})
    Execute(ctx, func() error {/*业务代码在这里*/})
}
```

不过，没有银弹。这样做又带来了一个副作用，由于闭包是可以访问外层函数的变量的，也就意味着还是有可能产生变量混用的可能性，
但是我认为这个可以人为消除，也就是说，写的时候注意，不是公共变量的，不要乱用，尽量在匿名函数内使用内部变量，当然也不是
没有解决方案，那就是外边包装一层匿名函数+调用，不过那样的缺点就是更复杂了，没有关系，我们手动控制好即可。

所以说是 "善用"闭包，任何东西，都要拿捏好，不能滥用的。

---

参考资料：

- https://github.com/jiajunhuang/gotasks/blob/master/loop/loop.go
