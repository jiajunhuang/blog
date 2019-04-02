# Golang/Python最佳实践

## 统一好返回格式

- 使用gRPC
- 如果使用RESTful风格，那么返回结果无论成功还是失败，都应该遵循如下格式：

```js
{
    "code": 200, // 业务代码code，可以用于详细错误判断
    "msg": "succeed", // 给开发人员看的错误提示
    "data": {} // 无论成功或者失败，有无数据，此处都应当返回一个dict/map而不应该是nil/None
}
```

为什么要这么做呢？因为使用动态语言开发时不会有太大差别，使用静态语言时，数据类型不一将会导致不好定义返回结果。

## 数据库

- PG比MySQL好用，数据类型支持更加丰富，坑也更少，唯一的缺点是流行程度不如MySQL
- 使用Redis作缓存，Redis的数据结构更加丰富，而且数据结构可以用作其他用途

## Python

- gunicorn + gevent异步比asyncio/tornado等更加方便
- [Flask](http://flask.pocoo.org/) 总体来说是一个好用的web框架
- ORM使用 [SQLAlchemy](https://www.sqlalchemy.org/) ，migration使用 [alembic](https://alembic.sqlalchemy.org/en/latest/)
- 参数校验使用 [marshmallow](https://github.com/marshmallow-code/marshmallow)
- 异步任务使用 gevent + python-rq，celery并发一高就容易遇到莫名假死的问题
- 善用decorator，例如参数校验时，可以写如下代码：

```python
@hello_bp.route("/")
@binding_schemma(HelloworldSchema)
def hello(data):
    return "hello world"
```

`binding_schemma` 的实现如下：

```python
def binding_schemma(schema):
    def wrapper(func):
        def inner(*args, **kwargs):
            # 获取参数
            arguments = request.get_json() or request.form or request.args
            if arguments is None:
                return failed(msg="请检查参数")

            # 校验
            try:
                data = schema().load(arguments)
            except ValidationError as err:
                return failed(msg=err.messages)

            # 执行函数
            return func(data.data, *args, **kwargs)

        return inner
    return wrapper
```

## Golang

- Echo比GIN体验上好用一些
- ORM还是pg好用一些，此外更喜欢 sqlx + SQLAlchemy + alembic
