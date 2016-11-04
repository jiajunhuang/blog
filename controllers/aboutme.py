# coding: utf-8

from controllers.article import ArticleHandler


class AboutMeHandler(ArticleHandler):
    def get(self):
        super().get("aboutme.rst")
