# 一起来做贼：Goroutine原理和Work stealing

最近在看Goroutine的实现原理，用户空间实现的并发模型。在用户空间进行调度不比
在内核级别，因为内核可以通过CPU中断夺回控制权，但是用户空间把控制权交给一段
代码之后，需要那段代码主动交出权力才可以。当然也可以通过一些trick，例如编译
的时候检测，然后自动插入某些让出CPU的代码。或者在生成的指令中插入。

关于Goroutine网上已经有很多篇很棒的文章了，我认为看这几篇就够了：

- 设计文档：https://docs.google.com/document/d/1TTj4T2JO42uD5ID9e89oa0sLKhJYD0Y_kqxDv3I3XMw/edit

- 比设计文档更详细并且配图的博客：http://morsmachine.dk/go-scheduler

- Goroutine背后的系统知识：http://webcache.googleusercontent.com/search?q=cache:AoAQmcVigUkJ:www.sizeofvoid.net/goroutine-under-the-hood/+&cd=1&hl=en&ct=clnk&gl=hk

其中当一个P的G队列消耗完了，就会尝试去其他P的G队列里偷一些任务过来。看看wikipedia：

- https://en.wikipedia.org/wiki/Fork%E2%80%93join_model

- https://en.wikipedia.org/wiki/Work_stealing
