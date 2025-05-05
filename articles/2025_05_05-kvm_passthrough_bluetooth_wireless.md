# KVM 穿透板载蓝牙和无线网卡

我的Host一直是Linux，但是偶尔会打游戏，因此有一台 Windows 虚拟机，最近想要把板载蓝牙和无线网卡穿透进去，但是都遇到了一些
小困难需要解决，因此记录成文。

## 穿透无线网卡

无线网卡穿透起来和穿透显卡等硬件是一样的，由于无线网卡在一个单独的IOMMU组，直接加到 `/etc/modprobe.d/vfio.conf`
中即可，此外，我发现我本地的无线网卡总是被 `igb` 驱动优先占用导致 `vfio-pci` 无法占用网卡，我在 `/etc/modprobe.d/vfio.conf`
中加上：

```
softdep igb pre: vfio-pci
options vfio-pci ids=...无线网卡的硬件ID
```

然后重启以后，就可以穿透了。

## 穿透蓝牙

一开始我以为蓝牙是和无线网卡在一起的，后来发现蓝牙是和USB控制器在一起的，在 `virt-manager` 中，点击 `Add Hardware` - `USB Host Device`,
选择带 `Intl...Bluetooth` 的那个，穿透进去以后，如果直接开机的话，会发现这样的现象：

- Windows 可以发现蓝牙硬件，但是无法使用，安装驱动以后，点开 "我的电脑" 查看硬件详情，驱动里报错 "code: 10"

经过搜索发现，这是 `libvirt` 的一个改动导致的问题，要解决这个问题，还需要编辑XML：

```bash
$ sudo virsh edit --domain windows
```

在 `</domain>` 上面加上 `<qemu:capabilities>` 这一节：

```XML
<domain>
    <devices>
        ...
    </devices>

    <qemu:capabilities>
        <qemu:del capability="usb-host.hostdevice"/>
    </qemu:capabilities>

...
</domain>
```

此时如果保存你会发现无法通过xml校验，还需要跳到顶部，在 `<domain type='kvm'>` 改成

```xml
<domain type='kvm' xmlns:qemu='http://libvirt.org/schemas/domain/qemu/1.0'>
```

然后保存，然后重启以后，就可以穿透蓝牙了。

---

参考文档：

- [KVM显卡穿透](https://jiajunhuang.com/articles/2022_03_15-kvm_windows_gpu_passthrough.md.html)
- [OnBoard Intel Bluetooth Error Code 10 on Windows KVM Guest](https://www.reddit.com/r/VFIO/comments/wbsqy1/how_to_fix_onboard_intel_bluetooth_error_code_10/)
- [PCI passthrough via OVMF]https://wiki.archlinux.org/title/PCI_passthrough_via_OVMF
