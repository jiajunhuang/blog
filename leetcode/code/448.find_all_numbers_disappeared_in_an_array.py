class Solution(object):
    def findDisappearedNumbers(self, nums):
        for i, v in enumerate(nums):
            if v == -1:
                continue

            mark(nums, nums[i])

        return [i+1 for i, v in enumerate(nums) if v != -1]


def mark(nums, current):
    fut = nums[current - 1]

    if fut == -1:
        return

    nums[current - 1] = -1
    mark(nums, fut)


if __name__ == "__main__":
    s = Solution()
    assert s.findDisappearedNumbers([1, 2, 3, 5, 2]) == [4]
