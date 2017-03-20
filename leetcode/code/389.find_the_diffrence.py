class Solution(object):
    def findTheDiffrence(self, s, t):
        temp_str = "".join([s, t])
        result = ord(temp_str[0])

        for c in temp_str[1:]:
            result ^= ord(c)

        return chr(result)


if __name__ == "__main__":
    s = Solution()
    assert s.findTheDiffrence("abcd", "abcde") == "e"
