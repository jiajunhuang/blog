# coding: utf-8

"""
- 从articles目录读取文件，生成目录，并将目录缓存在内存中
- 当请求文件时，提取出文件名并且尝试读取文件，生成html
- 当推送代码到Github时，由Github发出Webhook请求，响应请求并且拉取最新代码
  更新内存中缓存的目录，重启当前tornado进程

TODO:
    增加redis缓存支持，使博客响应更快
"""

import glob
import hashlib
import hmac
import logging
import operator
import os
import subprocess
import sys

from docutils.core import publish_parts

import tornado.autoreload
import tornado.ioloop
import tornado.web

from tornado.options import (
    define,
    options,
    parse_command_line,
)
define("debug", default=False, type=bool, help="debug is set to True if this option is set")
define("port", default=8080, type=int, help="port=8080")
# define("redis", default=False, type=bool, help="use redis as cache system")
parse_command_line()

# constants
PROJ_PATH = os.path.dirname(__file__)

MAIN_FILE_PATH = os.path.join(PROJ_PATH, __file__)

ARTICLE_PATH = os.path.join(PROJ_PATH, "articles")
ARTICLE_IMG_PATH = os.path.join(ARTICLE_PATH, "img")

TPL_PATH = os.path.join(PROJ_PATH, "templates")
STATIC_PATH = os.path.join(PROJ_PATH, "static")
SECRET_TXT_PATH = os.path.join(PROJ_PATH, "secret.txt")

HEADER = dict(
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


# utils
def gen_catalog():
    catalog = []
    for fullpath in glob.glob("%s/[0-9]*.rst" % ARTICLE_PATH):
        with open(fullpath) as f:
            title = f.readline()
        _, filename = fullpath.split("/")
        date, _ = filename.split("-")
        date = date.replace("_", "-")

        catalog.append((title, date, filename))

    return sorted(catalog, key=operator.itemgetter(1), reverse=True)


# handlers
class GithubWebHooksHandler(tornado.web.RequestHandler):
    def get(self):  # github webhooks ping, http status 200 is Okay
        self.finish()

    def post(self):
        if not self.__validate_signature(self, self.request.body):
            logging.error("github signature not match")
            raise tornado.web.HTTPError(400, "the given signature is invalid")

        # run git pull, we do not use GitPython anymore.
        subprocess.Popen(
            "git -C %s pull" % PROJ_PATH,
            shell=True
        )
        self.application.CATALOG = gen_catalog()

    def __validate_signature(self, data):
        sha_name, signature = self.request.headers.get('X-Hub-Signature').split('=')
        if sha_name != 'sha1':
            return False

        # HMAC requires its key to be bytes, but data is strings.
        mac = hmac.new(
            bytes(self.application.SECRET_TXT, "utf-8"),
            msg=data,
            digestmod=hashlib.sha1,
        )
        return hmac.compare_digest(mac.hexdigest(), signature)


class IndexHandler(tornado.web.RequestHandler):
    def get(self):
        self.render(
            "index.html",
            top_part=HEADER,
            catalog=self.application.CATALOG,
            article_url=self.__article_url,
        )

    def __article_url(self, filename):  # generate url of articles in index page
        return os.path.join("./article", '.'.join([filename, 'html']))


class ArticleHandler(tornado.web.RequestHandler):
    def get(self, filename):
        if not os.path.exists(self.__article_path(filename)):
            raise tornado.web.HTTPError(404)

        with open(self.__article_path(filename)) as f:
            html_body = publish_parts(f.read(), writer_name="html")["html_body"]

        self.render("article.html", top_part=HEADER, article=html_body)

    def __article_path(self, filename):
        return os.path.join(ARTICLE_PATH, filename)


class AboutMeHandler(ArticleHandler):
    def get(self):
        super(AboutMeHandler, self).get("aboutme.rst")


# app
class Application(tornado.web.Application):
    def __init__(self):
        try:
            with open(SECRET_TXT_PATH, "r") as f:
                self.SECRET_TXT = f.readline()[:-1]
        except IOError:
            logging.error("secret.txt not found, reject to boot")
            sys.exit()

        self.CATALOG = gen_catalog()

        handlers = [
            (r"/", IndexHandler),
            (r"/aboutme\.rst\.html/?", AboutMeHandler),
            (r"/article/img/(.+)", tornado.web.StaticFileHandler, {"path": ARTICLE_IMG_PATH}),
            (r"/article/(.+)\.html/?", ArticleHandler),
            (r"/webhooks/?", GithubWebHooksHandler),
        ]
        settings = {
            "template_path": TPL_PATH,
            "static_path": STATIC_PATH,
            "cookie_secret": "b6c20d57-958c-40ee-be9b-5a0f71a86285",
            "debug": options.debug,
        }
        tornado.web.Application.__init__(self, handlers, **settings)

        # we set tornado to watch main.py, restart when this file changes
        tornado.autoreload.start()
        tornado.autoreload.watch(MAIN_FILE_PATH)


if __name__ == "__main__":
    app = Application()
    app.listen(options.port)
    logging.warn("server has been listen at 127.0.0.1:%s with debug set to %s." % (options.port, options.debug))
    tornado.ioloop.IOLoop.current().start()
