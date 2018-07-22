# 我心目中的配置中心应该怎么做？

最近有使用携程开源的 [apollo](https://github.com/ctripcorp/apollo)。Golang的客户端的使用用法大概是
`agollo.GetStringValue("GRPCAddr", "")` 这样。

apollo的好处在于方便管控，热更新。但是实际开发过程中，更新配置非常麻烦，需要去配置中心更新配置，然后发布，然后等待
5-10s等配置热更新，但是通常来说，诸如数据库等配置是不会更新的，因为通常数据库连接池会在程序启动的时候初始化，之后
就不会再更新。此外，使用方式也不是很方便，这样程序中充斥着上面的代码，反而不如 `config.GRPCAddr` 这样来的简介明了。

所以我心目中的配置中心应当是：

    - 遵守 12 factors，所有配置从环境变量中读取
    - 配置写在配置文件中例如 `k8s/deploy.yml` 或者 `docker-compose.yml` 中，然后进行加密，对应工具拉起程序之后会把配置
    写入环境变量， 程序启动之后会从环境变量读取
    - 配置文件需要加密，重要的事情再强调一遍
    - 配置使用一个单例，将配置写入到单例的某个属性中

例如Python：

```python
from loka import LokaConfig


class Config(LokaConfig):
    MYSQL_URI = "mysql://root@127.0.0.1:3306"
    LISTEN_PORT = 8080


config = Config()

# below is just for test
print(config.MYSQL_URI, type(config.MYSQL_URI))
print(config.LISTEN_PORT, type(config.LISTEN_PORT))
```

例如Golang:

```go
main() {
	// Config is test demo
	type Config struct {
		Foo        string
		Bar        string
		Boolean    bool
		ReplicaNum int
	}
	c := Config{Foo: "hello"}

	log.Printf("before, c: %+v", c)
	LoadFromEnv(&c)
	log.Printf("after, c: %+v", c)
}
```

------

参考资料：

- https://12factor.net
- https://github.com/jiajunhuang/loka
