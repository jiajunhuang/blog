---
layout: post
title: 获取磁盘的UUID
---

在编辑`/etc/fstab`的时候个人觉得还是用UUID比较好，这样子你改变其他的分区都不影响OS去找磁盘位置

获得`UUID`的方法就是使用`blkid`啦:

```bash
$ sudo blkid /dev/sda1
....(Output)
```
