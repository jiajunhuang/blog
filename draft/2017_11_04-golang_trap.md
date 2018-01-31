# Golang的一些坑

- 传给 `signal.Notify` 的channel必须是一个buffered channel, 否则收不到信号

- channel默认是unbuffered channel, 因此在没有消费者之前, 放入channel的动作都会被阻塞, 例如:

```go
func main() {
    c := make(chan int)

    for i := 0; i < 3; i++ {
        go func() {
            c <- 1
        }()
    }

    fmt.Println(<-c)
}
```

此函数退出时,会有两个goroutine被阻塞在channel上, 然而gc不会回收. 因此, 如果大量出现这种情况, 将会导致goroutine leak.

- `for...range` 语句会在执行之前执行一次。而且所有值都是拷贝，而非返回指针。

    - https://golang.org/ref/spec#For_statements
    - https://github.com/golang/go/wiki/Range
    - https://garbagecollected.org/2017/02/22/go-range-loop-internals/
    - https://github.com/golang/gofrontend/blob/master/go/statements.cc#L5343
