# PromQL 备忘

偶尔需要写PromQL，但是总是不记得，以下为备忘。

- http_requests_total 查询所有值
- http_requests_total{job="xxx"} 用大括号增加过滤条件，逗号 , 分割多个条件
    - = : 选择与提供的字符串完全相同的标签。
    - != : 选择与提供的字符串不相同的标签。
    - =~ : 选择正则表达式与提供的字符串（或子字符串）相匹配的标签。
    - !~ : 选择正则表达式与提供的字符串（或子字符串）不匹配的标签。

例如 `http_requests_total{environment=~"staging|testing|development",method!="GET"}`

- 用中括号选择时间，例如 `http_requests_total{job="prometheus"}[5m]`

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

- https://prometheus.fuckcloudnative.io/di-san-zhang-prometheus/di-4-jie-cha-xun/basics
