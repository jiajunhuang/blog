# Haskell简明教程（五）：处理JSON

> 这一系列是我学习 `Learn You a Haskell For Great Good` 之后，总结，编写的学习笔记。

这个系列主要分为五个部分：

- [从递归说起](./2017_09_11-learn_you_a_haskell_part_1.md.html)
- [从命令式语言进行抽象](./2017_09_17-learn_you_a_haskell_part_2.md.html)
- [Haskell初步：语法](./2017_09_18-learn_you_a_haskell_part_3.md.html)
- [Haskell进阶：Monoid, Applicative, Monad](./2017_09_25-learn_you_a_haskell_part_4.md.html)
- [实战：Haskell和JSON](#)

> 本文有参考 《Real World Haskell》

## Haskell如何表示JSON

[json](http://json.org/) 是一种数据表现形式，来源于 javascript。JSON中有四种基本的值的类型：

- number
- boolean: true/false
- string
- null

此外有两种容器类型：

- array
- object

那么Haskell我们需要如何表示成JSON呢？我们需要一个 value constructor 加上Haskell自带的类型来表示：

```haskell
data JValue = JNumber Integer
          | JString String
          | JBoolean Bool
          | JNull
          | JArray [JValue]
          | JObject [(String, JValue)]
            deriving (Show, Ord, Eq)
```

那么我们要如何输出JSON呢？其类型肯定是这样的：

`echoJValue :: JValue -> String`

我们分别为 string, bool, null, number 定义好如何输出成JSON：

```haskell
echoJValue (JNumber i) = show i
echoJValue (JString s) = show s
echoJValue (JBoolean True) = "true"
echoJValue (JBoolean False) = "false"
echoJValue JNull = "null"
```

array该怎么处理？输出 "[" 之后输出array中的内容，然后输出 "]"，所以是：

```haskell
echoJValue (JArray a) = "[" ++ handleArray a ++ "]"
    where handleArray [] = ""
          handleArray a = intercalate ", " (map echoJValue a)
```

而object则是：

```haskell
echoJValue (JObject a) = "{" ++ handleObject a ++ "}"
    where handleObject [] = ""
          handleObject a = intercalate ", " (map handleKV a)
          handleKV (k,v) = k ++ ": " ++ echoJValue v
```

此处我们需要导入 `Data.List` 中的 intercalate 函数：

```haskell
Prelude> :m Data.List
Prelude Data.List> :t intercalate 
intercalate :: [a] -> [[a]] -> [a]
Prelude Data.List> intercalate ", " ["hello", "world"]
"hello, world"
```

于是我们便可以输出自定义的JSON了。

```haskell
Prelude> :l JSON
[1 of 1] Compiling Main             ( JSON.hs, interpreted )
Ok, 1 module loaded.
*Main> JNull
JNull
*Main> JNumber 1
JNumber 1
*Main> JString "hello"
JString "hello"
*Main> JBoolean True
JBoolean True
*Main> JBoolean False
JBoolean False
*Main> JArray [JBoolean True, JBoolean False]
JArray [JBoolean True,JBoolean False]
*Main> echoJValue $ JArray [JBoolean True, JBoolean False]
"[true, false]"
*Main> echoJValue $ JObject [("hello", JNumber 1), ("world", JString "world")]
"{hello: 1, world: \"world\"}"
```

我们可以把整个文件做成一个Module：

```haskell
module JSON (
    JValue(..),
    echoJValue
    ) where 

import Data.List (intercalate)


data JValue = JNumber Integer
          | JString String
          | JBoolean Bool
          | JNull
          | JArray [JValue]
          | JObject [(String, JValue)]
            deriving (Show, Ord, Eq)


echoJValue :: JValue -> String
echoJValue (JNumber i) = show i
echoJValue (JString s) = show s
echoJValue (JBoolean True) = "true"
echoJValue (JBoolean False) = "false"
echoJValue JNull = "null"
echoJValue (JArray a) = "[" ++ handleArray a ++ "]"
    where handleArray [] = ""
          handleArray a = intercalate ", " (map echoJValue a)
echoJValue (JObject a) = "{" ++ handleObject a ++ "}"
    where handleObject [] = ""
          handleObject a = intercalate ", " (map handleKV a)
          handleKV (k,v) = k ++ ": " ++ echoJValue v
```

那我们该要怎样 pretty print呢？这就留作思考吧 :) (提示：Haskell的世界，递归总是那么重要)

## 总结

到此Haskell教程便结束了，我们从最开始的命令式语言如何抽象，到介绍Haskell的基本语法和要素，
TypeClass, Functor, Applicative, Monoid, Monad，到随后我们封装一个简单的JSON模块。简略的
浏览了一下Haskell的面貌，但是Haskell抽象程度很高，为此付出的代价便是不那么容易一眼就看出来，
有时候甚至需要来回看，来回琢磨才能领会其中的意义。

之后我将会写一些散篇，主要是Haskell和现实世界的应用，或者是Haskell的源代码分析。

参考资料：

- [Learn You a Haskell For Great Good](learnyouahaskell.com/chapters)

- [Real World Haskell](http://book.realworldhaskell.org/read/)
