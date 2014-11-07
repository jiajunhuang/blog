---
layout: post
title: Mysql manual 笔记
---

### SELECT 相关

* 表名是大小写敏感的；`SELECT`语句中的String是大小写**不**敏感的

* `SELECT`语句中`AND`具有比`OR`更高的优先级, 尽管如此， 用括号来表明优先级会更好

* `SELECT` 只会简单地取回数据， 会包含其中重复的数据，想要不重复， 加上`DISTINCT`关键字:

```sql
SELECT DISTINCT <column-name> FROM <table-name>;
```

* `AS` 关键字只对它前面一个列名有效

* 判断值不为空用`<column-name> IS NOT NULL`而**不是**`<column <> NULL`因为`NULL` 是一个不能用`<>`不等号比较的特殊值

### ORDER BY 相关

* 对`SELECT`语句选择出来的结果进行排序使用`ORDER BY`， 默认是按照升序排列的， 想降序， 在最后加`DESC`, **`ORDER BY`语句应该放在一个SQL语句的最后面， 基本上逗号就跟在`ORDER BY`后面, **注意`DESC`关键字只对它前面一个列名有效**

* `ORDER BY`的结果也是大小写不敏感的， 也就是说当出现除大小写外相同的列时， 结果是未定义的， 可以加上`BINARY`关键字强制进行大小写敏感排序：`ORDER BY BINARY <column-name>`, 在mysql-5.7上进行实验， 加不加`BINARY`关键字结果都是大写在前， 小写在后



### 其它语句

* 常见语句如下

```sql
LOAD DATA LOCAL INFILE '/path/to/your/file' INTO TABLE <table-name>;
SHOW DATABASES; -- 显示数据库
SHOW TABLES: -- 显示表
DESCRIBE <table-name>; -- 显示表的详细结构
```
### Mysql 函数

* `TIMESTAMPDIFF()`函数接受参数如下：你想要表达的那一部分(例如YEAR)， 再加两个日期

* `CURDATE()` 获取当前日期
