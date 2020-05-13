# Debian 使用NetworkManager之后networking.service崩溃

NetworkManager更好用，但是直接安装NetworkManager之后，原有的networking.service会崩溃。

安装当然很简单：

```bash
$ sudo apt install network-manager
```

就可以了。但是如果安装了node exporter的话，会不断的报 SystemD Service Crashed 错误，而报错的就是 networking.service。
解决办法，把 `/etc/network/interfaces.d/setup` 和 `/etc/network/interfaces` 里的东西全部注释掉(加 #号)。
