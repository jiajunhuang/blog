# ElasticSearch 学习笔记

最近在学习使用ES，这篇博客记录一下相关的知识。在此之前，我都是使用PG的全文索引
来做搜索。

ES和PG一样，自带的插件对于中文的支持都不能用，之前使用PG做全文索引，是在应用内
进行分词，而ES则是使用ik插件。

## 安装

ES 是Java写的，可以直接安装在机器上，我选择通过Docker安装：

```bash
$ docker run --restart=always -d --name elasticsearch -p 9200:9200 -p 9300:9300 \
    -e "xpack.security.enabled=true" \
    -e "xpack.security.http.ssl.enabled=false" \
    -e "xpack.security.transport.ssl.enabled=false" \
    -e ELASTIC_USERNAME=elastic \
    -e ELASTIC_PASSWORD=<密码> \
    -e "discovery.type=single-node" \
    -v <PATH TO DATA>:/usr/share/elasticsearch/data/ \
    -v <PATH TO PLUGIN>:/usr/share/elasticsearch/plugins \
    elasticsearch:8.4.1
```

这样安装比较方便，注意要把 `<PATH TO DATA>` 和 `<PATH TO PLUGIN>` 替换成本地目录，
`<密码>` 换成自定义密码。这两个目录用于保存数据和插件。这样，我们就可以开始使用单节点的ES了。

## 基本概念

ES 有几个基本概念我们需要了解，如果你使用过其它数据库，那么就很容易理解了：

- Index: 一个索引，就是一堆 `Documents` 的集合。可以理解为关系型数据库里的 `database`，MongoDB 里的 `collection`
- Mapping: 每个索引，都会有多个字段，而 `Mapping` 就是索引的结构。可以类比为关系型数据库里的 `Schema`
- Documents: 一个文档，就是ES中被索引的基础信息单元，在ES里是一个JSON。可以类比为 MongoDB 里的 `Document`

理解了这三个概念，就可以开始使用ES了。我们再来复习一遍：

每一个Document，是ES中可以被索引的最小单元，其实就是一个JSON。多个Documents组成
一个Index。虽然我们使用时，可以不告诉ES Document中有什么字段是什么类型，但是Index
其实是需要这些信息的，我们不提供时，ES会自己猜测类型。而这个Document结构，就是 Mapping。

## 集成kibana

为了方便使用，我们再起一个Kibana。Kibana是一个管理ES数据的图形界面，准确来说是一个
Web服务。为了让Kibana能连接上ES，我们得为Kibana创建一个 `service account`：

```bash
$ http POST http://elastic:密码@127.0.0.1:9200/_security/service/elastic/kibana/credential/token/kibana
```

接着用上一步返回的token，启动kibana：

```bash
$ docker run --restart=always -d --name kibana --network=host -p 5601:5601 \
    -e ELASTICSEARCH_HOSTS='http://127.0.0.1:9200' \
    -e ELASTICSEARCH_SERVICEACCOUNTTOKEN='<上一步生成的token>' \
    docker.elastic.co/kibana/kibana:8.4.1

```

接着，我们就可以打开Kibana，输入ES的用户名密码登录，接着在左侧导航栏，找到 `Dev Tools`，
接着我们就可以开始操作了。

## 安装中文分词插件

在真正开始索引数据之前，我们需要安装一下插件：

```bash
$ docker exec -it elasticsearch bash
$ elasticsearch-plugin install https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.4.1/elasticsearch-analysis-ik-8.4.1.zip
```

注意，如果你的版本不是 `8.4.1`，那么需要更换成你的版本。安装完之后，重启ES。

## 索引与检索

### 创建索引

首先第一步，我们需要创建一个Index。在我们不指定Index结构的情况下，ES会根据输入的JSON自己猜测结构。但是实际上，ES猜测出来的
结构未必是我们想要的，生产上一般不这么干。我们一般会自己指定结构，在 `Dev Tools` 中输入如下内容，然后点击右上角的绿色三角形，就会执行：

```bash
PUT /words
{
  "mappings": {
    "properties": {
      "content": {
        "type": "text",
        "analyzer": "ik_max_word",
        "search_analyzer": "ik_smart"
      },
      "age": {
        "type": "integer",
        "index": false
     }
    }
  }
}
```

可以看到右侧返回信息：

```json
{
  "acknowledged": true,
  "shards_acknowledged": true,
  "index": "words"
}
```

说明索引成功创建。这个索引中，我们创建了两个字段：

- `content` 类型为 `text`，写入时使用 `ik_max_word` 做分词，搜索时使用 `ik_smart` 分词。这两个的区别在于，前者产生尽可能多的分词，后者产生粗粒度的分词。
- `age` 类型为 `integer`，`index: false` 表示不做索引。

文档参考：https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-create-index.html

