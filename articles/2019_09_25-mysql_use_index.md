# MySQL指定使用索引(使用索引提示)

> 也叫index hint

可以看到官方文档中，BNF如下：

```
index_hint_list:
    index_hint [index_hint] ...

index_hint:
    USE {INDEX|KEY}
      [FOR {JOIN|ORDER BY|GROUP BY}] ([index_list])
  | {IGNORE|FORCE} {INDEX|KEY}
      [FOR {JOIN|ORDER BY|GROUP BY}] (index_list)

index_list:
    index_name [, index_name] ...)
```

也就是说，指定索引的时候，可以同时指定多个，但是MySQL一定只会在指定的索引中选一个使用，所以一定要考虑清楚SQL语句是否
能利用上索引：

```
UPDATE xxx_table USE INDEX (index_a, index_b) SET updated_at='2019-09-25 00:00:00' where xxx
```

与之相反的使用，是 `IGNORE INDEX (index_a, index_b)` 这会告诉MySQL，避免使用这几个索引。

使用索引还有一种用法，那就是使用 `FORCE INDEX`，与 `USE INDEX` 的区别在于，这会告诉MySQL，扫表的代价非常昂贵，因此，除非
用不到索引，否则MySQL一定不会扫表，至少会用到其中提供的一个索引。

---

参考资料：

- [MySQL官方文档](https://dev.mysql.com/doc/refman/8.0/en/index-hints.html)
