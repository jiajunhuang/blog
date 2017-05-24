# HTTP 路由的两种常见设计形式

- 线性型

基本上是类似于用一个list保存路由的正则表达式，每次将url拿去匹配然后找到handler

- radix-tree型

radix-tree基于trie树，使得路由可以共用共同的前缀，查找效率高于线性
