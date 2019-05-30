# Ubuntu 18.04 dhcp更换新IP

今天在准备弄一个虚拟机集群，自然是装好一个，然后clone成三份。但是有一个问题，clone的时候虽然选择了更换 `MAC` 地址，但是
起来之后发现ip地址还是没变。原来是 `systemd-networkd` 的老bug。它不是根据 `MAC` 地址来决定是否换IP，而是根据 `/etc/machine-id`
来计算出来一个值，如果这个值发生了变化，那么就更换IP地址。

所以就需要把 `/etc/machine-id` 给换一下。

```bash
$ sudo su
# uuidgen | sed 's/-//g' > /etc/machine-id
# reboot
```

即可。

---

参考资料：

- https://www.freedesktop.org/software/systemd/man/networkd.conf.html
- https://unix.stackexchange.com/questions/456763/new-ip-address-whereas-dhcp-lease-time-not-out
