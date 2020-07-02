# 拯救删除ZFS之后的分区表

本想安装在笔记本上安装FreeBSD+Linux双系统，不过总是引导不起来，遂放弃。删除分区表之后，准备调整分区大小，结果发现
gparted里，显示的只有zfs一个分区，但是lsblk却又是正常的。经过搜索发现是ZFS写入了metainfo，于是就看怎么删除。

删除是这样的：

```bash
# zpool clearlabel /dev/sda
```

然而，这会带来一个严重后果：分区表被破坏了。我试着用Linux的live磁盘进去，但是Linux已经认不出来了。一般Linux的live系统
不带gpart这个程序，因此我用FreeBSD live系统进去，发现还可以认出来：

```bash
# gpart disk list
...
```

不过输出里，会显示 `GPT Corrupt`，幸好gpart特别强大，可以直接修复：

```bash
# gpart recover /dev/sda
```

呼，搞定，有惊无险，系统数据得以保存。要是真的把分区表给完全摧毁了，数据就没了。

---

Refs:

- https://forums.freebsd.org/threads/gpt-table-corrupt.52102/
- https://unix.stackexchange.com/questions/270595/how-do-i-expand-a-file-system-to-fill-a-partition
- https://forums.freebsd.org/threads/gpt-rejected-how-to-wipe-for-zfs.54187/
