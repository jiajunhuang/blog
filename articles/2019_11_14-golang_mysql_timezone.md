# Go语言MySQL时区问题

最近由于我要统一技术栈，因此把原本使用SQLite做存储的数据全部迁移到MySQL。博客也是。不过当我检查数据库时，发现时间和
我产生数据的时间相差8小时。

首先检查机器的时间：

```bash
$ date
Thu 14 Nov 2019 11:13:59 AM CST
```

检查MySQL的时间：

```sql
> select now();
+---------------------+
| now()               |
+---------------------+
| 2019-11-14 11:14:42 |
+---------------------+
1 row in set (0.000 sec)
```

检查Go的时间：

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Printf("now: %s\n", time.Now())
}
```

执行：

```bash
$ go run main.go 
now: 2019-11-14 11:16:44.277801507 +0800 CST m=+0.000070270
```

检查数据时间：

```sql
> select * from issue order by id desc limit 1;
+-----+---------------------+---------------------+------------+------------+----------------------------------+
| id  | created_at          | updated_at          | deleted_at | content    | url                              |
+-----+---------------------+---------------------+------------+------------+----------------------------------+
| 285 | 2019-11-14 01:54:53 | 2019-11-14 01:54:57 | NULL       | htop详解   | https://peteris.rocks/blog/htop/ |
+-----+---------------------+---------------------+------------+------------+----------------------------------+
1 row in set (0.000 sec)
```

1点？这是不可能的。

---

综合上面的检查结果，我们知道：

- Go的程序输出的时间没有问题
- 系统的时间也没有问题
- 数据库的时间也没有问题
- 本地开发机器的时间也没有问题

那么问题会在哪里呢？我的经验告诉我，可能是数据库驱动的锅。检查一下配置：

```
SQLX_URL="user:abcdefg@(localhost:3306)/blog?parseTime=true"
```

发现我之前加了 `parseTime=true`，如果不加的话，就无法让驱动把MySQL的 `DATETIME` 类型和Go的 `time.Time` 互转。于是查了一下
文档，发现需要用loc来指定时区。我选择和机器一致，因此改成下面即可：

```
SQLX_URL="user:abcdefg@(localhost:3306)/blog?parseTime=true&loc=Local"
```

---

参考资料：

- https://github.com/go-sql-driver/mysql
