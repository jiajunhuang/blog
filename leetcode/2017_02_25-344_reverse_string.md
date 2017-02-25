# 344. Reverse String

> https://leetcode.com/problems/reverse-string/?tab=Description

给一个字符串，返回反序字符串。Python版本的那可是相当简单。。因为有内置的，
而且官方为了防止使用标准库，特意把reversed给去掉了。但是还是没有办法去掉
`[::-1]` 这种形式的代码，哈哈。

其实我们不用作弊啦，思路就是，用两个指针，一个从左到右，一个从右到左，依次交换
所指向的字符。或者另一种方式，用快行指针的方法，让一个指针指向中间字符，然后
分别向两端展开，然后交换对应的字符。

----

- [Go](./code/344.reverse_string.go)
- [Python](./code/344.reverse_string.py)
