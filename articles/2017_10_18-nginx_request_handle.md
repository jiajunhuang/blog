# Nginx 请求匹配规则

当一个请求到来，nginx首先会检查请求的目标ip地址和端口与哪一个规则里listen的部分相匹配。
如果同一个ip和端口中匹配了多个虚拟的server块，nginx将会测试HTTP请求中 `Host`头部的值和
nginx配置文件中 `server_name` 的值。如果一个 `Host` 匹配到了多个 `server_name`，那么nginx
将会用以下顺序匹配，并且一旦匹配上就用那个 `server_name` 所在的server块来处理：

- `Host`和`server_name`完全匹配
- 以 `*` 开头的，最长的匹配，例如 `*.example.org`
- 以 `*` 结尾的，最长的匹配，例如 `mail.*`
- 配置文件里，出现顺序中，第一个匹配的正则表达式
- 如果以上规则都没有匹配到，那么将会把请求打到设有 `default_server` 的那一个server里去

> nginx中 `include/*.conf` 会是什么顺序呢？这将如何应用于上述第四条规则呢？

当请求进入server块之后，nginx又将如何匹配到哪一个location呢？以下是执行顺序：

- nginx将测试所有的写好前缀的location，例如 `location /test`, `location /`，并且保存能匹配的最长的
那个。如果有 `location = /` 并且匹配，那么匹配到此结束，进入该location块。
- 如果最长的能匹配的location带有 `^~` ，那么将不进行下一步，到此为止，进入该location块
- 依次检查带正则表达式的location块看是否匹配，如果匹配，则匹配到此结束，进入该块
- 如果正则表达式没有匹配到，则使用之前所保存的最长的前缀的块

参考：

- https://www.nginx.com/resources/admin-guide/nginx-web-server/
- http://nginx.org/en/docs/http/ngx_http_core_module.html#location
