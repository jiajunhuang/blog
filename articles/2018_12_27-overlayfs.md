# 耍耍OverlayFS

自Linux3.4之后，就可以使用overlay了。来，耍耍。

```bash
jiajun@ubuntu:~/playoverlay$ mkdir lower{1,2,3} upper work merged  # 创建文件夹
jiajun@ubuntu:~/playoverlay$ touch lower1/l1.txt lower2/l2.txt lower3/l3.txt  # 创建文件
jiajun@ubuntu:~/playoverlay$ tree  # 这是目前的结构
.
├── lower1
│   └── l1.txt
├── lower2
│   └── l2.txt
├── lower3
│   └── l3.txt
├── merged
├── upper
└── work

6 directories, 3 files
jiajun@ubuntu:~/playoverlay$ sudo mount -t overlay overlay -o lowerdir=./lower1:./lower2:./lower3,upperdir=./upper,workdir=./work ./merged/ # 挂载
jiajun@ubuntu:~/playoverlay$ tree  # 这是目前的结构
.
├── lower1
│   └── l1.txt
├── lower2
│   └── l2.txt
├── lower3
│   └── l3.txt
├── merged
│   ├── l1.txt
│   ├── l2.txt
│   └── l3.txt
├── upper
└── work
    └── work [error opening dir]

7 directories, 6 files
jiajun@ubuntu:~/playoverlay$ touch merged/hahahahhha  # 创建一个文件
jiajun@ubuntu:~/playoverlay$ rm merged/l1.txt  # 删除一个文件
jiajun@ubuntu:~/playoverlay$ tree
.
├── lower1
│   └── l1.txt
├── lower2
│   └── l2.txt
├── lower3
│   └── l3.txt
├── merged
│   ├── hahahahhha
│   ├── l2.txt
│   └── l3.txt
├── upper
│   ├── hahahahhha
│   └── l1.txt
└── work
    └── work [error opening dir]

7 directories, 8 files
jiajun@ubuntu:~/playoverlay$
```

解释两点：

- `sudo mount -t overlay overlay -o lowerdir=./lower1:./lower2:./lower3,upperdir=./upper,workdir=./work ./merged/`:
    - lowerdir后边，冒号分割的目录，深度是从右往左挂载的，也就是说，越在右边的目录，越在深处
    - workdir是overlayfs实现要用的一个目录，我也不知道为啥要有这玩意儿
    - upperdir就是可写层，所有的修改都会保存在里面
    - 最后的 `./merged/` 就是最终可以看到的overlay目录，我们在这里面操作，所有的操作都会保存在upper这个文件夹里
- 如果我们直接删除lower2里的文件会怎样呢？

    ```bash
    jiajun@ubuntu:~/playoverlay$ tree
    .
    ├── lower1
    │   └── l1.txt
    ├── lower2
    │   └── l2.txt
    ├── lower3
    │   └── l3.txt
    ├── merged
    │   ├── hahahahhha
    │   ├── l2.txt
    │   └── l3.txt
    ├── upper
    │   ├── hahahahhha
    │   └── l1.txt
    └── work
        └── work [error opening dir]

    7 directories, 8 files
    jiajun@ubuntu:~/playoverlay$ rm lower2/l2.txt
    jiajun@ubuntu:~/playoverlay$ tree
    .
    ├── lower1
    │   └── l1.txt
    ├── lower2
    ├── lower3
    │   └── l3.txt
    ├── merged
    │   ├── hahahahhha
    │   └── l3.txt
    ├── upper
    │   ├── hahahahhha
    │   └── l1.txt
    └── work
        └── work [error opening dir]

    7 directories, 6 files
    jiajun@ubuntu:~/playoverlay$
    ```

    当然是上边就看不到了。

---

实现，就去看下边的 overlay.txt，用法，就看ArchWiki和Docker那个文档，就不赘述了。另外：

- 可以使用mount打印出当前挂载的overlayfs
- 用Docker起容器之后，可以用mount打印出挂载的overlayfs，然后探索探索Docker是怎么使用overlayfs的

---

- https://docs.docker.com/storage/storagedriver/overlayfs-driver/
- https://www.kernel.org/doc/Documentation/filesystems/overlayfs.txt
- https://wiki.archlinux.org/index.php/Overlay_filesystem
