package main

import "fmt"

func findTheDiffrence(s, t string) rune {
	cache := make(map[rune]bool)

	for _, r := range s {
		cache[r] = true
	}

	var result rune
	for _, r := range t {
		if _, ok := cache[r]; !ok {
			result = r
			break
		}
	}

	return result
}

func main() {
	fmt.Printf("%c\n", findTheDiffrence("abcd", "abcde"))
}
