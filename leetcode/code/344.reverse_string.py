class Solution(object):
    def reverseString(self, s):
        return s[::-1]


if __name__ == "__main__":
    assert Solution().reverseString("hello") == "olleh"
