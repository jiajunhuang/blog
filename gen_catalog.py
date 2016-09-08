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
            date = result.group(1)
            with open(filename) as f:
                title = f.readline()
            catalog.append((title, date, filename))

    catalog = sorted(catalog, key=operator.itemgetter(2), reverse=True)  # sort by filename, in a reverse order

    with open("./README.rst", "w") as f:
        for line in f.readlines():
            if line == "目录":
                f.readline()  # move to next line

        # write catalog
        for item in catalog:
            title, date, filename = item
            f.write(
                "{date} - {title} `<{filename}>`__\n".format(
                    date=date,
                    title=title,
                    filename=filename,
                )
            )

        # append LICENSE
        f.write(
            """
            CC-BY <http://opendefinition.org/licenses/cc-by/>__
            """
        )


if __name__ == "__main__":
    gen_catalog()
