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

当然，也可以使用这个 [在线工具](https://tools.jiajunhuang.com) 当然了，其实就是这个脚本的在线版。
