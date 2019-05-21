import re
import os
import datetime
from collections import Counter

import jieba

IGNORE_WORDS = {
    "(", "的", "阅读", "：", "-", ")", "和", "，", " ", ",", "？", "）", "（", "《", "》", "如何", "一个", "什么",
    "怎么", "一些", "一年", "in", "完成", "就要",
}


def load_mds(posts_dir, title_prefix="", path="articles"):  # it's a little duplicate with function in `gen_catalog.py`
    # e.g. 2014_06_17-use_cron.rst, 2014_06_17-use_cron.md
    r = re.compile(r"(\d{4}_\d{2}_\d{2})-.+\..+")

    words = Counter()
    articles = []

    for filename in sorted(os.listdir(posts_dir), reverse=True):
        result = r.match(filename)
        if result:
            date = result.group(1).replace("_", "/")
            with open(os.path.join(posts_dir, filename)) as f:
                title = f.readline().strip()
                if filename.split(".")[-1] == "md":
                    title = title_prefix + title.lstrip("# ")

                for word in jieba.cut(title, cut_all=False):
                    if word not in IGNORE_WORDS and len(word) > 1:
                        words[word] += 1  # 统计分词

            articles.append((title, datetime.datetime.strptime(date, "%Y/%m/%d"), filename, path))

    return articles, words
