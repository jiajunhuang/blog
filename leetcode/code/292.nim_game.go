package main

func canWinNim(n int) bool {
    return n % 4 != 0
}

func main() {
    if !(canWinNim(1) && canWinNim(2) && canWinNim(3) && canWinNim(5)) {
        println("buggy")
    }
    if canWinNim(4) && canWinNim(8) {
        println("buggy")
    }
}
