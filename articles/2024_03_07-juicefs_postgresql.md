# 使用 PostgreSQL 搭建 JuiceFS

首先我们来看一下大概的步骤：

1. 安装PG和JuiceFS
2. 数据库创建好用户名密码
3. 创建好 s3 的bucket和access key + secret key
4. juicefs foramt 创建好文件系统
5. 编辑开机自启挂载文件系统
6. 校验和使用

## 原理简介

使用juicefs之前，还是需要大概的了解一下原理，我没有仔细阅读源码，但是基本把官方文档阅读了一遍，下面是我的理解：

> 更具体的细节，还需要阅读源码，这个只是我的一个粗浅理解

1. 首先最简单的来说，我们可以理解为 s3fs，也就是把 s3 通过fuse挂载到本地路径，从而你读写，数据就会自动同步到 s3
2. 光是 s3fs，或者像NFS那样，把本地POSIX读写转成网络读写的话，性能肯定好不起来，因此juicefs使用了一些技术来加速，包括但不限于：
    a. 文件分块：文件分块以后，读写直接使用偏移量，网络I/O也就会变的更快，一次可以只读一个块
    b. 预读：当检测到连续读时，如果本地缓存不存在，则可以直接将接下来的文件块提前下载下来
    c. 本地磁盘当缓存：例如写文件时，可以先写到本地磁盘，再同步到s3等
3. juicefs 将文件的属性和数据本身进行了分离，也就是元数据和数据本身分离，元数据可以存到数据库里，从而实现分布式部署，这样其他节点只要能连接到数据库，就可以知道文件的信息，数据本身则是存储到例如s3等云存储中

在企业中，很多场景下我们需要把文件备份到s3等存储，但是有时候则需要读写，例如：日志存储、HDFS、文件备份等，这种时候，如果可以
把s3等挂载到本地，就会比较方便。当然，如果不能挂载，我们还可以使用类似rclone等工具来实现备份(rclone其实也可以挂载)。

## 搭建步骤

1. 安装PG和JuiceFS

```bash
# sudo apt install postgresql
# 默认安装到 /usr/local/bin
# curl -sSL https://d.juicefs.com/install | sh -
```

2. PG创建好数据库、用户、密码

```bash
$ sudo -u postgres psql
$ psql
> CREATE USER 用户名 WITH ENCRYPTED PASSWORD '密码';
> CREATE DATABASE 数据库名 OWNER 用户名;
```

3. 创建好 s3 的bucket和access key + secret key

这个就直接去网页上操作，记得保存好 bucket 的 endpoint，access key, secret key 就行

4. juicefs foramt 创建好文件系统

执行命令

```bash
juicefs format --storage=s3 --bucket=https://<bucket endpoint> --access-key=<access key> --secret-key=<secret key> postgres://<数据库创建好的用户名>:<数据库创建好的密码>@127.0.0.1:5432/<创建好的database> <你想把这个叫做啥这里就写啥>
```

当然，上面的数据库地址，如果你不是使用本地连接的话，就得改成对应的地址

5. 编辑开机自启挂载文件系统

juicefs 本身提供一个编辑fstab的方式来实现开机自动挂载，但是缺点就是数据库的密码等等都在里面。而我更倾向于使用 systemd 挂载的方式。
首先，确定你要挂载的路径，例如我想挂载到 `/data/bitful` 这个路径，那么我要编辑的systemd mount文件的路径就是 `/etc/systemd/system/data-bitful.mount`，你是什么路径，就用中划线连接这个路径。

然后下面是内容：

```systemd
[Unit]
Description=Juicefs
Before=docker.service

[Mount]
Environment="ACCESS_KEY=<access key>" "SECRET_KEY=<secret key>" "STORAGE=s3" "BUCKET=https://<bucket endpoint>" "META_PASSWORD=<数据库密码>"
What=postgres://<数据库创建好的用户名>@127.0.0.1:5432/<你创建的database>
Where=/data/bitful
Type=juicefs
Options=_netdev,allow_other,writeback_cache

[Install]
WantedBy=remote-fs.target
WantedBy=multi-user.target
```

> ⚠️ 例如加密、设置缓存目录等，juicefs文档没写，我也暂时还没探索，所以如果你想使用这些设置选项的话，可能暂时还只能使用fstab文件的形式。

6. 校验和使用

启用：

```bash
$ sudo ln -s /usr/local/bin/juicefs /sbin/mount.juicefs
$ sudo systemctl enable --now data-bitful.mount
```

然后就可以去 `/data/bitful` 路径下随便写一个文件，去s3上看看有没有对应的新文件产生就可以了

## ⚠️注意事项

使用juicefs这样的工具，很重要的一点就是一定要把保存元数据的数据库保护好，要不然元数据损毁，整个文件就很难找回来了，像我
本地使用，一般Redis我都当作缓存或者队列使用，很有可能会操作 `FLUSH` 命令，因此我就选择关系型数据库来保存元数据，即使慢
一些，而且关系型数据库是做了备份的。

---

Refs:

- https://juicefs.com/docs/zh/community/mount_juicefs_at_boot_time
