# coding: utf-8

import os
import re
import operator


def gen_catalog():
    r = re.compile(r"(\d{4}_\d{2}_\d{2})-.+\..+")  # e.g. 2014_06_17-use_cron.rst

    catalog = []
    for filename in os.listdir("./"):
        result = r.match(filename)
        if result:
            date = result.group(1).replace("_", "/")
            with open(filename) as f:
                title = f.readline().strip()
            catalog.append((title, date, filename))

    catalog = sorted(catalog, key=operator.itemgetter(2), reverse=True)  # sort by filename, in a reverse order

    with open("./README.rst", "r+") as f:
        # clear all the contents in file
        f.truncate()

        # write title, aboutme
        f.write("Jiajun's Blog\n================\n\n")
        f.write("会当凌绝顶，一览众山小。\n\n")
        f.write("关于我\n------------\n\n")
        f.write("`点我 <aboutme.rst>`__\n\n")
        f.write("目录\n---------\n\n")

        # write catalog
        for item in catalog:
            title, date, filename = item
            f.write(
                "- {date} - `{title} <{filename}>`__\n".format(
                    date=date,
                    title=title,
                    filename=filename,
                )
            )

        # append LICENSE
        f.write("\n")
        f.write("--------------------------------------------\n\n")
        f.write("`CC-BY <http://opendefinition.org/licenses/cc-by/>`__\n")


if __name__ == "__main__":
    gen_catalog()
