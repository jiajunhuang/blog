# Crontab + Sendmail实现定时任务并且通知

systemd timers真是不好用。或者说，我还是更喜欢crontab，简单易懂。最近我有个需求是定时从云主机把数据
备份到笔记本上，一开始用的systemd timers，但是出错了也不通知我，于是用回crontab，而且还发先可以使用
Linux自带的本地邮件的功能来实现提醒的功能。

```bash
$ sudo pacman -S cronie opensmtp
$ sudo systemctl enable cronie
$ sudo systemctl enable smtp
$ sudo systemctl start cronie
$ sudo systemctl start smtp
```

然后就 `crontab -e` 编辑自己的定时任务，之后只要有邮件，你就会收到一个通知。例如编写下面这样一个crontab：

```
MAILTO=jiajun

* * * * * root
```

> 注意，MAILTO=<your username> 是必要的，否则不会发送邮件。

```bash
$ date
Thu 25 Apr 2019 09:01:32 PM CST
$ cd
You have new mail in /var/spool/mail/jiajun
```

切换目录的时候，就会提醒我有邮件。当然，这还有一个重要原因是设置了 `MAIL` 这个环境变量。如果没有，需要检查一下：

```bash
$ echo $MAIL
/var/spool/mail/jiajun
```
