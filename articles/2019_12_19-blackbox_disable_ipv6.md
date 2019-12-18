# Blackbox禁用IPv6

使用Blackbox Exporter对我的主站进行健康检查，但是由于本地的网络对IPv6支持并不好，而Blackbox默认优先使用ipv6来访问，
这样就导致老是报警，加入以下配置来避免使用IPv6：

```
      preferred_ip_protocol: ip4
      ip_protocol_fallback: false
```

于是配置长这样：

```
modules:
  http_2xx:
    prober: http
    http:
      method: GET
      preferred_ip_protocol: ip4
      ip_protocol_fallback: false
```

接着，重启Blackbox。

另外我把网卡的IPv6地址也一起禁用了：

```bash
sudo nmcli connection modify Wired ipv6.method disabled
```

其中Wired是网络名称。
