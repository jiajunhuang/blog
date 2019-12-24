# Matebook X Pro 2019安装Debian 10

最近着手把手头上的三台ArchLinux都换成Debian 10, 更换的原因是不想再频繁更新了，ArchLinux很好用，唯一的缺点就是更新太
频繁了，其实大多数软件的新特性我都用不上，比如内核，我就要一个bbr，最新的LTS内核已经有了，再新的也只是更新驱动而已。

再加上最近碰到一回内核更新之后滚挂了。仔细想了一下，不想折腾了，还是把时间放在更重要的事情上吧。回头想想，大学最开始
学Linux的时候用的Ubuntu，后来切换到Arch之后，到现在都五六年了。不想再每周都把时间花在重复又无意义的 `pacman -Syu` 上
了。

回到正题，周末把Y400从ArchLinux切换到Debian没有遇到什么问题，无非是先备份，然后重装之后把所有东西 `cp -r --preserve`
回来。

不过Matebook X Pro 2019倒是遇到一些问题，安装的时候，Debian 10 live提醒我：

```
Some of your hardware need non-free firmware iwlwifi blablabla
```

原因是Debian live默认不带non free的固件和驱动，在 [这里](https://cdimage.debian.org/cdimage/unofficial/non-free/cd-including-firmware/)
下载带 non free 固件和驱动的镜像，重新刻录即可。

此外安装 libvirt 后新建虚拟机时，提示 `Host Host does not support any virtualization options`，我这都2019年的机器，
早就有了虚拟化指令了，最后发现是因为 `qemu-kvm` 这个包没有自动装上，坑了一把。

有人问我，Debian的软件不是很老吗？没错，仓库里的软件的确老，这也是 Debian相对ArchLinux更稳定的原因嘛，有的时候确实
需要更新的软件，咋办呢？使用 `snapd` 就好了，它会把软件安装在单独的目录 `/snap` 下，这样就不会污染系统的软件目录。

```bash
$ sudo apt install -y snapd
$ sudo snap install go --classic
```

比如这样就可以把最新的Go装上，最后记得把 `/snap/bin` 加到 `PATH` 里。

好了，三台机器已经有两台机器重装为Debian 10了，还剩下一台远程备份机器，春节回家把系统换了 :-) 
以后就可以高枕无忧，愉快的开机几个月再更新重启一次了。
