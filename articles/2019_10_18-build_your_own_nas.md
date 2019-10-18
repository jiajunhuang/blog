# 旧电脑也不能闲着：家用备份方案

手头上有三台本子，一台X61，一台Y400，一台超级本，外加一个阿里云小水管vps。老机器常年放着，电池报废了，国庆回去换了个电池，
遂决定拿来做远程备份。超级本由于日常要用，而且因为划分了双系统，硬盘不是很足，因此不加入家庭备份方案(重要资料除外，如下述)。

首先我们来看需求，需要备份的主要有：

- 收藏的电子书，操作系统镜像，这些资料的特点是最好不要丢失，否则重新去找很费时间
- 生产服务器数据、两步验证码和手机上的照片，这些资料的特点是完全不能接受丢失，否则会导致严重后果例如再也看不到照片了
- 自己搭建的私有git服务器数据和自用数据库，这些资料也最好能完全不丢失，否则够搞了

因此，解决方案就是多重、异地备份，传说中，备份有个3-2-1法则？3个备份，2种介质，1份异地。我的这个方案除了2种介质外，基本满足。

此外我们分析一下机器的特性以及我准备的用法：

- X61 2C2G，机械硬盘1T，无公网IP。用来做监控、自用MySQL、自用git服务器等不需要高性能的服务器
- Y400，4C12G，SSD 400G，无公网IP。用来做虚拟机Host等需要高性能的服务器
- vps，1C2G，硬盘20G，有公网IP。主要是用来做中转服务器用来保证异地服务器可以互相连接组成局域网。

## 组建异地局域网

由于我们很难拿到公网IP，因此我们需要自行组建局域网，方案可以采用zerotier，如何搭建可以参考 [此处](https://jiajunhuang.com/articles/2019_09_11-zerotier.md.html)。

但是注意，zerotier我们的确可以打洞成功，尤其是因为有一个国内的vps节点做moon，但是我们可能经常要上异地备份服务器进行操作，
对于ssh我们还需要使用一个备份登录方式：frp。我在X61和Y400上都使用了frp将本地ssh端口转发到公网vps，这样就解决了zerotier
连接不一定稳定的问题。

> 注意，一定要开防火墙。如果觉得iptables难用的话，可以用ufw，参考 [ufw简明教程](https://jiajunhuang.com/articles/2019_09_14-ufw.md.html)

> 此外，记得禁用root用户远程登录等选项，至少不能通过密码登录，两台本地服务器也需要使用防火墙关掉ssh端口，并且只允许zerotier
> 所分配的网段进行登录。

## 本地磁盘备份

我的vps使用的是debian，文件系统使用ext4，由于vps对磁盘的可靠性提供保证，我就不对vps进行操作了，而X61和Y400则是安装的ArchLinux，
他们的文件系统都是使用btrfs，但是由于 `$$$` 的原因，没有开软RAID。如果 `$$$` 足够的话，建议买硬盘做 `RAID1`。

## 异地磁盘备份

接下来，就要做异地磁盘备份。我使用 `rsync` 进行操作。下面是一些常见数据的备份方式：

- 对于MySQL和Gitea，使用crontab进行备份：

```crontab
@daily /usr/bin/mysqldump --single-transaction --quick --lock-tables=false --all-databases | gzip -c > /data/backup/mysql/full-backup-$(date +\%F).sql.gz
```

MySQL在X61上，因此在Y400上进行同步：

```crontab
# every 5.am, sync backup files from lan.think
0 5 * * * rsync -a --delete x61:/data/backup /data/backup/
0 5 * * * rsync -a --delete x61:/data/docker /data/backup/
```

所以MySQL和Gitea的数据是备份两份，异地。Gitea的数据也是，并且于此同时，repo还分散在各地。

- 生产服务器数据：

在生产上进行备份后，使用rsync同步到X61，随后同步到Y400。步骤与上述相同，故不赘述。

- 对于照片、电子书、重要资料等，使用Syncthing，在超级本、X61、Y400之间互备。由于这些资料不大，因此在这三台机器上互备。
Syncthing还是很好用的，唯一的缺点就是慢一些，但是不怕，我可以接受慢。

## 监控和报警

本地服务器常年开着，也不知道啥时候会出点啥问题，因此监控和报警必不可少。我采用的是：

- Prometheus + Alert Manager 监控 + 报警
- Grafana查看数据
- Slack接收从Alert Manager来的报警
- Node Exporter + text-file collectors + `smartmon.py` + `btrfs.py` 收集节点信息、磁盘健康信息、btrfs错误信息
- Nginx Exporter 监控Nginx
- MySQL Exporter 监控MySQL
- [Battery Exporter](https://github.com/jiajunhuang/battery_exporter) 监控电池信息
- Libvirt Exporter 监控虚拟机

然后再配上报警规则。

到此，大功告成！
