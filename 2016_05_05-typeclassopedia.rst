Haskell TypeClass 笔记
========================

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

Functor Laws
~~~~~~~~~~~~~

- fmap id = id

- fmap (g . h) = (fmap g) . (fmap h)

这两条在一起保证了fmap g不会改变容器而只改变其中的内容。
`其中第一条是第二条的充分不必要条件 <https://github.com/quchen/articles/blob/master/second_functor_law.md>`__ 。

Applicative
-------------

Applicative 介于Functor和Monad之间，对于Functor，fmap可以
把 *普通函数* 作用于容器内的内容。而Applicative为我们提供了
``<*>`` 来一个 *在容器内的函数* 作用于容器内的内容
(``<*>`` 读作"apply")，另外还有一个函数，``pure`` 可以把
参数加入到容器内。看Applicative的定义：

.. code:: haskell

    class Functor f => Applicative f where
        pure :: a -> f a
        infixl 4 <*>
        (<*>) :: f (a -> b) -> f a -> f b

如上，Applicative必定属于Functor。这里的f也是一个类型构造子
(type constructor)。

.. code:: haskell

    data MyMaybe a = MyNothing | MyJust a deriving (Show, Eq)

    instance Functor MyMaybe where
        fmap g MyNothing = MyNothing
        fmap g (MyJust a) = MyJust (g a)

    instance Applicative MyMaybe where
        pure = MyJust
        MyNothing <*> _ = MyNothing
        (MyJust f) <*> box = fmap f box

.. code:: bash

    [root@arch haskell]# ghci
    GHCi, version 7.10.3: http://www.haskell.org/ghc/  :? for help
    Prelude> :l fun.hs
    [1 of 1] Compiling Main             ( fun.hs, interpreted )
    Ok, modules loaded: Main.
    Main> pure (+) <*> MyJust 1 <*> MyJust 2
    MyJust 3

Applicative Laws
~~~~~~~~~~~~~~~~~

- The identity law::

    pure id <*> v = v

- Homomorphism(同态性)::

    pure f <*> pure x = pure (f x)

- Interchange(交换性)::

    u <*> pure y = pure ($ y) <*> u

- Composition(组合)::

    u <*> (v <*> w) = pure (.) <*> u <*> v <*> w

另外，在 ``Control.Applicative`` 中定义了 ``<$>`` ，相当于
fmap::

    g <$> x = pure g <*> x

It says that mapping a pure function g over a context x
is the same as first injecting g into a context with pure,
and then applying it to x with (<*>).

Monad
-------

首先来看Monad的定义:

.. code:: haskell

    class Applicative m => Monad m where
        return :: a -> m a
        (>>=) :: m a -> (a -> m b) -> m b
        (>>) :: m a -> m b -> m b
        m >> n = m >>= \_ -> n

        fail :: String -> m a

其中的return就是pure，从其他编程语言过来的人一定要注意不要混淆。
``>>`` 是 ``>>=`` 的一种特殊情况，看上面的默认实现就知道了，另外，
``m >> n`` ignores the result of m, but not its effects.

    在这里我们可以对比一下Functor，Applicative，Monad三者，继续以
    容器为例，fmap的类型为 ``fmap :: (a -> b) -> f a -> f b`` 即把
    一个容器内的内容作用于普通函数；而 ``<*>`` 的类型为
    ``(<*>) :: Applicative f => f (a -> b) -> f a -> f b`` 即把一个
    在容器内的内容作用于在容器内的函数；而 ``>>=`` 的类型为
    ``(>>=) :: Monad m => m a -> (a -> m b) -> m b`` 即把一个容器内的内容
    作用于一个接受普通参数但产生容器类型的函数。Haskell的好处便在这里有了
    体现，我们可以直接通过看函数的类型签名来判断函数的作用，却不需要查看
    函数的具体实现。

接下来把上面自己定义的MyMaybe实现为Monad的实例:

.. code:: haskell

    data MyMaybe a = MyNothing | MyJust a deriving (Show, Eq)

    instance Functor MyMaybe where
        fmap g MyNothing = MyNothing
        fmap g (MyJust a) = MyJust (g a)

    instance Applicative MyMaybe where
        pure = MyJust
        MyNothing <*> _ = MyNothing
        (MyJust f) <*> box = fmap f box

    instance Monad MyMaybe where
        return a = MyJust a

        MyNothing >>= _ = MyNothing
        (MyJust a) >>= f = f a

.. code:: bash

    Main> MyJust 1 >>= (\p -> (MyJust (10+p)))
    MyJust 11
    Main> MyNothing >>= (\p -> (MyJust (10+p)))
    MyNothing
    Main> MyNothing >> MyJust 1 >>= (\p -> (MyJust (10+p)))
    MyNothing

more about >>=
~~~~~~~~~~~~~~~~

``>>=`` 通常读作 "bind"。The basic intuition is that it combines two
computations into one larger computation. In other words, x >>= k is
a computation which runs x, and then uses the result(s) of x to decide
what computation to run second, using the output of the second
computation as the result of the entire computation.

详见原文5.3节，其中有包括用 fmap, pure, <*> 实现 >>= 的介绍。

Monad Laws
~~~~~~~~~~~~

- ``return a >>= k = k a``

- ``m >>= return = m``

- ``m >>= (\k -> k x >> h) = (m >>= k) >>= h``

do notation
~~~~~~~~~~~~~

do 是Haskell中提供的一种语法糖，让程序看起来更像是命令式编程风格。
原文中介绍了do的缩进或者括号风格，还是直接看 `Haskell Report 2010中的相关介绍吧。 <https://www.haskell.org/onlinereport/haskell2010/haskellch10.html#x17-17800010.3>`__
