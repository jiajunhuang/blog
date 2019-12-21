# SQLAlchemy使用主从与数据库autocommit

autocommit，意思就是自动提交。它代表着这么一个行为，如果autocommit设置为True（或1），意味着，每一个传输到DBMS的
SQL都会被当作一个事务来执行，并且提交。用MySQL举个例子：

```sql
-- client 1
> select @@autocommit;
+--------------+
| @@autocommit |
+--------------+
|            1 |
+--------------+
1 row in set (0.000 sec)

> create table user (id INTEGER PRIMARY KEY, name VARCHAR(32));
Query OK, 0 rows affected (0.510 sec)

> INSERT INTO user(id, name) VALUES (1, "jhon");

-- client 2
> INSERT INTO user(id, name) VALUES (1, "jhon");
ERROR 1062 (23000): Duplicate entry '1' for key 'PRIMARY'
```

上面的例子中，client 1所执行的SQL，没有 `COMMIT` 语句，但是数据库会自动为它执行 `COMMIT` 操作，因此，当client 2
执行插入相同主键的数据时，MySQL会报错。

而当关闭 autocommit 时会发生什么呢？我实验了一下，关闭autocommit之后，如果插入相同主键的数据，client 2会卡住，一
直到client 1执行 `COMMIT` 之后，然后就开始报错。

然而当插入不同主键时，则不会出现卡住的问题，但是两个client也互相看不到对方插入的数据，这是MVCC隔离层级带来的效果。

而当 autocommit 为1时，两个client则互相可以看到对方插入的数据，因为数据库为每一条SQL都执行了隐式提交(implict commit)。

client 1 的SQL：

```sql
> create table user (id INTEGER PRIMARY KEY, name VARCHAR(32));
Query OK, 0 rows affected (0.752 sec)

> select @@autocommit;
+--------------+
| @@autocommit |
+--------------+
|            1 |
+--------------+
1 row in set (0.000 sec)

> INSERT INTO user(id, name) VALUES (1, "jhon");
Query OK, 1 row affected (0.399 sec)

-- 此处去执行client 2的SQL

> select * from user;
+----+------+
| id | name |
+----+------+
|  1 | jhon |
|  2 | jhon |
+----+------+
2 rows in set (0.001 sec)
```

client 2 的SQL：

```sql
> use stock;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed
> select @@autocommit;
+--------------+
| @@autocommit |
+--------------+
|            1 |
+--------------+
1 row in set (0.000 sec)

> select * from user;
+----+------+
| id | name |
+----+------+
|  1 | jhon |
+----+------+
1 row in set (0.001 sec)

> INSERT INTO user(id, name) VALUES (2, "jhon");
Query OK, 1 row affected (0.064 sec)

> select * from user;
+----+------+
| id | name |
+----+------+
|  1 | jhon |
|  2 | jhon |
+----+------+
2 rows in set (0.000 sec)
```

---

实验就做到这里。结论就是：

- 当打开autocommit时，DMBS会为每一条SQL执行一个事务，也就是说，每一条SQL都是立即生效的。
- 当关闭autocommit时，客户端必须手动显式声明事务的开始和结束，具体能不能读到其它客户端产生的数据得看MVCC隔离层级的设置。

---

那为什么突然查起autocommit呢？因为和同事的聊天中提到这个，发现以前用SQLAlchemy一直无法用主从，原因是我们的主从架构是
这样的：

```
SQLAlchemy -> Kingshard/阿里数据库中间件 -> MySQL
```

而SQLAlchemy推荐将autocommit关闭，并且默认也是关闭的。其文档中有这么一段话：

```
“autocommit” mode is a legacy mode of use and should not be considered for new projects.
If autocommit mode is used, it is strongly advised that the application at least ensure
that transaction scope is made present via the Session.begin() method, rather than using
the session in pure autocommit mode.
```

那为何无法使用Kingshard呢？Kingshard这类主从中间件的原理是，将每一个SQL，如果是写的，那么转发到master执行，如果是
读，那么转发到slave执行。这里有一个问题，只有当 `AUTOCOMMIT` 为1时，才能很方便的转发，因为每一条SQL都是一个单独的
事务。当 `AUTOCOMMIT` 为0时，一个事务里可能有多个语句，而这些个语句可能既有读，又有写，因为Kingshard无法在事务
开始的时候就判断未来到底有没有读和写，因此不好转发到slave，所以就干脆转发到master。

这就会导致SQLAlchemy无法愉快的使用主从。当然了，选项就是，你可以在SQLAlchemy中把 autocommit 设置为True，或者重写
`get_bind` 函数来自动转发主从，或者初始化数据库的时候，区分好读和写，然后在使用的时候用不同的即可，我比较推荐
第三种方式，因为可以写这么一个函数来帮助我们：

```python
@contextlib.contextmanager
def get_session(rw=True):
    s = Session() if rw else ReadOnlySession()
    try:
        yield s
        s.commit()
    except Exception:
        s.rollback()
        raise
    finally:
        s.close()
```

---

参考资料：

- [维基百科词条](https://en.wikipedia.org/wiki/Autocommit)
- [MySQL autocommit相关文档](https://dev.mysql.com/doc/refman/5.7/en/innodb-autocommit-commit-rollback.html)
- [SQLAlchemy相关文档](https://docs.sqlalchemy.org/en/13/orm/session_transaction.html)
