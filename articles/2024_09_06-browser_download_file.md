# 让浏览器下载文件

不知道你是不是遇到过这样的需求，想要从服务器将一个文件发送给用户。原来我以为，直接遵循这么几步就可以：

- 将文件内容写入到响应体中
- 设置响应头中的 `Content-Type` 为 `application/octet-stream`
- 设置响应头中的 `Content-Disposition` 为 `attachment; filename=xxx`

搞定，从后端来说，的确是这样就可以。

但是真正使用的时候，不一定能触发下载。因为浏览器是否会自动触发文件下载，有多种因素影响：

- `Content-Disposition` 头，服务端需要设置为 `attachment`，这样浏览器才会提示下载
- 浏览器本身的设置，如果设置了不自动下载，那么即使服务端设置了 `attachment`，也不会自动下载
- 如果前端代码是通过AJAX请求的，那么即使服务端设置了 `attachment`，也不会自动下载，必须要在前端代码中做特殊处理
- 如果服务器防火墙对下载做了限制，可能会被阻止下载

不过最常见的因素，是第一点和第三点。第一点是服务端设置不对，第三点是前端代码没有处理好。

## 前端的处理

写前端代码之前，我是不知道前端还需要做特殊处理的。我以为只要服务端设置了 `attachment`，浏览器就会自动下载。
不过实际上并非如此。

前端需要做如下处理：

```javascript
$http.get('/download')
    .then(function (response) {
        var retData = response.data;

        const blob = new Blob([retData], {type: 'application/octet-stream'});
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'audit_log.csv';
        a.click();
    });
```

这里的关键点是：

- 将返回的二进制数据转换为 `Blob` 对象
- 通过 `URL.createObjectURL` 创建一个 URL
- 创建一个 `a` 标签，设置 `href` 为上面创建的 URL，设置 `download` 为文件名
- 触发 `a.click()` 事件

这样就可以触发下载了。

> 什么是Blob对象？Blob对象表示一个不可变、原始数据的类文件对象。Blob表示的数据不一定是一个JavaScript原生格式。
> File接口基于Blob，继承了blob的功能并将其扩展使其支持用户系统上的文件。

## 总结

多学点前端知识，对后端开发也是有好处的。以前我就是纯粹的后端开发，对前端的知识了解的不多，以为只要服务端设置了对应的
头部，服务器就可以自动控制下载。实际上也不是这么简单的。
