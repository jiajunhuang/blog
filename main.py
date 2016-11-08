# coding: utf-8

"""
- 从posts目录读取文件，生成目录，并将目录缓存在内存中
- 当请求文件时，提取出文件名并且尝试读取文件，生成html
- 当推送代码到Github时，由Github发出Webhook请求，响应请求并且拉取最新代码
  更新内存中缓存的目录，重启当前tornado进程

TODO:
    增加redis缓存支持，使博客响应更快
"""

import hashlib
import hmac
import logging
import os
import sys
import subprocess

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
define("redis", default=False, type=bool, help="use redis as cache system")
parse_command_line()

# constants
PROJ_PATH = os.path.dirname(__file__)

POSTS_PATH = os.path.join(PROJ_PATH, "posts")
POST_IMG_PATH = os.path.join(POSTS_PATH, "img")

TPL_PATH = os.path.join(PROJ_PATH, "templates")
STATIC_PATH = os.path.join(PROJ_PATH, "static")
SECRET_TXT_PATH = os.path.join(PROJ_PATH, "secret.txt")
MAIN_FILE_PATH = os.path.join(PROJ_PATH, __file__)


# utils
def gen_catalog():
    return []


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
        self.app.CATALOG = gen_catalog()

    def __validate_signature(self, data):
        sha_name, signature = self.request.headers.get('X-Hub-Signature').split('=')
        if sha_name != 'sha1':
            return False

        # HMAC requires its key to be bytes, but data is strings.
        mac = hmac.new(bytes(self.app.SECRET_TXT, "utf-8"), msg=data, digestmod=hashlib.sha1)
        return hmac.compare_digest(mac.hexdigest(), signature)


class IndexHandler(tornado.web.RequestHandler):
    pass


class ArticleHandler(tornado.web.RequestHandler):
    pass


class AboutMeHandler(ArticleHandler):
    pass


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
            (r"/article/(.+)\.html/?", ArticleHandler),
            (r"/article/img/(.+)", tornado.web.StaticFileHandler, {"path": POST_IMG_PATH}),
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
