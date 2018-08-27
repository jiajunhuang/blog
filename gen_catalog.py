# coding: utf-8

import os
import re
import operator


def gen_catalog(posts_dir, output_file, headers, footers, relative_path):
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
                "- {date} - [{title}](https://jiajunhuang.com/{relative_path}/{filename}.html)\n".format(
                    date=date,
                    title=title,
                    relative_path=relative_path,
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
        "- [关于我](articles/aboutme.md)",
        "- 微信联系我",
        "![](./articles/img/wechat_qrcode.png)",
        "## 目录",
    ]
    readme_footers = [
        "\n",
        "--------------------------------------------",
        "禁止转载",
    ]
    gen_catalog(
        "articles",
        "./README.md",
        readme_headers,
        readme_footers,
        "articles",
    )

    # leetcode.md
    leetcode_headers = [
        "# Leetcode in Golang, Python",
    ]
    leetcode_footers = [
        "\n",
        "--------------------------------------------",
        "禁止转载",
    ]
    gen_catalog(
        "leetcode",
        "./leetcode/README.md",
        leetcode_headers,
        leetcode_footers,
        "."
    )
