# ArchLinux忽略某个包的升级

最近滚挂了，滚了两年，终于挂了一次，挂在啥地方呢？从5.3的内核升级到5.4之后，wifi用不了了。去报了个bug，但是
尝试了几个更新版本都没有修复，因此准备先忽略内核的更新：

```bash
$ sudo pacman -Syu
:: Synchronizing package databases...
 core is up to date
 extra is up to date
 community is up to date
:: Starting full system upgrade...
warning: linux: ignoring package upgrade (5.3.13.1-1 => 5.4.6.arch1-1)
 there is nothing to do
```

咋做的呢？往 `/etc/pacman.conf` 里，有这么一行： `#IgnorePkg = ` ，给它取消注释，改成：`IgnorePkg   = linux`，
然后就可以了。

ArchLinux当家用操作系统还是很爽的，软件足够新，像我这样每周至少滚一次的，没啥大问题，当然，得订阅一下官网的
更新通知。不过我趁着这个机会把我的家用服务器从ArchLinux换成了Debian，之前作死给家用服务器装的Arch，经过一段时间
的滚动升级，我发现已经厌倦了这种生活，服务器的话，还是老老实实稳定的给我安安静静的跑，别出岔子比较好。
