# agentzh的Nginx教程阅读笔记

## Nginx变量

- Nginx的配置有点像编程语言，不过不图灵完备
- Nginx的配置有点像bash，比如:

```nginx
server {
    listen 8080;

    location /test {
        set $foo hello;
        echo "foo: $foo";
    }

    location /test2 {
        set $first "hello ";
        echo "${first}world";
    }
}
```

可以进行变量设置，以及字符串拼接。

- Nginx中变量的作用域是全局(全部配置文件可见)，但是只有在有 `set` 的那一节才会得到变量值，其他地方是空值。

```nginx
server {
    listen 8080;

    location /foo {
        echo "foo = [$foo]";
    }

    location /bar {
        set $foo 32;
        echo "foo = [$foo]";
    }
}
```

访问：

```bash
$ curl 'http://localhost:8080/foo'
foo = []

$ curl 'http://localhost:8080/bar'
foo = [32]

$ curl 'http://localhost:8080/foo'
foo = []
```

- Nginx中变量的生命周期是跟着整个请求的，例如：

```nginx
server {
    listen 8080;

    location /foo {
        set $a hello;
        echo_exec /bar;
    }

    location /bar {
        echo "a = [$a]";
    }
}
```

请求 `/foo` 会输出 "a = [hello]"，但是请求 `/bar` 只会输出 "a = []"

- Nginx中的变量名，除了用户 `set` 定义的之外，还有很多内置的。例如 `$uri`。此外还有变量名组，例如 `$arg_XXX`，其中 "XXX"
代表了query string中的具体的key(Nginx对query string中的值的匹配不区分大小写)。

- Nginx的请求分为主请求和子请求。前者是HTTP请求，后者是主请求在Nginx内部流转，不涉及网络的一种“抽象”。

## Nginx配置执行顺序

- Nginx中的指令是根据请求处理的阶段来处理的，例如：

```nginx
location / {
    set $a 32;
    echo $a;

    set $a 56;
    echo $a;
}
```

的实际处理顺序是先执行两个 `set`，然后执行两个 `echo`。原因是set在 `rewrite` 阶段，而echo在 `content` 阶段。`rewrite`
阶段是发生在 `content` 阶段之前的。

- 指令执行顺序只跟请求处理阶段的顺序有关。不一定和书写顺序有关。

- Nginx处理请求的11个阶段，按照执行顺序依次是：

    - post-read:在Nginx读取并且处理完请求头(request headers)之后就开始运行
    - server-rewrite:写在server块的配置基本都在这个阶段执行
    - find-config:由Nginx核心来完成当前请求与location配置块之间的配对工作
    - rewrite
    - post-rewrite: Nginx完成内部子请求的阶段(如果rewrite阶段有此要求的话)
    - preaccess
    - access
    - post-access
    - try-files
    - content
    - log
