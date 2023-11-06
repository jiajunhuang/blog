# 解决 macOS 终端hostname一直变化问题

macOS 终端有时候名字会变化，网上查到的原因说是因为 macOS 会从 DHCP 服务器取值，解决的办法很简单，执行如下命令(假设你准备
把hostname改成 mbp)：

```bash
$ sudo scutil --set HostName mbp
$ sudo scutil --set LocalHostName mbp
$ sudo scutil --set ComputerName mbp
```

然后重启。
