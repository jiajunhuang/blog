package main

import (
    "fmt"
    "strconv"
)

func fizzBuzz(n int) []string {
    var result = make([]string, 0)

    for i := 1; i <= n; i++ {
        if i % 15 == 0 {
            result = append(result, "FizzBuzz")
        } else if i % 5 == 0 {
            result = append(result, "Buzz")
        } else if i % 3 == 0 {
            result = append(result, "Fizz")
        } else {
            result = append(result, strconv.Itoa(i))
        }
    }

    return result
}


func main() {
    expected := []string{"1", "2", "Fizz", "4", "Buzz"}
    result := fizzBuzz(5)

    for i := 0; i < 5; i ++ {
        if expected[i] != result[i] {
            fmt.Println("buggy")
        }
    }
}
