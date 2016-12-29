# Python零碎小知识

今天读Mock的源代码，发现的几个新玩意儿。以后读代码发现了也记这里吧 :smile:

- dict有个 `setdefault` 方法，`dict.setdefault(key, default)` 相当于
`dict.get(key, default)` 并且如果 `dict[key]` 为空还会 `dict.set(key, default)`

```bash
In [1]: a = dict()

In [2]: a.setdefault("A", 1)
Out[2]: 1

In [3]: a.setdefault("A", 2)
Out[3]: 1
```

- 判断一个方法是否为 magic 方法的小trick：

```python
def _is_magic(name):
    return '__%s__' % name[2:-2] == name
```

```bash
In [1]: def _is_magic(name):
   ...:     return '__%s__' % name[2:-2] == name
   ...:

In [2]: _is_magic("__magic__")
Out[2]: True

In [3]: _is_magic("name")
Out[3]: False
```

原理很简单，只是觉得很有趣啦，哈哈。:smile:

- lazy形式格式化字符串。平时我们格式化字符串都是 `'%s %s' % ('hello', 'world')`
这样子。但是如果有需求说要先填充一个，然后后边再填充呢？当然我们可以先格式化
一部分，然后后面再append新的字符串。但是也可以这样：

```python
>>> '%s %%s' % 'hello'
'hello %s'
>>> '%s %%s' % 'hello' % 'world'
'hello world'
>>>
```

哈哈哈，是不是也很简单。不过这只能算是一个trick，不能满足更灵活的要求，例如
要格式化更长的字符串咋办。

- 比较两个数大小。这个是从unittest的utils里看到的，感觉还是蛮优雅的，利用了
Python中布尔值也可以做数值运算的特性：

```python
In [1]: def cmp(x, y):
   ...:     return (x > y) - (x < y)
   ...:

In [2]: cmp(1, 2)
Out[2]: -1

In [3]: cmp(2, 1)
Out[3]: 1

In [4]: cmp(2, 2)
Out[4]: 0

In [5]:
```

因为C89没有定义布尔类型，0就是false，1就是true。所以c也可以这样搞。

```c
$ cat ~/tests/test.c
#include <stdio.h>

int cmp(int x, int y) {
    return (x > y) - (x < y);
}

int main(void) {
    int a = 1;
    int b = 2;
    printf("cmp(%d, %d) = %d\n", a, b, cmp(a, b));
}
```

```bash
$ ~/tests/a.out
cmp(1, 2) = -1
```

贴这个出来，主要是平时写惯了 `if...else...` 固化了思维，突然看到这个有点
眼前一亮的感觉。。。

- fnmatch 这是在看unittest源代码的时候翻到的，c里面也有这个函数，`man fnmatch` 一下就知道，主要是用unix下shell的通配符来匹配，文档在 https://docs.python.org/2/library/fnmatch.html

```python
import fnmatch
import os

for file in os.listdir('.'):
    if fnmatch.fnmatch(file, '*.txt'):
        print file
```
