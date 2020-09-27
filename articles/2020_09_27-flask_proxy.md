# Flask和requests做一个简单的请求代理

有的时候，我们需要做一些简单的代理工作，比如，把一个内部系统，通过已有的鉴权方式暴露出去。

代码如下：

```python
# 代理接口
import logging

import requests
from flask import Blueprint, request, Response

proxy_bp = Blueprint("proxy_bp", __name__, url_prefix="/proxy")


BASE_URL = "代理目标地址"


def get_token():
    return "已有系统的token获取"


@proxy_bp.route("/<path:url>", methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"])
def proxy(url):
    url = "{}/{}?{}".format(BASE_URL, url, request.query_string.decode("utf8"))
    method = request.method
    json_body = request.get_json()
    headers = {"Authorization": "Bearer {}".format(get_token())}

    resp = requests.request(method, url, json=json_body, headers=headers)
    logging.info("proxy got result: %s", resp.text)
    content_type = resp.headers.get("Content-Type", "text/html")

    return Response(resp.text, status=resp.status_code, content_type=content_type)
```

当然，这个只支持JSON，不过改成支持form也不难。
