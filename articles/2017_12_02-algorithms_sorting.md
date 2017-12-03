# 算法导论阅读笔记 --- 排序算法

> 这篇文章是我在学习算法导论之后的笔记

排序算法，无论是在工作上还是面试中，都是经常遇到的问题。话不多说，我们一个一个来回忆。为了方便理解，我们总是假设输入的
数据是[9..0]逆序排列的。而我们要输出的则是升序排列。

## 插入排序

插入排序，就是用的一种抽象的方式。什么抽象的方式呢？就是假设左边的已经是排好顺序的数组，而下一步要做的事情，就是从右边
拿一个数字，在左边找一个合适的位置把它插进去。仔细想想，是不是和打扑克牌时，我们把刚抓起的牌插到合适的位置一样呢？
为了插到合适的位置，我们必需从已经排好序的这一段里，从右往左依次查找，我们把当前正在遍历的元素的下标叫做cursor好了，
cursor的初始值是我们准备插入的数字的下标。如果发现我们要插入的数字比 `cursor - 1` 所在的数字大，那么就不动，否则，就把
`cursor - 1`上的元素和 `cursor` 所在的位置上的元素互换，然后继续这一步，进行查找。

我们来看例子：输入是 `[9, 8, 7, 6, 5, 4, 3, 2, 1, 0]`。

- 最开始我们假设，已经排好序的那部分是 `[9]`，而我们当前要进行的这一步是把 `8` 插入到合适的位置。进行比较，
发现8比9更小，所以把9和8互换，于是整个数组变成了 `8, 9, 7, 6, 5, 4, 3, 2, 1, 0`，然后发现到头了，所以这一步到此为止，
我们进行下一个动作。

- 然后我们要对7进行插入。首先往左看一下，发现9比7大，于是互换两者，数组变成了 `8, 7, 9, 6, 5, 4, 3, 2, 1, 0`，继续往左
探，发现8仍然比7大，于是继续互换两者，数组变成了 `7, 8, 9, 6, 5, 4, 3, 2, 1, 0`，发现到头了，所以这一步也到此为止，
我们进行下一个动作。

- 。。。依次类推，一直到右边所有的元素都弄完了，排序也就完成了。

- 那我们想想是否有特殊情况需要处理呢？因为我们从数组的第二个元素，也就是下标为1的元素开始操作，也就是说，我们假设数组
至少有两个元素。所以如果只有一个元素，我们需要进行判断，直接原样返回。而如果数组为空数组，则需要抛出一个错误，表明不能
对空数组进行排序。

下面我们来看代码：

```python
def insertion_sort(array):
    length = len(array)  # 因为要用两次，所以用变量村起来，Don't Repeat Yourself

    if length == 0:
        raise ValueError("请不要在空数组上使用排序")

    if length == 1:
        return array

    for i in range(1, length):
        for cursor in range(i, 0, -1):
            if array[cursor] < array[cursor - 1]:
                # 交换
                array[cursor], array[cursor - 1] = array[cursor - 1], array[cursor]

    return array


if __name__ == "__main__":
    # 写几个简单地测试用例
    assert insertion_sort([1]) == [1]
    assert insertion_sort([i for i in range(9, -1, -1)]) == [i for i in range(10)]
    print(insertion_sort([i for i in range(9, -1, -1)]))
```

执行一下就会发现，输出是正确的：

```bash
$ python insertion.py 
[0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
```

可以看出，插入排序里，有 `for...for...` 这样的嵌套循环，它的最差时间复杂度是 `O(n^2)`，这种情况就出现在我们所给的测试
用例上：逆序输入。最好的情况则是顺序：O(n)。平均情况为：O(n ^ 2)。具体的证明请阅读算法导论，下同。

## 堆排序

堆排序是借助堆的特殊性质，因为我们要输出正序排列，也就是从小到大输出。因此，我们采用小堆。小堆有这样的性质：它的每个
子节点，都会比它大。而对于它的子节点，也同样遵守这样的规定。在此，我们采用二叉堆，即，每个节点至多有两个子节点。我们
需要实现两个重要的操作：

- 把指定的某个值堆化

