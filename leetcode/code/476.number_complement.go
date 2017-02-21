package main

import "fmt"


func findComplement(num int) int {
    var i = 1

    for i <= num {
        num ^= i
        i <<= 1
    }

    return num
}


func main() {
    if (findComplement(5) != 2) {
        fmt.Println("output should be 2")
    }
}
