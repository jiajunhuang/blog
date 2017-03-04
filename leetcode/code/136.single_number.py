class Solution(object):
    def singleNumber(self, nums):
        result = 0

        for i in nums:
            result ^= i

        return result


if __name__ == "__main__":
    s = Solution()
    assert s.singleNumber([1, 2, 3, 2, 1]) == 3
