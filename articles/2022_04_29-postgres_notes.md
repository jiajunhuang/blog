# PostgreSQL 操作笔记

之前我写了一篇 [MySQL操作笔记](./2020_05_01-mysql_notes.md.html) 用于记录我常用的MySQL操作，事实证明还是
很有用处的，我经常需要翻到。由于今年我的目标之一是用熟PG，我也开始使用PG，但是很多时候不记得，还需要临时去查，因此
再来记录一篇PG的操作笔记，方便以后查阅。

## 创建用户并授权

```bash
$ sudo -u postgres psql
$ psql
> CREATE USER 用户名 WITH ENCRYPTED PASSWORD '密码';
> CREATE DATABASE 数据库名 OWNER 用户名;
```

## 更改用户密码

```bash
$ psql
> ALTER USER 用户名 WITH PASSWORD '新密码';
```

## 允许远程访问

可能创建完数据库之后，不仅仅要在本地能访问，还希望提供服务让其它机器远程访问，编辑 `/etc/postgresql/13/main/pg_hba.conf`，
当然如果你的版本不是13，那么路径里的版本号就要对应替换，在最后添加一行：

```bash
host all all 0.0.0.0/0 md5
```

此外还需要更改配置文件 `/etc/postgresql/13/main/postgresql.conf`，将 `#listen_addresses = 'localhost'` 取消注释，改为：

```bash
listen_addresses = '*'
```

## 设置 postgres 用户密码

连接上去之后，执行 `\password` 命令。

## 备份

我在 `postgres` 用户下，加了一个crontab每天全量备份一次：

```bash
30 3 * * * /usr/bin/pg_dumpall | gzip -c > /data/backup/postgres/full-backup-$(date +\%F).sql.gz
```

## 常见命令

```bash
$ psql
> \l 列出数据库
> \c dbname 切换数据库
> \d 列出当前数据库所有表
> \d tablename 列出当前数据库中tablename表的表结构
> \du 列出所有用户
```
