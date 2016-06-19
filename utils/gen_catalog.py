# coding: utf-8

import os
import re
import operator

from config import Config
from utils.singleton import Singleton


class Catalog(metaclass=Singleton):
    def __init__(self):
        catalog = []
        r = re.compile(Config().filename_format)

        for filename in os.listdir(Config().posts_path):
            result = r.match(filename)  # match or not
            if result:
                date = result.group(1)
                with open(os.path.join("./posts", filename)) as f:
                    date = date.replace("_", "-")
                    title = f.readline()
                    catalog.append((title, date, filename))

        self.catalog = sorted(catalog, key=operator.itemgetter(1), reverse=True)

    def __iter__(self):
        for item in self.catalog:
            yield item
