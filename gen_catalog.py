# coding: utf-8

import os
import re
import operator


def gen_catalog(posts_dir, output_file, headers, footers):
    # e.g. 2014_06_17-use_cron.rst, 2014_06_17-use_cron.md
    r = re.compile(r"(\d{4}_\d{2}_\d{2})-.+\..+")

    catalog = []
    for filename in os.listdir(posts_dir):
        result = r.match(filename)
        if result:
            date = result.group(1).replace("_", "/")
            with open(os.path.join(posts_dir, filename)) as f:
                title = f.readline().strip()
                if filename.split(".")[-1] == "md":
                    title = title.lstrip("# ")
            catalog.append((title, date, filename))

    # sort by filename, in a reverse order
    catalog = sorted(catalog, key=operator.itemgetter(2), reverse=True)

    with open(output_file, "w+") as f:
        # clear all the contents in file
        f.truncate()

        for header in headers:
            f.write(header)
            f.write("\n\n")

        # write catalog
        for item in catalog:
            title, date, filename = item
            f.write(
                "- {date} - [{title}]({posts_dir}/{filename})\n".format(
                    date=date,
                    title=title,
                    posts_dir=posts_dir,
                    filename=filename,
                )
            )

        for footer in footers:
            f.write(footer)
            f.write("\n\n")


if __name__ == "__main__":
    # README.md
    readme_headers = [
        "# Jiajun's Blog",
        "会当凌绝顶，一览众山小。",
        "## 关于我",
        "[点我](articles/aboutme.md)",
        "## 目录",
    ]
    readme_footers = [
        "--------------------------------------------",
        "[CC-BY](http://opendefinition.org/licenses/cc-by/)",
    ]
    gen_catalog(
        "articles",
        "./README.md",
        readme_headers,
        readme_footers,
    )

    # leetcode.md
    leetcode_headers = [
        "# Leetcode in Golang, Python",
    ]
    leetcode_footers = [
        "--------------------------------------------",
        "[CC-BY](http://opendefinition.org/licenses/cc-by/)",
    ]
    gen_catalog(
        "leetcode",
        "./leetcode/leetcode.md",
        leetcode_headers,
        leetcode_footers,
    )
