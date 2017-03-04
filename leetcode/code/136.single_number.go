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
