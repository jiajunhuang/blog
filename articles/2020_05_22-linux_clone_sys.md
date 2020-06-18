# Linux系统迁移记录(从HDD到SSD)

我把HDD上的Linux迁移到SSD上，重装系统太麻烦了，所以我直接拷贝整个系统，然后重建引导恢复，以下是记录。

## 两块硬盘

首先把SSD换上去，HDD用硬盘盒连接。

从U盘启动后，将SSD分区做好，分别将SSD系统盘和HDD系统盘挂载到某个路径，例如：

对于SSD，我分为 `/dev/sda1` 和 `/dev/sda2`，前者用于EFI引导，后者为系统盘。

- SSD系统盘挂载至 /data
- HDD系统盘挂载至 /mnt

然后用rsync同步数据：

```bash
# rsync -av --progress /mnt/ /data/
```

等待同步完成之后，进入 /data/etc 编辑 `fstab` 文件，将原有路径替换为新的。可以使用 blkid 或者 lsblk -f来查看UUID。

接下来就是重建引导。

## 一块硬盘

> 这一段加于2020.06.18

如果是一块硬盘，想要做到换个文件系统之类的，比如我就从btrfs切换到了ext4，方案如下：

- 首先看下当前数据有多少，把硬盘一分为二，留出足够的大小
- 给新分区创建好文件系统
- 进入live系统，将btrfs分区挂载至 /mnt，新分区挂载至 /opt
- 将 /mnt 压缩并将文件保存到 /opt：`cd /opt && tar cvf data.tgz /mnt`
- 将btrfs所在分区卸载，并且重建文件系统：`sudo umount /mnt && mkfs.ext4 /dev/<btrfs文件系统所在盘符例如我的是sda2>`
- 将格式化后的分区重新挂载到/mnt 下，并将 /opt 下压缩文件解压回去：`tar xvf data.tgz -C /`

等待解压完成之后，进入 /mnt/etc 编辑 `fstab` 文件，将原有路径替换为新的。可以使用 blkid 或者 lsblk -f来查看UUID。

接下来就是重建引导。

## 更新grub

更新grub之前要先挂载一些东西，否则会报错：

```bash
$ sudo mount --bind /dev /data/dev
$ sudo mount --bind /sys /data/sys
$ sudo mount --bind /proc /data/proc
$ sudo mount /dev/sda1 /data/boot/efi
```

然后chroot进去，并且更新EFI引导：

```bash
$ sudo chroot /data
$ sudo apt-get install --reinstall grub-efi
$ sudo grub-install /dev/sda
$ sudo update-grub
```

## 为SSD优化文件挂载选项

由于我用的ext4，为SSD优化，更改fstab，加入以下选项 `discard,noatime`，所以大概长这样：

```bash
UUID=刚才更新的UUID /               ext4    discard,noatime,errors=remount-ro 0       1
```

## 如果启动遇到 Gave up waiting for suspend/resume device

首先检查fstab是不是有问题，如果有的话，改之；然后检查 `/etc/initramfs-tools/conf.d/resume` 的UUID有没有问题，有的话，改之。

然后执行 `sudo update-initramfs -u` 之后重启。

搞定！这样比重装系统还是快多了，懒得重装然后重新配置。

---

参考资料：

- https://lists.debian.org/debian-user/2017/09/msg00866.html
- https://wiki.debian.org/GrubEFIReinstall
- https://askubuntu.com/questions/145241/how-do-i-run-update-grub-from-a-livecd
- https://wiki.debian.org/SSDOptimization#A.2Fetc.2Ffstab_example
