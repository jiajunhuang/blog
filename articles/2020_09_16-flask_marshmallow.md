# 结合Flask 与 marshmallow快速进行参数校验

在Flask里如何方便快速的进行参数校验呢？如下，我们通过组合Flask提供的工具函数，以及marshmallow，来完成一个方便快捷的参数
校验函数。

```python
from flask import abort, make_response, jsonify
from marshmallow import Schema, fields, ValidationError


def check_data(schema, data):
    try:
        return schema().load(data)
    except ValidationError as e:
        abort(make_response(jsonify(code=400, message=str(e.messages), result=None), 400))


class ReportSchema(Schema):
    app_type = fields.Int(missing=0)
    app_version = fields.Str(required=True)
    model = fields.Str(missing="Unknow")
    os_type = fields.Int(required=True)
    os_version = fields.Str(required=True)
```

使用的时候，就只需要导入Schema和 check_data 函数，例如：

```python
from flask import request

from schemas import (
    check_data,
    ReportSchema,
)

@app.route("/")
def get_report():
    qs_dict = check_data(ReportSchema, request.args)
    pass

@app.route("/report", methods=["POST"])
def report():
    json_dict = check_data(ReportSchema, request.get_json())
    pass
```

这样如果参数不满足的话，就会自动返回400，并且将错误信息打印在返回的JSON里，而且不用手动return，通过abort函数，自动终止
流程。

如果参数满足的话，则会通过 `check_data` 函数返回，之后则只需要使用即可。
