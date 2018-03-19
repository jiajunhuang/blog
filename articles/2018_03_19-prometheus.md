# Prometheus 数据类型

> 写博客记一下，老是忘记

- Counter: 适用于只增不减的数值，例如请求数，错误数

- Guage: 适用于可增可减的数值，例如内存量，goroutine的数量

- Histogram: 适用于样本统计，例如响应时间，响应大小

- Summary: 和Histogram类似，不过同时还提供总数和和
