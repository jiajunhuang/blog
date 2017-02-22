package main

import (
    "fmt"
)

func initializeAlphabet() map[rune]int {
    var alphabet = make(map[rune]int)

    alphabet['q'] = 1
    alphabet['w'] = 1
    alphabet['e'] = 1
    alphabet['r'] = 1
    alphabet['t'] = 1
    alphabet['y'] = 1
    alphabet['u'] = 1
    alphabet['i'] = 1
    alphabet['o'] = 1
    alphabet['p'] = 1
    alphabet['a'] = 2
    alphabet['s'] = 2
    alphabet['d'] = 2
    alphabet['f'] = 2
    alphabet['g'] = 2
    alphabet['h'] = 2
    alphabet['j'] = 2
    alphabet['k'] = 2
    alphabet['l'] = 2
    alphabet['z'] = 3
    alphabet['x'] = 3
    alphabet['c'] = 3
    alphabet['v'] = 3
    alphabet['b'] = 3
    alphabet['n'] = 3
    alphabet['m'] = 3
    alphabet['Q'] = 1
    alphabet['W'] = 1
    alphabet['E'] = 1
    alphabet['R'] = 1
    alphabet['T'] = 1
    alphabet['Y'] = 1
    alphabet['U'] = 1
    alphabet['I'] = 1
    alphabet['O'] = 1
    alphabet['P'] = 1
    alphabet['A'] = 2
    alphabet['S'] = 2
    alphabet['D'] = 2
    alphabet['F'] = 2
    alphabet['G'] = 2
    alphabet['H'] = 2
    alphabet['J'] = 2
    alphabet['K'] = 2
    alphabet['L'] = 2
    alphabet['Z'] = 3
    alphabet['X'] = 3
    alphabet['C'] = 3
    alphabet['V'] = 3
    alphabet['B'] = 3
    alphabet['N'] = 3
    alphabet['M'] = 3

    return alphabet
}

func findWords(words []string) []string {
    var result []string
    var alphabet = initializeAlphabet()

    for _, word := range(words) {
        if len(word) == 0 {
            continue
        }

        inSameRow := true
        // because alphabet row doesn't contain 0, so ignore the zero value check
        var firstRuneRow = alphabet[rune(word[0])]

        for _, arune := range(word[1:]) {
            if alphabet[arune] != firstRuneRow {
                inSameRow = false
                break
            }
        }

        if inSameRow {
            result = append(result, word)
        }
    }

    return result
}

func testSliceEqual(a, b []string) bool {
    if a == nil && b == nil {
        return true;
    }
    if a == nil || b == nil {
        return false;
    }
    if len(a) != len(b) {
        return false
    }
    for i := range a {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}


func main() {
    var testData []string
    testData = append(testData, "hello")
    testData = append(testData, "Qwerty")

    var resultExpected []string
    resultExpected = append(resultExpected, "Qwerty")

    result := findWords(testData)

    if !testSliceEqual(resultExpected, result) {
        fmt.Println("There must be bug(s) in your code!")
    }
}
