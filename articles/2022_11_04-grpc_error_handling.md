# Golang gRPC 错误处理

gRPC 最常见的错误处理方式就是直接返回错误，例如 `return nil, err`，但是实际业务中，我们还有业务码需要返回，常见的方式
是在返回的结构体里定义一个错误码，但是这样写起来又很是麻烦，例如，你可能需要这样写：

```go
user, err := dao.GetUserByEmail(ctx, email)
if err != nil  {
    if err == gorm.RecordNotFound {
        return &GetUserResp{Code: USER_NOT_FOUND, Msg: "user not found"}, nil
    }
    return nil, err
}
```

这里有几个问题：

1. 返回错误写起来很麻烦，因为需要每次都判断错误，然后转换成对应的错误码写在 Code, Msg 两个字段中
2. 如果直接返回 `err`，而不是 grpc 自定义的 `codes.NotFound` 等错误，无法在客户端中进行识别
3. 如果使用了 gRPC Gateway，对于返回了非 grpc 自定义错误的地方，统统会表示成500

上面几个问题，分开来都有解法，例如对于1，可以直接返回err，但是会导致问题2；对于问题2，可以使用1，但是写起来麻烦；对于
问题3，可以使用grpc内置的错误，但是表达能力非常受限，无法传达业务错误码。

因此，为了解决这一系列问题，在比较多个错误处理库之后，我们整理了一整套结合他们优点同时又适配业务需求的错误处理体系。

## 错误处理库

Python的异常体系是一个非常值得借鉴的设计。首先我们会将程序的非正常执行分为错误和异常，在Go语言中，错误是我们希望能够
进行检查和处理的，而异常是只能通过 recover 尝试进行恢复的。

我们首先将错误，分为错误的类型，和错误的实例。定义错误时，我们定义的是错误的类型，这里就携带了它所应该展示的HTTP状态码
和业务错误码。当抛出错误时，也就是实例化错误的时候，此时携带上错误的栈信息、执行信息等。

例如，定义错误：

```go
ErrBadRequest       = RegisterErrorType(BaseErr, http.StatusBadRequest, ErrCodeBadRequest)             // 400
ErrUnauthorized     = RegisterErrorType(BaseErr, http.StatusUnauthorized, ErrCodeUnauthorized)         // 401
ErrPaymentRequired  = RegisterErrorType(BaseErr, http.StatusPaymentRequired, ErrCodePaymentRequired)   // 402
ErrForbidden        = RegisterErrorType(BaseErr, http.StatusForbidden, ErrCodeForbidden)               // 403
ErrNotFound         = RegisterErrorType(BaseErr, http.StatusNotFound, ErrCodeNotFound)                 // 404
ErrMethodNotAllowed = RegisterErrorType(BaseErr, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed) // 405
```

实例化错误：

```go
err = validateReq(req)
if err != nil {
    return nil, errs.NewBadRequest(err.Error(), err)
}
```

检测错误类型：

```go
if errs.IsError(err, ErrBadRequest) {
    //
}
```

提取错误：

```go
if baseErr, ok := errs.AsBaseErr(err); ok {
    //
}
```

有了上面这一套错误库以后，我们就可以愉快的携带错误栈信息、错误类型、错误业务码、错误HTTP状态码、错误信息、导致错误发生
的元错误在整个代码体系中流转，并且还可以进行类型判断、信息提取。那么怎么和 gRPC 结合在一起呢？

## gRPC 错误处理

上面我们说过，如果直接使用 `return nil, err` 的形式，客户端无法准确识别，而如果使用 `return Resp{Code, Msg}, nil` 的形式，
写起来又很麻烦，而且 gRPC gateway 无法准确翻译成对应的 HTTP 状态码。

我们的解决方案就是，直接返回上一节描述的错误体系，例如：

```go
func (s *service) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserResp, error) {
    err = validateReq(req)
    if err != nil {
        return nil, errs.NewBadRequest(err.Error(), err)
    }
}
```

然后在中间件中，提取 `Resp` 并且将 `code` 和 `msg` 进行赋值：

```go
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        resp, err := handler(ctx, req)
        if err == nil {
            return resp, err
        }

        if errs.IsError(err, errs.BaseErr) {
            return resp, err
        }

        if val := reflect.ValueOf(resp); !val.IsValid() || val.IsNil() {
            tp := getRespType(ctx, info)
            if tp == nil {
                return resp, err
            }
            resp = reflect.New(tp).Interface()
        }

        if be, ok := errs.AsBaseErr(err); ok {
            grpc.SetHeader(ctx, metadata.Pairs("x-http-code", fmt.Sprintf("%d", be.HTTPCode())))
            return baseErrSetter(resp, be)
        }
    }
}
```

这样我们就可以将返回的错误自动序列化到Resp中对应的字段。

## gRPC gateway 状态码

如果上一步，我们处理完了错误之后，直接返回error，由于不是gRPC体系内的错误码，gRPC gateway会返回500，但如果我们返回nil，
gRPC gateway又会返回200，这两者都不符合预期。既然我们错误体系已经包含了HTTP状态码，是否可以直接使用呢？答案是是的，看上
面的代码中，最后我们设置了一个metadata `x-http-code`，我们可以在 gRPC gateway 中注册一个中间件，用这里传递的状态码：

```go
mux := runtime.NewServeMux(
    runtime.WithForwardResponseOption(GRPCGatewayHTTPResponseModifier),
)

func GRPCGatewayHTTPResponseModifier(ctx context.Context, w http.ResponseWriter, p proto.Message) error {
    md, ok := runtime.ServerMetadataFromContext(ctx)
    if !ok {
        return nil
    }

    // set http status code
    if vals := md.HeaderMD.Get(httpStatusCodeKey); len(vals) > 0 {
        code, err := strconv.Atoi(vals[0])
        if err != nil {
            return err
        }
        // delete the headers to not expose any grpc-metadata in http response
        delete(md.HeaderMD, httpStatusCodeKey)
        delete(w.Header(), grpcHTTPStatusCodeKey)
        w.WriteHeader(code)
    }

    return nil
}
```

这样，我们在 gRPC 中返回的是 `ErrBadRequest` 的实例，最终体现在 gRPC gateway 的响应中就会是400，返回的是 `ErrForbidden`，
在 gRPC gateway 中就会体现为 403，我们的目的就成功达成了。

## 监控

同时我们还提供了一套中间件，能够结合 sentry 收集错误栈。

## 总结

这一套整套体系，最终可以达到的效果是：

- 能结合 gRPC 与 HTTP，符合对应规范，能充分支持业务需求
- 错误有分级和分类，能组成错误树
- 能识别和判断类型，能包含足够多的信息，能够自定义错误和错误类型
- 能结合 sentry 和监控系统进行错误收集和监控
- 使用简单方便，通俗易懂
- 能够保持 grpc gateway 与 grpc 中状态码、错误码一致

---

ref:

- https://docs.python.org/3/tutorial/errors.html
- https://docs.python.org/3/library/exceptions.html
- https://grpc-ecosystem.github.io/grpc-gateway/docs/mapping/customizing_your_gateway/#controlling-http-response-status-codes