- 把堆顶pop，并且使得堆的性质仍然成立

通过第一个操作，我们可以把一个数组改造成一个堆。此外，第二个操作也是建立在第一个操作的基础之上，我们首先把堆定弹掉，
然后依次向下，把较大的子节点拉上来，然后继续继续继续。到了头之后，我们把堆里的最后一个元素填到上一步产生的空洞里，
然后对这个值进行堆化。

我们来看看代码（需要提到的是，假设某个节点index为i，则它的左边子节点的index总是2i + 1，右边的子节点是2i + 2，而父节点的index则为 int((i - 1)/2)）：

```python
def merge(c, i):
    """c is container, i is index to heapify"""
    while i > 0:
        parent_index = int((i - 1) / 2)
        if c[parent_index] < c[i]:
            c[i], c[parent_index] = c[parent_index], c[i]
        i = parent_index
```

然后我们看看怎么借助这个操作构成构建堆的操作：

```python
def insert(c, v):
    c.append(v)
    merge(c, len(c) - 1)
```

我们接下来看如何弹掉堆顶并且继续维持堆的性质：

```python
def extract(c):
    """pop the upper one and re-heapify"""
    upper = c[0]
    length = len(c)

    # find the bigger one in child
    i = 0
    left = 2 * i + 1
    right = 2 * i + 2
    while left < length and right < length:
        if c[left] > c[right]:
            c[i] = c[left]
            i = left
        else:
            c[i] = c[right]
            i = right

        left = 2 * i + 1
        right = 2 * i + 2

    # so in here, i is the place we need to re-insert
    c[i] = c.pop()
    merge(c, i)

    return upper
```

我们试试效果：

```python
if __name__ == "__main__":
    c = []
    [insert(c, i) for i in range(10)]
    print(c)
    extract(c)
    print(c)
```

输出：

```python
$ python heap.py 
[9, 8, 5, 6, 7, 1, 4, 0, 3, 2]
extract: 9
[8, 7, 5, 6, 2, 1, 4, 0, 3]
```

当然，这只是一个简单地堆的示范。在Python中利用堆来进行排序则很简单，因为有标准库：

```python
In [1]: import heapq

In [2]: array = [i for i in range(9, -1, -1)]

In [3]: array
Out[3]: [9, 8, 7, 6, 5, 4, 3, 2, 1, 0]

In [4]: heapq.heapify(array)

In [5]: array
Out[5]: [0, 1, 3, 2, 5, 4, 7, 9, 6, 8]

In [6]: heapq.nsmallest(len(array), array)
Out[6]: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
```

堆排序的最佳，平均和最差时间复杂度都为 `O(nlog(n))`。

## 快速排序

快排。非常经典，优美的一个排序算法，我们首先介绍朴素算法，然后介绍怎么进行原地快排，然后引入随机化算法。

快排的核心思想是，选一个数字，然后把小于等于它的数字放左边，大于它的放右边。然后分别对左边和右边进行同样的操作。
我们用最简单的Python代码来表示：

```python
def qsort(array):
    length = len(array)

    if length == 0:
        return []

    if length == 1:
        return array

    pivot = array[0]
    left = qsort(list(filter(lambda i: i <= pivot, array[1:])))
    right = qsort(list(filter(lambda i: i > pivot, array[1:])))

    return left + [pivot] + right


if __name__ == "__main__":
    print(qsort([i for i in range(9, -1, -1)]))
```

当然，这个实现很浪费空间。那怎么进行原地快排呢？首先，我们需要一个pivot，就选定array[0]好了。然后我们用两个下标i和j，
一个下标i用来记录上一个比pivot小的位置；另一个下标j用来记录下一个要遍历的数字。我们的算法是这样的：从左往右遍历
`array[1:]`，因为 `array[0]`已经被选为了pivot，如果发现当前指向的那个数字比pivot小，或者等于，我们就把 i++ 所在
的数字和j所在的数字交换。这样下来，遍历完成之后，i左边的数字就都小于或者等于pivot，右边则大于。最后我们还需要一步，
就是把pivot和i所在的值交换。

来看看代码：

