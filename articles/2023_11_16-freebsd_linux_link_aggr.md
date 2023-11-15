# FreeBSD 和 Linux 网卡聚合实现提速

今天折腾的内容是网卡聚合提速。由于家里的交换机是2.5G，但是有两个机器，板载网卡都是千兆网卡，我想提升一下访问该机器的
传输速率，因此想到了可以叠加两个网卡，从而提升带宽，由于两台机器的操作系统不同，因此FreeBSD和Linux都折腾了一下。

## FreeBSD

- 首先编辑 `/etc/rc.conf`，注释掉原来的网卡配置，新增如下配置：

```conf
cloned_interfaces="lagg0"
ifconfig_em0="up"
ifconfig_em1="up"
ifconfig_lagg0="laggproto roundrobin laggport em0 laggport em1 DHCP"
```

注意，注释掉原来的配置，不要删掉，这样方便出故障的时候回滚。另外就是把上面的网卡名称替换成你机器上实际出现的网卡名称，
网卡名称可以用 `ifconfig` 命令查看。

- 重启网卡 `service netif restart` 或者重启系统即可

## Linux

Linux 相对来说配置麻烦一些。

- 首先编辑网卡配置文件 `sudo vi /etc/network/interfaces`

注释掉原来的单网卡配置，增加如下内容：

```conf
auto bond0
iface bond0 inet dhcp
  bond-mode balance-rr
  bond-miimon 100
  bond-slaves eth0 eth1
```

`eth0` 和 `eth1` 分别替换成你的网卡名称，下同，Linux下，网卡名称可以用 `ip addr` 查看。

然后分别创建网卡配置文件 `/etc/network/interfaces.d/eth0` 和 `/etc/network/interfaces.d/eth1`，内容如下但是注意也要替换成
你本机的网卡名称：

```conf
auto eth1
iface eth1 inet manual
  bond-master bond0
```

- 重启网络 `sudo systemctl restart networking` 或者重新启动

## 配置

注意，我这里因为是家用环境，所以全都配置成了DHCP，另外因为是家用环境，所以聚合的模式全都是选择了 roundrobin，因此可以
获得更高的带宽而不提供冗余，测试了其他几种模式，只有这一种模式最快，当然前提就是网卡不能挂，挂一个估计网络就要挂。

最后用 iperf 测速，服务端关闭防火墙（或者开启对应端口，一般是5001），然后执行 `iperf -s`。客户端执行 `iperf -c <服务端IP>`
就可以看到速度了。我这里两个网卡叠加之后，iperf跑出来大概是1.53Gbps，虽然比2Gbps的理论速度打折不少，但是比单网卡的 960Mbps
又要好上不少了。
