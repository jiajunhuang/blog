Haskell TypeClass 笔记
============================

这是Haskell Wiki上的 `一篇文章 <http://wiki.haskell.org/Typeclassopedia>`_ ，以下是我的学习笔记。

学好Haskell有两个重点:

    - Understand the types

    - Gain a deep intuition for each type class and its
      relationship to other type classes, backed up by
      familiarity with many examples.

Haskell的类型系统如下图所示：

.. image:: https://wiki.haskell.org/wikiupload/d/df/Typeclassopedia-diagram.png

其中：

    - 实线箭头，如图中Functor指向Applicative的箭头，表示
      Applicative包含于Functor

    - 虚线箭头，如图中Applicative指向Traversable，表示
      其他的某种关系

    - Monad和ArrowApply是相等的

    - Semigroup, Apply, Comonad 尚未包含在标准库中

另外，不要被Haskell中的class关键字迷惑了，其实不是类，更像是
Java中的Interface。

Functor
--------

Functor的定义如下：

.. code:: haskell

    class Functor f where
        fmap :: (a -> b) -> f a -> f b

个人以为，用盒子来描述Functor非常合适，符合Functor的实例
都能进行fmap，而这个fmap能改变盒子里的内容但是不改变盒子的
结构，例如Maybe是Functor：

.. code:: haskell

    data MyMaybe a = MyNothing | MyJust a deriving (Show, Eq)

    instance Functor MyMaybe where
        fmap g MyNothing = MyNothing
        fmap g (MyJust a) = MyJust (g a)

.. code:: bash

    # ghci
    Prelude> :l fun.hs
    Main> fmap (+1) (MyJust 1)
    MyJust 2

现在我们回到fmap上来，其类型为::

    fmap :: (a -> b) -> f a -> f b

其中 ``f a`` 整体作为一个类型，那么 ``f`` 便是一个
类型构造子(type constructor)。也就是说，作为Functor
的instance，必须是接受且仅接受一个参数。

常见的instance有 ``Either e`` , ``((,) e)`` , `` ((->) e)`` , IO , 还有许多容器类型如 Tree, Map等。

Functor Law
~~~~~~~~~~~~~

- fmap id = id

- fmap (g . h) = (fmap g) . (fmap h)

这两条在一起保证了fmap g不会改变容器而只改变其中的内容。
`其中第一条是第二条的充分不必要条件 <https://github.com/quchen/articles/blob/master/second_functor_law.md>`__
