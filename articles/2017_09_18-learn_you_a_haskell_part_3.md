# Haskell简明教程（三）：Haskell语法

> 这一系列是我学习 `Learn You a Haskell For Great Good` 之后，总结，编写的学习笔记。

这个系列主要分为五个部分：

- [从递归说起](./2017_09_11-learn_you_a_haskell_part_1.md.html)
- [从命令式语言进行抽象](./2017_09_17-learn_you_a_haskell_part_2.md.html)
- [Haskell初步：语法](#)
- [Haskell进阶：Monoid, Applicative, Monad](./2017_09_25-learn_you_a_haskell_part_4.md.html)
- [实战：Haskell和JSON](./2017_09_26-learn_you_a_haskell_part_5.md.html)

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

## 模式匹配

我们在命令式语言中接触的最多的表示分支的方式，估计就是 `if ... else ...` 了。
只不过在不同的语言中，表述也不一样，表现形式也略有不同，例如 `if...then...`，
例如 `switch...case...`。其实就是决策树的分支表现形式。在Haskell中其实也有
这样的东西，我们先来看第一种，模式匹配，为了好好的和模式匹配玩耍，我们要先和
List玩耍。我们来看看List的几个常见的操作：`head`, `tail`。

先看看类型：

```haskell
Prelude> :t head
head :: [a] -> a
Prelude> :t tail
tail :: [a] -> [a]
```

我们就这样猜测吧：`head` 是取出List中的第一个元素，而 `tail` 是取出除了第一个元素
的其他所有剩余元素。

我们新建 `MyList.hs` 然后把类型填进去：

```haskell
head :: [a] -> a
tail :: [a] -> [a]
```

然后填上实现：

```haskell
head :: [a] -> a
head [] = 
head (x:_) = x

tail :: [a] -> [a]
tail [] = []
tail (_:xs) = xs
```

然后我们去ghci里执行一下：

```bash
Prelude> :l MyList.hs 
[1 of 1] Compiling Main             ( MyList.hs, interpreted )
Ok, modules loaded: Main.
*Main> head []

<interactive>:2:1: error:
    Ambiguous occurrence ‘head’
    It could refer to either ‘Prelude.head’,
                             imported from ‘Prelude’ at MyList.hs:1:1
                             (and originally defined in ‘GHC.List’)
                          or ‘Main.head’, defined at MyList.hs:2:1
*Main> head [1, 2, 3]

<interactive>:3:1: error:
    Ambiguous occurrence ‘head’
    It could refer to either ‘Prelude.head’,
                             imported from ‘Prelude’ at MyList.hs:1:1
                             (and originally defined in ‘GHC.List’)
                          or ‘Main.head’, defined at MyList.hs:2:1
*Main> 
```

嗯。。。看报错是说我们定义的head函数和标准库Prelude预先加载定义的名字冲突了，
如果写过Python我们可以参考看有没有类似 `import as` 的功能，所以我们搜一下：

http://lmgtfy.com/?q=haskell+import+as

不过我试了一下，在ghci中这么干并不行。所以只能用一种比较老土的方法，就是把函数
名字给改了。

```haskell
Prelude> :l MyList.hs 
[1 of 1] Compiling Main             ( MyList.hs, interpreted )
Ok, modules loaded: Main.
*Main> myHead []
*** Exception: bad operation on empty list
CallStack (from HasCallStack):
  error, called at MyList.hs:2:13 in main:Main
*Main> myHead [1, 2, 3, 4]
1
*Main> myTail [1, 2, 3, 4]
[2,3,4]
*Main> 
```

要注意Haskell和Golang中有一点比较相像，那就是首字符大小写有不同的意义。

模式匹配就是把某个参数，强行拆解，看能不能匹配上该行的拆解形式，如果能，那么就执行
后面的代码，要不然就跳到下一个模式里去。

## if, else, let, where, guard

在平时的编程中我们经常要干的事情就是做判断，如果那么，要不然就怎样。那么为什么
Haskell作为一门函数式编程不能脱离这些呢？说好的函数式编程和命令式编程不一样呢？
是啊，函数式编程应该是描述这是什么，而不是具体怎么做。其实这两者并不冲突，例如
模式匹配其实也是一种分支，不过模式匹配要更加强大。接下来我们看看Haskell中的分支
是怎么表达的。

```haskell
biggerThan100 :: Int -> Bool
biggerThan100 x = if x > 100 then True else False
```

和我们往日所写的其实差不多，所以就不赘述了。

接下来我们介绍一种新的写法，guard。有中文翻译成守卫表达式，不过我还是更喜欢直接
用英文。把上面的改成guard会是这样：

```haskell
biggerThan100' :: Int -> Bool
biggerThan100' x
    | x > 100 = True
    | otherwise = False
```

有两点需要注意，第一，最后一个参数后面是 `|` 而不是 `=`，第二，otherwise相当于
switch语句里的 `default`。

平时我们定义函数和变量都是在最外层定义，然后函数里引用，那么有没有更小的作用域呢？
有，我们看看 `let` 和 `where` 的示例。

```haskell
printHello :: String -> String
printHello x = let finalPrint = "Hello! " ++ x in finalPrint

printHello' :: String -> String
printHello' x = finalPrint
    where finalPrint = "Hello! " ++ x
```

执行一下：

```haskell
Prelude> :l LetWhere.hs 
[1 of 1] Compiling Main             ( LetWhere.hs, interpreted )
Ok, modules loaded: Main.
*Main> printHello "World"
"Hello! World"
*Main> printHello' "World"
"Hello! World"
*Main> 
```

## Haskell的pure所在和如何递归的写程序

有没有发现到目前为止我们都没有真正的简单的原始的写一个 `HelloWorld` 出来呢？而是
一直在写一些 `String -> String` 啊 `Char -> String -> Bool` 的函数呢？为什么Haskell中
这么多这样的函数（最少到目前我们接触的为止）？因为Haskell有一个很大的特点是pure。
什么是纯函数？就是无论在什么情况下，只要给定输入，那么输出一定是同样的。那什么是
不纯的函数？举个例子，网络IO，硬盘IO，标准输出也是。我们暂时可以这样想像：凡是
和现实世界接触的东西，都是不纯的，凡是可以抽象成数学理论可以解释的，都是纯的。

Haskell把不纯的东西也做了抽象，叫做Monad，你可以把它理解成一个盒子，就是我们以前
所说的黑盒子，它把不纯的东西包在里面，并且提供一些接口来操作它。这些我们会在下一篇
看到。

接下来我们将要看看如何递归的写程序，和第一篇一样，我们简单地来看看如何递归的遍历
列表。现在我们有一个列表 `[1, 3, 2, 5, 6]`，我们想要将偶数过滤掉，只留下奇数。
但是我们不能用for循环。列表，如果用递归的方式去看，就是一个表头+一个列表，拿
`[1, 3, 2, 5, 6]`来看，就是 `1` 和 `[3, 2, 5, 6]`。在Haskell中我们可以这样写：
`1:[3, 2, 5, 6]`，其中 `:` 是列表连接符，读作 `Cons`。我们可以看看它的类型：

```haskell
Prelude> :t (:)
(:) :: a -> [a] -> [a]
```

很显然，要过滤偶数，我们可以这样写：

```haskell
filterEven :: [Int] -> [Int]
filterEven [] = []
filterEven (x:xs) = if x `mod` 2 == 0 then filterEven xs else x:filterEven xs

filterEven' :: [Int] -> [Int]
filterEven' [] = []
filterEven' (x:xs)
    | x `mod` 2 == 0 = filterEven' xs
    | otherwise = x:filterEven' xs


-- 把判断是否是偶数抽离出来
isEven :: Int -> Bool
isEven x = x `mod` 2 == 0

filterEven'' :: [Int] -> [Int]
filterEven'' [] = []
filterEven'' (x:xs)
    | isEven x = filterEven'' xs
    | otherwise = x:filterEven'' xs
```

执行一下：

```haskell
Prelude> :l FilterEven.hs 
[1 of 1] Compiling Main             ( FilterEven.hs, interpreted )
Ok, modules loaded: Main.
*Main> filterEven [1, 3, 2, 5, 6]
[1,3,5]
*Main> filterEven' [1, 3, 2, 5, 6]
[1,3,5]
*Main> filterEven'' [1, 3, 2, 5, 6]
[1,3,5]
*Main>
```

我们其实就是顺着递归的思路去写代码，如何过滤偶数呢？就是我们把列表看成是一个
元素+一个列表的结构，我们每次都看当前元素是否是偶数，如果是，那么就忽略，直接
考虑下一个列表，要不然的话，我们要把现在这个元素追加到最前面，然后才开始考虑下一个
列表。

此外，在写示例的时候我犯了一个错误，就是忘记了写边界条件：

```haskell
filterEven :: [Int] -> [Int]
filterEven (x:xs) = if x `mod` 2 == 0 then filterEven xs else x:filterEven xs
```

结果运行就会报错：

```haskell
Prelude> :l FilterEven.hs 
[1 of 1] Compiling Main             ( FilterEven.hs, interpreted )
Ok, modules loaded: Main.
*Main> filterEven [1, 3, 2, 5, 6]
[1,3,5*** Exception: FilterEven.hs:2:1-77: Non-exhaustive patterns in function filterEven

*Main>
```

为什么呢？初看你可能觉得这是因为ghci就像python解释器一样，执行到了对应的报错然后才
处理。这么说的话，似乎也对，其实真正的原因是因为Haskell是一门惰性语言，英文的说法
叫做 `lazy`。`lazy`?

```python
def iter_array(array):
    for i in array:
        yield i
```

什么叫做lazy呢？就是延迟计算，在Python中可以是迭代器，也可以是重写 `__call__`来
造成延迟计算，或者类似的手法。其核心思想就是，并不是一开始就计算好，而是等到真正
要用的时候才去计算。

## [高阶函数](http://learnyouahaskell.com/higher-order-functions)

> Haskell functions can take functions as parameters and return functions as return values. A function that does either of those is called a higher order function. 

所谓的高阶函数其实就是能接受函数作为参数，或者返回一个函数作为结果的函数。举个例子，map函数，map的类型为：

```haskell
Prelude> :t map
map :: (a -> b) -> [a] -> [b]
```

map接受一个函数 `a->b`，然后接受一个列表 `[a]`，返回一个列表 `[b]`。我们可以看看示例：

```haskell
Prelude> map (+1) [1..5]
[2,3,4,5,6]
```

## 模块

什么是模块呢？模块就是一组函数或者数据类型。是一种集合，Haskell中的模块用 `module` 关键字声明。我们常用的模块例如预先导入的 `Prelude`，例如 `Data.List`, `Data.Map`等。我们可以把上面的 MyList.hs 包装成一个模块。

```haskell
module MyList (
      myInit
    , myTail
    ) where

myInit :: [a] -> a
myInit [] = error "cannot operate on empty list"
myInit (x:xs) = x

myTail :: [a] -> [a]
myTail [] = error "cannot operate on empty list"
myTail (x:xs) = xs
```

然后在导入的时候，可以有如下写法：

```haskell
import XMonad.Util.Run (spawnPipe)
import qualified Data.Map as M
```

第一个是类似于Python中 `from xxx import yyy` 的写法，只引入模块中的部分。
第二个是类似于Python中 `import xxx as yyy` 的写法，引入模块，但是创建别名。

## 自己创建类型和TypeClass

Haskell中还有一个关键字是 `data`。是用来创建我们的类型的。

```haskell
data Bool = False | True 
```

左边的 `Bool` 是类型(type)，右边的用 `|` 分隔的叫做 `value constructors`。有的需要带值，有的不需要。例如上面的Bool就不需要，而Maybe则需要：

```haskell
data Maybe a = Nothing | Just a
```

此处左边的a叫做 `type parameter`。而Maybe叫做 `type constructor`。为什么又有一个新名字而不是也叫type呢？因为 `Maybe`不是一个具体的类型，而像 `Maybe Int` 才是。

```haskell
Prelude> :t Nothing
Nothing :: Maybe a
Prelude> :t Just 1
Just 1 :: Num a => Maybe a
Prelude> :t Just "haha"
Just "haha" :: Maybe [Char]
```

可能看到这里还比较迷糊，但是在之后我们看到Monad的时候，就知道为什么要把这些分得这么清楚了。把有共同特征的一类事物提取出来就是抽象。
对于具体类型例如 `Bool`，是一种类型。而像 `Maybe` 这样需要接受一个具体类型才能产生出一个具体类型的类型，就是另一种类型了。

我们可以把Maybe看作是一个盒子，这个盒子需要填充一个东西，才算是一个有意义的盒子，否则
知识一个空荡荡的i空盒子。对于这样的想象成盒子的想法，我们在后面学习Monad的时候还会用到。

## 总结

在这一篇中我们简略的过了一下Haskell的基本语法，学习一门新语言首先要把基本的语法搞懂，然后更重要的是拿基本语法开始开发，
然后不断的搜索，写，看资料，查文档，写。这样才能更快的掌握一门语言。

参考资料：

- [Learn You a Haskell For Great Good](http://learnyouahaskell.com/chapters)
- [Haskell WikiBook](https://en.wikibooks.org/wiki/Haskell)
- [一步一步学Haskell（一）：从递归说起](https://jiajunhuang.com/articles/2017_09_11-learn_you_a_haskell_part_1.md.html)
- [一步一步学Haskell（二）：从命令式语言进行抽象](https://jiajunhuang.com/articles/2017_09_17-learn_you_a_haskell_part_2.md.html)
