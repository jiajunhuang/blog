# coding: utf-8

import os

from utils.singleton import Singleton

# 这里是首页的大部分内容，例如箴言，等等。将在启动项目的时候导入，并且
# 缓存在内存中。


class Config(metaclass=Singleton):
    # 顶栏所需要的信息
    top_part = dict(
        navbar=[
            ("首页", "/"),
            ("Github", "https://github.com/jiajunhuang"),
            ("关于我", "/aboutme.rst.html")
        ],  # navbar这一栏的内容，将按照列表的顺序生成
        index_title="Jiajun's Blog",  # 网站首页的标题，以及顶部的标题
        subtitle="你的眼睛能看多远",  # 网站顶部的标题下面的话
        avatar_img="static/img/avatar.png",  # 网站顶部的头像
        announcement="会当凌绝顶，一览众山小。",  # 网站首页旁边的公告栏
        disqus_site_name="gansteedeblog",  # disqus site name
        github="https://github.com/jiajunhuang",  # footer
        username="jiajunhuang",  # footer
    )

    template_path = os.path.join(os.path.dirname(__file__), "templates")  # 模板的路径
    static_path = os.path.join(os.path.dirname(__file__), "static")  # 静态文件的路径
    posts_path = os.path.join(os.path.dirname(__file__), "posts")  # posts文件夹的路径
    article_img_path = os.path.join(posts_path, "img")  # 纯文本文件中 .. image:: img.png 的存储路径
    text_type = "rst"  # 设置写作所用标记文本类型
    filename_format = r"(\d{4}_\d{2}_\d{2})-.+\..+"  # 文件名的正则表达式，默认为 年_月_日-标题.后缀 可以更改日期等的规则，但捕获组只能有一个而且是日期。
    github_webhook_secret_path = os.path.join(os.path.dirname(__file__), "secret.txt")
    repo_path = os.path.dirname(__file__)

    def article_path(self, filename):  # 获取某篇文章的具体路径
        return os.path.join(self.posts_path, filename)

    @staticmethod
    def article_url(filename):  # 生成文章url所用的函数
        return os.path.join("./article", '.'.join([filename, 'html']))

    @property
    def github_secret_key(self):
        with open(self.github_webhook_secret_path) as f:
            return f.readline()[:-1]
