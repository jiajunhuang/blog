# Web开发系列(六)：关系型数据库，ORM

数据库，web开发中总是离不开这个核心应用，可以说web开发的核心就是数据库。但数据库是一个泛称，通常我们说的
数据库是指关系型数据库，此外还有非关系型数据库。在这篇文章中，我们单指关系型数据库。

相信大家都看过excel表格，在PostgreSQL中，每个数据库包含多个schema，每个schema中可以包含多个表，而这个表，
就好比是我们平时见到的excel表格，但是它的特点是列数是固定的，行数是可变的。当然，列也可以通过关系型数据库管理系统
来更改。什么叫做关系型数据库管理系统呢？

我们要这样想，数据库是用来管理数据的，数据是要存储在操作系统的磁盘上的，而我们平常所说的数据库，实际上就是帮我们
管理这些文件（以及其他相关东西）的管理系统。而我们操纵数据库便是通过SQL来完成的。

常见的关系型数据库管理系统（以下简称数据库）有SQLite，MySQL，Postgesql等。他们各有特点，例如SQLite一般用于移动设备，
而MySQL和PostgreSQL则用于服务器。

在日常开发中有一个必须要搞清楚的概念，就是 [表的连接](https://en.wikipedia.org/wiki/Join_%28SQL%29) 如果你去面试，
那么将会有很大概率被问到：什么是内连接，什么是外连接。。。等等。

## SQLAlchemy

在日常开发中，一般我们很少直接写SQL，一般都会用ORM：Object-relational mapping，通过ORM我们可以把类和数据库中的表
对应起来，例如Python中ORM的黄金标准便是 [SQLAlchemy](https://www.sqlalchemy.org/)。

学习SQLAlchemy我推荐阅读官方的教程：http://docs.sqlalchemy.org/en/latest/orm/tutorial.html

我们来看一个SQLAlchemy的例子，声明一个表：

```python
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy import Column, Integer, String


Base = declarative_base()


class User:
    __tablename__ = 'users'

    id = Column(Integer, primary_key=True)
    name = Column(String(32))
    fullname = Column(String(32))
    nickname = Column(String(64), nullable=False)
    passwd = Column(String(128), nullable=False)
```

而查询，则是类似这样：

```python
>>> session.query(User).filter(
        User.id == 1
    ).first()
```

SQLAlchemy会自动把上述的Python代码翻译成SQL，执行，然后映射成User类的示例，于是结果便可以 `user.name` 这样用。

## migration工具

用SQLAlchemy的时候我们还能获得一个好处，便是我们可以很方便的知道表是怎么定义的，此外，我们还可以借助
[alembic](http://alembic.zzzcomputing.com/en/latest/) 来做数据库变更记录，但是用alembic的时候我们也许不想手动
写变更，想让alembic自己生成，那么我们要好好地组织SQLAlchemy的代码，让SQLAlchemy的metadata可以记录下所有的表，然后
在alembic的配置里引入这个metadata，便可以通过alembic生成。
