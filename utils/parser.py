# coding: utf-8

from config import Config


class RestParser:
    def gen(self, filename):
        from docutils.core import publish_parts
        with open(Config().article_path(filename)) as f:
            return publish_parts(f.read(), writer_name="html")["html_body"]


class MarkdownParser:
    def gen(self, filename):
        from markdown2 import markdown
        with open(Config().article_path(filename)) as f:
            return markdown(f.read())


class AsciiParser:
    def gen(self, filename):
        with open(Config().article_path(filename)) as f:
            return f.read()


class TextParser:
    __parser_mapper = {
        "rst": RestParser,
        "md": MarkdownParser,
        "ascii": AsciiParser,
    }

    def publish_html(self, filename):
        return self.__parser_mapper[Config().text_type]().gen(filename)
