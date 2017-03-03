# 520. Detect Capital

> https://leetcode.com/problems/detect-capital/?tab=Description

只有以下三种情况的字符串才是合法的：

    - 全大写
    - 全小写
    - 仅首字母大写


遇到字符串匹配问题，首先就应该想到的是正则表达式（没错，这回可学乖了）！

```go
package main

import (
    "regexp"
)

var validString = regexp.MustCompile(`\b([a-z]+|[A-Z]+|[A-Z][a-z]+)\b`)

func detectCapitalUse(word string) bool {
    return validString.MatchString(word)
}

func main() {
    if !(detectCapitalUse("USA") && detectCapitalUse("Title") && detectCapitalUse("hello")) {
        println("buggy")
    }

    if detectCapitalUse("wOrld") {
        println("buggy")
    }
}
```

其次第二种就是常规解法，即，从第二个字符开始必须保持大小写一致，第一个字母可以是大写，可以是小写，但是不能首字母小写后面是大写。然后对这种特殊情况作出区分即可。

```python
class Solution(object):
    def detectCapitalUse(self, word):
        result = self.sameCase(word[1:])

        if len(word) > 1 and word[0].islower() and word[1].isupper() and result:
            return False

        return result

    def sameCase(self, chars):
        return (
            len(chars) == 0 or
            len(list(filter(lambda x: x.islower(), chars))) == 0 or
            len(list(filter(lambda x: x.isupper(), chars))) == 0
        )


if __name__ == "__main__":
    s = Solution()
    assert s.detectCapitalUse("USA") is True
    assert s.detectCapitalUse("Little") is True
    assert s.detectCapitalUse("world") is True
    assert s.detectCapitalUse("hLLo") is False
```

当然，Python还有强大的标准库。。。

```python
class Solution():
    def detectCapitalUse(self, word):
        return word.isupper() or word.islower() or word.istitle()
```

----------

- [Go](./code/520.detect_capital.go)
- [Python](./code/520.detect_capital.py)
