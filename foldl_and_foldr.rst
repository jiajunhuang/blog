
foldl 和 foldr 的变换
========================

foldl 和 foldr 的类型签名为:

.. code:: haskell

    Prelude> :t foldl
    foldl :: Foldable t => (b -> a -> b) -> b -> t a -> b
    Prelude> :t foldr
    foldr :: Foldable t => (a -> b -> b) -> b -> t a -> b

可以看出来他们接受的参数为：一个函数，一个initial value，一个列表，
然后把整个列表变成和列表类型相同的value。举个例子:

.. code:: haskell

    foldl (+) 0 [1..5]
    -- 展开成
    = ((((1 + 2) + 3) + 4) + 5)
    = 15

    foldr (+) 0 [1..5]
    -- 展开成
    = (1 + (2 + (3 + (4 + 5))))
    = 15

从结果来看foldl和foldr似乎可以相互替换，但并不是这样的，上例结果相同只是因为
加法(+)满足结合律，同样是这么几个数，从左往右加和从右往左加结果是一样的。
但对于减法就不是这样了，我们继续看代码:

.. code:: haskell

    Prelude> foldl (-) 0 [1..5]
    -15
    Prelude> foldr (-) 0 [1..5]
    3

那么foldl和foldr长得那么像，能不能互相实现呢？我们来用foldl实现foldr，
首先我们从类型签名看起。

- ``foldl :: Foldable t => (b -> a -> b) -> b -> t a -> b``

- ``foldr :: Foldable t => (a -> b -> b) -> b -> t a -> b``

可以看出来，首先接受的第一个函数的参数是相反的，所以对于
``foldr func init alist`` 首先要把函数的参数反过来，所以我们用了 flip,
``foldl (flip func) init alist`` (flip 的类型为 ``flip :: (a -> b -> c) -> b -> a -> c`` )。
但是这还不够，foldl是从左边向右两个两个处理，而foldr是从右向左，我们需要将
alist反转，所以我们用了 reverse，他的类型是 ``reverse :: [a] -> [a]`` :

.. code:: haskell

    foldr func init alist = foldl (flip func) init (reverse alist)
    -- 可以化简成：
    foldr func init = (foldl (flip func) init) . reverse

python中的foldl
-----------------

python中有reduce，作用相当于Haskell中的foldl:

.. code:: python

    >>> import operator
    >>> from functools import reduce
    >>> reduce(operator.add, list(range(1, 6)), 0)
    15

.. [#] https://wiki.haskell.org/Foldr_Foldl_Foldl'
