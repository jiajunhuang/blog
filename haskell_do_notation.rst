:Date: 06/04/2016

Haskell do notation
======================

从普通的语言切换过来的时候，看到return总是让人费解。因为Haskell中的return
根本就不是普通语言中的return。这篇文章是看完Haskell Wiki [#]_ 以后写的一
篇笔记，当然主要是自己的理解，下面的代码部分来自wiki，部分属于自己捏造。

do-notation
------------

通常我们学习Haskell都是看 ``Learn You a Haskell for Great Good`` [#]_ 。这篇
教程在第12章第一次引入了Haskell的do-notation，他说，有这样一个函数:

.. code:: haskell

    foo :: Maybe String
    foo = Just 3   >>= (\x ->
        Just "!" >>= (\y ->
        Just (show x ++ y)))

我们可以用do-notation写成下面这样以减少lambda函数的使用:

.. code:: haskell

    foo :: Maybe String
    foo = do
        x <- Just 3
        y <- Just "!"
        Just (show x ++ y)

Haskell Report 2010 [#]_ 上是这么说的: "A do expression provides a more
conventional syntax for monadic programming." do作为一种语法糖，让
Haskell写起来更方便。

``then`` operator
~~~~~~~~~~~~~~~~~~~

为了完全理解do，我们首先要把它变回去。举个例子:

.. code:: haskell

    manyPutStrs = do
        putStr "Hello"
        putStr " "
        putStr "world!"
        putStr "\n"

其类型为 ``manyPutStrs :: IO ()`` 。它其实相当于:

.. code:: haskell

    manyPutStrs = putStr "Hello" >> putStr " " >> putStr "world!" >> putStr "\n"

我们知道 ``>>`` 的类型为: ``(>>) :: Monad m => m a -> m b -> m b`` ，即，
执行 ``m a`` ，但是丢弃其结果，并且执行 ``m b`` ，最终结果的类型就是 ``m b``
的类型。 ``manyPutStrs`` 的类型就是 ``m b`` 也就是:

.. code:: bash

    Prelude> :t putStr "!"
    putStr "!" :: IO ()

``bind`` operator
~~~~~~~~~~~~~~~~~~

只进行输出的程序好像没有太大的用处，那么我们来举个能够和用户交互的例子:

.. code:: haskell

    helloBody = do
        putStrLn "what's your name?"
        name <- getLine
        putStrLn $ "Hello " ++ name

首先我们来看一下这个函数的类型:

.. code:: bash

    Prelude> :l fun.hs
    [1 of 1] Compiling Main             ( fun.hs, interpreted )
    Ok, modules loaded: Main.
    Main> :t helloBody
    helloBody :: IO ()
    Main> helloBody
    what's your name?
    jhon
    Hello jhon

这个函数的最终类型为 ``helloBody :: IO ()`` 的原因也是因为最后一条语句的类型
为 ``putStrLn "hello" :: IO ()`` 。但是对于 ``name <- getLine`` 好像并不是
和上面的那个例子那样简单，首先来看一下 ``getLine`` 的类型：

.. code:: bash

    Prelude> :t getLine
    getLine :: IO String

而 ``<-`` 的作用就是把 ``String`` 从 ``IO String`` 中取出来并且给 ``name`` 绑上。
而且 ``name`` 在后面还用上了。其实上面的函数就相当于:

.. code:: haskell

    helloBody'' = putStrLn "what's your name?" >>
                getLine >>= (\name ->
                             putStrLn $ "Hello " ++ name
                            )

来看看他的类型和效果是否一样:

.. code:: bash

    Prelude> :l fun.hs
    [1 of 1] Compiling Main             ( fun.hs, interpreted )
    Ok, modules loaded: Main.
    Main> :t helloBody''
    helloBody'' :: IO ()
    Main> helloBody''
    what's your name?
    jhon
    Hello jhon

其中 ``>>=`` 的类型为 ``(>>=) :: Monad m => m a -> (a -> m b) -> m b`` 。理解
了这个bind操作符理解上面的代码也就没问题了。

return
----------

首先来看一下 ``return`` 的类型: ``return :: Monad m => a -> m a`` 其实在Haskell
中，return的作用就是将数据塞到一个盒子里，这里所说的盒子也就是我们的Monad。
我们来举个例子：

.. code:: haskell

    foo = do
        return "hi"
        putStrLn "foo"

foo的类型为 ``foo :: IO ()`` ，这是因为如上面我们所说的，这相当于::

    foo = return "hi" >> putStrLn "foo"

.. [#] https://en.wikibooks.org/wiki/Haskell/do_notation
.. [#] http://learnyouahaskell.com/chapters
.. [#] https://www.haskell.org/onlinereport/haskell2010/haskellch3.html#x8-470003.14
