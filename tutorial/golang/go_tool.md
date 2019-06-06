# Go tools

## 目录

- [安装](./installation_linux.md)
    - [Windows](./installation_windows.md)
    - [Linux](./installation_linux.md)
    - [macOS](./installation_mac_os.md)
- [Hello, World](./hello_world.md)
- [Go语言简介](./intro.md)
- [基本类型](./basic_types.md)
- [容器类型](./composite_types.md)
- [函数](./function.md)
- [流程控制](./flow.md)
- [错误处理](./errors.md)
- [面向对象编程](./oo.md)
- [面向接口编程](./interface.md)
- [指针](./pointers.md)
- [Goroutine](./goroutine.md)
- [Channel](./channel.md)
- [并发编程](./concurrency.md)
- [go tools](./go_tool.md)

在命令行里输入 `go` 我们就可以看到go提供的工具：

```bash
$ go
Go is a tool for managing Go source code.

Usage:

	go <command> [arguments]

The commands are:

	bug         start a bug report
	build       compile packages and dependencies
	clean       remove object files and cached files
	doc         show documentation for package or symbol
	env         print Go environment information
	fix         update packages to use new APIs
	fmt         gofmt (reformat) package sources
	generate    generate Go files by processing source
	get         download and install packages and dependencies
	install     compile and install packages and dependencies
	list        list packages or modules
	mod         module maintenance
	run         compile and run Go program
	test        test packages
	tool        run specified go tool
	version     print Go version
	vet         report likely mistakes in packages
```

## go build

`go build` 是用来编译Go代码的，常见用法是：

```bash
$ go build
$ go build -o main
```

其中 `-o main` 是用来指定编译出来的可执行文件名。

## go test

`go test` 是用来跑单元测试的，这个需要阅读这里：https://golang.org/pkg/testing/

## go vet

`go vet` 是用来检验代码中常见错误的，用法为： 

```bash
$ go vet ./...
```

## go mod

`go mod` 是go官方的依赖管理工具，可以参考这里：https://jiajunhuang.com/articles/2018_09_03-go_module.md.html

---

- 上一篇：[并发编程](./concurrency.md)
- 下一篇：恭喜你已经完成了 Go语言简明教程
