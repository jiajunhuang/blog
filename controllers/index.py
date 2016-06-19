# coding: utf-8

import tornado.web

from config import Config
from utils.gen_catalog import Catalog


class IndexHandler(tornado.web.RequestHandler):
    def get(self):
        self.render("index.html", top_part=Config().top_part, catalog=Catalog(), article_url=Config().article_url)
