# MySQL charset不同导致无法使用索引的坑

今天排查了一个MySQL Charset不同导致无法使用索引的坑。

当然最开始我是不知道的，同事碰到一个性能问题，我也在群里，因此就捞过界了一把，一起看了一下问题。但是从SQL来说应该是
充分利用了SQL才对，所有该有索引的地方都有。原SQL经脱敏简化后如下：

```sql
SELECT A.* FROM A LEFT JOIN B ON A.xxx=B.xxx AND B.yyy='yyy' INNER JOIN A AS A_1 ON A.xxx=A1.xxx WHERE A.yyy='yyy';
```

不过原SQL因为有left join又有inner join，看得我有点晕，于是我直接问原始需求是什么。原来是想看在A表不在B表，所以我改写
成这样：

```sql
SELECT A.* FROM A LEFT JOIN B ON A.xxx=B.xxx WHERE B.id IS NULL;
```

EXPLAIN一下，发现Explain输出里，ref是空，rows却是接近全表的行数，这说明虽然key和extra上表明用上了索引，但实际上没有。
实际执行一下也验证了，确实没有用上索引，查询时间非常久。

逐一检查各表之后，发现该加索引的地方都加了，但是却没有用上索引，最后发现是因为字符集不一样。`SHOW CREATE TABLE xxx`
之后发现，一张表默认为 `utf8`，另外一张表默认为 `utf8mb4`，把字符集都改成utf8mb4之后，就可以正常工作了。

---

参考链接：

- https://stackoverflow.com/questions/18660252/mysql-why-does-left-join-not-use-an-index
