# 设计一个路由

简单地说一下常见的路由形式。

- 数组存储。按照添加路由的顺序存储在数组中，查找时依次匹配。这种路由效率比较低。tornado就是这么设计的。

- 字典存储。把路由存储在map里，查找效率很高，但是不支持URI中有参数。Golang中默认的mux就这么设计的。

- 树。一般使用前缀树，空间更紧凑的用radix tree。如 httprouter 就是这种设计。

一般就这么几种，可能会有些许细节上的不同。
