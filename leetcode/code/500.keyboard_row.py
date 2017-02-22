class Solution(object):
    def findWords(self, words):
        alphabet_sets = [set("qwertyuiop"), set("asdfghjkl"), set("zxcvbnm")]
        result = []

        for word in words:
            for alphabet_set in alphabet_sets:
                if set(word.lower()).issubset(alphabet_set):
                    result.append(word)
                    break

        return result


if __name__ == "__main__":
    assert Solution().findWords(["Hello", "Qwerty"]) == ["Qwerty"], "buggy"
