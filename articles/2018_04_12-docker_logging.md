# 容器时代的日志处理

> 这是对目前接触到的日志处理的一个总结

## 日志收集

以前的架构一般是：

```
app -> stdout/stderr -> 重定向到文件 -> 定期将文件整理到某处
```

后来先进一点：

```
app -> stdout/stderr -> 重定向到文件 -> logstash等文件处理程序 -> | redis集中收发 | -> logstash收取到某处
```

再后来：

```
app -> stdout/stderr -> Docker -> fluentd -> fluentd收取到某处
```

## 日志搜索

以前一般都是grep，在没有集中处理日志的时候，可能还需要配合ansible等工具，多处grep。缺点是得手工执行，另外数据备份是个问题。

现在一般都是ELK镇场，logstash收到日志之后，顺便解析一下，丢到ElasticSearch中，由Kibana来搜索。

    - 优点是：ElasticSearch一般会配上replica，日志不会丢掉
    - 缺点是：正是因为replica，吃硬盘。另外还有一个缺点，logstash ruby写的，慢。所以现在收集这一步，改用filebeat了。

## 日志归档

- 现在都上云了，用的云服务进行归档保存。当然，如果要用的日志（比如最近一个月的），还是会丢到hdfs等处。

## 没尝试过的

- NSQ/Kafka：很多公司用这两个来传递日志。不过暂时还没尝试过，没有那么大量级。
