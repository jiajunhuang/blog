# 461. Hamming Distance

> https://leetcode.com/problems/hamming-distance/

对于两个整数的Hamming Distance是指在二进制表示中不同的位的个数。

首先观察完问题，我们要把问题精确化，缩小解题范围：

- 1，是整数之间的Hamming Distance
- 2，二进制位数是否相同？如果不是，是否需要在更短的整数左侧补0？

## 解法1：遍历二进制位进行比较。时：O(n)；空：O(1)

此解法需要对齐整数。思路为遍历，Python描述：

```python
def ham(x, y):
    x = "{0:032b}".format(x)
    y = "{0:032b}".format(y)

    result = 0
    for i in range(32):
        if x[i] != y[i]:
            result += 1
    return result


if __name__ == "__main__":
    assert ham(1, 4) == 2, "your buggy code"
```

## 解法2：位运算。时：O(n)；空：O(1)

能不能想办法把二进制位标记出来？回想了一下，异或便从脑海里跳了出来，我们复习一下
异或：

```
1 ^ 1 = 0
1 ^ 0 = 1
0 ^ 1 = 1
0 ^ 0 = 0
```

如果两个位相同，最后会变成0。反之，会变成1。于是我们可以做完异或之后，把结果保存
后再去算有多少个位是1，那么结果就是多少。

计算的方式有两种：

- xor & (1 << i) == (1 << i): 将1向左偏移i位，得到从右往左第i个位是1，其余都是0
然后和xor做与运算，如果xor的该位为1，那么结果为1，否则为0。所以再和(1 << i)判断
是否为1，从而得出结果。

- 将xor和xor - 1做与运算，一直到xor变为0为止。其原理是，xor - 1便会消除最右边的
一个1，因为减法需要借位。例如4的二进制表示为 `00000100`，那么4-1的二进制表示为
`00000011` 做一次与运算之后，便成了0。

-------

- [Go](./code/461.hamming_distance.go)
- [Python](./code/461.hamming_distance.py)
