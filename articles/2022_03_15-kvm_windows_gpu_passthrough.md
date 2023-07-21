# KVM 显卡穿透给 Windows

很久以前就想这么玩，但是碍于各种条件，一直没有实现，直到最近，得到一块额外的亮机卡，于是就有了这次折腾。
那么通过显卡穿透，我们可以做啥呢？比如：我们可以用Linux做宿主机，虚拟化一个Windows出来，但是Windows使用独显打游戏；
或者我们可以把显卡透传给虚拟机挖矿，等等。我的宿主机是AMD平台，宿主机使用AMD显卡，虚拟机使用Nvidia显卡。

## 准备

- 首先我们需要有一个可扩展性较强的电脑，必须要有两个PCIe插槽，分别插上两个显卡，当然如果是Intel的卡，那么可以让宿主机
使用集成显卡，而虚拟机使用独立显卡
- 要主板和CPU都支持IOMMU，如果不了解这个是什么，那么简单一句话来说，就是支持高性能虚拟化和透传的功能，基本上近几年生产
的主板和CPU都支持，如果想要确认一下，可以参考 wiki 中的做法
- 安装 libvirt、OVMF 和 virt-manager，我们使用 virt-manager 来配置，这样会方便很多；OVMF是开源的UEFI固件
- 如果可以的话，准备两个显示器，如果没有的话，那就只能自己把 HDMI 线来回插拔了

## 开始

### 设置BIOS并且编辑GRUB

首先我们要设置BIOS，打开IOMMU功能，不过我的是AMD的主板，所以默认就是打开的，接下来则需要编辑GRUB，在启动的时候就告诉
Linux内核要启用透传。

编辑 `/etc/default/grub` 文件，找到 `GRUB_CMDLINE_LINUX_DEFAULT`，添加 `quiet iommu=pt`，因为我的是AMD，所以只要这样
就可以，Intel的我没试过，但是网上说要编辑成：`quiet intel_iommu=on iommu=pt`。

接着重新生成一下GRUB配置文件：

```bash
$ sudo grub-mkconfig -o /boot/grub/grub.cfg
```

### 将硬件加入黑名单

首先我们要做的事情就是，让 Linux 宿主机忽略某一个显卡，通常，显卡在 Linux 中都会有两个设备，一个 Video 输出，一个 Audio
输出，Linux会为每一个硬件都分配一个ID，我们要找到他们，使用这个脚本，他会自动把 IOMMU 分组情况打印出来：

```bash
#!/bin/bash
shopt -s nullglob
for g in $(find /sys/kernel/iommu_groups/* -maxdepth 0 -type d | sort -V); do
    echo "IOMMU Group ${g##*/}:"
    for d in $g/devices/*; do
        echo -e "\t$(lspci -nns ${d##*/})"
    done;
done;
```

样例输出：

```bash
IOMMU Group 1:
	00:01.0 PCI bridge: Intel Corporation Xeon E3-1200 v2/3rd Gen Core processor PCI Express Root Port [8086:0151] (rev 09)
IOMMU Group 2:
	00:14.0 USB controller: Intel Corporation 7 Series/C210 Series Chipset Family USB xHCI Host Controller [8086:0e31] (rev 04)
IOMMU Group 4:
	00:1a.0 USB controller: Intel Corporation 7 Series/C210 Series Chipset Family USB Enhanced Host Controller #2 [8086:0e2d] (rev 04)
IOMMU Group 10:
	00:1d.0 USB controller: Intel Corporation 7 Series/C210 Series Chipset Family USB Enhanced Host Controller #1 [8086:0e26] (rev 04)
IOMMU Group 13:
	06:00.0 VGA compatible controller: NVIDIA Corporation GM204 [GeForce GTX 970] [10de:13c2] (rev a1)
	06:00.1 Audio device: NVIDIA Corporation GM204 High Definition Audio Controller [10de:0fbb] (rev a1)
```

