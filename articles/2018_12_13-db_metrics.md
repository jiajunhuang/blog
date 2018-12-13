# MySQL性能指标

最近在魔改MySQL性能收集器 [mysqld-exporter](https://github.com/prometheus/mysqld_exporter)，接触到一些MySQL常见的性能指标，好好地记录下来学习学习：

`SHOW GLOBAL STATUS` 中：

- `Slow_queries` 是慢查询的数量。具体的慢查询，还需要开慢查询日志(https://dev.mysql.com/doc/refman/5.7/en/slow-query-log.html)：

    ```
    mysql root@192.168.175.132:(none)> show variables  like '%slow_query_log%';
    +---------------------+--------------------------------+
    | Variable_name       | Value                          |
    +---------------------+--------------------------------+
    | slow_query_log      | OFF                            |
    | slow_query_log_file | /var/lib/mysql/ubuntu-slow.log |
    +---------------------+--------------------------------+
    2 rows in set
    Time: 0.017s
    ```

- `Innodb_row_lock_current_waits` 是InnoDB当前被等待的行锁的数量
- `Threads_connected` 和 `Threads_created` 是当前打开的线程数和总共创建的线程数
- `Questions` 是总共执行的语句数
- `Connections` 是总共的连接数
- `Com_select`, `Com_insert`, `Com_update`, `Com_delete`, `Com_replace` 则是分别对应 `SELECT`, `INSERT`, `UPDATE`, `DELETE`, `REPLACE` 语句的数量
- `Qcache_hits` 是缓存命中量
- `Select_full_join` 是全表连接的数量

`SHOW VARIABLES\G` 中：

- `max_connections` 是最大连接数

`SHOW SLAVE STATUS\G`中：

- `Seconds_Behind_Master` 是主从之间的延时

----

当然了，这些指标其实文档上全都有，但是很久不读文档，或者没有DBA那么熟悉文档的话，这样记录一下就还是有用的。
