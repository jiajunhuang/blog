我不理解asyncio
==================

最近通篇读了两遍asyncio的源代码和文档。给我的感受是：这玩意儿就像是拿几条
蟒蛇打成了死结。

从设计角度上来看，asyncio作为标准库，想要抽象UNIX的select异步模型和Windows的
IOCP，从这点上来说，还可以理解。但是整个asyncio为我们带来了以下概念：

- Event Loop

- Event Loop Policy

- ``asyncio.coroutine`` ``yield from``

- Future

- Task

- Handle

- ThreadPoolExecutor/ProcessPoolExecutor

- ``__await__`` ``__aenter___`` ``__aexit__``

看完以后表示脑子都涨了，画了我好几页草稿才基本理清楚这些东西之间的关系。
但是asyncio里 transport 和 protocol 的抽象层面是这一次读代码比较眼前一亮的。
但是这都是从 ``Twisted`` 里抄来的吧。。。论代码，还是 ``Tornado`` 漂亮一些，
只可惜为了兼容python2和python3中yield的语义， ``Tornado`` 的代码也不是那么
干净整洁。当然了， ``Tornado`` 和 ``asyncio`` 的抽象层次不在同一个级别，也
不能完全拿来做比较。

这篇文章是吐槽专用的，来，我们再来一篇，讲讲python中借助协程，用同步的方式
写异步。
