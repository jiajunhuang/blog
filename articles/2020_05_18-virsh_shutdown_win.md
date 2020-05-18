# virsh自动关闭windows虚拟机

大家都知道，这种定时任务是通过crontab来做，但是，如果你直接尝试在crontab里关机，你会发现有时候是关不掉windows虚拟机的。
这是为啥呢？这个问题困扰了我好久，因为每次我直接输入命令关机的时候，都关机成功了，我要是放到crontab里等，它也能关机，
但是我设置crontab在晚上1点关机，第二天发现，它就是没有关机。

最后发现，原来是windows有这么一个锅：在息屏之后，如果你输入 `virsh shutdown --domain win` 的话，它只会激活屏幕，此时
如果你输入第二次这个命令的话，就可以成功关机。

原来问题是这样：因为我每次尝试的时候，都不是息屏状态。

解决方案有三种：

- 禁用Windows的息屏：Control Panel -> System and Security -> Power Options --> Click "Change plan Settings".
Set "Turn off the display" to Never (default is 10 minutes)

- [更改注册表，允许息屏时关机](https://serverfault.com/questions/844188/shut-down-windows-server-2012r2-kvm-vm)

- 还有就是我选择的最土的方式：执行两遍shutdown(不喜欢改注册表，也不想让它一直亮屏)

```bash
# 工作日早上9点，自动开启虚拟机
0 9 * * 1-5 virsh start --domain win
# 工作日晚上18点，自动关闭虚拟机
0 18 * * 1-5 virsh shutdown --domain win && virsh shutdown --domain win
```

---

参考资料：

- https://www.suse.com/support/kb/doc/?id=000018975
- https://serverfault.com/questions/844188/shut-down-windows-server-2012r2-kvm-vm
