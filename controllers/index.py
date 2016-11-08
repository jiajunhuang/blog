# coding: utf-8

import tornado.web

from config import (
    ARTICLE_IMG_PATH,
    TOP_PART,
    Config,
)


class IndexHandler(tornado.web.RequestHandler):
    def get(self):
        self.render("index.html", top_part=TOP_PART, catalog=Config().catalog, article_url=Config.article_url)
