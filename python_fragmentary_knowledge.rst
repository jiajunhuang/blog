:Date: 02/13/2016

Python零碎知识汇总
===================

    这些都是我在PEP里阅读到,而以前不知道或者没有注意或者我觉得仍然需要
    注意的知识.

- 字符格式化: '%s'会先把参数转换成unicode然后再填充进去.

- 嵌套list comprehension: ``[(i, f) for i in i_list for f in f_list]``
