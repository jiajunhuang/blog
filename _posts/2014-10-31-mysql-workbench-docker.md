---
layout: post
title: use mysql-workbench with docker
---

本来想说明一下为什么用docker装mysql的， 但是想了一下， 做个笔记而已啦啦啦

首先， 把mysql镜像pull下来， 可以用`sudo docker pull`也可以直接`sudo docker run`

```bash
sudo docker run --name learnmysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=mysecretpassword -d mysql
```

注意要加`-p 3306:3306`绑定端口, 要不然没有绑定到本机端口是访问不了的。

接下来下载`mysql-workbench`

```bash
wget http://cdn.mysql.com/Downloads/MySQLGUITools/mysql-workbench-community-6.2.3-1ubu1404-amd64.deb
```

linux下没有迅雷啥的， 下载很慢怎么办？ 百度云离线下载嘛， 我就是这么做的， 可不是打广告， 速度还是很赞的。

安装完workbench之后启动， 新建连接， 填入第一条命令中的`mysecretpassword`， 这是可以改的， 改成什么就填什么, 
然后启动就可以用啦~

暂停容器

```bash
sudo docker stop learnmysql
```

启动容器

```bash
sudo docker start leanrmysql
```
