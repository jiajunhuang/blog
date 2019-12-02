# SQLAlchemy快速更新或插入对象

写脚本刷数据的时候，常常有这样的需求：如果对象存在，那么更新数据，否则，插入数据。
有可能数据源的数据比schema的字段更多，这种时候，就需要想办法把SQLAlchemy中的schema字段取出来，
只取需要的字段了。假设我们的schema叫做 `User`：

```python
valid_columns = set(User.__table__.columns.keys())  # 获取model中定义的columns
```

这样就能获取到合法的schema字段，然后我们把数据源的数据过滤：

```python
valid_attrs = {k: v for k, v in i.items() if k in valid_columns}
```

如果其中有需要单独处理的字段，我们进行特殊处理：

```python
adjust_time(valid_attrs)


def adjust_time(attrs):
    # mongo存储的是毫秒时间戳
    attrs["created_at"] = datetime.datetime.fromtimestamp(attrs["created_at"] / 1000.0)
    attrs["updated_at"] = datetime.datetime.fromtimestamp(attrs["updated_at"] / 1000.0)
    if attrs.get("deleted_at"):
        attrs["deleted_at"] = datetime.datetime.fromtimestamp(attrs["deleted_at"] / 1000.0)
```

之后我们就可以愉快的对SQLAlchemy对象进行更新或者插入了：

```python
already_exist = User.get_by_id(s, i["user_id"])
if already_exist:
    logging.info("update item: %s", i)
    User.update_by_user_id(s, already_exist.user_id, valid_attrs)
else:
    logging.info("insert item: %s", i)
    s.add(User(**valid_attrs))
```

其中，model中的 `update_by_user_id`定义如下：

```python
@classmethod
def update_by_user_id(cls, session, user_id, attr_map):
    session.query(cls).filter(cls.user_id == user_id).update(attr_map)
```

搞定！
