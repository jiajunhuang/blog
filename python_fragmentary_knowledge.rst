:Date: 02/13/2016

Python零碎知识汇总
===================

    这些都是我在文档里阅读到或者是在网上看到的,不足以成文却值得注意的知识.

- 字符格式化: '%s'会先把参数转换成unicode然后再填充进去.

- 嵌套list comprehension: ``[(i, f) for i in i_list for f in f_list]``

- The name binding operations are argument declaration, assignment,
  class and function definition, import statements, for statements,
  and except clauses.  Each name binding occurs within a block
  defined by a class or function definition or at the module level
  (the top-level code block).

  If a name is bound anywhere within a code block, all uses of the
  name within the block are treated as references to the current
  block.  (Note: This can lead to errors when a name is used within
  a block before it is bound.)

- An assignment operation can only bind a name in the current scope
  or in the global scope.

- ``filter(function, iterable)`` 相当于 ``(item for item in iterable if function(item))``
  如果function为None的话，相当于 ``(item for item in iterable if item)``. 另外，
  `itertools.filterfalse() <http://localhost/py35/library/itertools.html#itertools.filterfalse>`__ 可以把结果为false的过滤出来。

- ``iter(object[, sentinel])`` 要求 object 必须实现 ``__iter__()`` 或者 ``__getitem__()``
  之一，第二个参数是停止提示。

- ``round(number, [ndigits])`` 给小数截断:

.. code:: python

    >>> round(10.12345, 2)
    10.12

- python 2 中的 ``hasattr`` 比较坑,原因如下:

    The arguments are an object and a string. The result is True if the string
    is the name of one of the object’s attributes, False if not. (This is
    implemented by calling getattr(object, name) and seeing whether it raises
    an exception or not.)

  坑就坑在它接住的异常不是 ``AttributeError`` 而是 ``Exception``

- ``sum(iterable, [start])`` 可以接受字符串，数字等。For some use cases,
    there are good alternatives to sum(). The preferred, fast way to
    concatenate a sequence of strings is by calling ''.join(sequence).
    To add floating point values with extended precision, see math.fsum().
    To concatenate a series of iterables, consider using itertools.chain().

- ``type(name, bases, dict)`` 可以用来动态创建类，其中name是类名，bases是基类，
  dict是类的 ``__dict__`` 属性
