# Nginx设置Referer来防止盗图

如果不是图床，还真怕别的网站直接使用本站图片。这样子流量很可能一下子就被刷光了，毕竟CDN都是白花花的银子买来的。
因此，还是设置一个防盗链吧，Nginx就可以完成这个功能了。

一般来说，遵照HTTP协议实现的浏览器，在从A网站访问B网站时，都会带上当前网站的URL，以表明此次点击是从何而起的。因此，
Nginx的这个模块也是依靠这个来实现，所以，如果骇客不加此头部，还是没法愉快的防盗图。

Nginx官网文档如下：

```
Syntax:	valid_referers none | blocked | server_names | string ...;
Default:	—
Context:	server, location
```

因此，我们可以在 `server` 或者 `location` 块加上代码，我是保存为 `valid_referers.conf`：

```nginx
valid_referers none blocked server_names;

if ($invalid_referer) {
    return 403;
}
```

然后在对应需要的地方加上 `include /etc/nginx/valid_referers.conf`，当然，执行这个的前提是已经把 `valid_referers.conf`
放到对应机器上的 `/etc/nginx/valid_referers.conf` 路径下。

示例：

```nginx
    location /articles/img {
        include /etc/nginx/valid_referers.conf;
        root /data/blog/code;
    }
```

---

- http://nginx.org/en/docs/http/ngx_http_referer_module.html
- https://en.wikipedia.org/wiki/HTTP_referer
- https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referer
