# coding: utf-8

from controllers.article import ArticleHandler


class AboutMeHandler(ArticleHandler):
    def get(self, suffix):
        super().get("aboutme.rst")
