# 设计一个分布式块存储

https://github.com/jiajunhuang/hfs

最近读了GFS论文，然后想自己造一个轮子出来，毕竟自己的轮子圆又圆。

## 设计思路

最开始其实是脑子一团浆糊，没想好要怎么完成这个设计。不过仔细想了一下之后，逐步把大的系统拆解之后，然后依次实现和迭代，
最终还是成功的做出来了。

- 封装一套POSIX API的操作文件的API，用于操作本地文件，包括CRUD
- 封装一套操作文件(file)和chunk的API，并且提供gRPC接口，称之为chunkserver
- 使用etcd来存储关于chunk和file的信息，例如file有哪些chunk组成，顺序是如何，每个chunk大小是多少，实际上写入的数据是多少，把这些信息称之为meta data
- 给chunkserver加上服务注册的功能，使得在etcd中可以读取到worker的信息
- 增加一个监听chunk变化的worker，当发现本机有新建的chunk时，就挑选可用的其他chunkserver对该chunk进行同步
- 更改删除文件的API，删除文件时，删除所有节点上的chunk

差不多就是这样一个顺序。

还可以做的事情：

- 读文件的API需要改成从metadata中获取信息，然后从对应的chunkserver中获取chunk
- 读文件的客户端 `hfsclient` 可以改成并发下文件
- 增加检查chunk和file是否数据损坏的worker
- ...
