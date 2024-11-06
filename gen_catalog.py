# coding: utf-8

from utils import load_mds


def gen_catalog(posts_dir, output_file, headers, footers):
    articles = load_mds("./articles")

    with open(output_file, "w+") as f:
        # clear all the contents in file
        f.truncate()

        for header in headers:
            f.write(header)
            f.write("\n\n")

        # write catalog
        for title, date, filename, path in articles:
            f.write(
                "- {date} - [{title}](https://jiajunhuang.com/{path}/{filename}.html)\n".format(
                    date=date.strftime("%Y/%m/%d"),
                    title=title,
                    path=path,
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
        "- [关于我](https://jiajunhuang.com/aboutme)",
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
    )
