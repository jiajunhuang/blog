class Solution(object):
    def islandPerimeter(self, grid):
        """
        :type grid: List[List[int]]
        :rtype: int
        """
        result = 0

        for ri, row in enumerate(grid):
            for ci, col in enumerate(row):
                if col == 1:
                    if ri == 0 or grid[ri - 1][ci] == 0:
                        result += 1
                    if ci == len(row) - 1 or grid[ri][ci + 1] == 0:
                        result += 1
                    if ri == len(grid) - 1 or grid[ri + 1][ci] == 0:
                        result += 1
                    if ci == 0 or grid[ri][ci - 1] == 0:
                        result += 1
        return result


if __name__ == "__main__":
    grid = [
        [0, 1, 0, 0],
        [1, 1, 1, 0],
        [0, 1, 0, 0],
        [1, 1, 0, 0],
    ]
    assert Solution().islandPerimeter(grid) == 16
