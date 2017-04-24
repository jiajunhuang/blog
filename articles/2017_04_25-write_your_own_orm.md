# 写一个简单的ORM

最近在研究SQLAlchemy源码，自己造个轮子先（当然，SQLAlchemy远比这里复杂）

```python
import logging

logging.basicConfig(level=logging.DEBUG)


class Column:
    def __init__(self, type, default=None):
        self.col_name = None
        self.col_type = type
        self.col_default = default

    def __eq__(self, value):
        logging.debug("__eq__ of {} && {}".format(self, value))
        return "{} = {}".format(self.col_name, value)


class Base(type):
    def __new__(cls, *args, **kwargs):
        logging.debug("creating class {}, args {}, kwargs {}".format(
            cls, args, kwargs
        ))

        if len(args) > 2:
            for k, v in args[2].items():
                if k.startswith("_"):
                    continue

                v.col_name = k

        return type.__new__(cls, *args, **kwargs)


class User(metaclass=Base):
    __tablename__ = "user"

    id = Column(int)
    name = Column(str)
    passwd = Column(str)


class Queryable:
    def __init__(self, table):
        self.table = table

    def filter(self, text):
        return "select * from {} where {}".format(
            self.table.__tablename__,
            text,
        )


class Session:
    def query(self, table):
        logging.debug("returning Queryable({})".format(table))
        return Queryable(table)


session = Session()
print(session.query(User).filter(User.id == 1))
```

运行结果：

```bash
jiajun@debian test: python3 orm.py
DEBUG:root:creating class <class '__main__.Base'>, args ('User', (), {'__qualname__': 'User', '__tablename__': 'user', 'id': <__main__.Column object at 0x7f8174b107b8>, 'passwd': <__main__.Column object at 0x7f8174b10940>, 'name': <__main__.Column object at 0x7f8174b10828>, '__module__': '__main__'}), kwargs {}
DEBUG:root:returning Queryable(<class '__main__.User'>)
DEBUG:root:__eq__ of <__main__.Column object at 0x7f8174b107b8> && 1
select * from user where id = 1
```
