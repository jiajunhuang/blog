# KVM 共享 Intel 集成显卡

我比较喜欢折腾虚拟化，最近在折腾的就是这个：把宿主机的 Intel 集成显卡 给 Windows 虚拟机使用。Intel 平台做这个事情非常
方便，因为官方有 Intel GVT-g 项目，就是用来把宿主机生成多个虚拟 GPU 给虚拟机用的。

## 共享集成显卡

1. 首先我们需要安装 `virt-manager`
2. 编辑宿主机 `/etc/default/grub` 文件，在 `GRUB_CMDLINE_LINUX` 中，添加 `i915.enable_gvt=1 i915.enable_fbc=0`
3. 重新生成GRUB 配置 `sudo update-grub`
4. 编辑 `/etc/modules-load.d/kvm-gvt-g.conf` 配置文件，内容为：

```
kvmgt
vfio-iommu-type1
mdev
```

5. 宿主机安装Mesa库 `apt-get install mesa-utils`
6. 重启宿主机，执行 `lspci | grep VGA`，看看输出是否有 `00:02.0 VGA compatible controller: Intel Corporation UHD Graphics 630 (Mobile) (rev 02)` 类似这么一行
7. 进入文件夹 `cd /sys/bus/pci/devices/0000\:00:02.0`
8. 执行 `ls -l mdev_supported_types/` 看是否有输出，每一个都是一种GPU的类型，数字越大，显存越小，可以通过执行 `cat /sys/devices/pci0000\:00/0000\:00\:02.0/mdev_supported_types/<类型>/description` 查看每种类型的描述
9. 选择一个类型，生成一个uuid，创建一个虚拟GPU

```bash
# uuidgen
945248c6-8d80-44a3-9a19-d85e22b19a7f

# echo 945248c6-8d80-44a3-9a19-d85e22b19a7f | tee mdev_supported_types/i915-GVTg_V5_4/create
```

10. 上述配置每次重启都会丢失，为了让宿主机每次启动都生成上述GPU，编辑一个systemd文件 `/etc/systemd/system/setup-gvt.service`

```
[Unit]
Description=Setup GVT

[Service]
Type=oneshot
ExecStart=/usr/bin/bash -c 'echo 945248c6-8d80-44a3-9a19-d85e22b19a7f > /sys/devices/pci0000:00/0000:00:02.0/mdev_supported_types/i915-GVTg_V5_4/create'

[Install]
WantedBy=multi-user.target
```

11. 安装Windows 10虚拟机，记得安装 virtio 驱动，这个我在 [KVM 显卡穿透给 Windows](https://jiajunhuang.com/articles/2019_10_08-linux_windows.md.html) 讲过，此处不赘述，也可以看参考文档中第一篇
12. virsh 编辑虚拟机 xml
    - 第一行 `<domain type='kvm'>` 改成 `<domain type='kvm' xmlns:qemu='http://libvirt.org/schemas/domain/qemu/1.0'>`
    - 找到 `<graphics ...>` 那一行，改成

    ```xml
    <graphics type='spice'>
        <listen type='none'/>
        <gl enable='yes' rendernode='/dev/dri/by-path/pci-0000:00:02.0-render'/>
    </graphics>
    ```

    - 找到 xml 中的 video 那一块，原来的内容可能是

    ```xml
    <video>
        <model type='qxl' ram='65536' vram='65536' vgamem='16384' heads='1' primary='yes'/>
        <address type='pci' domain='0x0000' bus='0x00' slot='0x01' function='0x0'/>
    </video>
    ```

    改成

    ```xml
    <video>
        <model type='none'/>
    </video>
    ```

13. 编辑文件，在 `<video>` 后面，增加

```xml
<hostdev mode='subsystem' type='mdev' managed='no' model='vfio-pci' display='on'>
  <source>
    <address uuid='945248c6-8d80-44a3-9a19-d85e22b19a7f'/>
  </source>
  <address type='pci' domain='0x0000' bus='0x06' slot='0x00' function='0x0'/>
</hostdev>
```

14. 启动虚拟机

## Windows RDP 远程桌面配置优化

1. `Windows-R` 输入 gpedit.msc，找到 `Computer Configuration > Policies > Administrative Template > Windows Components > Remote Desktop Services > Remote Desktop Session Host > Connections`
    - "RDP Transfer Protocols" 设置为 True，下面选择 "Use both UDP and TCP"
2. `Computer Configuration > Policies > Administrative Template > Windows Components > Remote Desktop Services > Remote Desktop Session Host > Remote Session Enviorment` 编辑
    - Use hardware graphics adapters for all Remote Desktop Services Sessions = Enabled
    - Prioritize H.264/AVC 444 graphics mode for Remote Desktop Connections = Enabled
    - Configure H.264/AVC Hardware encoding for Remote Desktop Connections = Enabled, Set "Prefer AVC hardware encoding" to "Always attempt"
    - Configure compression for Remote FX data = Enabled, Set RDP compression algorithem: "Do not use an RDP compression algorithm"
    - Configure image quality for RemoteFX Adaptive Graphics = Enabled, Set Image Quality to "High" (lossless seemed too brutal over WAN connections.)
    - Enable RemoteFX encoding for RemoteFX clients designed for Windows Server 2008R2 SP1 = Enabled.
3. `Computer Configuration > Policies&amp;gt;Administrative Template > Windows Components > Remote Desktop Services > Remote Desktop Session Host > Remote Session Enviorment > Remote FX for Windows Server 2008R2`
    - Configure Remote FX = Enabled
    - Optimize visual experience when using Remote FX = Enabled
    - Set Screen capture rate (frames per second) = Highest (best quality)
    - Set Screen Image Quality = Highest (best quality)
    - Optimize visual experience for remote desktop sessions = Enabled
    - Set Visual Experience = Rich Multimedia

4. `Windows-R` 输入 `regedit`，找到 `[HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations]`
    - 新建 Dword(32) 类型，名为 `DWMFRAMEINTERVAL`，值为 `0000000f`

5. 重启

---

参考资料：

- https://blog.tmm.cx/2020/05/15/passing-an-intel-gpu-to-a-linux-kvm-virtual-machine/
- https://wiki.archlinux.org/title/Intel_GVT-g
- https://www.reddit.com/r/sysadmin/comments/fv7d12/pushing_remote_fx_to_its_limits/
