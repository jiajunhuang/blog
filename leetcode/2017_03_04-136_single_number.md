# 136. Single Number

> https://leetcode.com/problems/single-number/?tab=Description

最简单的方法应该就是使用一个集合来判断，不过鉴于题目不允许我们使用额外的存储
空间，所以这个思路报废。

判断数字是否重复，可以用异或运算，因为两个相同的数字异或之后所有位会变成0。
但是这里无法准确定位相同的两个数在哪里，所以好像不太好用异或。

不过说到无法定位两个数在哪里，我们可以先进行排序，然后再进行定位（但是常见的
比较排序算法时间复杂度都为O(lgN)，但是目前没想到更好的方法）。

```go
package main

import "sort"

type Nums []int

func (nums Nums) Len() int {
    return len(nums)
}

func (nums Nums) Swap(i, j int) {
    nums[i], nums[j] = nums[j], nums[i]
}

func (nums Nums) Less (i, j int) bool {
    return nums[i] < nums[j]
}

func singleNumber(nums []int) int {
    sort.Sort(Nums(nums))

    for i := 0; i < len(nums) - 1; i += 2 {
        if nums[i] != nums[i+1] {
            return nums[i]
        }
    }
    return nums[len(nums) - 1]
}

func main() {
    expected := Nums{1, 2, 3, 2, 1}
    if singleNumber(expected) != 3 {
        println("buggy")
    }
}
```

看到题解中用异或的方案了，膜拜！

```python
class Solution(object):
    def singleNumber(self, nums):
        result = 0

        for i in nums:
            result ^= i

        return result


if __name__ == "__main__":
    s = Solution()
    assert s.singleNumber([1, 2, 3, 2, 1]) == 3
```

解释一下，首先初始化为0，0和任何数字做异或都等于那个数，所以遍历刚开始的时候不
会出错。然后每次做异或，只要遇到了相同的数，原先做异或的时候该数字所产生的影响
都会被重置为0。所以最后保存的结果就是那个只出现过一次的数。

----------

- [Go](./code/136.single_number.go)
- [Python](./code/136.single_number.py)
