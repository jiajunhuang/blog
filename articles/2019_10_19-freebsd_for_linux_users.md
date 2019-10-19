# 给Linux用户的FreeBSD快速指南

可能现在很多人甚至没有听过FreeBSD，但是不得不说，FreeBSD虽然市场份额变小了，
仍然是UNIX家族中重要的一支。

我给自己定的目标是Linux和FreeBSD都要玩熟悉，原因有下：

- Linux占据服务器市场主要份额，日常生产中也是经常用到，所以必须熟练使用
- FreeBSD的协议非常好，如果哪天有需要，基于FreeBSD做产品不需要考虑被开源的问题

更何况他们同宗同源，都是类UNIX呢，两者切换起来并不难。接下来，我们就看看Linux用户怎么切换到FreeBSD：

首先，在国内拉取FreeBSD的二进制包非常慢，所以我们要使用一个镜像：

```bash
$ sudo mkdir -p /usr/local/etc/pkg/repos
$ sudo su
# cat > /usr/local/etc/pkg/repos/FreeBSD.conf << EOF
FreeBSD: {
  url: "pkg+http://mirrors.ustc.edu.cn/freebsd-pkg/${ABI}/quarterly",
}
EOF
# pkg update
```

- shell。服务器我们最常用的软件估计就是shell了，Linux中默认的一般是 `bash`，FreeBSD默认的一般是 `tcsh`，我不想再学一个shell怎么用，我的配置都是基于 `bash` 的，因此，我使用 `bash`：

```bash
$ sudo pkg install -y bash bash-completion
$ chsh  # 使用chsh把自己的默认shell改成 `/usr/local/bin/bash`。官方推荐别把root的改了。
```

- 安装常用的包。Linux中常用的包FreeBSD中都有，并且可以直接使用pkg安装：

```bash
$ sudo pkg install -y python3 tmux git bash neovim htop bash-completion
```

Linux要用到的包，FreeBSD中一般都有，如果名字不对，可以进行搜索：`pkg search xxx`。此外有一些Linux独有的软件例如kvm，
FreeBSD中也有对应，是bhyve。其它的连 `libvirt`, `virt-manager` 都有。

当然FreeBSD还可以使用ports，就跟gentoo那样，自己编译，当然了，我是不喜欢自己编译，费时费力。

- 软件自启。现在Linux中我们一般都使用systemd，不过FreeBSD还是使用BSD-style init，
也就是说，编辑文本，例如：

```bash
$ cat /etc/rc.conf
sendmail_enable="NONE"
hostname="freebsd"
keymap="us.kbd"
ifconfig_vtnet0="DHCP"
sshd_enable="YES"
ntpd_enable="YES"
powerd_enable="YES"
# Set dumpdev to "AUTO" to enable crash dumps, "NO" to disable
dumpdev="AUTO"
```

FreeBSD中要启动软件，就编辑各种 `rc.conf` 和 `rc.d/` 下的文件，然后使用 `service xxx start` 启动。

如果不是想开机自启，只是想启动一次，那么可以使用 `service sshd onestart` 这样的用法。

- 现在的Linux都是使用 `ip` 和 `ss` 命令来查看和管理网络相关的东西，而FreeBSD使用
`ifconfig` 和 `netstat`。

- 防火墙。Linux中使用iptables来做防火墙，不过那个太复杂了，我一般使用 `ufw`。
FreeBSD中可以使用 `ipfw` 来管理防火墙，它的用法和ufw差不多：

```bash
$ ipfw add allow tcp from any to me 22 in via $ext_if
```

- 更新。FreeBSD和Linux中更新的不同之处在于，内核和软件包是分开更新的，内核使用
`freebsd-update` 更新，而软件包使用 `pkg upgrade` 更新。

- 常用命令。

    - 安装软件：`sudo pkg install xxx`
    - 安装本地包：`sudo pkg add ./xxx`
    - 列出安装的包：`sudo pkg info`
    - 列出设备：`sudo pciconf -l`
    - 列出内核模块：`sudo kldstat`

其实Linux和FreeBSD的用法相当接近，总体来说，就这么几个区别：

- Linux迁移到了systemd而FreeBSD仍然是用init
- FreeBSD默认使用tcsh而Linux发行版一般使用bash
- FreeBSD安装的软件包全都在 `/usr/local` 下而Linux则一般都在 `/usr/` 下

其它几乎没有，就是一些边边角角的区别和命令的用法区别而已。FreeBSD为什么不如Linux流行呢？查了一下，据说是因为
90年代的官司，而恰好那个时候开发了Linux，这就是时代的选择。

但这些都不能掩盖FreeBSD是一个好的服务器系统的真相。

---

参考资料：

- [FreeBSD Quickstart Guide for Linux Users](https://www.freebsd.org/doc/en_US.ISO8859-1/articles/linux-users/)
