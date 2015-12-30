Python lib浏览笔记 - 类型
=========================

-  | Concatenating immutable sequences always results in a new object.
     This means
   | that building up a sequence by repeated concatenation will have a
     quadratic
   | runtime cost in the total sequence length. To get a linear runtime
     cost, you
   | must switch to one of the alternatives below:

       | if concatenating str objects, you can build a list and use
         str.join() at
       | the end or else write to a io.StringIO instance and retrieve
         its value
       | when complete.
       | if concatenating bytes objects, you can similarly use
       | bytes.join() or io.BytesIO, or you can do in-place
         concatenation with a
       | bytearray object. bytearray objects are mutable and have an
         efficient
       | overallocation mechanism.
       | if concatenating tuple objects, extend a list instead for other
         types,
       | investigate the relevant class documentation.

-  | The only operation that immutable sequence types generally
     implement that is
   | not also implemented by mutable sequence types is support for the
     hash()
   | built-in.

-  | list.sort() modifies the sequence in place for economy of space
     when
   | sorting a large sequence.use sorted() to explicitly request a new
     sorted
   | list instance.

-  | it is actually the comma which makes a tuple, not the parentheses.
   | The parentheses are optional, except in the empty tuple case, or
     when
   | they are needed to avoid syntactic ambiguity. e.g.

.. code:: python3

    >>> a = 1,2,3
    >>> a
    (1,2,3)

-  | Ranges implement all of the common sequence operations except
     concatenation
   | and repetition. range 类实现了collections.abc.Sequence
     定义的方法里除
   | 连接(concat)和重复(repetition)之外所有操作.

-  print(): '%r' use ``repr()``, '%s' use ``str()``, '%a' use
   ``ascii()``

-  set is muable, frozenset is immutable.