### 更改索引

ES中，只能增加删除字段，不能更改已有字段的类型，所以要求我们创建的时候，就确定好类型：

```bash
PUT /words/_mapping
{
  "properties": {
    "content": {
      "type": "keyword"
    }
  }
}
```

返回报错：

```json
{
  "error": {
    "root_cause": [
      {
        "type": "illegal_argument_exception",
        "reason": "mapper [content] cannot be changed from type [text] to [keyword]"
      }
    ],
    "type": "illegal_argument_exception",
    "reason": "mapper [content] cannot be changed from type [text] to [keyword]"
  },
  "status": 400
}
```

但是我们可以增加字段：

```bash
PUT /words/_mapping
{
  "properties": {
    "name": {
      "type": "keyword"
    }
  }
}
```

返回成功：

```json
{
  "acknowledged": true
}
```

文档参考：https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-put-mapping.html

### 获取索引信息

```bash
GET /words  # 获取索引的信息
GET /words/_mapping  # 获取索引结构信息
GET /words/_settings  # 获取索引相关设置信息
```

文档参考：https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-get-index.html

### 索引一个文档

```bash
POST /words/_doc/1
{
  "content": "你好世界世界你好",
  "age": 10,
  "name": "测试1"
}
```

返回成功：

```json
{
  "_index": "words",
  "_id": "1",
  "_version": 1,
  "result": "created",
  "_shards": {
    "total": 2,
    "successful": 1,
    "failed": 0
  },
  "_seq_no": 0,
  "_primary_term": 1
}
```

ES 的行为是：如果没有就创建，有的话，就更新。如果不填 `_id`，也就是 `_doc/1` 中的1这一段，那么ES就会自动生成一个 Docuemtn ID。

文档参考：https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html

## 加快索引的速度

如果大量导入文件，为了提高索引速度，我们可以做的事情有：

- 批量导入，使用 `_bulk` 接口
- 客户端多线程导入，提高请求并发量
- 调大文档刷新间隔，默认情况下，ES每1秒刷新一次文档，让新索引进来的文档可以被搜索到，但是这样会降低索引速度。可以这样操作：

```bash
PUT /words/_settings
{
    "index.refresh_interval": "10m"
}
```

将刷新间隔设置为10分钟。

- 写入时，如果是集群，可以将replicas设置为1，避免多次写入
- 禁用所在机器的swap
- 加大文件系统的缓存
- 不指定ID，让ES自动生成。这一点，需要结合实际业务需求来做
- 加硬件！更好的硬件当然更快
- 调大 `indices.memory.index_buffer_size`
- 使用多节点，避免大量创建索引时导致搜索卡住
- 不需要索引的字段不索引
- 显示指定Mapping，不要让ES猜测。ES猜测出来的类型，通常不是我们想要的，它会同时创建为 `text` 和 `keyword` 两种，比较浪费磁盘
- 每个 `shard` 不要过大
- 禁用 `_source`。不过我一般是留着。`_source` 保存了原始输入的JSON
- 考虑启用 `best_compression`，可以节省磁盘，但是会更费CPU
- 采用可以表示的最小类型，比如如果 `integer` 可以表示，就不要用 `long`，因为他们占用的字节数不一样
- 定期归档旧数据

文档参考：https://www.elastic.co/guide/en/elasticsearch/reference/current/tune-for-indexing-speed.html 和
https://www.elastic.co/guide/en/elasticsearch/reference/current/tune-for-disk-usage.html

## 数据类型

最后，我们来看看ES的常用数据类型：

- `binary` 二进制数据，encode成base64保存
- `boolean` true或者false
- `keyword` 和 `text` 差不多，但是不会分词，必须精确匹配才能搜索到
- `numbers` 分为 `integer`, `long`, `double` 等，表示具体数据类型
- `date` 存储日期
- `alias` 为已有字段创建别名
- `ip` 存储IP地址
- `text` 存储文本，会进行分词

其余类型，参考：https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-types.html

## 总结

ES无疑是强大的，但是它也有缺点，那就是很吃硬件，像2C4G的机器，跑ES本身就已经很吃力，更别提批量导入数据的时候了。这里我们
记录了ES的一些基本知识，业内，ES通常是使用ELK全家桶，用来存储和搜索日志。当然，如果有搜索需求的话，也会使用ES做搜索。
除此之外，elastic 还在探索把ES作为APM的数据存储系统，但是目前看来还没有完全推广开来。

ES目前已经和其它数据库一样，成为程序员必须掌握的一个工具，希望这篇文章能够给你带来一定的帮助。

---

ref:

- https://jiajunhuang.com/articles/2022_04_12-postgresql_fulltext_search.md.html
- https://www.elastic.co/guide/en/elasticsearch/reference/current/elasticsearch-intro.html
