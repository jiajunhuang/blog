# ArchLinux 怎么降级 package ？

今天我照常升级：`pacman -Syyu`。升级之后发现没法儿愉快的ssh登录了，不论是我从host登录到虚拟机，还是从虚拟机登录到
远端VPS。首先我做的事情是确认是谁的问题。

- 首先我发现无法从XShell登录到本地ArchLinux虚拟机
- 于是我打开虚拟机软件，直接登录，发现OK
- 然后查看 `/var/log/pacman.log`，发现 openssh 的确有升级，而 `/etc/ssh/sshd_config` 也没啥显著变化
- 于是去网上搜了一下，发现没啥类似问题，于是去ArchLinux Bug Tracker [搜到了](https://bugs.archlinux.org/task/59838?string=ssh&project=1&type%5B0%5D=&sev%5B0%5D=&pri%5B0%5D=&due%5B0%5D=&reported%5B0%5D=&cat%5B0%5D=&status%5B0%5D=open&percent%5B0%5D=&opened=&dev=&closed=&duedatefrom=&duedateto=&changedfrom=&changedto=&openedfrom=&openedto=&closedfrom=&closedto=)
- 不过看起来他们也没有什么好的解决方案，然后我决定降级ssh包


怎么愉快的降级呢？pacman会把老的包放在 `/var/cache/pacman` 这个文件夹下，所以直接去找就可以了：

```bash
# pacman -U /var/cache/pacman/openssh-7.7p1-2-x86_64.pkg.tar.xz
# systemctl restart sshd
```

搞定！

----------

当你在 `/var/cache/pacman` 下找不到包时，可以参考 https://wiki.archlinux.org/index.php/downgrading_packages 去ArchLinux
Archive里找对应的包。
