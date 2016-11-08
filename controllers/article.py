# coding: utf-8

import os
import tornado.web

from config import (
    ARTICLE_IMG_PATH,
    TOP_PART,
)
from utils.parser import RestParser


class ArticleHandler(tornado.web.RequestHandler):
    def get(self, filename):
        if not os.path.exists(ARTICLE_IMG_PATH):
            raise tornado.web.HTTPError(404)
        article = RestParser().gen(filename)
        self.render("article.html", top_part=TOP_PART, article=article)
