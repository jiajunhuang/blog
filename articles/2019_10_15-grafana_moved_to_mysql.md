# 将SQLite的数据迁移到MySQL

这几天将Grafana的数据库 `/var/lib/grafana/grafana.db` 迁移到了 MySQL，原因是不想维护多个数据库备份，全都丢MySQL里，统一
管理，统一备份即可。

首先将 `grafana` 服务暂停一下，省得写新数据进去：

```bash
$ sudo systemctl stop grafana
```

然后执行命令把SQLite的数据导出来：

```bash
#!/bin/bash
DB=$1
TABLES=$(sqlite3 $DB .tables | sed -r 's/(\S+)\s+(\S)/\1\n\2/g' | grep -v migration_log)
for t in $TABLES; do
    echo "TRUNCATE TABLE $t;"
done
for t in $TABLES; do
    echo -e ".mode insert $t\nselect * from $t;"
done | sqlite3 $DB
```

```bash
$ ./export_sqlite.sh grafana.db > grafana.sql
```

接着改 grafana 的配置文件 `/etc/grafana.ini` 把MySQL中创建好的数据库、用户名、密码写进去：

> 如果没有创建好，那么应当先创建一个。

```ini
[database]
# You can configure the database connection by specifying type, host, name, user and password
# as separate properties or as on string using the url properties.

type = mysql
host = 127.0.0.1:3306
name = grafana
user = grafanauser
password = grafana123
url = mysql://grafanauser:grafana123@127.0.0.1:3306/grafana
```

接着启动一下grafana，让它在MySQL中创建好表： `sudo systemctl start grafana`，使用 `journalctl -u grafana.service -f`
看着日志，当日志显示数据库migration已经做完了之后，就可以再次停用grafana，然后把数据导进去：

```bash
$ mysql -u grafanauser -p -D grafana < grafana.sql
```

最后，启动grafana，大功告成！

---

参考资料：

- [https://community.grafana.com/t/migrating-grafana-data-to-new-database/2454](https://community.grafana.com/t/migrating-grafana-data-to-new-database/2454)
