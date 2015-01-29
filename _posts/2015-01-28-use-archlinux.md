---
layout: post
title: "Archlinux"
tags: [linux]
---

今天花了一傍晚+一晚上装完了`Archlinux`， 安装系统过程倒是相当的愉快， 也非常的顺利， 但是最后却在fcitx输入法上出了一个小小的问题卡住了(pkill fcitx; fcitx&没有加载gtk模块， 重启了才加载)， 在IRC上和网友试了好一阵子才解决问题。

就目前从安装到使用几小时(以前虚拟机里用的不计入)来说i， 发现`Archlinux`确实是K.I.S.S的， 在我淘宝上淘来的x61上启动速度和配有SSD的小y速度相当， 当然这其中有`systemd`的大功劳。

`Archlinux`和`Ubuntu LTS`相比由于是滚动更新， 所以没有发行版的概念， 也正因为如此总能享受到最新软件的特性和bug，`Ubuntu LTS`以及`Debian`、`CentOS`都很稳定， 稳定的原因是因为每一个发行版本发布以后软件主版本就固定了，以后只提供安全更新没有特性更新，这种情况下用刚发布的版本还好一点，但是等到过了几年了（等过了2年软件就很老了）里面的软件就太老了。

另外`Archlinux`内存占用那是相当的小啊， 我特意看了一下， `Archlinux`上开机以后只占110多M， 但是`Ubuntu 14.04`占用了780多M， 安装的需要自启动的软件基本相当(Xmonad+Fcitx+LightDM+系统)， `Ubuntu`就多了一个`samba`而已也不要大这么多吧。

另外`Archlinux`有很多在`Ubuntu`里需要ppa安装的软件都在官方源里，直接`sudo pacman -S ***`安装即可，还是最新的 爽啊！如果x61上试用一段时间不崩溃的话， 有必要考虑把小y也从`Ubuntu`切过来啊！
