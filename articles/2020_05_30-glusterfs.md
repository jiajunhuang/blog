# glusterfs 笔记

我在自己的集群里搭了一个分布式文件系统，起初在选型的时候，纠结选择ceph还是glusterfs，因为听ceph多一些，且看到有测评
ceph写入速度快一点，不过后来想了想，对于我的需求，因为我的集群是跨公网的异地节点组成的，带宽非常有限，会直接成为
性能瓶颈，因此选择了glusterfs，因为搭建起来特别简单。

首先当然是在三个节点都安装glusterfs(我不喜欢改hosts文件，所以直接用的IP代替的hostname)：

```bash
$ sudo apt install -y glusterfs-server
```

然后选择其中一个节点(192.168.1.1，其它两个节点分别为192.168.1.2, 192.168.1.3)，执行：

```bash
$ sudo gluster peer probe 192.168.1.2
$ sudo gluster peer probe 192.168.1.3
```

然后分别在其它两个节点上执行 `sudo gluster peer probe 192.168.1.1`。然后就会看到成功的信息，我们来看看集群状态：

```bash
$ sudo gluster peer status
Number of Peers: 2

Hostname: 192.168.1.2
Uuid: <一个UUID>
State: Peer in Cluster (Connected)

Hostname: 192.168.1.3
Uuid: <一个UUID>
State: Peer in Cluster (Connected)
```

接下来要做的事情，就是起一个volume，在三个节点上分别创建想要挂载的目录，最好是一致的，比如：

```bash
$ sudo mkdir -p /data/brick1/gv0
```

然后在一个节点上执行：

```bash
$ sudo gluster volume create gv0 replica 3 192.168.1.1:/data/brick1/gv0 192.168.1.2:/data/brick1/gv0 192.168.1.3:/data/brick1/gv0
$ sudo gluster volume start gv0
$ sudo gluster volume info  # 就可以看到volume的信息
```

由于我并没有给glusterfs一个单独的分区来存储数据，而是直接在现有文件系统上创建，所以上述 `gluster volume create` 的命令
还需要在最后面加一个 `force` 才行。

然后就可以开始使用了，要咋使用呢？挂载之：

```bash
$ sudo mkdir -p /data/sync
$ sudo mount -t glusterfs 192.168.1.1:/gv0 /data/sync
```

接着就可以把 `/data/sync` 当本地盘使用了，然而实际上它是分布式文件系统提供的盘，并且由于我创建volume的时候，选择的类型
是replica为3，也就是每一份数据，都会存储为3份，所以是相当可靠的盘。

```bash
$ sudo dd if=/dev/zero of=a.out bs=1M count=2 && sudo rm a.out
2+0 records in
2+0 records out
2097152 bytes (2.1 MB, 2.0 MiB) copied, 17.0367 s, 123 kB/s
```

瞧，由于我这是跨公网的异地集群，所以最开始考虑性能什么的，完全是多虑了。

## 开机自动挂载

由于我使用tinc组网，开机之后，并不能立刻就连上其它节点，所以不能在 `/etc/fstab` 里写挂载，我的方案是创建 `/etc/systemd/system/glusterfsmounts.service`：

```bash
[Unit]
Description=glusterfs mounting
Requires=glusterd.service

[Service]
Type=simple
RemainAfterExit=true
ExecStartPre=sleep 30
ExecStart=mount -t glusterfs 192.168.1.1:/gv0 /data/sync
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
```

所以在开机以后，等待30秒才会去挂载，有30秒，tinc都已经连接成功了。

## 性能优化

```bash
$ sudo gluster volume set gv0 performance.cache-size 256MB  # 读缓冲区大小，默认是32M
$ sudo gluster volume set gv0 performance.write-behind-window-size: 512MB  # 写缓冲区，默认是1M
```

还有一些参数比如 `performance.io-thread-count`，默认是16，对于我的机器来说够用了（4超线程），所以没调。

## 用途

有了分布式文件系统，那么用途是啥呢？刚才看了，由于是跨网络的异地节点组成的集群，所以写入速度非常慢，我目前主要是用作
备份，于是把我的crontab备份脚本全部改成往本地的 `/data/sync/backup` 目录写即可。

还有一些用途，比如把docker的volume挂载到 `/data/sync/docker` 里，这样可以让他们直接同步，但是需要考虑到一点，即如果
他们是大批量IO的，那么显示不适合这种操作，但是一般来说，自用的Redis和MySQL，还是够用的。

---

参考资料：

- https://docs.gluster.org/en/latest/Quick-Start-Guide/Quickstart/
- https://www.jamescoyle.net/how-to/559-glusterfs-performance-tuning
