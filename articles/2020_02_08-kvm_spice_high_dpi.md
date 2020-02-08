# KVM spice协议在高分屏上的分辨率问题

高分屏用起来很爽，但是目前还是经常会遇到各类问题，比如，最近我发现如果我使用扩展显示器，virt manager可以完美的切换guest
的分辨率，但是一旦我切回笔记本的屏幕（高分屏），guest就无法自动更新分辨率，而是会有一个最大上限比如 `1600x1020` 之类的。

最后发现是QXL的内存限制只有16M，莫非是不够用？计算一下在 `3000x2000` 的屏幕上需要多少内存来当显存：

```python
In [1]: 3000 * 2000 * 32 / 8 / 1024 / 1024                                                                                              
Out[1]: 22.88818359375
```

其中 32 是32位色深，除以8是把单位从bit转换为byte，除以 `(1024 * 1024)` 是为了把单位从byte转换为MB。所以给22.88M内存就够了，
不过我直接给了32M

```bash
$ sudo virsh
# edit --domain thinpc
```

找到qxl一行：

```
<model type='qxl' ram='65536' vram='65536' vgamem='16384' heads='1' primary='yes'/>
```

改为

```
<model type='qxl' ram='65536' vram='65536' vgamem='32768' heads='1' primary='yes'/>
```

---

参考资料：

- https://stafwag.github.io/blog/blog/2018/04/22/high-screen-resolution-on-a-kvm-virtual-machine-with-qxl/
