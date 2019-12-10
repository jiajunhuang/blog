# Grafana Gravatar头像显示bug修复

给Grafana提了个PR修复了这个问题，现在PR已经合并了，估计会在下个版本发布。Grafana的响应还是很快的，全程大概2d。

事情的起因是这样的，我发现自建的Grafana登录之后，左下角的头像裂了，于是就把请求抓出来，发现：

```bash
$ http http://<Grafana Server所在机器>/avatar/abcd
HTTP/1.1 200 OK
Cache-Control: private, max-age=3600
Cache-Control: no-cache
Connection: keep-alive
Content-Length: 95
Content-Type: image/jpeg
Date: Tue, 10 Dec 2019 11:55:59 GMT
Expires: -1
Pragma: no-cache
Server: nginx
X-Frame-Options: deny

<html><body><script language='javascript'>location.href='http://mmhr.tv'</script></body></html>

```

咦？这不是妥妥的被劫持了吗？于是我就开始排查，由于我这是虚拟局域网，而我确定所使用的组网软件是基于TLS的，因此
这个过程不可能出问题。
于是我跳到Grafana Server所在机器上请求：

```bash
$ http 127.0.0.1:3000/avatar/abcd
HTTP/1.1 200 OK
Cache-Control: private, max-age=3600
Cache-Control: no-cache
Content-Length: 95
Content-Type: image/jpeg
Date: Tue, 10 Dec 2019 11:58:07 GMT
Expires: -1
Pragma: no-cache
X-Frame-Options: deny

<html><body><script language='javascript'>location.href='http://mmhr.tv'</script></body></html>

```

既然直接请求Grafana都出问题，那么问题肯定就出在Grafana身上了。起初我以为是Grafana插入了广告，但是去翻翻他们也没有这样的
协议呀，于是我就去翻了一下源码，看看Grafana是怎么请求Gravatar拿头像的，从 `pkg/api/avatar/avatar.go` 找到了源码：

```go
var gravatarSource string
 
func UpdateGravatarSource() {
    srcCfg := "//secure.gravatar.com/avatar/"
 
    gravatarSource = srcCfg
    if strings.HasPrefix(gravatarSource, "//") {
        gravatarSource = "http:" + gravatarSource
    } else if !strings.HasPrefix(gravatarSource, "http://") &&
        !strings.HasPrefix(gravatarSource, "https://") {
        gravatarSource = "http://" + gravatarSource
    }
}
```

看了半天，这段逻辑写的绕来绕去，最后就是用了 `http://` 协议，由于我服务器所在环境是被劫持的，所以Grafana就被劫持了。来
验证一下：

```bash
$ http http://secure.gravatar.com/avatar/abcd
HTTP/1.1 200 OK
Cache-Control: no-cache
Connection: close
Content-Length: 95

<html><body><script language='javascript'>location.href='http://mmhr.tv'</script></body></html>

$ http https://secure.gravatar.com/avatar/abcd
HTTP/1.1 200 OK
Accept-Ranges: bytes
Access-Control-Allow-Origin: *
Cache-Control: max-age=300
Connection: keep-alive
Content-Disposition: inline; filename="abcd.jpg"
Content-Length: 2637
Content-Type: image/jpeg
Date: Tue, 10 Dec 2019 12:01:50 GMT
Expires: Tue, 10 Dec 2019 12:06:50 GMT
Last-Modified: Wed, 11 Jan 1984 08:00:00 GMT
Link: <https://www.gravatar.com/avatar/abcd>; rel="canonical"
Server: nginx
Source-Age: 112943
X-nc: HIT hkg 1



+-----------------------------------------+
| NOTE: binary data not shown in terminal |
+-----------------------------------------+
```

于是提了个 [PR](https://github.com/grafana/grafana/pull/20964/files)，就又混到了一个开源项目的Contributor...
