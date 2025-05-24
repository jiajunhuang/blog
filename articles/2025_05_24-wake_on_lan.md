# Wake on LAN 实现工作机的自动开关机

我的工作机是一台配置较高的台式机，同时也就意味着，开机以后，功率比较高。之前还不觉得，买了一个统计功率的插座发现，待机
都能80多瓦，为了避免不必要的浪费，因此决定折腾一下自动开关机。

## 自动关机

这个简单，root用户设置一个crontab就可以：

```bash
0 22 * * * /usr/sbin/shutdown -h now
```

## 自动开机

这就需要用 Wake on LAN。原理就是，在主板打开这个功能，网卡设置好的情况下，局域网内一台机器向指定的MAC地址发送一个魔术包，
机器收到以后就会自启动。

首先，使用 NetworkManager 设置对应的网卡，启用 Wake on LAN。然后，在主板上找到对应的开关，打开它。

然后，由于我的路由器是 OpenWRT，所以可以直接也添加一个 crontab:

```bash
0 8 * * * /usr/bin/etherwake -D -i br-lan <MAC地址>
```

这样就可以在早上 8 点自动开机了。

---

Ref:

- https://wiki.archlinux.org/title/Wake-on-LAN
