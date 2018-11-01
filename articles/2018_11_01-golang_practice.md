# Golang 实践经验

- 编码风格：https://github.com/golang/go/wiki/CodeReviewComments
- 编码规范：代码提交前使用 go fmt 格式化代码
- 虽然Go推荐使用比较短的名字来命名，但是不要太短，例如l, a, r，容易看不懂，尤其是当嵌套层次深了之后。例如gRPC实现中的一段代码：

```go
frame, err := t.framer.fr.ReadFrame()
if err == io.EOF || err == io.ErrUnexpectedEOF {
    return nil, err
}
```

如果对实现不够熟悉，很难知道t是什么，fr是什么。

- 使用linter: go vet，或者 https://godoc.org/golang.org/x/lint
- 使用 https://github.com/pkg/errors 代替标准库中的errors: https://banzaicloud.com/blog/error-handling-go/
- 有逃逸分析，不要滥用指针，否则代价是GC，而GC则是影响Go高性能的常见原因：http://www.agardner.me/golang/garbage/collection/gc/escape/analysis/2015/10/18/go-escape-analysis.html
- 如果是Web应用，尽可能的遵守：https://12factor.net/
