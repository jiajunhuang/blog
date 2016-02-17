:Date: 02/13/2016

Python零碎知识汇总
===================

    这些都是我在PEP里阅读到,而以前不知道或者没有注意或者我觉得仍然需要
    注意的知识.

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