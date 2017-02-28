package main

func islandPerimeter(grid [][]int) int {
    if grid == nil {
        return 0
    }

    var result int

    for ri, row := range grid {
        for ci, point := range row {
            if point == 1 {
                result += 4

                if ri > 0 && grid[ri - 1][ci] == 1 {
                    result -= 2
                }
                if ci > 0 && grid[ri][ci - 1] == 1 {
                    result -= 2
                }
            }
        }
    }

    return result
}


func main() {
    grid := [][]int{
        {0, 1, 0, 0},
        {1, 1, 1, 0},
        {0, 1, 0, 0},
        {1, 1, 0, 0},
    }
    if islandPerimeter(grid) != 16 {
        println("buggy")
    }
}
