# MySQL Index Condition Pushdown Optimization

比如数据库中有如下表:

```sql
show create table people\G
*************************** 1. row ***************************
       Table: people
Create Table: CREATE TABLE `people` (
  `zipcode` varchar(16) NOT NULL,
  `lastname` varchar(32) NOT NULL,
  `address` varchar(32) NOT NULL,
  KEY `idx_people_zipcode_lastname_address` (`zipcode`,`lastname`,`address`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci
1 row in set (0.001 sec)
```

我们执行如下查询：

```sql
SELECT * FROM people
  WHERE zipcode='95054'
  AND lastname LIKE '%etrunia%'
  AND address LIKE '%Main Street%';
```

如果没有 index condition pushdown optimization 的话，那么执行步骤如下：

1. 对比 `idx_people_zipcode_lastname_address` 索引中 `zipcode` 是否能匹配，如果可以匹配，获取整个数据行；
2. 对比 `WHERE` 中的条件，看是否可以全部匹配；

而有了 index condition pushdown optimization 之后，步骤则变成了：

1. 对比 `idx_people_zipcode_lastname_address` 索引中的 `zipcode` 是否能匹配，同时对比 `WHERE` 语句中，index中包含的部分，看是否可以通过；
2. 如果可以通过，获取整个数据行；
3. 对比 `WHERE` 中的条件，看是否可以全部匹配；

这两者的区别就在于，是否充分利用了索引中的值，提前进行了数据的过滤。我们在 `EXPLAIN` 的时候，如果当前数据库查询使用了这项优化，则会在 `Extra`
那一列显示 `Using index condition`。


---

ref:

- https://dev.mysql.com/doc/refman/5.7/en/index-condition-pushdown-optimization.html
