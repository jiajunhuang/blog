class Solution(object):
    def hammingDistance(self, x, y):
        xor = x ^ y
        result = 0

        while xor != 0:
            xor &= xor - 1
            result += 1

        return result


if __name__ == "__main__":
    assert Solution().hammingDistance(1, 4) == 2, "your buggy code"
