# 删除分区表

之前删除分区表，都是用很傻的方式，例如，gparted一个一个分区删掉，然后最后把分区表删了，或者是fdisk去删除。但是！

读一下 [GPT](https://zh.wikipedia.org/wiki/GUID%E7%A3%81%E7%A2%9F%E5%88%86%E5%89%B2%E8%A1%A8) 的维基百科定义，就会发现，
其实这些信息就是记录在磁盘最开始的512字节里（MBR分区就更小了），而且，如果删除分区表，那么所有的数据都会找不到（普通方式下）。

所以，直接摧毁这512字节就ok了呀！

```bash
jiajun@ubuntu:~$ lsblk
NAME   MAJ:MIN RM  SIZE RO TYPE MOUNTPOINT
loop0    7:0    0 89.5M  1 loop /snap/core/6130
loop1    7:1    0 88.2M  1 loop /snap/core/5897
loop2    7:2    0 89.5M  1 loop /snap/core/6034
loop3    7:3    0 78.8M  1 loop /snap/go/3095
sda      8:0    0   20G  0 disk
├─sda1   8:1    0    1M  0 part
└─sda2   8:2    0   20G  0 part /
sr0     11:0    1  812M  0 rom
jiajun@ubuntu:~$ sudo dd if=/dev/zero of=/dev/sda bs=512 count=1
```
