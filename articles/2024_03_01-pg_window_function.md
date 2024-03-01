# 窗口函数的使用(以PG为例)

很少做数据统计，之前一直没有接触和使用窗口函数。今天看了一下文档，发现在统计领域，窗口函数非常强大，当然，缺点就是把计算量
移到了数据库这一层，但是没关系，对于少量数据，直接一条SQL解决，cool！

在 SQL 中，窗口函数是一种特殊类型的函数，可以在一组相关的行（称为"窗口"）上执行计算。窗口函数可以解决很多数据统计的功能，
例如包括计算移动平均、总计、累计和排名等。

首先我们看下语法：

```SQL
SELECT depname, empno, salary, avg(salary) OVER (PARTITION BY depname) FROM empsalary;
```

它的主要特点，就是有一个 `OVER` 子句，子句里一般会包含 `PARTITION BY` 语句，当然也可以不包含。我们来建一个表，并且插入一些
数据：

```SQL
postgres=# create table empsalary(depname text, empno int, salary float);
CREATE TABLE
postgres=# insert into empsalary(depname, empno, salary) values ('develop', 11, '5200');
INSERT 0 1
postgres=# insert into empsalary(depname, empno, salary) values ('develop', 7, '4200');
INSERT 0 1
postgres=# insert into empsalary(depname, empno, salary) values ('develop', 9, '4500');
INSERT 0 1
postgres=# insert into empsalary(depname, empno, salary) values ('develop', 8, '6000');
INSERT 0 1
postgres=# insert into empsalary(depname, empno, salary) values ('develop', 10, '5200');
INSERT 0 1
postgres=# insert into empsalary(depname, empno, salary) values ('personnel', 5, '3500');
INSERT 0 1
postgres=# insert into empsalary(depname, empno, salary) values ('personnel', 2, '3900');
INSERT 0 1
postgres=# insert into empsalary(depname, empno, salary) values ('personnel', 3, '4800');
INSERT 0 1
postgres=# insert into empsalary(depname, empno, salary) values ('sales', 1, '5000');
INSERT 0 1
postgres=# insert into empsalary(depname, empno, salary) values ('sales', 4, '4800');
INSERT 0 1
```

然后执行上述SQL：

```SQL
postgres=# SELECT depname, empno, salary, avg(salary) OVER (PARTITION BY depname) FROM empsalary;
  depname  | empno | salary |        avg
-----------+-------+--------+--------------------
 develop   |    11 |   5200 |               5020
 develop   |     7 |   4200 |               5020
 develop   |     9 |   4500 |               5020
 develop   |     8 |   6000 |               5020
 develop   |    10 |   5200 |               5020
 personnel |     5 |   3500 | 4066.6666666666665
 personnel |     2 |   3900 | 4066.6666666666665
 personnel |     3 |   4800 | 4066.6666666666665
 sales     |     1 |   5000 |               4900
 sales     |     4 |   4800 |               4900
(10 rows)
```

可以看到，输出中，前三列是数据库里原来的数据，第四列是 avg(salary)，整个数据已经按 `depname` 分区，然后区域内再计算avg。

再看一个例子：

```SQL
postgres=# SELECT depname, empno, salary,
       rank() OVER (PARTITION BY depname ORDER BY salary DESC)
FROM empsalary;
  depname  | empno | salary | rank
-----------+-------+--------+------
 develop   |     8 |   6000 |    1
 develop   |    10 |   5200 |    2
 develop   |    11 |   5200 |    2
 develop   |     9 |   4500 |    4
 develop   |     7 |   4200 |    5
 personnel |     3 |   4800 |    1
 personnel |     2 |   3900 |    2
 personnel |     5 |   3500 |    3
 sales     |     1 |   5000 |    1
 sales     |     4 |   4800 |    2
(10 rows)
```

可以看到，输出仍然是按 `depname` 分区，然后区域内进行排名。

再来看一个例子：

```SQL
postgres=# SELECT salary, sum(salary) OVER () FROM empsalary;
 salary |  sum
--------+-------
   5200 | 47100
   4200 | 47100
   4500 | 47100
   6000 | 47100
   5200 | 47100
   3500 | 47100
   3900 | 47100
   4800 | 47100
   5000 | 47100
   4800 | 47100
(10 rows)
```

这个例子里，`OVER()` 里没有子句，因此是对全局产生作用，整个作为一个窗口，然后计算 `sum(salary)`。

通过这三个简单的例子，可以一窥窗口函数的强大，一些常规的计算和统计任务，可以一条SQL直接解决，例如年级成绩排名，按科目排名等等。


---

Refs:

- https://www.postgresql.org/docs/current/tutorial-window.html
