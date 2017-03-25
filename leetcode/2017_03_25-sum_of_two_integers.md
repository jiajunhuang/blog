# 371. Sum of Two Integers

> https://leetcode.com/problems/sum-of-two-integers/#/description

这题没做出来，只能处理出加法，减法忘记要怎么处理了。《编码的奥秘》里面是有讲
这个的，但是上次看还是五年前，看一下别人的解答吧。

```python
class Solution(object):
    def getSum(self, a, b):
        MAX_INT = 0x7FFFFFFF
        MIN_INT = 0x80000000
        MASK = 0x100000000

        while b:
            a, b = (a ^ b) % MASK, ((a & b) << 1) % MASK

        return a if a <= MAX_INT else ~((a % MIN_INT) ^ MAX_INT)
```

另外就是可以分别处理加法和减法，但是判断很麻烦。
