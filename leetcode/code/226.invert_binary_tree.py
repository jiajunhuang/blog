class TreeNode(object):
    def __init__(self, x):
        self.val = x
        self.left = None
        self.right = None


class Solution(object):
    def invertTree(self, root):
        if root is None:
            return

        self.invertTree(root.left)
        self.invertTree(root.right)

        root.left, root.right = root.right, root.left

        return root


if __name__ == "__main__":
    s = Solution()
    tree = TreeNode(0)
    tree.left = TreeNode(1)
    tree.right = TreeNode(2)

    s.invertTree(tree)

    assert tree.val == 0
    assert tree.left.val == 2
    assert tree.right.val == 1