其中的 `8086:0151`，`8086:0e31`，`10de:13c2` 这些就是硬件ID，我们之后要屏蔽硬件，就靠这个ID。我这里是宿主机使用AMD显卡，
虚拟机使用Nvidia显卡，分别找到它们的硬件ID。

接下来我们编辑文件 `/etc/modprobe.d/vfio.conf`：

```
softdep nouveau pre: vfio-pci
options vfio-pci ids=8086:0151,8086:0e31,10de:13c2
```

第一行表示，让 `vfio-pci` 驱动先于 `nouveau` 加载，`vfio-pci` 加载以后，就会把硬件占着，`nouveau` 是Nvidia显卡的开源驱动，
第二行表示，我们要占用的硬件ID，多个ID用英文逗号连接。

接着重新生成启动镜像：

```bash
$ sudo update-initramfs -u
```

接下来很重要的一步，就是重启。这样我们第一步和第二步所做的更改，才能产生效果。

### 创建Windows虚拟机

我使用的是Windows 10 LTSC，就和普通创建流程一样，但是在最后一步，勾选 `Customize before install`，然后在Overview里，
把BISO改成 `UEFI`，Chipset 改成 `i440`，UEFI固件选择 `/usr/share/OVMF/OVMF_CODE_4M.fd`。将磁盘驱动改成VirtIO，添加
`virtio-win.iso` ，这里面包含的是VirtIO的各种驱动，如果没有的话，从 https://fedorapeople.org/groups/virt/virtio-win/direct-downloads/archive-virtio/ 找到最新的下载。

> 一定要UEFI，否则虚拟机能检测到显卡，但是无法输出信号。我的尝试结果是这样。

接下来就是正常安装了，这里我不会事无巨细的贴出来，因为如果你不会安装的话，穿透是基本没戏了。

### 透传硬件

安装完成之后，关闭虚拟机，点开详情，点添加，然后选择 `PCI Host Device`，把前面我们要透传的设备添加进去，记得一定要同时
添加 Video 和 Audio 设备。

接下里很重要的一步，就是要隐藏虚拟机特性，因为Nvidia不让在虚拟机里使用，但是我们可以通过隐藏虚拟机特性规避。
执行 `sudo virsh edit --domain win` 来编辑你的虚拟机，记得替换成你自己的虚拟机名字。

- 在 `<os>` 标签里面，增加 `<smbios mode='host'/>`
- 在 `<features>` 里，增加

```xml
<kvm>
    <hidden state='on'/>
</kvm>
```

- 在 `<hyperv>` 标签里，删掉其他的内容，增加 `<vendor_id state='on' value='1234567890ab'/>`

然后保存配置，启动虚拟机。打开虚拟机之后，在Windows里安装驱动，或者点击驱动管理，让 Windows 10 自己安装驱动，基本上就搞定了。

## 穿透声卡和USB

按照上面的步骤，是没有声音的，有两个方案：

- 继续查看声卡和USB是否是多个IOMMU组，我的是，所以我直接穿透了，虚拟机关机的时候，宿主机照用，虚拟机开机的时候，虚拟机占用
- 仅穿透USB，或者选择 `Add Hardware` -> `USB Host Device` 然后把具体的声音输入输出、USB传进去

## 最后

最后我们要做什么呢？前文说到，我们如果有两个显示器的话，就比较方便，每个显卡上插一个显示器，如果没有的话，就只能把HDMI
线插到你要用的那个显卡上，或者有的显示器支持两个输入线，然后在显示器的设置里切换信号源。

对于键鼠，如果不想用两套的话，可以在 VirtManager 里分别把USB设备传进去，又或者可以参考 [这一篇文章](https://jiajunhuang.com/articles/2021_11_26-use_barrier.md.html)
使用 barrier 来共享一套键鼠。

大功告成！接下里就是愉快的折腾和玩耍了。

---

参考资料：

- https://wiki.archlinux.org/title/PCI_passthrough_via_OVMF
- https://gist.github.com/davesilva/445276f9157e7cb3a4f6ed2fe852b340
