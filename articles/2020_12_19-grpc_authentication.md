# gRPC鉴权方案

昨日与群友讨论，gRPC比较优雅的鉴权方案应该怎么做，本文记录一下最终相对优雅的实践方案。

不过我们仍然要从故事的开头讲起。

## 起因

我在使用gRPC的时候，一开始是不想使用header的，尽量避免往gRPC里修改一些东西，因此最开始，我是想着，在每个请求的message里，
都带上access_token，例如：

```go
message PingPong {
    string access_token = 1;
}
```

因此，如果想要做认证的话，就得在每个gRCP实现的方法里，对 `access_token` 进行校验，例如：

```go
userID, err := getUserIDByAccessToken(req.AccessToken)
if err != nil {
    logrus.Errorf("failed to get userID by accessToken %s: %s", req.AccessToken, err)
    return nil, status.Errorf(codes.Unauthenticated, "AccessToken expired")
}
```

但是由于Go既没有像Python那样灵活的decorator，可以在函数执行前执行一些代码，又没有像Java那样的annotation可以注入一些
metadata以便根据这些metadata来执行一些操作。所以导致这种方案必须在每一个gRPC方法里都加上上面这样的代码，这就很麻烦。

## 解决方案1，注入metadata

要想在gRPC里，针对每一个方法都做一点事情，那么最简单的，自然是使用中间件，在gRPC里，叫做 `interceptor` 。gRPC是可以注入
metadata的，链接见 [这里](https://developers.google.com/protocol-buffers/docs/proto3#options)，gRPC提供的options，其中有
一种就叫做 `google.protobuf.MethodOptions`:

```go
extend google.protobuf.MethodOptions {
  optional MyMessage my_method_option = 50007;
}
```

然后就可以使用它：

```go
service MyService {
  option (my_service_option) = FOO;

  rpc MyMethod(RequestType) returns(ResponseType) {
    // Note:  my_method_option has type MyMessage.  We can set each field
    //   within it using a separate "option" line.
    option (my_method_option).foo = 567;
    option (my_method_option).bar = "Some string";
  }
}
```

不过我最终没有选择这种方案，原因是文档极少，而且需要在中间件里把方法的options拿出来，但是中间件函数的签名是：

```go
// UnaryHandler defines the handler invoked by UnaryServerInterceptor to complete the normal
// execution of a unary RPC. If a UnaryHandler returns an error, it should be produced by the
// status package, or else gRPC will use codes.Unknown as the status code and err.Error() as
// the status message of the RPC.
type UnaryHandler func(ctx context.Context, req interface{}) (interface{}, error)

// UnaryServerInfo consists of various information about a unary RPC on
// server side. All per-rpc information may be mutated by the interceptor.
type UnaryServerInfo struct {
	// Server is the service implementation the user provides. This is read-only.
	Server interface{}
	// FullMethod is the full RPC method string, i.e., /package.service/method.
	FullMethod string
}

// UnaryServerInterceptor provides a hook to intercept the execution of a unary RPC on the server. info
// contains all the information of this RPC the interceptor can operate on. And handler is the wrapper
// of the service method implementation. It is the responsibility of the interceptor to invoke handler
// to complete the RPC.
type UnaryServerInterceptor func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (resp interface{}, err error)
```

想要拿到option不容易。不过，这倒是引出了方案2。

## 解决方案2: 手动维护白名单/黑名单

经群友建议，我们可以把该要鉴权，或者不需要鉴权的方法，根据方法名放在一个列表（我使用的是map）里。看上面的 `UnaryServerInfo`，
此方法可行。还有，与此配合的就是，把 `access_token` 放在metadata里（其实就是HTTP/2 里的header）传输。

我这里是大部分接口都需要鉴权，因此我把不需要鉴权的接口放在map里：

```go
// 所有不需要用户登录的接口，就放在这里，否则则必须登录
var publicAPIMapper = map[string]bool{
	"/cashapp.CashApp/PingPong": true,
	"/cashapp.CashApp/Register": true,
	"/cashapp.CashApp/Login":    true,
}

func IsPublicAPI(fullMethodName string) bool {
	return publicAPIMapper[fullMethodName]
}
```

然后就可以在interceptor里做统一处理：

```go
func SentryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		fullMethodName := info.FullMethod

		if !utils.IsPublicAPI(fullMethodName) {
			accessToken := utils.GetAccessToken(ctx)
			userID, err := utils.GetUserIDByAccessToken(accessToken)
			if err != nil || userID == 0 {
				logrus.Infof("failed to find user by %s: %s", accessToken, err)
				return nil, status.Errorf(codes.Unauthenticated, err.Error())
			}

			ctx = context.WithValue(ctx, "access_token", accessToken)
			ctx = context.WithValue(ctx, "user_id", userID)
		}

		logrus.Infof("got request with %v", req)
		result, err := handler(ctx, req)
		if err != nil {
			sentry.CaptureException(fmt.Errorf("%v", err))
			sentry.Flush(time.Second * 5)
		}
		return result, err
	}
}
```

其中 `utils.GetAccessToken` 是从ctx里拿metadata，然后从中拿 `access-token` 的，代码如下：

```go
func GetAccessToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	accessTokenList := md.Get("access-token")
	if len(accessTokenList) == 1 {
		return accessTokenList[0]
	}

	return ""
}
```

这个时候，客户端只需要统一添加一个 `Access-Token` 的头部，值为对应用户的 `access_token` 即可。

> 注意踩坑，Nginx会把头部做一次转换，比如 `Access-Token` 会转换成 `access-token`(HTTP/2 规范要求，
> 详见RFC：https://tools.ietf.org/html/rfc7540#section-8.1.2 )，而带下划线的头，默认会过滤掉，
> 详见：http://nginx.org/en/docs/http/ngx_http_core_module.html#underscores_in_headers ，我一开始传
> 头部传了 `access_token`，所以排查了好一会儿。

## 总结

上面我们看了两种gRPC做鉴权的方案，首先我们还是要承认，access_token这种东西，放在头部还是比较合适的，
并不需要为了不动gRPC的metadata而去迁就。其次，简单粗暴的方案也可以是相对优雅的，所以最后我们选择了
实现上更简单但是代码可读性，维护行好很多的方案2。

---

ref:

- https://developers.google.com/protocol-buffers/docs/proto3#options
- https://developers.google.com/protocol-buffers/docs/proto#customoptions
