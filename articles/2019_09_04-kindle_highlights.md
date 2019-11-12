# 提取kindle笔记

首先在Kindle上通过邮件把笔记分享到自己的邮箱，然后把html下载到本地，执行下面的脚本：

```python
#!/usr/bin/env python
import os
import sys

from lxml import html


def extract_notes(s):
    etree = html.fromstring(s)

    for i in etree.find_class("noteText"):
        yield i.text


if __name__ == "__main__":
    notes = []

    with open(sys.argv[1]) as f:
        s = f.read()
        for i in extract_notes(s):
            notes.append(i)

    print("".join(notes))
```

执行之后，就会输出提取之后的文档。

```bash
$ python kindle.py 读书笔记.html
```

---

2019.11.12注：

网页版已经移除，请使用脚本。先安装Python3，然后pip安装lxml，之后即可执行此脚本。
