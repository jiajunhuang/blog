# 我的技术栈选型

工作已经几年了，逐步摸索到了自己的技术上限 --- 还是计算机五大件。不断的追新和扩展广度已经没有太大意义，它们的实现原理也
都了解的差不多，因此现在是时候开始收缩技术栈，缩小需要不断更新的知识范围，节省精力去做更有意义的事情，把常用技术栈固定
下来，并且保持小量更新的状态。

> 注：我的技术选型风格偏向于UNIX的KISS风，喜欢小而美的东西 :-)

## 编程语言

编程语言Python和Golang：

- Python主要写Web服务、小工具、脚本等等。Python的优点是写起来非常快，比如我写了一个删除过量的MySQL备份数据的脚本，核心代码
6行就搞定了。如果是Go来写的话，估计要2-3倍，其它语言就更多了。Python的缺点就是速度慢些，但是我写了一个QPS数百的应用，跑在
生产服务器上，响应速度也在20ms左右，所以对于绝大多数网站来说，完全够用。如果一定要挑一个Python的缺点的话，那就是部署没有Go方便。

> 其实绝大部分web应用，瓶颈都不是编程语言。Python Flask框架平均响应时间15-20ms左右，相比公网传输速度来说，完全可以接受。
> 当然了，Go固然更快，一般1ms以内，内存占用也小很多。

```python
def find_outdated():
    for i in glob.glob("/data/backup/mysql/*.sql.gz"):
        year, month, day = i.split(".")[0].split("-")[-3:]
        t = datetime.datetime(year=int(year), month=int(month), day=int(day))
        if t < datetime.datetime.now() - datetime.timedelta(days=7):
            os.remove(i)
```

- Go主要写系统工具和需要高并发的软件。比如如果我要自己实现一个网关，那么我就会考虑使用Golang；另外，处理二进制协议的时候
Golang也方便点。还有就是如果瓶颈在内存使用上，我也会考虑用Go而不是Python。

> Rust我也考虑过要不要学，但是最终我发现我的需求用不上Rust，不需要那么高的速度，主要就是写web，此外写点小工具。Rust还是太啰嗦了。
> C++ 和Java以前也考虑过，一个太复杂，一个太啰嗦，略。有Python和Golang，快速和高性能都够了，还不够的话，那说明真的该加机器了。

## 数据库

数据库使用MySQL，MySQL应当玩的很很熟悉才行。原因是MySQL比PG简单得多(配置管理)，而且在主从上有更好地支持。此前我也有使用SQLite，
但是说实话，脑子里同时维护更新多个数据库的知识，太费力，大部分都是相通的，但是细节处又不相同，每次还要查文档，太麻烦了，所以
还是用最熟悉的那个。

此外MySQL最强大的优点就是社区，碰到什么问题，基本都能够搜到。

## 备份工具

MySQL使用mysqldump和crontab，如下：

```bash
@daily /usr/bin/mysqldump --single-transaction --quick --lock-tables=false --all-databases | gzip -c > /data/backup/mysql/full-backup-$(date  + \%F).sql.gz
```

然后再使用rsync进行异地同步。

## 缓存、队列

缓存和队列统一使用Redis。用不上RabbitMQ这类重量级的消息队列。Redis一箭双雕，而且好用。

## Web服务器

Web服务器使用Nginx，Nginx还可以作为端口转发工具(使用Nginx的stream模块)。

> 端口转发也可以使用frp。

## 部署

部署统一使用supervisor。之前我也有使用Docker，但是越发觉得，我一个人维护的系统，不必用Docker，因为系统依赖等全都是我可以
控制的。Go基本无依赖，Python可以使用virtualenv或者venv解决依赖问题。

> 什么？k8s？个人，甚至是小团队根本没有必要上这个玩意儿。不过如果是团队，倒是可以上上Docker。

自动化部署使用ansible，我自己的系统就懒得搭CI/CD了，每次要部署的时候直接跑ansible-playbook即可。

## Web框架

Python使用Flask，Go使用GIN。Python需要搭配gunicorn + gevent使用。

## ORM和数据库migration

Python使用SQLAlchemy + alembic，Go暂时没找到顺手的。

## 异步任务框架

Python使用python-rq作为异步任务队列，Golang暂时没有找到好的。

> 为啥不用celery？celery太大了，而且遇到过多次worker无故死掉的情况，包括很多网友也遇到过，但是一直没找到原因，后来换了rq
> 之后发现rq简单好用，就切换了。

## 监控

监控使用Prometheus + Grafana + AlertManager + Exporter全家桶。

## 日志收集

我自己的服务现在不进行日志收集，之前用过EFK/ELK，但是太占内存，而我的场景直接去服务器上tail + grep就够了，用EFK完全是杀鸡用牛刀。
看了一下Grafana设计的loki，设计挺不错，但是现在还不够成熟。

## 搜索引擎

目前暂时使用MySQL来做，以后如果有需要，再学习ElasticSearch，之前用过ES，但是频率太低，学了又忘记了，MySQL足矣。

## 错误追踪

错误追踪使用Sentry，直接用 [官网](https://sentry.io) 就够了。

## 总结

以上就是我这几年接触过然后筛选下来的技术栈，有一些可能漏掉了，有一些被我放弃的原因也没写上，欢迎一起交流 :-)
