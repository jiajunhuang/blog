# Sentry 自建指南

之前一直使用官方的免费服务，最近想要自己托管一个玩玩，于是就折腾了起来。Sentry 是一个开源的异常收集工具，据我所知
好像很多公司都在用它，而且讲真确实挺好用的，我也是老用户。

使用 docker-compose + docker 的部署方式比较简单，首先要安装 docker 和 docker-compose ，它对 docker 的版本有一定的
要求，所以我直接官网安装最新的：

https://docs.docker.com/engine/install/debian/

接着安装 docker-compose：

https://docs.docker.com/compose/install/

之后，就可以把官网的部署仓库 clone 下来：

```bash
$ mkdir sentry
$ cd sentry
$ git clone https://github.com/getsentry/onpremise.git
$ cd onpremise
$ ./install.sh
... 提示创建用户，那就创建
$
```

然后启动：

```bash
$ docker-compose up -d
...
```

接下来，就是配置一个域名，将请求代理到此服务，接下来就可以访问对应域名，然后更改发送邮件相关的配置了，页面上就可以
操作，但是如果页面上没有，那么就需要改配置文件，然后重新启动。

我是使用的 Mailgun 的 SMTP 服务，配置如下：

```bash
###############
# Mail Server #
###############

mail.backend: 'smtp'  # Use dummy if you want to disable email entirely
mail.host: 'smtp.mailgun.org'
mail.port: 25
mail.username: 'Mailgun 系统内的 SMTP 用户名'
mail.password: '密码'
# mail.use-tls: true
# mail.use-ssl: false
# The email address to send on behalf of
mail.from: 'sentry@<你的域名>'
```

此处注意，我一开始走了 TLS，结果一直报错 `Connection timeout`，加大 socket 超时时间也没用。估计是伟大的墙做的好事吧。

更改配置之后，就需要更新一下应用：

```bash
$ docker-compose build
$ docker-compose run --rm web upgrade
$ docker-compose down && docker-compose up -d
```

接下来就可以愉快的使用自己的 Sentry 服务了。
