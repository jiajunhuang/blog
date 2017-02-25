class Solution(object):
    def fizzBuzz(self, n):
        result = []

        for i in range(1, n + 1):
            if i % 15 == 0:
                result.append("FizzBuzz")
            elif i % 5 == 0:
                result.append("Buzz")
            elif i % 3 == 0:
                result.append("Fizz")
            else:
                result.append(str(i))

        return result


if __name__ == "__main__":
    assert Solution().fizzBuzz(3) == ["1", "2", "Fizz"]
    assert Solution().fizzBuzz(5) == ["1", "2", "Fizz", "4", "Buzz"]
