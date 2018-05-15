# MySQL EXPLAIN中的filesort是什么？

- https://dev.mysql.com/doc/refman/5.7/en/order-by-optimization.html#order-by-filesort
- https://www.percona.com/blog/2009/03/05/what-does-using-filesort-mean-in-mysql/
- http://s.petrunia.net/blog/?p=24

最开始我一为filesort是外排。结果不是，EXPLAIN中这个提示非常具有误导性。只要排序的时候不能用上索引，就会显示成filesort。
MySQL所谓的"filesort"进行的动作：

- Read the rows that match the `WHERE` clause
- For each row, store in the sort buffer a tuple consisting of the sort key value and the columns referenced by the
query
- When the sort buffer becomes full, sort the tuples by sort key value in memory and write it to a temporary file
- After merge-sorting the temporary file, retrieve the rows in sorted order, but read the columns required by the
query directly from the sorted tuples rather than by accessing the table a second time.

注意，当内存放不下所有的数据时，数据会被分块，对于每个块应用快排，对于多个块应用合并排序。
