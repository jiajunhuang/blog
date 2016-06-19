# coding: utf-8

import os
import logging

import tornado.web
import tornado.ioloop
import tornado.autoreload

from config import Config
from controllers.aboutme import AboutMeHandler
from controllers.index import IndexHandler
from controllers.article import ArticleHandler
from controllers.webhooks import GithubWebHooksHandler

from tornado.options import define, options, parse_command_line
define("debug", default=False, type=bool, help="debug is set to True if this option is set")
define("port", default=8080, type=int, help="port=8080")
parse_command_line()


class Application(tornado.web.Application):
    def __init__(self):
        handlers = [
            (r"/", IndexHandler),
            (r"/article/img/(.+)", tornado.web.StaticFileHandler, {"path": Config().article_img_path}),
            (r"/article/(.+)/?", ArticleHandler),
            (r"/aboutme(.rst)?/?", AboutMeHandler),
            (r"/webhooks/?", GithubWebHooksHandler),
        ]
        settings = {
            "template_path": Config().template_path,
            "static_path": Config().static_path,
            "cookie_secret": "cfHo1VmQ8z9kut.wMVwympjbM",
            "debug": options.debug,
        }
        if os.path.exists(Config().posts_path):
            tornado.autoreload.watch(Config().posts_path)
            settings.update({
                "autoreload": True,
            })
        tornado.web.Application.__init__(self, handlers, **settings)


if __name__ == "__main__":
    app = Application()
    app.listen(options.port)
    logging.warn("server has been listen at port %s with debug set to %s." % (options.port, options.debug))
    tornado.ioloop.IOLoop.current().start()
