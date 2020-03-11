# 一个Gunicorn worker数量引发的血案

最近大佬想要我重写一个应用，为嘛呢？因为发现这个应用内存占用非常高，每个pod(我们部署在k8s里)占用1.2-1.3G，一共放了6个pod。
但是按道理来说，这个应用并不复杂，不应该占用如此多的内存。

为啥要重写呢？我很好奇，最终说服大佬，先让我尝试去分析问题，找到root cause，然后再决定是否重写。这个应用最开始是
在虚拟机上部署的，因此为了充分利用核心数，我们参照 gunicorn 的官方文档，把worker数量设置为 `2 * CPU 核心数 + 1`：

```python
workers = multiprocessing.cpu_count() * 2 + 1
```

> Gunicorn relies on the operating system to provide all of the load balancing when handling requests.
> Generally we recommend (2 x $num_cores) + 1 as the number of workers to start off with. While not overly scientific,
> the formula is based on the assumption that for a given core, one worker will be reading or writing from the
> socket while the other worker is processing a request.

然而大家都知道，Python是很吃内存的，起一个Python解释器，就要 100M+ 的内存，而Gunicorn采取pre-fork的模式，设置多少个
worker就会fork出多少个进程来处理任务，所以这个配置在迁移到 k8s 之后，出现一个大问题：内存占用非常高。

因为在k8s里，我们实际并没有分配那么多核心数，而是例如 `1000m`, `2000m` 等值，但是上面的代码算出来确是 k8s 节点的CPU
数量。

因此就陷入这么一个性能瓶颈：根据cpu数量出来的worker数很高；实际分配的cpu比较少；当并发来临，进程上下文切换非常消耗性能；
多个进程占用了很多内存。

> Always remember, there is such a thing as too many workers. After a point your worker processes will start
> thrashing system resources decreasing the throughput of the entire system.

最后，把worker数量调低即可解决这个问题，比如如果分配的是 `2000m`，那么worker数量设置为4或者5是比较好的选择。

---

参考资料：

- [Gunicorn官方文档](https://docs.gunicorn.org/en/latest/design.html)
