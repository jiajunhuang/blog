# 使用PostgreSQL做搜索引擎

最近我在研究用PostgreSQL来做搜索引擎，简单来说，搜索引擎主要使用的就是倒排索引，也就是把一篇文章或者语句，首先进行分词，
将文章分成N个词语，每一个词语都有一定的权重，这一步里有很多地方可以优化，将文章切分成准确含义的分词，对后续搜索的影响
十分大，这个可以参考 TF-IDF 或者 TEXTRANK 算法；第二步就是建立倒排索引，也就是将分词和该词语在文章中的位置关联起来；
有了倒排索引，第三步就是进行搜索，同样的，我们将输入进行分词，并且组装成一定的搜索条件，放到搜索引擎里进行搜索；第四步
就是处理搜索结果。

不过，业界常用的搜索引擎都是ES，为啥我想用一下PG呢？海量数据情况下的搜索也许ES会是更好的方案，但是一般情况用ES太重了，
PG是否可以胜任常见情况呢？如果我们的业务数据和搜索都能由PG处理，就不需要在数据库和ES之间进行同步了，整体的复杂度会下降
很多。

这就是我做这次尝试的起因。

## PG 内部支持

要在 PostgreSQL(简称PG，后同) 中使用全文索引，需要借助 PG 提供的一个内置数据结构，叫做 `tsvector`：

```
A tsvector value is a sorted list of distinct lexemes, which are words that have been normalized to merge different
variants of the same word (see Chapter 12 for details). Sorting and duplicate-elimination are done automatically
during input, as shown in this example:

SELECT 'a fat cat sat on a mat and ate a fat rat'::tsvector;
                      tsvector
----------------------------------------------------
 'a' 'and' 'ate' 'cat' 'fat' 'mat' 'on' 'rat' 'sat'
```

`tsvector` 是用来存储分词向量的，我们来看一个简单的例子：

```sql
# SELECT tokens FROM cargos;

'一':66B,70B '一部分':68B,72B '上':35B '不是':62B '两极':29B '两极分化':31B '中国':44B,54B '丰':19B '丰衣足食':22B...
```

可以看到，这里就是我们前面说的倒排索引，每一个都是一个词语以及词语的位置组成的，里面的 `A` 和 `B` 其实是词语的权重，
PG有 ABCD 4种权重，A的权重最高，D的最低。权重越高，后续搜索的时候，排名就可以越前。

## 构建倒排索引

我们拿到文章之后，要先进行分词，我使用的是 `sego`，原因是 `gojieba` 总是莫名panic，所以为了简单，我先使用 `sego`，假设
后续发现 `sego` 优化空间不足，那么可以使用 `jieba` Python版封装成服务提供出来。

我的数据库设计为：

- `title` 标题
- `description` 文章或者描述
- `tokens` 存储分词向量

这样应该可以适用于绝大部分场景，例如搜索文章、商品、歌词、帖子等，代码如下：

```go
var (
    titleWords       []string
    descriptionWords []string
)
for _, i := range sego.SegmentsToSlice(segmenter.Segment([]byte(cargo.Title)), true) {
    if len(i) <= 1 {
        continue
    }

    titleWords = append(titleWords, i)
}
for _, i := range sego.SegmentsToSlice(segmenter.Segment([]byte(cargo.Description)), true) {
    if len(i) <= 1 {
        continue
    }

    descriptionWords = append(descriptionWords, i)
}
```

下一步就是把分词结果更新进去：

```go
sql := `UPDATE cargos SET tokens = setweight(to_tsvector('simple', $1), 'A') || setweight(to_tsvector('simple', $2), 'B') WHERE id = $3`

_, err = tx.Exec(ctx, sql, strings.Join(titleWords, " "), strings.Join(descriptionWords, " "), cargo.ID)
```

使用 `to_tsvector` 来设置向量，前面的 `simple` 表示按照空格切分，如果不给的话，默认按照 `english` 来进行。`setweight` 函数
用来将 `to_tsvector` 的结果设置权重，权重选择有 `ABCD` 四种，如上文所说。`||` 用来将多个分词向量合并。

这种模式其实就是完全在应用层控制分词，网上可以搜索到的案例，大部分都是基于编译PG插件的形式。我更倾向于应用层分词，这样的
好处在于：

