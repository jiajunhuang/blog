# Nginx作为TCP/UDP的负载均衡

从Nginx 1.9开始，nginx也可以支持TCP/UDP的负载均衡，当然前提是编译的时候
把stream这个模块编译进去了，可以通过 `nginx -V` 来查看是否有：

```
$ nginx -V 2>&1 | grep stream
configure arguments: --prefix=/etc/nginx --conf-path=/etc/nginx/nginx.conf --sbin-path=/usr/bin/nginx --pid-path=/run/nginx.pid --lock-path=/run/lock/nginx.lock --user=http --group=http --http-log-path=/var/log/nginx/access.log --error-log-path=stderr --http-client-body-temp-path=/var/lib/nginx/client-body --http-proxy-temp-path=/var/lib/nginx/proxy --http-fastcgi-temp-path=/var/lib/nginx/fastcgi --http-scgi-temp-path=/var/lib/nginx/scgi --http-uwsgi-temp-path=/var/lib/nginx/uwsgi --with-cc-opt='-march=x86-64 -mtune=generic -O2 -pipe -fstack-protector-strong -fno-plt -D_FORTIFY_SOURCE=2' --with-ld-opt=-Wl,-O1,--sort-common,--as-needed,-z,relro,-z,now --with-compat --with-debug --with-file-aio --with-http_addition_module --with-http_auth_request_module --with-http_dav_module --with-http_degradation_module --with-http_flv_module --with-http_geoip_module --with-http_gunzip_module --with-http_gzip_static_module --with-http_mp4_module --with-http_realip_module --with-http_secure_link_module --with-http_slice_module --with-http_ssl_module --with-http_stub_status_module --with-http_sub_module --with-http_v2_module --with-mail --with-mail_ssl_module --with-pcre-jit --with-stream --with-stream_geoip_module --with-stream_realip_module --with-stream_ssl_module --with-stream_ssl_preread_module --with-threads
```

如果有，则使用如下配置便可以：

```
stream {
    server {
        listen 12345;
        proxy_connect_timeout 1s;
        proxy_timeout 3s;
        proxy_pass 192.168.1.1:8080;
    }
}
```

其中listen是监听本地的什么IP以及端口，proxy_pass则是需要代理的目标服务器IP和端口。具体的
参数可以看参考中列出的文档。

Nginx真是越来越强大了！

参考：

- https://www.nginx.com/resources/admin-guide/tcp-load-balancing/
- http://nginx.org/en/docs/stream/ngx_stream_core_module.html
