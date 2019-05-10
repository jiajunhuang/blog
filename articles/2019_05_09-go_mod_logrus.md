# go mod 和 logrus 路径大小写的问题

前段时间遇到了这么个问题：

```bash
$ go get -u
...
parsing go.mod: unexpected module path "github.com/sirupsen/logrus"
...
```

原因就是，sirupsen 老大哥的名字从 `Sirupsen` 变成了 `sirupsen`，而go的库写的是包的URL路径。
当依赖的依赖使用的是老版本的时候，就要找不到这个包了。解决方案是在 `go.mod` 最后加上：

```go
replace (
    github.com/Sirupsen/logrus v1.4.1 => github.com/sirupsen/logrus v1.4.1
)
```
