使用Git Hooks
===============

最近的工作就是，写文档！用markdown来写，然后要遵循一定的格式。
之前都是我写好我负责的一部分，然后打包发送给项目负责人，因为
一开始以为没多少改动，写完就可以了。没想到，要改的还挺多。
今天就把流程自动化了一下。

整体工作流程为：

    ``编辑文档`` -> ``git push`` -> ``服务端通过hoos执行脚本，自动部署`` -> ``浏览器刷新看效果``

好，现在来看一下GitHooks文档 [#]_ 。我们需要在客户端push之后触发，所以在
服务器端的项目里设置hook。我最开始犯了一个错，就是以为是在客户端git目录一的
``.git/hooks`` 下设置。

.. code:: bash

    # cat post-receive
    #!/bin/bash

    LOG=/var/log/api-generator.log
    unset GIT_DIR

    echo "lastest generating " `grep '^time' $LOG | tail -n 1`
    git -C /srv/api_docs_generator/smartx/input_api_docs pull >> $LOG 2>&1
    cd /srv/api_docs_generator/slate
    bundle exec middleman build --clean >> $LOG 2>&1 &
    echo 'time:' `date` >> $LOG
    echo "it's generating html, please visit http://192.168.49.22/docs/dev/"

上面这个hook解决了两个问题：

- ``git -C`` 后面可以指定git目录，一开始我用的是 ``cd xxx && git pull`` 完全没效果，
  后来改成 ``git -C`` 也没效果，原因是，githook调用的时候会设置变量 ``GIT_DIR`` 。
  ``unset GIT_DIR`` 之后直接用 ``git -C`` 的方式可以一行搞定。

- ``bundle exec middleman build --clean >> $LOG 2>&1 &`` 让bash进程独立执行，从
  而不会阻塞客户端。

我们来看一下客户端输出吧：

.. code:: bash

    ➜  api-docs git:(master) git push
    root@192.168.49.22's password: 
    Counting objects: 3, done.
    Delta compression using up to 2 threads.
    Compressing objects: 100% (3/3), done.
    Writing objects: 100% (3/3), 365 bytes | 0 bytes/s, done.
    Total 3 (delta 2), reused 0 (delta 0)
    remote: lastest generating  time: Tue Nov 8 16:42:34 CST 2016
    remote: it's generating html, please visit http://192.168.49.22/docs/dev/
    To root@192.168.49.22:/srv/api-docs
    236dexx..10xxxx9  master -> master

额，其实没什么技术含量可言。顺便吐槽一句，在开发的时候，chrome的缓存真的很讨人厌。

.. [#] https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks
