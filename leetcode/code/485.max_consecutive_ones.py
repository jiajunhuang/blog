class Solution(object):
    def findMaxConsecutiveOnes(self, nums):
        result = temp = 0
        nums.append(0)

        for i in nums:
            if i:
                temp += 1
            else:
                result = max(temp, result)
                temp = 0

        return result


if __name__ == "__main__":
    assert Solution().findMaxConsecutiveOnes([1, 1, 0, 1, 1, 1]) == 3
