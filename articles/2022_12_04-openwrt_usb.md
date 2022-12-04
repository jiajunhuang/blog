# OpenWRT 使用 Android/iOS USB 网络

我所在的地方网络不好，于是计划用 4G/5G 网络，但是开热点有个缺点，那就是WiFi本身新号未必稳定，第二无法让它经过OpenWRT
实现全局科学上网。一个方案是购买 CPE，也就是 4G路由器，或者5G路由器，将手机信号转换成有线信号，然后作为 OpenWRT 的
网络入口。

但是，作为一个折腾党(qiong)，怎么会花 400/1200 去买一个CPE呢？于是，我将目光转向了旧的安卓手机。

> 我试过，iPhone 也可以，不过最后我用旧安卓来做了，因此下文使用安卓作为示例。

由于我的OpenWRT是在虚拟机里，所以我可以直接将手机通过数据线插到主机的USB接口上，然后在virt manager上，将宿主机的USB
设备穿透给虚拟机。如果你使用的是实体刷了OpenWRT的机器，就需要你的机器上有一个USB接口。

然后要做的事情是在 OpenWRT 上安装软件包，用于识别安卓/iOS设备：

```bash
# opkg update
# opkg install kmod-usb-net-rndis kmod-usb-net-cdc-ncm kmod-usb-net-huawei-cdc-ncm kmod-usb-net-cdc-eem kmod-usb-net-cdc-ether kmod-usb-net-cdc-subset kmod-nls-base kmod-usb-core kmod-usb-net kmod-usb-net-cdc-ether kmod-usb2
```

> 直接全装了，因为我是虚拟机，磁盘足够大。如果是路由器，那么请酌情减少，主要还是看你的手机需要哪个包。

如果是 iOS 设备，执行：

```bash
opkg update
opkg install kmod-usb-net-ipheth usbmuxd libimobiledevice usbutils
 
# Call usbmuxd
usbmuxd -v
 
# Add usbmuxd to autostart
sed -i -e "\$i usbmuxd" /etc/rc.local
```

接着就可以在手机上打开USB共享网络，然后在 luci 页面上，增加设备和接口，并且将 `usb0` 设置为 `WAN`。

> 在 Network - Interfaces - Devices 查看是否有 usb0 的网络设备，如果没有，说明没有设置成功

在 Network - Interface - Interfaces 上点击左下角 "Add New Interface"，名字自己选，设备选择 `usb0`。

这个时候，就可以通过安卓设备，用USB数据线传输数据，来进行上网了。

---

ref:

- https://openwrt.org/docs/guide-user/network/wan/smartphone.usb.tethering
