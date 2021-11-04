# SQL 防注入及原理

SQL 注入一直是 Web 安全中非常常见的攻击手段，也是非常严重的安全漏洞，我在使用 SQLAlchemy 的时候，想要防止 like 语句被
注入，最后发现 SQLAlchemy 已经做了这一层的处理。我们来看看其原理。

## 如何注入

防止 SQL 注入，我们首先要知道什么是 SQL 注入。SQL 注入的先决条件，就是服务端没有仔细校验用户的输入，并且直接把用户的
输入拼接到 SQL 中然后去执行，例如 `"SELECT * FROM user WHERE name='{}'".format(user_name)`，如果用户输入的正常的数据，
例如 `admin`，那么这条语句拼接出来，就会变成：

`SELECT * FROM user WHERE name='admin'`

但是如果用户输入的是 `' OR '1'='1`，那么这条语句就会变成：

`SELECT * FROM user WHERE name='' OR '1' = '1'`

整个 SQL 的语义就变了，相当于加上了一个 OR 条件，而 `'1' = '1'` 永远为真，我们来执行一下，为了便于观察，我把 `SELECT *`
改成 `SELECT COUNT(*)`：

```bash
mysql root@(none):gitea> select COUNT(*) from `user` where name='admin';        
+----------+
| COUNT(*) |
+----------+
| 0        |
+----------+
1 row in set
Time: 0.035s
mysql root@(none):gitea> select COUNT(*) from `user` where name='admin' OR '1'='
                     ->  1';            
+----------+
| COUNT(*) |
+----------+
| 1        |
+----------+
1 row in set
Time: 0.036s
```

那么有同学可能会问，这样的危害是什么？危害可大了，如果可以通过构造输入的数据，实现执行自己想要的 SQL 语句，那么攻击者
完全可以不要输入 `' OR '1'='1` 这样的简单语句，可以输入例如 `DROP`，或者是给数据库改密码，把整个数据库都 dump 下来等等。

## SQL 的执行过程

我们先来了解一下 SQL 语句的执行过程，基本上就是：

- 解析。由于我们输入的 SQL 语句，本质上是一个字符串，所以首先要解析成语法树，这样 db server 才能理解
- 语法检查。db server 开始对输入的 SQL 语句进行语法检查，看是否符合数据库支持的语法规范、是否有权限、是否存在数据库和表等
- SQL优化。优化引擎对 SQL 语句进行优化
- 执行。

结合上面一节，可以发现，如果想要注入，那么就需要把构造的语句让数据库解析成语法树去执行，因此有了
[Prepared statement](https://en.wikipedia.org/wiki/Prepared_statement)，通过这个，可以提前把 SQL 语句编译好，通过
替换符来替代之前的拼接，这样子用户的输入就可以被当作普通字符串来处理。

使用示例为：

```go
// AlbumByID retrieves the specified album.
func AlbumByID(id int) (Album, error) {
    // Define a prepared statement. You'd typically define the statement
    // elsewhere and save it for use in functions such as this one.
    stmt, err := db.Prepare("SELECT * FROM album WHERE id = ?")
    if err != nil {
        log.Fatal(err)
    }

    var album Album

    // Execute the prepared statement, passing in an id value for the
    // parameter whose placeholder is ?
    err := stmt.QueryRow(id).Scan(&album.ID, &album.Title, &album.Artist, &album.Price, &album.Quantity)
    if err != nil {
        if err == sql.ErrNoRows {
            // Handle the case of no rows returned.
        }
        return album, err
    }
    return album, nil
}
```

## SQLAlchemy 的 like

我们来看下 like 方法的文档：

```python
    def like(self, other, escape=None):
        r"""Implement the ``like`` operator.

        In a column context, produces the expression::

            a LIKE other

        E.g.::

            stmt = select(sometable).\
                where(sometable.c.column.like("%foobar%"))

        :param other: expression to be compared
        :param escape: optional escape character, renders the ``ESCAPE``
          keyword, e.g.::

            somecolumn.like("foo/%bar", escape="/")

        .. seealso::

            :meth:`.ColumnOperators.ilike`

        """
        return self.operate(like_op, other, escape=escape)
```

官方文档说明已经可以处理用户拼接的输入，于是我打开执行日志来看看：

```log
2021-11-04 09:41:15,066 INFO sqlalchemy.engine.Engine BEGIN (implicit)
INFO:sqlalchemy.engine.Engine:BEGIN (implicit)
2021-11-04 09:41:15,085 INFO sqlalchemy.engine.Engine SELECT user.id AS user_id FROM user WHERE user.name LIKE %(name_1)s
INFO:sqlalchemy.engine.Engine:SELECT user.id AS user_id FROM user WHERE user.name LIKE %(name_1)s
2021-11-04 09:41:15,085 INFO sqlalchemy.engine.Engine [generated in 0.00030s] {'name_1': "%admin%'; DROP TABLE identifier;"}
INFO:sqlalchemy.engine.Engine:[generated in 0.00030s] {'name_1': "%admin%'; DROP TABLE identifier;"}
[]

```

确实就是使用了 prepared statements。

## 总结

这一篇文章我们首先看了一下什么是 SQL 注入，然后了解了一下 SQL 注入的先决条件，因此引入了如何防止 SQL 注入，以及原理。
最后我们看了一下 Go 和 Python 的两个例子，以及验证了 SQLAlchemy 的实现。希望对读者有所帮助。

---

ref:

- https://en.wikipedia.org/wiki/Prepared_statement
- https://www.hackedu.com/blog/how-to-prevent-sql-injection-vulnerabilities-how-prepared-statements-work
- https://snyk.io/blog/sql-injection-cheat-sheet/
