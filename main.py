# coding: utf-8

import logging

import tornado.web
import tornado.ioloop
import tornado.autoreload

from config import (
    ARTICLE_IMG_PATH,
    STATIC_PATH,
    TEMPLATE_PATH,
)
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
            (r"/article/img/(.+)", tornado.web.StaticFileHandler, {"path": ARTICLE_IMG_PATH}),
            (r"/article/(.+)\.html/?", ArticleHandler),
            (r"/aboutme\.rst\.html/?", AboutMeHandler),
            (r"/webhooks/?", GithubWebHooksHandler),
        ]
        settings = {
            "template_path": TEMPLATE_PATH,
            "static_path": STATIC_PATH,
            "cookie_secret": "cfHo1VmQ8z9kut.wMVwympjbM",
            "debug": options.debug,
        }
        tornado.web.Application.__init__(self, handlers, **settings)


if __name__ == "__main__":
    app = Application()
    app.listen(options.port)
    logging.warn("server has been listen at port %s with debug set to %s." % (options.port, options.debug))
    tornado.ioloop.IOLoop.current().start()