```python
def qsort(nums, left=None, right=None):
    length = len(nums)

    if left is None:
        left = 0

    if right is None:
        right = length

    if left == right:
        return

    pivot_index = partition(nums, left, right)
    qsort(nums, left, pivot_index)
    qsort(nums, pivot_index + 1, right)


def partition(nums, left, right):
    p = nums[left]
    i, j = left, left + 1

    for j in range(left + 1, right):
        if nums[j] <= p:
            i += 1
            nums[i], nums[j] = nums[j], nums[i]  # exchange these two elems

    nums[i], nums[left] = nums[left], nums[i]

    return i


if __name__ == "__main__":
    nums = [i for i in range(9, -1, -1)]
    qsort(nums)
    print(nums)
```

看运行结果：

```bash
$ python qsort.py 
[0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
```

而随机化快排则是在上述代码中的 `partition` 中加入随机化的步骤：生成一个随机数，然后把随机数所对应的下标和left所在的下标
所在的数字互换。其他的则一样。

快排的最佳和平均时间复杂度为 `O(n log(n))`，最差时间复杂度为 `O(n ^ 2)`。

## 基数排序

我们先来看看这样一种排序方式（我们假设参与排序的数字/字符串都是等长，为了方便取值，这一次我们对字符串进行排序）：
假设我们有一串字符串组 `array = ["zzwt", "uios", "abcd", "efgh", "xyrz"]`，我们首先对每一个字符串的
第一个字母进行排序，然后对第二个字母进行排序，然后对第三个。。。依次类推。很显然，最终我们是可以得到有序字符串组的。
而问题在于，每一次我们都需要记好哪几个字符串是有相同字母的，我们需要借助外部存储来记住。那有没有方法可以不用呢？有。

如果我们从后面往前面排序，保持这样一个规则：如果已经是升序，那么我们就不打乱它。只要这样从右往左过一遍，最终我们就可以
得到一个 有序的数组。参见代码：

```python
def insertion(array, index):
    length = len(array)

    if length == 0:
        raise ValueError("请不要在空数组上使用排序")

    if length == 1:
        return array

    for i in range(1, length):
        for cursor in range(i, 0, -1):
            if array[cursor][index] < array[cursor - 1][index]:
                array[cursor], array[cursor - 1] = array[cursor - 1], array[cursor]


def radix(array):
    for i in range(len(array[0]) - 1, -1, -1):
        insertion(array, i)


if __name__ == "__main__":
    array = ["zzwt", "uios", "abcd", "efgh", "xyrz"]
    radix(array)
    print(array)
```

仔细想想便知道为什么这样可以：因为我们保证了只要目前已经有序，我们就不会打乱它。所以每当我们从右往左排完一位之后，这一部分
已经有序的字符串组有序的性质会带到下一次排序中，当我们完成整个循环之后，所有的数字都遵循这个性质。

基数排序的最佳，平均和最差时间复杂度都是 `O(nk)`。

## 补充

### 选择排序

选择排序使用这样一种思想：从左往右开始遍历，假设当前正在遍历i，从i处到数组尾部，找一个最小的值，和i处交换。循环下来之后，
数组便是有序的。很容易理解：当最开始的时候，i = 0，也就是，我们从 `array[0:]` 中选择一个最小的值，和0处交换，依次类推。

看代码：

```python
def select_sort(array):
    length = len(array)

    for i in range(length):
        smallest = i
        for j in range(i, length):
            if array[j] < array[smallest]:
                smallest = j
        array[i], array[smallest] = array[smallest], array[i]


if __name__ == "__main__":
    array = [i for i in range(9, -1, -1)]
    select_sort(array)
    print(array)
```

选择排序的最佳，平均和最差时间复杂度都是 `O(n ^ 2)`。

### 归并排序

归并排序是很经典的分治思维。我想把这个留到我总结分治的时候再讲。此处略过。参考：https://en.wikipedia.org/wiki/Merge_sort

### TimSort

TimSort是结合了归并排序和插入排序的排序算法，目前在很多标准库中都使用它。不过我还没看完实现，所以在这里知识引申一下。
参考：https://en.wikipedia.org/wiki/Timsort
