import re
import os
import datetime


def load_mds(posts_dir, title_prefix=""):  # it's a little duplicate with function in `gen_catalog.py`...
    # e.g. 2014_06_17-use_cron.rst, 2014_06_17-use_cron.md
    r = re.compile(r"(\d{4}_\d{2}_\d{2})-.+\..+")

    articles = []
    for filename in sorted(os.listdir(posts_dir), reverse=True):
        result = r.match(filename)
        if result:
            date = result.group(1).replace("_", "/")
            with open(os.path.join(posts_dir, filename)) as f:
                title = f.readline().strip()
                if filename.split(".")[-1] == "md":
                    title = title_prefix + title.lstrip("# ")
            articles.append((title, datetime.datetime.strptime(date, "%Y/%m/%d"), filename))

    return articles
