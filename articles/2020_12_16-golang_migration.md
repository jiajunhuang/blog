# Golang里，数据库migration方案

Python中如果用SQLAlchemy的话，一般会用alembic来做migration。那么，Go呢？我一般用这个：

https://github.com/golang-migrate/migrate

官方的用法是生成migration文件（其实就是sql）时用命令行，升级时也可以用命令行。不过我更喜欢把升级文件和代码一起打包到二进制文件里。

我们依次来看。

## 创建migration文件

这个很简单，首先你要在自己电脑上装上 `migrate` 这个二进制：

```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/$version/migrate.$platform-amd64.tar.gz | tar xvz
```

我是放到了 `/usr/local/bin` 下，其实也可以放到 `$PATH` 的任一自定义目录里。

生成migration文件，则执行以下命令即可：

```bash
$ migrate create -ext sql -dir ./migrations -seq create_user_table
```

执行之后，就会在 `./migrations` 文件夹下，创建两个文件，都会包含 `-seq` 后的那一段，以 `up.sql` 或 `down.sql` 结尾。分别对应升级和降级操作该要执行的SQL。

migration操作，就写在对应的文件里。

## 升级

除了继续用 `migrate` 这个命令来进行升级降级操作，我们还可以把操作结合到程序里，我更喜欢这样做，好处如下：

- 不用每个地方都打包 `migrate` 这个命令，或者安装这个二进制文件
- 不用把migrations文件夹到处同步

咋做呢？这个时候就该祭出 `go-bindata` 这个包了，它的原理就是读文件，把文件的内容生成到一个 `bindata.go` 文件里，然后就可以通过内置的api，读取其中的内容。

那怎么结合到我们的应用里呢？直接上代码：

```go
package main

import (
    "flag"
)

var migrateUp = flag.Bool("migrateUp", false, "run migration")

func main() {
    // ...
	if *migrateUp {
		s := bindata.Resource(migrations.AssetNames(),
			func(name string) ([]byte, error) {
				return migrations.Asset(name)
			})

		d, err := bindata.WithInstance(s)
		if err != nil {
			logrus.Panicf("failed to get migrations: %s", err)
		}
		m, err := migrate.NewWithSourceInstance("go-bindata", d, config.MigrateDBURL)
		if err != nil {
			logrus.Panicf("failed to get migrations: %s", err)
		}
		err = m.Up()
		if err != nil {
			logrus.Panicf("failed to migrate: %s", err)
		}
		return
	}
    // ...
}
```

就是介样，我们就可以愉快的把migration也编译到应用里，然后继续愉快的一个二进制包到处丢了。

当然了，要想使用 `go-bindata`，也得安装：

```bash
$ go get -u github.com/go-bindata/go-bindata/...
```

然后我在Makefile里加上这么一个动作：

```bash
bindata:
	cd migrations && go-bindata -pkg migrations .
```

duang，大功告成。
