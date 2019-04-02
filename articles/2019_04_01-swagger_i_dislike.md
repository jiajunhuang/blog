# Swagger? 不好用

最近用了一下swagger，swagger是一个用于把代码和文档连接起来的工具，我们在注释里写好一些东西，然后swagger生成出一个网页，
这样可以方便的在网页上点一下测试，就可以发出一个请求。但是，实际体验中，并不好用，原因如下：

- 文档虽然可以自动生成，但是书写相当繁琐
- 虽然可以在网页上自动生成一个用例用于测试，但是参数意义、是否可选等不明确

例如Python的Flask-Swagger使用如下：

```python
class UserAPI(MethodView):
    def post(self):
        """
        Create a new user
        ---
        tags:
          - users
        definitions:
          - schema:
              id: Group
              properties:
                name:
                 type: string
                 description: the group's name
        parameters:
          - in: body
            name: body
            schema:
              id: User
              required:
                - email
                - name
              properties:
                email:
                  type: string
                  description: email for user
                name:
                  type: string
                  description: name for user
                address:
                  description: address for user
                  schema:
                    id: Address
                    properties:
                      street:
                        type: string
                      state:
                        type: string
                      country:
                        type: string
                      postalcode:
                        type: string
                groups:
                  type: array
                  description: list of groups
                  items:
                    $ref: "#/definitions/Group"
        responses:
          201:
            description: User created
        """
        return {}
```

我更倾向于：

- 使用Postman发请求
- 单独使用Markdown写API

---

- https://pypi.org/project/flask-swagger/
