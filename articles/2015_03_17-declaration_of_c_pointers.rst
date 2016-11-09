如何解读c的声明
===============

2015-03-17 更新：
~~~~~~~~~~~~~~~~~

最近在看《C专家编程》, 结合《征服C指针》对上次的笔记做了一部分修改:

如何阅读C的声明？ **用英语阅读**, 按照以下步骤：

1, 首先着眼于标识符（变量名或者函数名）

2, 从距离标识符最近的地方开始,
依照优先顺序解释派生类型（指针、数组和函数）.优先顺序如下：

::

    2.1 用于整理声明内容的括弧

    2.2 后缀操作符: 用于表示数组的[], 用于表示函数的()

    2.3 用于表示指针的*

3, 解释完派生类型, 使用"of"、"to"、"returning"将他们连起来

4,最后, 追加数据类型修饰符（在左边, int, double等）, 对于const关键字,
如果右边紧跟类型修饰符(如int, double等), 则它作用于类型修饰符,
否则作用于左边紧邻的\ ``*``\ 指针星号, 并且虽然const意思是常量,
把它解读为\ ``read-only``\ 才更加正确

5,英语不好的人, 可以倒序用中文解释

注意读完一部分就把那一部分换成()忽略掉.

比如：

.. code:: c

    int (*func_p)(double);

1, 首先着眼于标识符.那就是func\_p : ``func_p is``

2, 因为有括号, 所以先着眼于 *, 那就是 *\ func\_p：
``func_p is pointer to``

3, 解释用于函数的(), 参数是double, 那就是 (\*func\_p)(double) :
``func_p is pointer to function(double) returning``

4, 最后, 解释数据类型修饰符 int, 那就是 int (\*func\_p)(double) :
``fun_p is pointer to function(double) returning int``

5, 翻译成中文: ``func_p 是指向返回类型为int的函数的指针``

;p 怎么样, 这方法很有效吧？

再来一个例子：

.. code:: c

    char * const *(*next)();

1, 首先我们着眼于next, 英文是"next is ";

2, next在括号内, 右边既没有\ ``[]``\ 也没有\ ``()``, 左边有\ ``*``,
所以英文是"next is a pointer to";

3, 接下来上面的式子变成了这个\ ``char *const *()()``, 我们继续解读,
()右边还有 (), 所以是一个函数, 函数没有形参: "next is a ponter to a
function(which has no arguments)";

4, 上面读完以后左边有一个\ ``*``, 所以应该继续解读为 "next is a pointer
to a function(which has no arguments), and the function return a pointer
to";

5, 完成上面一步, 原来的语句就变成了\ ``char * const ()``,
接下来按照规则, const右边没有类型修饰符, 所以const修饰的是左边的星号:
"next ... , and the function return a pointer to a read-only pointer";

6, 这下式子已经很简单了, 可谓是"司马昭之心, 路人皆知"(诶？好像有点不对?)
``char ()``, 所以最终解读为 " next ..., and the function return a
pointer to a read-only pointer who's type is char";

7, 翻译成中文: next是一个指针, 它指向一个函数, 函数返回另一个指针,
该指针指向一个类型为char的只读指针(其中函数没有参数).

2014-10-17:
~~~~~~~~~~~

别人翻译的《征服C指针web版》:
`点我 <http://avnpc.com/pages/c-pointer>`__

2015-03-17:
~~~~~~~~~~~

`征服C指针 <http://www.amazon.cn/%E5%9B%BE%E7%81%B5%E7%A8%8B%E5%BA%8F%E8%AE%BE%E8%AE%A1%E4%B8%9B%E4%B9%A6-%E5%BE%81%E6%9C%8DC%E6%8C%87%E9%92%88-%E5%89%8D%E6%A1%A5%E5%92%8C%E5%BC%A5/dp/B00BKU37NG/ref=sr_1_1?ie=UTF8&qid=1426599089&sr=8-1&keywords=%E5%BE%81%E6%9C%8Dc%E6%8C%87%E9%92%88>`__

`C专家编程 <http://www.amazon.cn/C%E4%B8%93%E5%AE%B6%E7%BC%96%E7%A8%8B-%E8%8C%83%E5%BE%B7%E6%9E%97%E7%99%BB/dp/B00BHSPPDQ/ref=sr_1_1?ie=UTF8&qid=1426599140&sr=8-1&keywords=c%E4%B8%93%E5%AE%B6%E7%BC%96%E7%A8%8B+%E8%8B%B1%E6%96%87%E7%89%88>`__
(english version)
