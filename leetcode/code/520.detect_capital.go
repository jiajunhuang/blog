package main

import (
    "regexp"
)

var validString = regexp.MustCompile(`\b([a-z]+|[A-Z]+|[A-Z][a-z]+)\b`)

func detectCapitalUse(word string) bool {
    return validString.MatchString(word)
}

func main() {
    if !(detectCapitalUse("USA") && detectCapitalUse("Title") && detectCapitalUse("hello")) {
        println("buggy")
    }

    if detectCapitalUse("wOrld") {
        println("buggy")
    }
}
