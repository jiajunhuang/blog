# coding: utf-8

import os
import tornado.web

from config import Config
from utils.parser import TextParser


class ArticleHandler(tornado.web.RequestHandler):
    def get(self, filename):
        if not os.path.exists(Config().article_path(filename)):
            raise tornado.web.HTTPError(404)
        article = TextParser().publish_html(filename)
        self.render("article.html", top_part=Config().top_part, article=article)
