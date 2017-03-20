# 389. Find the Difference

> https://leetcode.com/problems/find-the-difference/#/description

拿到这个题目有两个想法：

- 因为字符也是二进制表示，我们可以利用二进制的异或运算，将两个字符连接起来之后
进行异或运算，便可以将不同的那个字符的二进制找出来。

```python
class Solution(object):
    def findTheDiffrence(self, s, t):
        temp_str = "".join([s, t])
        result = ord(temp_str[0])

        for c in temp_str[1:]:
            result ^= ord(c)

        return chr(result)
```

- 使用集合，或者哈希表，总之就是缓存住一个字符串，然后遍历另一个字符串，就可以
把不同的字符串找出来。

```go
package main

import "fmt"

func findTheDiffrence(s, t string) rune {
	cache := make(map[rune]bool)

	for _, r := range s {
		cache[r] = true
	}

	var result rune
	for _, r := range t {
		if _, ok := cache[r]; !ok {
			result = r
			break
		}
	}

	return result
}

func main() {
	fmt.Printf("%c\n", findTheDiffrence("abcd", "abcde"))
}
```

-----

- [Go](./code/389.find_the_diffrence.go)
- [Python](./code/389.find_the_diffrence.py)
