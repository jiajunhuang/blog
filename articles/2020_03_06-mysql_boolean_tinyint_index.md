# MySQL Boolean类型的坑

MySQL中，Boolean只是 tinyint(1) 的别名，也就是说，MySQL中并没有真正的bool类型。而SQLAlchemy生成SQL的时候并没有检测到
这一点，这就导致一个问题，当使用 bool 类型作为查询条件时，用不上索引，从而导致扫表的行为：

```sql
> SELECT COUNT(*) FROM message WHERE message.is_national = 1 AND message.updated_at > '2020-01-01 00:00:00' AND message.deleted_at IS NULL;
+----------+
| COUNT(*) |
+----------+
| 0        |
+----------+
1 row in set
Time: 0.018s
> SELECT COUNT(*) FROM message WHERE message.is_national is true AND message.updated_at > '2020-01-01 00:00:00' AND message.deleted_at IS NULL;
+----------+
| COUNT(*) |
+----------+
| 0        |
+----------+
1 row in set
Time: 2.162s
```

注意观察第一行和第二行的时间，很明显第二行没有用上索引，我们来看看 EXPLAIN 的结果便知道了：

```
> EXPLAIN SELECT COUNT(*) FROM message WHERE message.is_national = 1 AND message.updated_at > '2020-01-01 00:00:00' AND message.de
        leted_at IS NULL;
| id | select_type | table | type | possible_keys | key | key_len | ref | rows | Extra |
| 1  | SIMPLE | message | ref  | ix_message_updated_at,idx_updated_at_is_national,ix_message_is_national | ix_message_is_national | 1 | const | 1 | Using where |

> EXPLAIN SELECT COUNT(*) FROM message WHERE message.is_national is true AND message.updated_at > '2020-01-01 00:00:00' AND messag
        e.deleted_at IS NULL;
| id | select_type | table   | type | possible_keys | key    | key_len | ref    | rows    | Extra |
| 1  | SIMPLE | message | ALL  | ix_message_updated_at,idx_updated_at_is_national | <null> | <null>  | <null> | 一个很大的数字 | Using whe
re |
```

对此，我只想说，太坑了！

---

参考资料：

- https://dev.mysql.com/doc/refman/8.0/en/numeric-type-syntax.html
