class Solution(object):
    def findComplement(self, num):
        i = 1

        while i <= num:
            num ^= i
            i *= 2

        return num


if __name__ == "__main__":
    assert Solution().findComplement(5) == 2, "output should be 2"
