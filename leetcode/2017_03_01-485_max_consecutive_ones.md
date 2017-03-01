# 485. Max Consecutive Ones

> https://leetcode.com/problems/max-consecutive-ones/?tab=Description

给一个0和1组成的数列，求其中最长的连续为1的序列长度。

这个问题，看到的最初想法就是遍历。用一个变量存储最终结果，用另一个存储遍历过程
中的临时长度，如果临时长度比最终结果大，就替换之。但是，因为在for循环的过程中
进行判断上一个1是不是已经结束，所以这里有一个小坑，就是最后一个数字是无法判断
到的。

```python
class Solution(object):
    def findMaxConsecutiveOnes(self, nums):
        result = temp = 0
        nums.append(0)

        for i in nums:
            if i:
                temp += 1
            else:
                result = max(temp, result)
                temp = 0

        return result


if __name__ == "__main__":
    assert Solution().findMaxConsecutiveOnes([1, 1, 0, 1, 1, 1]) == 3
```

-----

- [Go](./code/485.max_consecutive_ones.go)
- [Python](./code/485.max_consecutive_ones.py)
