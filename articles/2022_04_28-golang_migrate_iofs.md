# Golang migrate 做数据库变更管理

最近在使用 [golang-migrate](https://github.com/golang-migrate/migrate) 做数据库变更管理，按照官方的教程，需要本地
先下载一个二进制，命令行生成变更文件：

```bash
$ migrate create -ext sql -dir db/migrations -seq create_article_table
...项目路径/db/migrations/000011_create_article_table.up.sql
...项目路径/db/migrations/000011_create_article_table.down.sql
```

这样，就会在 `./db/migrations` 下有一堆的 `.sql` 文件，其中 `000011` 是编号，从 `000001` 开始递增，
`create_article_table` 是传入的文件名，`.up.sql` 代表向前变更时执行的SQL，`.down.sql` 代表回滚时执行的SQL。

我们在编辑完之后，可以这样在本地变更数据库：

```bash
$ export POSTGRESQL_URL='postgres://postgres:密码@localhost:5432/dbname?sslmode=disable'
$ migrate -database ${POSTGRESQL_URL} -path db/migrations up
no change
```

但是在实际使用中，我们在发布项目之后，如果每次都还要去线上变更一下，那就比较麻烦了，能不能让代码自己来运行呢？
当然可以。

## 代码自动执行变更

migrate 既可以使用命令行，又可以以库的方式调用。migrate中，migration文件存放的地方，叫做 `source`，migrate支持多种
source例如 `iofs`, `github`, `gitlab`, `s3` 等等。

我想使用的就是官方的 `io/fs`：

```go
package main

import (
    "context"
    "embed"
    "time"

    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    "github.com/sirupsen/logrus"
)

var (
    //go:embed db/migrations/*.sql
    fs embed.FS
)

func initDB(config *Config) {
    var err error
    uri := fmt.Sprintf(
        "postgres://%s:%s@%s:%s/%s?sslmode=disable",
        config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName,
    )
    ctx := context.Background()
    db, err = pgxpool.Connect(ctx, uri)
    if err != nil {
        logrus.Fatalf("connect to db failed: %v", err)
    }

    if err = db.Ping(ctx); err != nil {
        logrus.Fatalf("could not connect to database: %v", err)
    }

    d, err := iofs.New(fs, "db/migrations")
    if err != nil {
        logrus.Fatalf("could not open migrations: %v", err)
    }
    m, err := migrate.NewWithSourceInstance("iofs", d, uri)
    if err != nil {
        logrus.Fatalf("could not init migrate: %v", err)
    }
    err = m.Up()
    if err != nil {
        logrus.Errorf("migrate up error: %v", err)
    }
}

func main() {
    config := GetConfig()
    initDB(config)
}
```

也就是在应用启动以后，立刻初始化数据库，连接到数据库之后立刻开始执行数据库变更，看最上面的 `var fs embed.FS`
以及上一行的注释，这是Go官方提供的将文件打包到二进制的方式，本质是编译时将文件内容读取，编译后，赋值到 `fs` 变量里。

这样就可以在每次启动之后，自动先做数据库变更然后才开始执行代码了。但是有一点值得注意，跑数据库变更的程序，最好只部署
一份，否则容易出现竞争问题。

## 总结

这篇文章记录了我使用migrate的方式，这种方式比所有操作都在命令行执行更加方便，也不再需要将sql文件同步到服务器，当然
缺点就是二进制文件会变的更大一些。有利有弊，不过我更倾向于这种方式。
