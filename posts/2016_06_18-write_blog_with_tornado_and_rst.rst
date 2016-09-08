使用Tornado和rst来写博客
==========================

为什么用rst
-------------

rst，全称 ``reStructuredText`` ，也是一种标记文档。在Markdown异常流行的今天，
为什么还要选择rst呢？Markdown的标准太多了，而rst只有这一种，sphinx等自定义配置
的当然不算啦。

效果
------

.. image:: ./img/screenshot0.png
.. image:: ./img/screenshot1.png

开始配置
-----------

首先我们需要把项目从Github拖下来::

    $ git clone https://github.com/jiajunhuang/blog

我们来看一下项目的结构::

    $ tree -d -L 1
    .
    ├── controllers
    ├── posts
    ├── static
    ├── templates
    └── utils

    5 directories

其中 ``posts`` 就是我们存放rst的文件夹啦。其中所有的想列在首页的文章，必须符合
``config.py`` 中定义的 ``filename_format = r"(\d{4}_\d{2}_\d{2})-.+\..+"`` 正则，
如果匹配不到，将忽略。当然，用户可以更改文件命名规范，但注意，目前的设置是
正则表达式需要用一个组把日期匹配出来，也仅仅可以匹配日期。

开始写作
------------

上面把项目托下来以后，接下来删除本人的一些信息然后替换自己的信息，之后进行推送和
部署:

.. code:: bash

    $ cd blog
    $ rm -rf posts && mkdir posts  # 删除我的博客内容
    $ rm static/favicon.ico static/img/avatar.png  # 删除我的头像
    $ rm -rf .git/  # 删除我的git提交记录
    $ git init

之后的操作就是写博客，然后其他的配置在 ``config.py`` 中都有注释说明。
阅读对应的条目，重新部署就可以生效啦～

部署
------

我们建议用一个Nginx挡在Tornado的前面，下面是我的Nginx配置的server部分::

    server {
        listen       80;
        server_name  _;

        location / {
            proxy_pass http://localhost:8080;
        }

        #error_page  404              /404.html;

        # redirect server error pages to the static page /50x.html
        #
        error_page   500 502 503 504  /50x.html;
        location = /50x.html {
            root   /usr/share/nginx/html;
        }
    }

在vps上只开了一个进程，并且该进程监听在8080端口上，如果有多个，则需要使用nginx的
upstream_ 模块。

.. _upstream: http://nginx.org/en/docs/http/ngx_http_upstream_module.html#upstream

原理
-----

原理其实很简单，就是在启动的时候，Tornado读取 ``config.py`` 中指定的目录下的所有文件，
并且解析文件名，然后生成一个文件目录的列表，并且缓存（因为Config类是单例）。

另外配置了Tornado的autoreload项为True，并且添加了对posts目录的监听，所以只要该目录
有改动，进程就会自动重启。

本项目还添加了对github webhooks的支持，当在github设置了git push的hook时，本地推送
到github以后，github就会对配置的vps发起POST请求，然后项目会从github拉取代码并且重启
进程，这样就不需要手动去pull代码了。详见 `这里 <2016_06_19-use-github-webhooks.rst>`__ 。
