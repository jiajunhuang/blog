---
layout:     post
title:      Python-Challenge第三题
tags: [python]
---

记录一下这一题:

* 首先, 把网页从网上下下来:
* 然后进行正则匹配, 把注释部分匹配出来
* 然后进行翻译

```python
>>> import urllib2, re
>>> url = 'http://www.pythonchallenge.com/pc/def/equality.html'
>>> html = urllib2.urlopen(url).read() # 读取网页
>>> comments = re.findall('<!--[^>]*-->', html) # 用正则表达式提取出网页注释
>>> print ''.join(x[4:5] for x in re.findall('[^A-Z][A-Z]{3}[a-z][A-Z]{3}[^A-Z]', ''.join(content)))
linkedlist
```

然后把网页中的equality换成linkedlist, 它提示linkedlist.php,改吧, 骚年~

当然这只是我的做法, 每次页面上都会有提示, 如果你想看前一题的解答, 就把网页上的pc改成pcc ;-) 

~~今天我要继续做题, 我想之后我会整理出一些非常棒的解题方法的~~

###2014-06-22:

哎呀, 官网已经有人整理出来了~, [点我](http://wiki.pythonchallenge.com/index.php?title=Level3:Main_Page), 但是答案不是给你白看的, 想看答案, 需要先做出这题来~ 

还有一个写得很棒的正则表达式:

```python
>>> import re
>>> .join(x[1] for x in re.findall('(^|[^A-Z])[A-Z]{3}([a-z])[A-Z]{3}([^A-Z]|$)', text))
```

还有一个:

```python
>>> .join(re.findall('(?:^|[^A-Z])[A-Z]{3}([a-z])[A-Z]{3}(?:[^A-Z]|$)',text))
```

one more(even shorter! with RE):

```python
import urllib, re
print ''.join(re.findall('[^A-Z][A-Z]{3}[a-z][A-Z][^A-Z]', 
      urllib.open("http://www.pythonchallenge.com/pc/def/equality.html").read()))
```


这两个都涉及到了捕获型变量和不捕获, 可能要等我再复习了正则表达式才能详细讲解, 现在了解一点点但不全面
