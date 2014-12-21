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
