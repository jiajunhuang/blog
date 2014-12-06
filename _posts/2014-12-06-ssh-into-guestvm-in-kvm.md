---
layout: post
title: 怎样ssh连接到kvm中的虚拟机
---

这几天玩kvm虚拟机， 却不知道怎么ssh进去(不懂得虚拟机中网桥和ip到底是什么关系)，[找到一个方法](http://rwmj.wordpress.com/2010/10/26/tip-find-the-ip-address-of-a-virtual-machine/):

```bash
arp -an
? (192.168.1.101) at 20:6a:8a:72:63:73 [ether] on eth0
? (192.168.1.1) at a8:57:4e:26:6f:dc [ether] on eth0
? (192.168.1.108) at b8:88:e3:e2:7c:42 [ether] on eth0
? (192.168.122.11) at 52:54:00:69:64:d8 [ether] on virbr0
? (192.168.1.104) at 68:5b:35:8a:ef:93 [ether] on eth0
```

于是就找到了kvm中系统的ip

当然， 还有另外一种方式就是用`virt-manager` 或者 `virt-viewer` vnc进系统， 直接输入`ifconfig`也是可以找到的。

kvm的默认磁盘格式raw不支持快照， 我想到一个办法：用版本控制系统`git`可以做到这一点， 因为`git`就是对当前已追踪的文件做快照 ;D
