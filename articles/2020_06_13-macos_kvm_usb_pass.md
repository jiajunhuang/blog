# 把USB设备穿透给虚拟机里的系统

最近买了一个NUC8，i5版本，因为听说好装黑苹果，买来之后，看到这么小一个电脑，真的是感叹科技进步速度之快。不过我并不喜欢
双系统，也不喜欢macOS，我已经用Linux桌面+Windows虚拟机这套组合8年了，一切都是得心应手。除了一个：最近我在学flutter，
没有macOS没法给iOS打包。

但是有没有办法可以通过虚拟机来完成呢？我知道有人弄显卡穿透，但是我用不着这个功能，我并不需要多好的图形性能，只要有一个
macOS可以用来编译即可。首先通过 [macOS-Simple-KVM](https://github.com/foxlet/macOS-Simple-KVM) 这个项目把macOS装虚拟机
里，我尝试过直接在 virt-manager 的菜单里，Virtual Machine -> Redirect USB Device 将iPhone传进去，但是不知为何，总是失败，
每次点传入，就会卡死，然后报错：Device is used by another application，接着会发现宿主机里，iPhone的设备号变了，相当于
把iPhone拔出来重新插进去。然而Windows虚拟机是可行的。

## 穿透USB设备

因此只好祭出穿透大法，首先你得确定CPU和主板都支持(3-5年内的基本都支持，这个得具体型号去查，看BIOS里是不是有一个Vt-d或者
AMD-Vi技术)，因为我们不是穿透GPU而是USB，这相对来说要简单很多。

首先在GRUB里，给内核加上IOMMU的参数，编辑 `/etc/default/grub`，找到 GRUB_CMDLINE_LINUX，改成(我们以Intel为例)：

```bash
GRUB_CMDLINE_LINUX="intel_iommu=on"
```

然后更新GRUB：

```bash
$ sudo update-grub
```

之后重启，执行命令看看是不是成功开启IOMMU：

```bash
$ sudo dmesg | grep -i -e DMAR -e IOMMU
...
[    0.000000] Intel-IOMMU: enabled
...
```

有这个就说明OK了。

接下来我们要穿透USB设备，IOMMU是会对设备分组的，一个组是穿透的一个基本单位，要么都传，要么都不传。执行下面的脚本来查看分组：

```bash
$ for usb_ctrl in /sys/bus/pci/devices/*/usb*; do pci_path=${usb_ctrl%/*}; iommu_group=$(readlink $pci_path/iommu_group); echo "Bus $(cat $usb_ctrl/busnum) --> ${pci_path##*/} (IOMMU group ${iommu_group##*/})"; lsusb -s ${usb_ctrl#*/usb}:; echo; done
Bus 1 --> 0000:00:14.0 (IOMMU group 4)
Bus 001 Device 003: ID 78jd:abcd Intel Corp. 
Bus 001 Device 001: ID efjk:abcd Linux Foundation 2.0 root hub

Bus 2 --> 0000:00:14.0 (IOMMU group 4)
Bus 002 Device 001: ID 12jh:8932 Linux Foundation 3.0 root hub

Bus 3 --> 0000:6c:00.0 (IOMMU group 15)
Bus 003 Device 005: ID dsac:892a Dell Computer Corp. 

Bus 4 --> 0000:6c:00.0 (IOMMU group 15)
Bus 004 Device 002: ID 89ds:jds2 Genesys Logic, Inc. Hub
Bus 004 Device 001: ID 879d:abcd Linux Foundation 3.0 root hub
```

这个时候，上面的哪个设备对应机器的哪个接口，就有点半蒙半猜了，主要看比如 `Dell Computer Corp`，我其实就一个鼠标是Dell的，
所以我知道这个是我的USB Hub，因此 `IOMMU group 15` 是我插了 Hub 的那个接口。

由此类推，NUC前面两个USB设备的组，是 `Bus 1 --> 0000:00:14.0 (IOMMU group 4)` 这个。

确定了分组之后，就好办了，打开 virt-manager，打开macOS的虚拟机，Details里面，添加PCI设备，把对应的硬件地址的那几行，
全部加进去，比如我这里是 `0000:00:14.0`，对应的就是 `PCI 0000:00:14.x`，x是一个数字，没关系，满足这个样子的全部加进去。

就这样，搞定，启动虚拟机，就会发现里面的macOS检测到了iPhone，duang！

---

参考资料：

- https://wiki.archlinux.org/index.php/PCI_passthrough_via_OVMF
