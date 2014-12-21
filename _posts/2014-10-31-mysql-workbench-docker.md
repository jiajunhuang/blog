---
layout: post
title: use mysql-workbench with docker
tags: [mysql, docker]
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

2014-11-04 更新:

今天做了个更无聊事情： 既然我把`mysql-server`放容器里了， 干脆把`mysql-client` 也放一个容器里好了。

开干： 首先， 打个小广告， 拉取ubuntu的镜像， 我自己做了一个， 只是在官方镜像上改了镜像源而已， 换成了163的， 其他的丝毫未动， 镜像地址[点我](https://registry.hub.docker.com/u/gansteed/docker-ubuntu-cn/):

```bash
sudo docker pull gansteed/docker-ubuntu-cn
```

当然， 最聪明的办法就是写个`Dockerfile`让它自己安装`mysql-client-5.6`， 但是就这一个， 咱直接自己动手好了：

```bash
sudo docker run --name mysql-client --link learnmysql:client -dti ubuntu:trusty /bin/bash
```

记得要获取`learnmysql`的ip地址， 要不然是没办法连上去的， 我一开始在这卡了好久 ;(

```bash
$ sudo docker inspect learnmysql | grep 'IPAddress'
        "IPAddress": "172.17.0.2",
```

接下来

```bash
sudo attach mysql-client

apt-get install -y mysql-client-5.6 #我一开始写成5.7了， 我说半天搜不到..
mysql -h 172.17.0.2 -u root -p
# 输入你的`mysql-server`密码
```

OK！
