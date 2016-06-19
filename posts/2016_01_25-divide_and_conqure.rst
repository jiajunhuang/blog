分治
======

Divide and Conquer. 在之前读 `Eric Roberts` 先生所著的
`Thinking Recursively` 之后写过几句小总结,那本小书介绍了如何以递归的角度去
看一些数据结构和算法,比如树和汉诺塔问题. 其中的三点, 也就是现在这篇文章需要
总结的, 分治.

概念
-----

`算法导论`_ 上是这么说用分治算法解决一个问题的:

- Divide the problem into a number of subproblems that are smaller instances
  of the same problem.

- Conquer the subproblems by solving them recursively. If the subproblem
  sizes are small enough, however, just solve the subproblems in a
  straightforward manner.

- Combine the solutions to the subproblems into the solution for the
  original problem.

准确的来说,分治是一种思想,而不是某一个具体的算法,只要符合上述思想,先找到形式
相同的子问题,然后依次解决子问题,再把子问题合起来,得到最终的结果的算法,都是分
治.

实例
-----

举个简单的例子,快排:

.. code:: python

    def qsort(alist):
        length = len(alist)
        if length <= 1:
            return alist

        mid = length // 2
        less = list(filter(lambda x: x < alist[mid], alist))
        more = list(filter(lambda x: x > alist[mid], alist))
        return qsort(less) + [alist[mid]] + qsort(more)

首先我们找到一个pivot, 把小于他的放到左边, 把大于他的放到右边, 并且分别对左
边和右边递归进行这个操作,然后把结果合起来,返回结果.

证明
-----

证明递归算法的O复杂度方法有三种:

- substitution method.

    先猜测算法复杂度,再用数学证明.

- recursion-tree method.

    画出递归树,然后把每层复杂度相加.

- master method.

    根据公式计算(公式见 `算法导论`_ ).

实例
-----

接下来我们拿最大子序列问题来练练手.

最大子序列问题: 有一串数字,找出其中和为最大的那一段.

- 对于这个问题,有暴力解法, 先for循环从左到右,再在for的里面从右到左for一遍,记住
  出现最大子序列的sum, left_index, right_index. 暴力解法的O一般都不低,这个例子
  是O(n^2).

- 分治解法,想到这个解法的关键点在于,最大子序列要么在中点(中间的那个值的index)
  的左边,要么在右边,也有可能横跨中点,是左右加起来.但即使是在左边或者右边,也是
  在左边一部分的中点上(右边同理).所以这里的关键是建立 `递归信任`_ .这个解法的
  O为O(nlgn). `O(nlgn)解法`_

- 是不是没有比O(nlgn)更低的算法了呢?不是,对于这个问题,还有更快的算法,就是先从
  左向右,累加,找到最大值的sum和right_index,然后再从左向右减,看是不是能找到更大
  的sum,并记录left_index.这个算法最多只要扫描两遍,算法复杂度为O(n). `O(n)解法`_

.. _`算法导论`: https://mitpress.mit.edu/books/introduction-algorithms
.. _`O(nlgn)解法`: https://github.com/jiajunhuang/intro_to_algorithms/blob/master/chap4/max_subarray/maxsub.c
.. _`O(n)解法`: https://github.com/jiajunhuang/intro_to_algorithms/blob/master/chap4/max_subarray/maxsub_linear.c
