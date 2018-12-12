# 使用Dropbox来备份服务器文件

服务器一直都没有备份，上面其实还有数据库，[Dropbox](https://db.tt/8enuh0Mf)这么好用的东西，当然就用它了，当然了，我的数据库
里没有机密信息，但是又要保证不丢，所以符合使用 [Dropbox](https://db.tt/8enuh0Mf) 来备份的要求。

- [注册一个全新的Dropbox账号](https://db.tt/8enuh0Mf)，然后下载服务器操作系统对应的Dropbox软件，例如我的是Ubuntu：

```bash
$ wget https://www.dropbox.com/download?dl=packages/ubuntu/dropbox_2018.11.28_amd64.deb -o dropbox.deb
```

- 安装Dropbox：

```bash
$ apt install ./dropbox.deb
$ dropbox start -i
```

- 然后就可以运行了

```bash
$ dropbox start
```

- 然后你会发现自己并没有登录，怎么登录呢？

```bash
$ dropbox status
```

然后他就会提醒你没登陆，并且给出一个链接，点击就可以登录所打开的浏览器里登录的账号。

- 然后就进入到 `Dropbox` 这个文件夹里，把你想同步的文件加个软链接就好了

```bash
$ cd ~/Dropbox
$ ln -s /data/tgbot.db
```

注意几点：

- 由于备份的是数据库，多少还是重要数据，所以要确保Dropbox账号密码安全性，例如开两步验证
- 如条件允许，可以写cronjob把文件加密之后再进行同步
