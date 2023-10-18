# PromQL 备忘

偶尔需要写PromQL，但是总是不记得，以下为备忘。

表达式语言数据类型包括以下四种：

- 瞬时向量 Instant vector:一组时间序列，每个时间序列包含单个样本，它们共享相同的时间戳。也就是说，表达式的返回值中只会包含该时间序列中的最新的一个样本值。而相应的这样的表达式称之为瞬时向量表达式。
- 区间向量 Range vector:- 一组时间序列，每个时间序列包含一段时间范围内的样本数据。
- 标量 Scalar:一个浮点型的数据值。
- 字符串 String:  - 一个简单的字符串值。

下面来看看查询语句：

- http_requests_total 查询所有值
- http_requests_total{job="xxx"} 用大括号增加过滤条件，逗号 , 分割多个条件
    - = : 选择与提供的字符串完全相同的标签。
    - != : 选择与提供的字符串不相同的标签。
    - =~ : 选择正则表达式与提供的字符串（或子字符串）相匹配的标签。
    - !~ : 选择正则表达式与提供的字符串（或子字符串）不匹配的标签。

例如 `http_requests_total{environment=~"staging|testing|development",method!="GET"}`

- 用中括号选择时间，例如 `http_requests_total{job="prometheus"}[5m]`，单位有：
    - ms - milliseconds
    - s - seconds
    - m - minutes
    - h - hours
    - d - days - assuming a day has always 24h
    - w - weeks - assuming a week has always 7d
    - y - years - assuming a year has always 365d

- `rate()` 函数用来计算变化率
- `increase()` 用来计算增长量
- `sum()` 用来聚合

`sum(rate(xxx)) / sum(rate(xxx)) = sum(increase()) / sum(increase)`

- `histogram_quantile()` 用来计算直方图里，百分位中的最大值

例如 `histogram_quantile(0.9, rate(http_requests_total[10m]))` 计算 过去10分钟内，http_requests_total 90% 分位的最大值。

- `$__interval` 是Grafana中选中的时间范围
- `$__rate_interval` 是选中时间范围，但是解决了 `$__interval` 的一些小问题，比如 `$__interval` 选择了15s，刚好Prometheus抓取间隔是15s，那么就会显示不出数据，而 `$__rate_interval` 保证至少会选择4个数据展示出来。

---

ref:

- https://prometheus.io/docs/prometheus/latest/querying/basics/
