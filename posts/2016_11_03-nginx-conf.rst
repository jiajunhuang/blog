nginx配置笔记
=================

nginx变量
-----------

nginx配置不常写，看了又忘，真是尴尬。那就再看一次，然后把不会的记录一下，下次
就只要看这篇笔记，就能快速想起来了。

``nginx`` 的变量名前面有一个 ``$`` ，与bash一致，例如：

.. code:: bash

    set $a "hello";
    set $b "$a, $a";

那么 ``$a`` 是 "hello"，而 ``$b`` 是 "hello, hello"

nginx变量的作用范围是整个全局的。但是每个请求都有自己的变量副本：

.. code:: nginx

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

.. code:: bash

    $ http localhost:8080/foo
    foo = []

    $ http localhost:8080/bar
    foo = [32]

请求bar虽然会给 ``foo`` 赋值，但并不影响另外一次请求里的 ``foo``

``set_unescape_uri`` 指令能够将uri中类似 ``%20`` 解码出来，例如：

.. code:: nginx

    location /echo {
        set_unescape_uri $name $arg_name;
        echo "uri = $uri";
        echo "request_uri = $request_uri";
        echo "name = $name";
    }

.. code:: bash

    $ http 'localhost:8080/echo?name=hi lo'
    uri = /echo
    request_uri = /echo?name=hi%20lo
    name = hi lo

其中 ``arg_xxx`` 就会对应uri中的某个key，所以会有无穷无尽个这种参数啦。

nginx中map的意思是映射，举个例子：

.. code:: nginx

    map $args $foo {
        default     0;
        debug       1;
    }

当 ``$args`` 匹配 debug 时， ``$foo`` 被设置成 1，否则为 0。

.. code:: nginx

        map $args $foo {
            default 0;
            debug 1;
        }

        server {
            listen 8080;
            location / {
                default_type text/html;
                content_by_lua '
                    ngx.say("<p>hello, world</p>")
                    ';
            }

            location /map {
                echo "foo = $foo";
            }
        }

.. code:: bash

    $ curl 'localhost:8080/map'
    foo = 0
    $ curl 'localhost:8080/map?debug'
    foo = 1

nginx中的子请求，实际上只调用了一些函数，而并没有另外新建socket连接等，例如：

.. code:: nginx

    location /main {
        echo_location /foo;
        echo_location /bar;
    }

    location /foo {
        echo foo;
    }

    location /bar {
        echo bar;
    }

.. code:: bash

    $ curl 'http://localhost:8080/main'
    foo
    bar

nginx变量在父请求和子请求之间是不相互影响的。但是有些nginx模块不遵循此规则。

nginx配置
---------

想要知道某条指令的将会在nginx的11个请求处理阶段的哪个阶段进行，
需要自行 翻阅对照文档和源码。

nginx 常用配置文件
------------------

-  静态文件web服务器

.. code:: nginx

    location /static/ {
        default_type text/html;
        root /var/;
        autoindex on;
    }

 .. [#] https://openresty.org/download/agentzh-nginx-tutorials-zhcn.html

 .. [#] http://www.nginxguts.com/2011/01/phases/
