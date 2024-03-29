# 单测时要不要 mock 数据库？

最近我们讨论了一个问题，要不要 mock 数据库？ 本文是我的一些个人观点。

首先，对于第三方调用进行 mock ，这是基本都能获得一致意见的，但是是否需要 mock 数据库，则各有分歧。我是赞成不 mock 数据库
的，原因如下：

- 对于后端业务系统来说，业务代码是围绕着数据库做增删改查从而支撑业务的，换句话说，数据库才是核心，如果 model 层没有得到
测试的覆盖，我认为是不可靠的
- 不把测试放到数据库跑一遍，可能会遇到一些数据库相关的错误却检查不出来，例如最简单的：SQL拼写错误
- 对数据库进行 mock 操作需要多写很多mock代码，一般都不会有这么多时间来做这个事情
- 简单粗暴，方便快捷

那么应该如何测试呢？我一般都是在跑CI的时候，起一个全新的数据库，然后跑数据库 migration，在每一个单元测试 `setUp` 时
组装数据，在 `tearDown` 时删除数据。如果是 Go 的单元测试，那么也是类似的，在 `Test` 函数开始组装数据，`defer` 删除数据。基本的流程是：

- 开始单元测试
- 起一个数据库，用于单元测试
- 构造单元测试用例所需要的数据，并且插入数据库
- 执行业务逻辑
- 检查返回结果以及相关调用是否符合预期
- 各种断言检查
- 销毁先前插入的数据，并且删除本次所产生的数据
- 销毁单元测试所用的数据库

当然，这种模式也有它的弊端：

- 如果单元测试没有销毁构造的数据以及产生的数据，有可能会对其它测试产生影响
- 由于需要连接数据库处理，速度上会稍微慢一点

对于第一点，我是真实经历过的，这一点只能靠好好的写测试，清理该清理的数据来做到；第二点，我认为问题不大，也没有那么多
测试代码来跑，数据库能扛住线上的请求，却扛不住单元测试（我知道我知道，配置虽然不太一样）？

而对于 mock 数据库连接这种方式，需要付出很大代价，却无法获得超额的收益，即收益抵不过付出。

这就是我的看法了，你呢？
