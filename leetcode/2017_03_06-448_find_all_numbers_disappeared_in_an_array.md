# 448. Find All Numbers Disappeared in an Array

> https://leetcode.com/problems/find-all-numbers-disappeared-in-an-array/?tab=Description

拿到题目的时候想了好一会儿，要怎么不借助空间并且大O为n。想到能不能类似哈希表
一样给他做标记呢？于是想到了，把数组对应的位标记成-1，最后不是-1的位的index + 1
就是我们要的答案了。

```python
class Solution(object):
    def findDisappearedNumbers(self, nums):
        for i, v in enumerate(nums):
            if v == -1:
                continue

            mark(nums, nums[i])

        return [i+1 for i, v in enumerate(nums) if v != -1]


def mark(nums, current):
    fut = nums[current - 1]

    if fut == -1:
        return

    nums[current - 1] = -1
    mark(nums, fut)


if __name__ == "__main__":
    s = Solution()
    assert s.findDisappearedNumbers([1, 2, 3, 5, 2]) == [4]
```

题解里有更高效更简单的做法，就是用对应位的负数来做标记。这样就解决了我所想到的
这种方案要借助递归来实现递进标记的问题。

```go
package main

import (
    "fmt"
    "math"
)

func findDisappearedNumbers(nums []int) []int {
    for i := 0; i < len(nums); i++ {
        val := int(math.Abs(float64(nums[i])) - 1)

        if nums[val] > 0 {
            nums[val] = -nums[val]
        }
    }

    result := []int{}
    for i := 0; i < len(nums); i++ {
        if nums[i] > 0 {
            result = append(result, i+1)
        }
    }

    return result
}


func main() {
    toFind := []int{2, 3, 5, 1, 2}

    result := findDisappearedNumbers(toFind)

    if !(len(result) == 1 && result[0] == 4) {
        fmt.Println("buggy")
    }
}
```

---

- [Go](./code/448.find_all_numbers_disappeared_in_an_array.go)
- [Python](./code/448.find_all_numbers_disappeared_in_an_array.py)
