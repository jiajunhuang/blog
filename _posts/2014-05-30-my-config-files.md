---
layout:     post
title:      我的配置文件们
date:       2014-05-30 12:00:00
---

linux 吸引我们的， 大概就是自由和可定制吧 

先把我的vim配置， xmonad配置地址贴在这里， 因为我也是从毫不懂配置走来的， 我知道一开始我们都是从别人的配置复制粘贴过来的

[vim配置 + xmonad配置地址](https://github.com/gansteed/configs)

apt-get (Ubuntu, Debian) 系列使用方法:

```bash
sudo apt-get install git
```

```bash
git clone https://github.com/jiajunhuang/configs .xmonad
```

```bash
cd .xmonad && cp .vimrc ~/.vimrc
```

当然这是把xmonad和vim配置一同clone了， 如果不想或者没有安装xmonad， 可以 

* 安装 `sudo apt-get install xmonad suckless-tools `

* 删除配置文件 ` rm -rf .xmonad/ `

yum (Redhat, Fedora, CentOS) 系列相差不大， 把上面 ` apt-get ` 改为 ` yum `即可， 但包名可能会不一样， 我没有去验证 ;(
