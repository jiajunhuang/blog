# VirtManager Windows自适应屏幕

此前折腾过一次，最近在折腾btrfs的时候，因为磁盘大小不够，把原本的虚拟机给删掉了，重新装一个，
因此又遇到了这个问题，因此记录一下解决方案。

问题：使用Virt Manager安装Windows 7，并且在Windows 7中安装了virtio guest tools之后，Windows 7仍然
无法自适应屏幕。

解决方案：Graphics选 `spice`，且添加一个Channel，值选 `com.redhat.spice.0`，重启。
