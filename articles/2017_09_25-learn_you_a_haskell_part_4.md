# Haskell简明教程（四）：Monoid, Applicative, Monad

> 这一系列是我学习 `Learn You a Haskell For Great Good` 之后，总结，编写的学习笔记。

这个系列主要分为五个部分：

- [从递归说起](./2017_09_11-learn_you_a_haskell_part_1.md.html)
- [从命令式语言进行抽象](./2017_09_17-learn_you_a_haskell_part_2.md.html)
- [Haskell初步：语法](./2017_09_18-learn_you_a_haskell_part_3.md.html)
- [Haskell进阶：Monoid, Applicative, Monad](#)
- [实战：Haskell和JSON](./2017_09_26-learn_you_a_haskell_part_5.md.html)

## 回忆TypeClass

TypeClass，我们在第二篇中就讲过，与命令式编程不同，Haskell中的class不是类，而是更像
"接口"这个概念，或者说，"类型类"。比如我们有个接口是能比较是否相等：

```haskell
class Equalable a where
    equal :: a -> a -> Bool
    uneuqal :: a -> a -> Bool

    equal x y = not $ uneuqal x y
    uneuqal x y = not $ equal x y
```

首先我们可以看到 Equalable 针对一个类a，其中类型声明，`equal :: a -> a -> Bool`表示
`equal`这个函数接受两个a类型的参数，然后返回一个布尔类型的值。并且我们提供了默认实现，
`equal`就是 `uneuqal` 的反，`uneuqal`就是`equal`的反。我们来看看Int是怎么实现这个接口的：

```haskell
instance Equalable Int where
    equal x y = x == y
```

运行一下：

```haskell
Prelude> :load Demo.hs
[1 of 1] Compiling Main             ( Demo.hs, interpreted )
Ok, 1 module loaded.
*Main> let a = 1 :: Int
*Main> let b = 1 :: Int
*Main> let c = 2 :: Int
*Main> a `equal` b
True
*Main> a `equal` c
False
```

热身完毕，接下来我们将要开始讲Monoid这个 `class`。

## Monoid

在讲述 `Monoid` 之前，我们需要先看看 `Functor` 和 `Applicative`热热身。

### Functor和Applicative

上定义：

```ghci
Prelude> :i Functor
class Functor (f :: * -> *) where
  fmap :: (a -> b) -> f a -> f b
  (<$) :: a -> f b -> f a
  {-# MINIMAL fmap #-}
  	-- Defined in ‘GHC.Base’
instance Functor (Either a) -- Defined in ‘Data.Either’
instance Functor [] -- Defined in ‘GHC.Base’
instance Functor Maybe -- Defined in ‘GHC.Base’
instance Functor IO -- Defined in ‘GHC.Base’
instance Functor ((->) r) -- Defined in ‘GHC.Base’
instance Functor ((,) a) -- Defined in ‘GHC.Base’
Prelude> :i Applicative
class Functor f => Applicative (f :: * -> *) where
  pure :: a -> f a
  (<*>) :: f (a -> b) -> f a -> f b
  GHC.Base.liftA2 :: (a -> b -> c) -> f a -> f b -> f c
  (*>) :: f a -> f b -> f b
  (<*) :: f a -> f b -> f a
  {-# MINIMAL pure, ((<*>) | liftA2) #-}
  	-- Defined in ‘GHC.Base’
instance Applicative (Either e) -- Defined in ‘Data.Either’
instance Applicative [] -- Defined in ‘GHC.Base’
instance Applicative Maybe -- Defined in ‘GHC.Base’
instance Applicative IO -- Defined in ‘GHC.Base’
instance Applicative ((->) a) -- Defined in ‘GHC.Base’
instance Monoid a => Applicative ((,) a) -- Defined in ‘GHC.Base’
```

`fmap`？似曾相识，不就是map吗？map是函数，但是 `Functor`是能应用到map函数
的东西抽象出来的接口。我们暂且把被map应用的东西叫做 "容器" 或者 "盒子"。

比如最典型的，`List`能被map(之后的例子我们都用List)：

```haskell
Prelude> map (+1) [1 .. 3]
[2,3,4]
```

`<$`？看样子是指把一个类型放到容器里，便能反推出该类型所对应的有容器的值，
说起来好绕，看一个具体例子好了：

```haskell
Prelude> fmap (+1) [1 .. 3]
[2,3,4]
Prelude> (<$) 1 [3]
[1]
```

果然是这样。瞧，这就是类型声明的好处，读代码靠看类型就能猜个大概出来 :joy:

至于 `Applicative`，我们则可以看到，首先Applicative是Functor，此外：

```haskell
class Functor f => Applicative (f :: * -> *) where
  pure :: a -> f a
  (<*>) :: f (a -> b) -> f a -> f b
  GHC.Base.liftA2 :: (a -> b -> c) -> f a -> f b -> f c
  (*>) :: f a -> f b -> f b
  (<*) :: f a -> f b -> f a
```

- `pure` 应该是拿一个具体的值，便能给出在容器中的值
- `<*>` 则是把从a到b的函数丢到容器里，然后取一个容器，并且将函数引用到那个容器
- `*>`　则是取两个容器但是舍弃前面一个容器
- `<*`　则是取两个容器但是舍弃后面一个容器
- `GHC.Base.liftA2` 则是一个从a和b到c的函数，然后取两个容器，其中容器内类型分别为a和b，产生装有c的容器

是不是更晕？没关系，Haskell就是这样，每一行代码你都需要仔细考虑。我们对这上面的讲解分别看下面五个例子：

```
Prelude> pure 1 :: [Int]
[1]
Prelude> (<*>) [\x -> x + 1] [1]
[2]
Prelude> (*>) [1] [2]
[2]
Prelude> (<*) [1] [2]
[1]
Prelude> GHC.Base.liftA2 (\a b -> a + b) [1] [2]
[3]
```

瞧，有感觉了吗？如果没有的话，我想可能需要重新一步一步跟着来，再读一遍此前的内容。接下来我们看`Monoid`。

### Monoid

首先我们打开ghci看看定义：

```haskell
Prelude> :m Data.Monoid
Prelude Data.Monoid> :i Monoid
class Monoid a where
  mempty :: a
  mappend :: a -> a -> a
  mconcat :: [a] -> a
  {-# MINIMAL mempty, mappend #-}
  	-- Defined in ‘GHC.Base’
instance Num a => Monoid (Sum a) -- Defined in ‘Data.Monoid’
instance Num a => Monoid (Product a) -- Defined in ‘Data.Monoid’
instance Monoid (Last a) -- Defined in ‘Data.Monoid’
instance Monoid (First a) -- Defined in ‘Data.Monoid’
instance Monoid (Endo a) -- Defined in ‘Data.Monoid’
instance Monoid a => Monoid (Dual a) -- Defined in ‘Data.Monoid’
instance Monoid Any -- Defined in ‘Data.Monoid’
instance GHC.Base.Alternative f => Monoid (Alt f a)
  -- Defined in ‘Data.Monoid’
instance Monoid All -- Defined in ‘Data.Monoid’
instance Monoid [a] -- Defined in ‘GHC.Base’
instance Monoid Ordering -- Defined in ‘GHC.Base’
instance Monoid a => Monoid (Maybe a) -- Defined in ‘GHC.Base’
instance Monoid a => Monoid (IO a) -- Defined in ‘GHC.Base’
instance Monoid b => Monoid (a -> b) -- Defined in ‘GHC.Base’
instance (Monoid a, Monoid b, Monoid c, Monoid d, Monoid e) =>
         Monoid (a, b, c, d, e)
  -- Defined in ‘GHC.Base’
instance (Monoid a, Monoid b, Monoid c, Monoid d) =>
         Monoid (a, b, c, d)
  -- Defined in ‘GHC.Base’
instance (Monoid a, Monoid b, Monoid c) => Monoid (a, b, c)
  -- Defined in ‘GHC.Base’
instance (Monoid a, Monoid b) => Monoid (a, b)
  -- Defined in ‘GHC.Base’
instance Monoid () -- Defined in ‘GHC.Base’
```

```haskell
class Monoid a where
  mempty :: a
  mappend :: a -> a -> a
  mconcat :: [a] -> a
```

我们看类型来推测：

- `mempty` 不取值，给出一个类型的 "空值"
- `mappend` 取一个一个值和另一个值，返回同类型的值
- `mconcat` 取一个列表的值，合成一个值

说实话，这些都是我猜的，所以得验证一下：

```haskell
Prelude Data.Monoid> mempty :: [Int]
[]
Prelude Data.Monoid> mappend [1] [2]
[1,2]
Prelude Data.Monoid> mconcat [[1], [2, 3], [4]]
[1,2,3,4]
```

原来是这样，第一个如我所说，第二个是把两个容器连起来，第三个是把容器的容器打散组合成
一个新的容器。原来是这样，那么有哪些类型实现了这个接口呢？我们可以看到上面，`Maybe a`，
`[a]`等都实现了，因为 `Monoid` 是针对操作容器自身的，所以感觉有些抽象，有点像 `Python`
里的 `metaclass`。这一节得要仔细消化。

## Monad

Monad和Monoid有什么关系吗？说实话，我个人认为是雷锋和雷峰塔的关系。国际惯例，我们来看看定义：

```haskell
Prelude> :i Monad
class Applicative m => Monad (m :: * -> *) where
  (>>=) :: m a -> (a -> m b) -> m b
  (>>) :: m a -> m b -> m b
  return :: a -> m a
  fail :: String -> m a
  {-# MINIMAL (>>=) #-}
  	-- Defined in ‘GHC.Base’
instance Monad (Either e) -- Defined in ‘Data.Either’
instance Monad [] -- Defined in ‘GHC.Base’
instance Monad Maybe -- Defined in ‘GHC.Base’
instance Monad IO -- Defined in ‘GHC.Base’
instance Monad ((->) r) -- Defined in ‘GHC.Base’
instance Monoid a => Monad ((,) a) -- Defined in ‘GHC.Base’
```

我们继续猜测Monad的几个接口：

- `return` 是取一个具体值，然后把它封到容器里
- `fail` 是取一个字符串，然后生成一个容器
- `>>=` 是取一个容器，取一个具体值到盒子的映射函数，生成一个容器
- `>>` 是取两个容器，舍弃前一个

看看具体例子：

```haskell
Prelude> return 1 :: [Int]
[1]
Prelude> fail "hello" :: [Int]
[]
Prelude> (>>=) [1] (\x -> [x + 1])
[2]
Prelude> (>>) [1] [2]
[2]
```

没了，这就是Monad。更深的Monad的内容需要到以后实际用到才更好理解。这里先讲到这个程度。

## 为什么总是讲到盒子？容器？抽象？

我们很久之前就讲到过抽象的好处，抽象使得我们不必关心具体实现细节，只需要知道有这么一个
方法，我们只要这样用就好。而所谓的盒子，所谓的容器其实是同样的想法，为了抽象。

什么是Monad？实现了这几个接口就可以是一个Monad。[XMonad](http://xmonad.org/)就因此得名，
因为他把核心实现了Monad这个接口(类型类)。
