# gRPC-gateway 源码阅读

> https://github.com/grpc-ecosystem/grpc-gateway

首先看一下要怎么用这个库，README里写着：

```bash
protoc -I/usr/local/include -I. \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    --grpc-gateway_out=logtostderr=true:. \
    path/to/your_service.proto
```

搜索 `protoc plugin` 可以得到：https://developers.google.com/protocol-buffers/docs/reference/cpp/google.protobuf.compiler.plugin

可以得到这几个结论：

- `protoc --plugin=protoc-gen-NAME=path/to/mybinary --NAME_out=OUT_DIR` 其中NAME是插件的名字，`=`后边接的是二进制的路径
- plugin 接受一个 `CodeGeneratorRequest`，返回一个 `CodeGeneratorResponse`

> https://github.com/google/protobuf/blob/master/src/google/protobuf/compiler/plugin.proto

从README中可以看到我们是这样安装 grpc-gateway 的： `go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway`

然后我们就到 `protoc-gen-grpc-gateway` 中看 `main.go`:

```go
func main() {
	flag.Parse()
	defer glog.Flush()

	reg := descriptor.NewRegistry()

	glog.V(1).Info("Parsing code generator request")
	req, err := codegenerator.ParseRequest(os.Stdin)
	if err != nil {
		glog.Fatal(err)
	}
	glog.V(1).Info("Parsed code generator request")
	if req.Parameter != nil {
		for _, p := range strings.Split(req.GetParameter(), ",") {
			spec := strings.SplitN(p, "=", 2)
			if len(spec) == 1 {
				if err := flag.CommandLine.Set(spec[0], ""); err != nil {
					glog.Fatalf("Cannot set flag %s", p)
				}
				continue
			}
			name, value := spec[0], spec[1]
			if strings.HasPrefix(name, "M") {
				reg.AddPkgMap(name[1:], value)
				continue
			}
			if err := flag.CommandLine.Set(name, value); err != nil {
				glog.Fatalf("Cannot set flag %s", p)
			}
		}
	}

	g := gengateway.New(reg, *useRequestContext, *registerFuncSuffix, *pathType)

	if *grpcAPIConfiguration != "" {
		if err := reg.LoadGrpcAPIServiceFromYAML(*grpcAPIConfiguration); err != nil {
			emitError(err)
			return
		}
	}

	reg.SetPrefix(*importPrefix)
	reg.SetImportPath(*importPath)
	reg.SetAllowDeleteBody(*allowDeleteBody)
	if err := reg.Load(req); err != nil {
		emitError(err)
		return
	}

	var targets []*descriptor.File
	for _, target := range req.FileToGenerate {
		f, err := reg.LookupFile(target)
		if err != nil {
			glog.Fatal(err)
		}
		targets = append(targets, f)
	}

	out, err := g.Generate(targets)
	glog.V(1).Info("Processed code generator request")
	if err != nil {
		emitError(err)
		return
	}
	emitFiles(out)
}
```

其中：

- `req, err := codegenerator.ParseRequest(os.Stdin)` 跳进去看，是从标准输入读取参数，然后返回一个 `*plugin.CodeGeneratorRequest`
对象。就是上面的proto文件中的CodeGeneratorRequest
- `g := gengateway.New(reg, *useRequestContext, *registerFuncSuffix, *pathType)` 生成一个 `gen.Generator` 对象
- `out, err := g.Generate(targets)` 生成代码
- 跟进去，`func (g *generator) Generate(targets []*descriptor.File) ([]*plugin.CodeGeneratorResponse_File, error)` 函数的实现，
发现调用了 `code, err := g.generate(file)`
- 调用了 `func (g *generator) generate(file *descriptor.File) (string, error)`, 调用了 `protoc-gen-grpc-gateway/gengateway/template.go`
中的代码
- template.go 下面就是我们要生成的代码的模板

如果生成一个demo看看，就知道，模板做的事情就是每接受一个HTTP请求，就生成一个gRPC请求。
