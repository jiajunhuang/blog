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
