# FreeBSD 使用 vm-bhyve 安装Debian虚拟机

首先需要安装 `bhyve` 和 `vm-bhyve`:

```bash
# pkg install vm-bhyve bhyve
```

加载对应的内核：

```bash
kldload vmm
kldload nmdm
kldload if_bridge if_tap
```

初始化，由于我不是使用ZFS，而是使用的UFS，所以命令如下:

```bash
# sysrc vm_enable="YES"
# mkdir -p /data/vms
# sysrc vm_dir="/data/vms"
# vm init
# cp /usr/local/share/examples/vm-bhyve/* /data/vms/.templates/
```

如果是ZFS，则是：

```bash
# zfs create pool/vm
# sysrc vm_enable="YES"
# sysrc vm_dir="zfs:pool/vm"
# vm init
# cp /usr/local/share/examples/vm-bhyve/* /mountpoint/for/pool/vm/.templates/
```

初始化网络：

```bash
# vm switch create public
# vm switch add public <你的网卡>
```

把ISO镜像传到对应目录：`scp debian-xxx.iso freebsd:/data/vms/.iso/`，记住ISO文件名。

创建并启动Debian虚拟机：

```bash
# vm create -t debian -s 20G debian
# vm install debian debian-12.1.0-amd64-netinst.iso
```

然后VNC连接到FreeBSD的5900端口安装。安装完，如果重启Debian，则会遇到报错：

```
BdsDxe: failed to load Boot0001 "UEFI BHYVE SATA DISK BHYVE-48FF-992B-D5E0" from PciRoot(0x0)/Pci(0x4,0x0)/Sata(0x0,0xFFFF,0x0): Not Found
>>Start PXE over IPv4.
```

这样的报错，出现问题的原因是 FreeBSD 去找EFI文件的路径，但是没找到，Debian默认安装到别的地方去了，修复方式如下：

```bash
Shell>FS0:
FS0:\>ls
Directory of: FS0:\
FS0:\>cd EFI\debian
FS:\EFI\debian\>grubx64.efi
```

然后就可以启动进入了Debian，然后需要修正一下EFI的路径，在Debian中，以root身份执行：

```bash
# mkdir /boot/efi/EFI/BOOT
# cp /boot/efi/EFI/debian/grubx64.efi /boot/efi/EFI/BOOT/bootx64.efi
```

就这样就修复完成了，以后就可以正常启动。

---

参考资料：

- https://npulse.net/en/blog/125-bhyve-uefi-drops-into-efi-shell-linux-wont-boot-easy-workaround
- https://www.davidschlachter.com/misc/freebsd-bhyve-uefi-shell
