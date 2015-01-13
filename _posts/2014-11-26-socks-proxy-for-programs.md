---
layout: post
title: 让程序使用socks5代理
tags: [socks5, linux]
---

首先得有shadowsocks代理到本地

然后`sudo apt-get install tsocks`

编辑`/etc/tsocks.conf`, 找到server这一行，修改为`server = 127.0.0.1`(你代理到本地的sock5地址)

然后就可以使用啦， 比如安装ppa里的软件， 或者是使用pip, gem, cabal之类的包管理器，示例：

```bash
sudo tsocks apt-get update

sudo tsocks pip install shadowsocks

...
```

2015-01-13 更新：

还发现了一个叫`proxychains`的程序， 用起来和`tsocks`一样， 配置见[这里](https://github.com/shadowsocks/shadowsocks/wiki/Using-Shadowsocks-with-Command-Line-Tools)

另外还有直接转成http proxy的软件， 可以供Android SDK Manager这样子只能使用http proxy的程序用， [点这里](https://github.com/shadowsocks/shadowsocks/wiki/Convert-Shadowsocks-into-an-HTTP-proxy)
