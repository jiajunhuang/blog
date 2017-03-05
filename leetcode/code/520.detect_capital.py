class Solution(object):
    def detectCapitalUse(self, word):
        result = self.sameCase(word[1:])

        if len(word) > 1 and word[0].islower() and word[1].isupper() and result:
            return False

        return result

    def sameCase(self, chars):
        return (
            len(chars) == 0 or
            len(list(filter(lambda x: x.islower(), chars))) == 0 or
            len(list(filter(lambda x: x.isupper(), chars))) == 0
        )


if __name__ == "__main__":
    s = Solution()
    assert s.detectCapitalUse("USA") is True
    assert s.detectCapitalUse("Little") is True
    assert s.detectCapitalUse("world") is True
    assert s.detectCapitalUse("hLLo") is False
