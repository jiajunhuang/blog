# 修复Linux下curl等无法使用letsencrypt证书

遇到一个奇怪的问题，由于我的博客使用的是letsencrypt，浏览器可以正常访问，但是命令行却不可以。

```bash
$ curl https://git.jiajunhuang.com
curl: (60) SSL certificate problem: unable to get local issuer certificate
More details here: https://curl.haxx.se/docs/sslcerts.html

curl failed to verify the legitimacy of the server and therefore could not
establish a secure connection to it. To learn more about this situation and
how to fix it, please visit the web page mentioned above.
```

除了curl，git和wget等也不可以用，这就很奇怪了，为什么浏览器可以但是命令行却不可以呢？

原因是Nginx配置中，`ssl_certificate` 我之前指向的是生成的中间的 `crt` 文件而不是最后的 `chained certificate`。
Nginx官方文档给出的方案是：

> 不过我们不必这么做，此处仅作阅读

```bash
$ cat www.example.com.crt bundle.crt > www.example.com.chained.crt
```

我们不必这么做，因为无论是 `certbot` 还是 `acme.sh` 都会帮我们处理好。把 `ssl_certificate` 的路径改成证书所在路径，
名称改为 `fullchain.crt` 即可。

例如：

```nginx
ssl_certificate /etc/acme.sh/jiajunhuang.com/fullchain.cer;
```

这下大功告成：

```bash
$ curl -I https://jiajunhuang.com
HTTP/2 302 
server: nginx/1.14.2
date: Tue, 26 Nov 2019 15:07:03 GMT
content-type: text/html; charset=utf-8
location: /404

```

---

参考资料：

- [Nginx官方文档](https://nginx.org/en/docs/http/configuring_https_servers.html)
