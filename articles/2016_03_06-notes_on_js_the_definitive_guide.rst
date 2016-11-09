JavaScript权威指南笔记
=========================

- 声明变量不带var，JS会在全局对象中创建一个同名属性，坑爹

- JS在函数作用域上倒是和Python有些相似，不过Python是直接抛出错误，而JS是打印出undefined.

.. code:: python

    >>> def foo():
    ...   print(a)
    ...   a = 1
    ...
    >>> foo()
    Traceback (most recent call last):
    File "<stdin>", line 1, in <module>
        File "<stdin>", line 3, in foo
    UnboundLocalError: local variable 'a' referenced before assignment

.. code:: javascript

    > function foo() {
        ...   console.log(x);
        ...   var x = 1;
        ...   console.log(x);
        ...
    }
    >
    > foo()
    undefined
    1

- JS的in操作符，``"x" in {x: 1}`` 表现还比较正常，但是对数组操作的时候，就不是正规军了:

.. code:: javascript

    > 1 in [1]
    false
    > 1 in [1,2]
    true
    > "0" in [1]
    true
    > "0" in [1, 2]

  原因在于，JS把前面转成索引。。。

- delete 只是删除引用
