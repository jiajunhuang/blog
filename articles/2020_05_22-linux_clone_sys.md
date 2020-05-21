# Linux系统迁移记录(从HDD到SSD)

我把HDD上的Linux迁移到SSD上，重装系统太麻烦了，所以我直接拷贝整个系统，然后重建引导恢复，以下是记录。

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
