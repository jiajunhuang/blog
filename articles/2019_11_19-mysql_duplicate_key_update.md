# MySQL的ON DUPLICATE KEY UPDATE语句

有这么一种场景：

- 查找记录
    - 如果存在：更新字段
    - 如果不存在：插入字段

如果使用ORM来表述的话，则比较长，而且会出现并发问题，即多个提交时，假设设置了唯一索引的情况下，会发生数据冲突，然后
就会隔三差五收到错误：duplicate key 'xxx'。

因此使用MySQL提供的 `INSERT ... ON DUPLICATE KEY UPDATE` 语句，这是MySQL的扩展语法，因此也就意味着，使用了这个语句之后，
数据库基本上就被绑定在MySQL上了，不过没有关系，一般谁会轻易更换数据库呢？

这个语句的语法是这样的：

```sql
INSERT INTO t1 (a,b,c) VALUES (1,2,3) ON DUPLICATE KEY UPDATE c=c+1, b=4;
```

分三段来理解：

- 第一段，常规的INSERT语句。`INSERT INTO <table>(col1, col2, ...) VALUES (val1, val2, ...)`
- 第二段，`ON DUPLICATE KEY`，表示后面的语句是当数据有冲突的情况下会执行的
- 第三段，UPDATE语句。`UPDATE a=1, b=2`

注意，由于有 `ON DUPLICATE KEY`，也就是说必须得有字段会发生冲突。什么属性的字段能冲突呢？

- 主键(Primary Key)
- 唯一索引(Unique Key)

把代码使用这个语句之后，世界都安静了 :-)

---

参考资料：

- [MySQL官方文档](https://dev.mysql.com/doc/refman/8.0/en/insert-on-duplicate.html)