- 维护简单，不用编译。如果是云托管的PG，可能无法加载自编译插件
- 便于扩展，应用层水平扩展要比数据库水平扩展简单很多。分词本身是一个CPU密集型工作，放在数据库很容易达到瓶颈
- 应用更新速度快，分词插件有了什么新功能新特性，更新起来会非常简单

到这一步，我们就已经把分词向量更新进去了。接下来我们还需要创建索引：

```sql
create extension pg_trgm;

CREATE INDEX IF NOT EXISTS cargos_token_idx ON cargos USING GIN(tokens);
```

我使用的是GIN索引，除此之外，还可以选择 GiST，具体区别见：https://www.postgresql.org/docs/current/textsearch-indexes.html

## 搜索

保存完了数据之后，接下来我们要做的事情就是搜索：

```go
var queryWords []string
for _, i := range sego.SegmentsToSlice(segmenter.Segment([]byte(query)), true) {
    if len(i) <= 1 {
        continue
    }

    queryWords = append(queryWords, i)
}

sql := `SELECT id, created_at, updated_at, title, ts_rank(tokens, query) AS score FROM cargos, to_tsquery('simple', $1) query WHERE tokens @@ query ORDER BY score DESC`

rows, err := tx.Query(ctx, sql, strings.Join(queryWords, " | "))
// ...
```

`to_tsquery` 是解析查询语句，有如下语法(参考 [文档](https://www.postgresql.org/docs/current/datatype-textsearch.html#DATATYPE-TSQUERY))：

- `&` 表示两个条件都要满足
- `|` 表示满足其中一个
- `!` NOT运算表示不匹配
- `<->` 表示A词语以后跟随B词语，差不多就是 `A...B` 这样的结构
- `blabla*` 表示符合前缀

`ts_rank` 就是根据权重进行计算，方便后续的 `ORDER BY` 排名。

## 性能测试

我把博客的所有文章都重复导入了很多次，凑足了1亿分词，以下是实测数据：

```sql
# \timing on
Timing is on.
# SELECT id, ts_rank(tokens, query) AS score FROM cargos, to_tsquery('simple', '共同 & 富裕') query WHERE tokens @@ query ORDER BY score DESC;
Time: 1.775 ms  ; 1 行结果
# SELECT id, ts_rank(tokens, query) AS score FROM cargos, to_tsquery('simple', 'Linux') query WHERE tokens @@ query ORDER BY score DESC;
Time: 317.306 ms  ; 47775 行结果
# SELECT id, ts_rank(tokens, query) AS score FROM cargos, to_tsquery('simple', 'Go & 并发') query WHERE tokens @@ query ORDER BY score DESC;
Time: 111.316 ms  ; 16170 行结果
# SELECT id, ts_rank(tokens, query) AS score FROM cargos, to_tsquery('simple', 'Windows & 虚拟机') query WHERE tokens @@ query ORDER BY score DESC;
Time: 57.231 ms  ; 8085 行结果

# SELECT COUNT(*) FROM cargos;
 count  
--------
 360152
(1 row)

Time: 48.547 ms
# SELECT SUM(LENGTH(tokens)) FROM cargos;
    sum    
-----------
 101115491
(1 row)

Time: 477.441 ms
```

可以看到，一共是 36 万篇文章，分词后一共约 1亿 词语，搜索时，响应时间与返回结果数量基本呈正比关系，如果搜索结果少时，
响应是非常快的。我认为常见场景PG完全足够覆盖。目前主要还是分词过于粗略，还有很多地方可以优化，例如，去除常见语气词，去除
标点符号、特殊字符，去除大部分的无用词汇，使用 TF-IDF 提取关键字赋予更高权重，其余词语降低权重，加上这些优化之后，
整体表现应该会好很多。

## 总结

这篇文章总结了一下我折腾PG当搜索引擎的经历，经过验证，PG是完全胜任常见场景的，以后我自己做一些什么需要搜索能力时，相信
这个方案可以让整体更加简单。

---

参考资料：

- https://en.wikipedia.org/wiki/Tf%E2%80%93idf
- https://zh.wikipedia.org/wiki/%E5%80%92%E6%8E%92%E7%B4%A2%E5%BC%95
