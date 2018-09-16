# Cookie 中的secure和httponly属性

Golang中 `net/http` 的Cookie结构体：

```go
// A Cookie represents an HTTP cookie as sent in the Set-Cookie header of an
// HTTP response or the Cookie header of an HTTP request.
//
// See https://tools.ietf.org/html/rfc6265 for details.
type Cookie struct {
	Name  string
	Value string

	Path       string    // optional
	Domain     string    // optional
	Expires    time.Time // optional
	RawExpires string    // for reading cookies only

	// MaxAge=0 means no 'Max-Age' attribute specified.
	// MaxAge<0 means delete cookie now, equivalently 'Max-Age: 0'
	// MaxAge>0 means Max-Age attribute present and given in seconds
	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite SameSite
	Raw      string
	Unparsed []string // Raw text of unparsed attribute-value pairs
}
```

其中有 `Secure` 和 `HttpOnly` 两个属性，我们打开结构体上附加的注释，可以查阅到：

- `Secure` 的作用是设置为True时，只有走HTTPS时才会带上此Cookie，如果是HTTP，则不会带上。
- `HttpOnly` 的作用是设置为True时，只有和服务器的HTTP请求会带上此Cookie，如果是AJAX请求，则不会带上。

-----------------

- https://tools.ietf.org/html/rfc6265
