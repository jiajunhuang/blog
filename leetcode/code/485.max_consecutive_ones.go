package main

func findMaxConsecutiveOnes(nums []int) int {
    var result, temp = 0, 0
    nums = append(nums, 0)

    for _, num := range(nums) {
        if num == 1 {
            temp++
        } else {
            if temp > result {
                result = temp
            }
            temp = 0
        }
    }

    return result
}

func main() {
    array := []int{1, 1, 0, 1, 1, 1}
    if findMaxConsecutiveOnes(array) != 3 {
        println("buggy")
    }
}
