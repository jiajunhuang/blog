---
layout: post
title: 《征服C指针》笔记 - 如何解读c的声明
date: 2014-09-27
---

如何阅读C的声明？ **用英语阅读**, 按照以下步骤：

1， 首先着眼于标识符（变量名或者函数名）

2， 从距离标识符最近的地方开始， 依照优先顺序解释派生类型（指针、数组和函数）。优先顺序如下：

    2.1 用于整理声明内容的括弧

    2.2 用于表示数组的[]， 用于表示函数的()

    2.3 用于表示指针的*

3, 解释完派生类型， 使用"of"、"to"、"returning"将他们连起来

4，最后， 追加数据类型修饰符（在左边， int, double等）

5，英语不好的人， 可以倒序用中文解释

比如：

```c
int (*func_p)(double)
```

1, 首先着眼于标识符。那就是func_p : `func_p is`

2, 因为有括号， 所以先着眼于 *， 那就是 *func_p： `func_p is pointer to`

3, 解释用于函数的(), 参数是double， 那就是 (*func_p)(double) : `func_p is pointer to function(double) returning`

4, 最后， 解释数据类型修饰符 int， 那就是 int (*func_p)(double) : `fun_p is pointer to function(double) returning int`

5, 翻译成中文: `func_p 是指向返回类型为int的函数的指针`

;p 怎么样， 这方法很有效吧？
