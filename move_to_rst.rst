:Date: 12/30/2015

RST大迁移
=========

About
-----

rst全称reStructuredText，`这里 <http://docutils.sourceforge.net/docs/user/rst/quickref.html>`__ 是一份简单教程。

今天花了一下午，把博客内容全部转移到了rst， 内容转换是使用神器``pandoc``，下面
是转换用的脚本

.. code:: bash

    for i in *.md; do
        pandoc -f markdown_github -t rst -o ${i%%.*}.rst $i;
    done

重新组织了一下博客，命名基本为::

    notes_on_xxxx: 一般是读书笔记，比如读python标准库，APUE等，
    后面紧接书名等，然后是具体位置，然后选填内容主题。

    xxxx: 这是根据自己心得写出来的文章，而不是读书笔记。

Why
---

为什么迁移到rst呢？目前好像是markdown才大行其道，不过markdown标准太多了，反而是
rst标准比较单一。不过从原来的博客转到纯文本(虽然原来也是静态页面，但好歹有Disqus)
等，丢失了评论和特效，以及漂亮的css样式。

不过基于统计数据标明，反正我的博客访问人数极少，所以丢掉这些功能影响不会很大。

Other thing
------------

gh-pages跳转到blog的repo代码:

.. code:: html
    <!DOCTYPE html>
    <html>
        <body>
            <script type="text/javascript">
                // Javascript URL redirection
                window.location.replace("https://github.com/jiajunhuang/blog");
            </script>
        </body>
    </html>

Trash
-----

删除掉了一些没什么意义的blog，比如ubuntu升级内核，切换到archlinux，docker资料收集等。
但毕竟是用git管理，如果有兴趣的话，还是可以在`这里 <https://github.com/jiajunhuang/blog/tree/874f3c336897623a07e82fa4eea99b038ab60e33/_posts>`__ 看到。
