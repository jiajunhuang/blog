# 一步一步学Haskell（三）：Haskell语法

> 这一系列是我学习 `Learn You a Haskell For Great Good` 之后，总结，编写的学习笔记。

这个系列主要分为五个部分：

- [从递归说起](./2017_09_11-learn_you_a_haskell_part_1.md.html)
- [从命令式语言进行抽象](./2017_09_17-learn_you_a_haskell_part_2.md.html)
- [Haskell初步：语法](#)
- [Haskell进阶：Moniod, Applicative, Monad](#)
- [实战：Haskell和JSON](#)

## 简介

但凡谈到Haskell，都会有这么一句话，`Haskell is an advanced, purely functional programming language.`，这是其 [官网](https://www.haskell.org/) 所标榜的。

至于什么是纯(pure)，什么是函数式，什么是高等/高阶。我们将会在之后的文章里一一看到。

## 安装GHC

如同Go需要编译器，Python需要解释器一样，Haskell也许要一个程序让它能跑起来，以前很多教学都用 [hugs](https://www.haskell.org/hugs/)，不过我们用 [GHC](https://www.haskell.org/ghc/)。

那么如何安装呢？我假设本文的读者都是 Linux 用户。

- ArchLinux: `sudo pacman -S ghc`
- Ubuntu/Debian: `sudo apt-get install ghc`

找准自己的发行版执行对应的命令，就可以了。安装好之后，在命令行中输入 `ghci` 会进入类似以下的界面：

```bash
ghci
GHCi, version 8.0.2: http://www.haskell.org/ghc/  :? for help
Prelude> 
```

我们常用的几个命令就是：

- ghci: 交互式的Haskell解释器，类似于 python 命令
- runhaskell: 解释性的执行Haskell程序
- ghc: 编译Haskell程序

## 第一个Haskell程序

我们先来看一个Haskell程序，蹭蹭脸熟。

```haskell
sayHello :: String -> String
sayHello = (++) "Hello! "

main = do
    fmap sayHello getLine >>= putStrLn
    main
```

保存成 `Hello.hs`，然后执行一下：

```bash
$ runhaskell Hello.hs # 或者ghc -dynamic Hello.hs 然后 ./Hello
world
Hello! world
Jhon
Hello! Jhon
Hello.hs: <stdin>: hGetLine: end of file
```

嗯，看样子我们可以看到：

- Haskell也是从main开始切入
- Haskell里有很多奇奇怪怪的符号，比如这里的 `>>=`

对于第二点，不用担心，以后还会看到更多的 :)

## 类型系统

### 普通的类型

看到了上面示例里的第一行吗？ `sayHello :: String -> String`，Haskell中有很多类似的代码，叫做 [类型签名](https://wiki.haskell.org/Type_signature)。

等我们对Haskell的类型系统稍微熟悉之后，我们便可以获得这样一种能力：根据类型签名就可以推断出这个函数大概是做什么用的，再加上比较好的函数命名，我们就可以不看实现便知道函数的作用。

不过在此之前我们需要先熟悉一些常见的类型：

- Char 字符，例如 `'a'`
- [] 列表，例如 `[1, 2, 3]`
- String 字符串，也是[Char]的别名，例如 `Hello` 其实是 `H:e:l:l:o:[]`
- Int 整数，通常是32位，所以取值范围是 -2147483648 ~ 2147483647
- Bool 布尔值，例如 `True`和`False`
- Integer 整数，没有取值范围，其范围只取决于内存大小
- Float 浮点数
- Double 双精度浮点数

如果我们不确定一个东西是什么类型怎么办呢？打开 `ghci`，然后输入 `:t`，ghci便会告诉我们：

```bash
Prelude> :t "Hello"
"Hello" :: [Char]
Prelude> :t 1
1 :: Num t => t
Prelude> :t 1.3
1.3 :: Fractional t => t
Prelude> :t 1.3 :: Double
1.3 :: Double :: Double
Prelude> :t True
True :: Bool
```

通常ghci都会给我们一个比较宽泛的类型，但是我们可以通过加入 `:: <类型>` 来指定一个更加具体的类型。

### 函数的类型

函数也有类型？对啊，没毛病，例如 Golang 里我们也要把函数的签名写出来，Haskell也是如此，但是我们可以省略，因为Haskell有一个
强大的功能，叫做 [类型推断](https://zh.wikipedia.org/wiki/%E7%B1%BB%E5%9E%8B%E6%8E%A8%E8%AE%BA)。

我们来看看最开始我们的函数类型 `sayHello :: String -> String`，其中 `->` 意思是给一个 String，得到一个 String。就算是 `sayHelloTo :: String -> String -> String` 我们也可以这样读，尽管Haskell中所有的函数都只有一个参数，但是为了简便，我们知道就好。

## 类型的共同特征-类型的类型？TypeClass

在 [第二篇](https://jiajunhuang.com/articles/2017_09_17-learn_you_a_haskell_part_2.md.html) 中我们知道了什么是抽象。那么我们想想，Haskell的类型中是否也能找出共同点呢？是否也能抽象出一系列的接口呢？是否也有类似接口的概念呢？

有。就是我们要讲的TypeClass。例如，`Char` 和 `Int` 都是可以比较是否相等的，那么我们可以抽象一个接口叫做 `Equal`，例如 `Int` 和 `Bool` 都是有取值范围限制的，我们可以抽象一个接口叫做 `Bounded`。

不过Haskell中的接口不叫 `interface`，而是叫做 `class`。没错，叫做`class`，在此需要重申一遍，以防止大家和面向对象编程中的 `class` 关键字搞混。

我们看看 `Equal` 在Haskell中是怎么定义的。

```bash
Prelude> :i Eq
class Eq a where
  (==) :: a -> a -> Bool
  (/=) :: a -> a -> Bool
  {-# MINIMAL (==) | (/=) #-}
  	-- Defined in ‘GHC.Classes’
instance Eq a => Eq [a] -- Defined in ‘GHC.Classes’
instance Eq Word -- Defined in ‘GHC.Classes’
instance Eq Ordering -- Defined in ‘GHC.Classes’
instance Eq Int -- Defined in ‘GHC.Classes’
instance Eq Float -- Defined in ‘GHC.Classes’
instance Eq Double -- Defined in ‘GHC.Classes’
instance Eq Char -- Defined in ‘GHC.Classes’
instance Eq Bool -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e, Eq f, Eq g, Eq h, Eq i,
          Eq j, Eq k, Eq l, Eq m, Eq n, Eq o) =>
         Eq (a, b, c, d, e, f, g, h, i, j, k, l, m, n, o)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e, Eq f, Eq g, Eq h, Eq i,
          Eq j, Eq k, Eq l, Eq m, Eq n) =>
         Eq (a, b, c, d, e, f, g, h, i, j, k, l, m, n)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e, Eq f, Eq g, Eq h, Eq i,
          Eq j, Eq k, Eq l, Eq m) =>
         Eq (a, b, c, d, e, f, g, h, i, j, k, l, m)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e, Eq f, Eq g, Eq h, Eq i,
          Eq j, Eq k, Eq l) =>
         Eq (a, b, c, d, e, f, g, h, i, j, k, l)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e, Eq f, Eq g, Eq h, Eq i,
          Eq j, Eq k) =>
         Eq (a, b, c, d, e, f, g, h, i, j, k)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e, Eq f, Eq g, Eq h, Eq i,
          Eq j) =>
         Eq (a, b, c, d, e, f, g, h, i, j)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e, Eq f, Eq g, Eq h, Eq i) =>
         Eq (a, b, c, d, e, f, g, h, i)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e, Eq f, Eq g, Eq h) =>
         Eq (a, b, c, d, e, f, g, h)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e, Eq f, Eq g) =>
         Eq (a, b, c, d, e, f, g)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e, Eq f) =>
         Eq (a, b, c, d, e, f)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d, Eq e) => Eq (a, b, c, d, e)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c, Eq d) => Eq (a, b, c, d)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b, Eq c) => Eq (a, b, c)
  -- Defined in ‘GHC.Classes’
instance (Eq a, Eq b) => Eq (a, b) -- Defined in ‘GHC.Classes’
instance Eq () -- Defined in ‘GHC.Classes’
instance (Eq b, Eq a) => Eq (Either a b)
  -- Defined in ‘Data.Either’
instance Eq Integer
  -- Defined in ‘integer-gmp-1.0.0.1:GHC.Integer.Type’
instance Eq a => Eq (Maybe a) -- Defined in ‘GHC.Base’
```

wow，原来在 `ghci` 中输入 `:i` 就可以知道了，应该还有很多其他功能，我们输入 `:help`，然后仔细看看，绝对会受益匪浅。

我们把 `Eq` 的定义抽出来看：

```haskell
class Eq a where
  (==) :: a -> a -> Bool
  (/=) :: a -> a -> Bool
```

a？a是什么？在Haskell的定义中，a常常用来代表不管是什么类型，也可以通过加类型限制的方式指定具体某种类型，为什么用a呢？其实用b，用c都可以，如果a不够表达，就可以用abcdefg等等等等了，只不过默认的约定俗成的就是a罢了。

我们来看看这里 `(==) :: a -> a -> Bool`，`(==)` 取任意类型，再取任意类型，然后返回一个Bool。没错，这就是是否相等的定义。

那我们怎么知道某个具体类型是否实现了这个接口呢？事实上，为了实现这个接口，我们需要在具体类型上这样写：`instance Eq Int where...`。这优点类似Java中的 `implemented` 关键字。但是对于这一点，我们暂时先按下不表。

我们先继续以浏览更多东西，增加广度为先。接下来我们看看长间的数据结构。

## List, Tuple

啊，学过Python的朋友们都知道List和Tuple的区别，在Haskell中也是一样，
只不过List中只能放同一种数据类型的数据，说起来有点像Golang中的slice。而Tuple则更像是Python中的Tuple了，长度固定，可以放多种类型在其中。

List用 `[]`表示而 Tuple用 `()`表示。

```bash
Prelude> :t [1, 2, 3]
[1, 2, 3] :: Num t => [t]
Prelude> :t (1, 2, 3)
(1, 2, 3) :: (Num t, Num t1, Num t2) => (t2, t1, t)
```

Haskell中的List有语法糖叫做List Comprehension，其实Python中也有，看看Python的：

```python
>>> [i for i in range(10)]
[0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
>>> [i + 1 for i in range(10)]
[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
>>> [i + j for i in range(10) for j in range(2)]
[0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10]
>>> [i + j for i in range(10) for j in range(2) if i % 2 == 0]
[0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
```

而Haskell版本的：

```haskell
Prelude> [i | i <- [0..9]]
[0,1,2,3,4,5,6,7,8,9]
Prelude> [i + 1 | i <- [0..9]]
[1,2,3,4,5,6,7,8,9,10]
Prelude> [i + j | i <- [0..9], j <- [0..1]]
[0,1,1,2,2,3,3,4,4,5,5,6,6,7,7,8,8,9,9,10]
Prelude> [i + j | i <- [0..9], j <- [0..1], i `mod` 2 == 0]
[0,1,2,3,4,5,6,7,8,9]
```

是不是感觉很像？

## 模式匹配，guards

## if, else, let, where

## Haskell的pure所在和如何递归的写程序

## 高阶

## 模块

## 自己创建类型和TypeClass

## 总结
