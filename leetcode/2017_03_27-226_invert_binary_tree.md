# 226. Invert Binary Tree

> https://leetcode.com/problems/invert-binary-tree/#/description

递归的话就很简单了：

```python
class Solution(object):
    def invertTree(self, root):
        if root is None:
            return

        self.invertTree(root.left)
        self.invertTree(root.right)

        root.left, root.right = root.right, root.left

        return root
```

非递归版本其实就是需要用到栈来手动模拟递归调用栈。

TODO

---

- [Python](./code/226.invert_binary_tree.py)
