# Python开发实践经验

> 这是我在公司项目中写下的约定, 该项目使用Python3开发, 分享出来, 希望对大家也会有一些帮助

- 项目杜绝循环引用, 包之间的引用关系为(`->` 表示被引用)

    ```
    models -> controllers -> application
       ^            ^
        \          / \
           utils       service
    ```

- 包内相对引用,包外绝对引用

- `controllers`, `models`, `utils` 都是 `package` ,使用 `__all__` 来管控其下面的成员, 所有代码使用如下方式导入:

    `from itachi.models import Base` 而不是 `from itachi.models.base import Base`

    > 单元测试除外, 单元测试可能需要mock包内成员, 所以可能需要跳过 `__all__` 的限制

- 每个单元测试用例必须自己清理自己创建的数据

- 单元测试必须继承 `tests/base.py -> BaseCase`, 使用 `prehook` 和 `posthook` 替代 `setUp` 和 `tearDown`,
    因为base中处理的app的context顺序会影响单元测试(因为这个项目用的是flask,所以这一点比较蛋疼)

- 每次修改完数据库,必须显示commit `session.commit()`

- 禁止使用 `lazy import` 这种方式来规避循环引用,正确的方式是合理的规划代码组织, 参见第一条

- 推送使用接口来定义, 见 `itachi/services/push/base.py`, 由于python没有明确声明接口的方式, 所以还请人为遵守

- 异步任务task应当是可重入的, 会配置为重试

- 显示优于隐式, 所以不要用各种trick

- `requirements.txt` 和 `requirements-dev.txt` 分别对应正式和开发环境的依赖, 其中后者仅包含前者的增量部分

- 严禁for循环查数据库,请使用连表代替,连表时请把一次性能过滤最多的条件放在上面(即区分度最大的条件)

- 请求第三方接口一定要设置超时

- 使用Docker容器部署,必须设置CPU和内存上限,且不得与所在机器内存大小相同或非常接近.

## 目前发现的可以改善的地方

- [x] session可以和request脱离,手动 `session = Session()` 而非 `app.before_request`

- [ ] constants 可以使用 `Enum` 类
