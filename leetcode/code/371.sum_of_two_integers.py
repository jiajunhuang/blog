class Solution(object):
    def getSum(self, a, b):
        if a < 0:
            a = self.getSum(~a, 1)
        if b < 0:
            b = self.getSum(~b, 1)

        sum = a
        carry = b

        while carry:
            temp = sum
            sum = temp ^ carry
            carry = (temp & carry) << 1

        return sum


if __name__ == "__main__":
    s = Solution()
    assert s.getSum(1, 2) == 3
    assert s.getSum(1, -1) == 0
    assert s.getSum(-1, 1) == 0
