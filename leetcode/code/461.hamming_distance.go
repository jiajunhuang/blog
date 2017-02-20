package main

import "fmt"

func hammingDistance(x int, y int) int {
    xor := x ^ y
    result := 0

    for xor != 0 {
        xor &= xor - 1
        result++
    }

    return result
}


func main() {
    if hammingDistance(1, 4) != 2 {
        fmt.Println("code here should never be executed")
    }
}
