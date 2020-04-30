# MySQL操作笔记

MySQL是我最常用的关系型数据库，不过运维相关的一些命令，我不是很常用，但是偶尔又要用，每次都要去搜索太麻烦了，遂作笔记。

首先安装完mysql之后，执行 ``

## 把默认编码设置为 utf8mb4

更改 `/etc/mysql/my.cnf`，ubuntu的话，更改 `/etc/mysql/conf.d/mysql.cnf`：

```ini
[client]
default-character-set = utf8mb4

[mysqld]
collation_server = utf8mb4_unicode_ci
character_set_server = utf8mb4

[mysql]
default-character-set = utf8mb4
```

## 创建用户，授权，删除用户，删除授权

```bash
> CREATE USER 'monty'@'localhost' IDENTIFIED BY 'some_pass';
> GRANT ALL PRIVILEGES ON mydb.* TO 'monty'@'localhost';
> FLUSH PRIVILEGES;
> quit
```

`@` 后面接地址，可以是IP地址，也可以是 `%` 代表所有地址，也可以是 `localhost` 代表本地。

删除用户之后，授权会被一起删掉：

```bash
> DROP USER dbadmin@localhost;
```

查看授权：

```bash
> SHOW GRANTS FOR rfc@localhost;
```

如果只想删除授权的话：

```bash
> REVOKE INSERT, UPDATE ON classicmodels.* FROM rfc@localhost;
```

## 更改密码

```bash
> ALTER USER 'user-name'@'localhost' IDENTIFIED BY 'NEW_USER_PASSWORD';
> FLUSH PRIVILEGES;
```

## 备份

我本地的MySQL有一个cronjob每天进行备份：

```bash
@daily /usr/bin/mysqldump --single-transaction --quick --lock-tables=true --all-databases | gzip -c > /backup/mysql-$(date +\%F).sql.gz
```

---

参考资料：

- https://wiki.archlinux.org/index.php/MariaDB#Using_UTF8MB4
