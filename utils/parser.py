# coding: utf-8

from config import Config


class RestParser:
    def gen(self, filename):
        from docutils.core import publish_parts
        with open(Config().article_path(filename)) as f:
            return publish_parts(f.read(), writer_name="html")["html_body"]
