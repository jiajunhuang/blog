package main

import "fmt"

func reverseString(s string) string {
    length := len(s)
    result := []rune(s)

    for i, j:= 0, length - 1; i < j; {
        result[i], result[j] = result[j], result[i]
        i++
        j--
    }

    return string(result)
}


func main() {
    expect := reverseString("hello")
    if expect != "olleh" {
        fmt.Println("buggy")
    }
}
