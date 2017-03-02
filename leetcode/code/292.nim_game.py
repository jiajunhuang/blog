class Solution(object):
    def canWinNim(self, n):
        return n % 4 != 0


if __name__ == "__main__":
    for i in [1, 2, 3, 5, 6, 7, 9]:
        assert Solution().canWinNim(i) is True

    for i in [4, 8, 12]:
        assert Solution().canWinNim(i) is False
