---
layout: post
title: "Ubuntu内核升级的那些事儿"
tags: [linux]
---

在google一搜“Ubuntu 内核升级”有366000条结果, 我为什么还要写这么一篇呢？因为搜索结果基本都是针对某一个内核写的文章， 我想写的是无论你是哪一个版本（不过我只确定LTS有效）看到以后都有用的文章。

那么在Ubuntu上升级内核(跨版本升级， 安全更新不在本文范围内)有哪些方式呢？ 下面就容我一一道来：

### 1, apt-get升级， 难度☆

Ubuntu的LTS版本都会提供新版本的内核， 而不是像RedHat那样把新内核的代码提取出来放到当前发布的内核中保持主版本号不变， Ubuntu可能是没有那么大人力物力财力？好吧， 闲话少说， 以14.04为例， 首先我们来看看源列表中存在的可安装内核(LTS支持)：

```bash
$ apt-cache search linux-generic-lts
linux-generic-lts-quantal - Generic Linux kernel image and headers
linux-generic-lts-quantal-eol-upgrade - Complete Generic Linux kernel and headers
linux-generic-lts-raring - Generic Linux kernel image and headers
linux-generic-lts-raring-eol-upgrade - Complete Generic Linux kernel and headers
linux-generic-lts-saucy - Generic Linux kernel image and headers
linux-generic-lts-saucy-eol-upgrade - Complete Generic Linux kernel and headers
linux-generic-lts-trusty - Generic Linux kernel image and headers
linux-generic-lts-utopic - Complete Generic Linux kernel and headers
```

看最下面， 哦， 目前除了随14.04发布的3.13版本的内核还可以选择安装随utopic发布的3.16版本的内核(更低版本的内核我就不说啦， 嗯， 这里是讲升级内核的嘛)， 所以如果想安装3.16的内核就执行

```bash
sudo apt-get install linux-generic-lts-utopic
```

好了， 坐等完成， 重启以后就是了。

### 2, 自行下载安装， 难度☆☆

Ubuntu官方发布到源里的内核优点就是有安全更新(`sudo apt-get dist-upgrade`)， 缺点是总不是最新的， 好吧， 我想体验一下最新内核！

首先访问[Ubuntu每日内核更新的站点](http://kernel.ubuntu.com/~kernel-ppa/mainline/daily/current/)， 下载三个包：

```bash
mkdir tmp_kernel
cd tmp_kernel

# 一个是linux-headers-xxxxx_amd64.deb结构命名的， 如果你是32位机器就选linux-headers-xxxxx_i386.deb， 下同
wget http://kernel.ubuntu.com/~kernel-ppa/mainline/daily/current/linux-headers-3.19.0-999-generic_3.19.0-999.201501100206_amd64.deb
# 这个是 linux-headers-xxxxx_all.deb结构命名的
wget http://kernel.ubuntu.com/~kernel-ppa/mainline/daily/current/linux-headers-3.19.0-999_3.19.0-999.201501100206_all.deb
# 这个是linux-image-xxxxx-_amd64.deb结构命名的
wget http://kernel.ubuntu.com/~kernel-ppa/mainline/daily/current/linux-image-3.19.0-999-generic_3.19.0-999.201501100206_amd64.deb
```

> 你可能还看到了 ****-lowlatency-**** 结构命名的内核， 说实话我没用过， 可以[看这里](http://askubuntu.com/questions/126664/why-to-choose-low-latency-kernel-over-generic-or-realtime-ones), 大概是像录音设备之类的需要这种低延迟的内核？这个内核更费电， 对于我们笔记本或台式机还是用不着的。

下载过来以后执行`dpkg -i *.deb`(你要保证该目录下没有其他.deb， 要不然就一起被安装了), 再执行`sudo update-grub`, 重启就可以了。

### 3， 自行编译内核, 难度☆☆☆

这种方式适用于需要高度定制内核(或精简内核)的人群， 但是本文目的是作为一篇通用的文章， 所以这种方法就不多说了。下面我针对几个特定内核给出一些链接吧：

* [编译内核大杂烩, 需梯子](http://lmgtfy.com/?q=ubuntu+compile+kernel)

* [编译3.14内核, 需梯子](http://blog.pinguinplanet.de/2014/04/building-custom-kernel-314-on-ubuntu.html)

* [12.04上编译3.2内核](http://mitchtech.net/compile-linux-kernel-on-ubuntu-12-04-lts-detailed/)
