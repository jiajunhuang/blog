---
layout:     post
title:      "FreeBSD"
tags: [unix, linux]
---

FreeBSD, NetBSD, OpenBSD, BSD家族的三个重要成员。

前不久装上了FreeBSD 10在公司机器上，主要也是为了体验一把传说中的ZFS。当然我的
Arch上使用的是btrfs。

FreeBSD 10 也提供了二进制安装包和源码编译两种方式。并且把以前的`pkg_add` 换成了
现在的`pkg`， 但是不知道为什么官方不允许镜像， 所以在国内速度那叫一个慢。所以
还是老老实实的用ports比较好。安装完系统以后记得使用`portsnap update`更新ports
源， 然后更新完成后它会提示你执行`portsnap extract`， 之后你就可以使用`whereis`
查询你要的软件在哪里，比如:

```bash
$ whereis python34
/usr/ports/lang/python34
```

然后进到该目录去执行`make install clean`安装， 如果你执行到一半，发现不想安装了
也可以ctrl-c中断掉， 然后执行`pkg autoremove`自动清除依赖， 另外如果你想卸载
某个软件，那么先执行`whereis python3`， 然后看最后面， 是在哪个ports目录下，进入
到该目录并且执行`make deinstall`， 之后`pkg autoremove`就可以了。

从Linux转到FreeBSD需要注意的是， FreeBSD 的root默认的shell是csh， 不习惯的话
可以把默认shell换成bash（我就换了。。。）但是一定要小心，比如不要换了shell之后
缺把它删了，那你就得重启然后进单用户模式，把shell改回来了。所以一般来说，root
还是不改为妙。

另， FreeBSD 安装软件的默认目录是在`/usr/local/`下， 所以如果自己编译了bash，
那么目录应该是在`/usr/local/bin/bash`， 这也是不建议更改root的shell的原因之一：
如果你没有挂载`/usr/local/`，那么又要单用户模式了。

与大多数Linux发行版不一样， FreeBSD很多软件甚至是不带默认配置的，比如samba4，
并且FreeBSD想要使用`service sshd start`类似的命令，一定要在`rc.d`里有或者在
`/etc/rc.conf`里存在， 例如想要开机自启ssh， 就要在里面写

```bash
sshd_service="yes"
```

据官方文档所说， "yes"可以换成"true", "1"都是可以的， 但是我没有尝试。还需要
注意的是， FreeBSD默认不是utf-8编码，需要在`.bash_profile`或者`/etc/locale`(这个忘了
		是不是这个目录) 里导入LANG变量， 后者还需要执行一条什么指令更新一下
数据库。

最后吐槽一下：FreeBSD对网卡支持不好。Intel网卡大都没有问题，螃蟹卡就。。看运气了。
