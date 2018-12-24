# Docker组件介绍（二）：shim, docker-init和docker-proxy

[上一篇](https://jiajunhuang.com/articles/2018_12_22-docker_components.md.html) 文章中，我们简单的介绍了 `runc`和`containerd`。这一篇文章中，
我们分别看看 `docker-containerd-shim`, `docker-init` 和 `docker-proxy` 的作用。

## docker-containerd-shim

shim的翻译是垫片，就是修自行车的时候，用来夹在螺丝和螺母之间的小铁片。关于shim本身，网上介绍的文章很少，但是作者在 Google Groups
里有解释到shim的作用：

> https://groups.google.com/forum/#!topic/docker-dev/zaZFlvIx1_k

- 允许runc在创建&运行容器之后退出
- 用shim作为容器的父进程，而不是直接用containerd作为容器的父进程，是为了防止这种情况：当containerd挂掉的时候，shim还在，因此可以保证容器打开的文件描述符不会被关掉
- 依靠shim来收集&报告容器的退出状态，这样就不需要containerd来wait子进程

因此，使用shim的主要作用，就是将containerd和真实的容器（里的进程）解耦，这是第二点和第三点所描述的。而第一点，为什么要允许runc退出呢？
因为，Go编译出来的二进制文件，默认是静态链接，因此，如果一个机器上起N个容器，那么就会占用M*N的内存，其中M是一个runc所消耗的内存。
但是出于上面描述的原因又不想直接让containerd来做容器的父进程，因此，就需要一个比runc占内存更小的东西来作父进程，也就是shim。但实际上，
shim仍然比较占内存（[参考这里](https://github.com/moby/moby/issues/21737)），因此，比较好的方式是：

- 用C重写并且默认使用动态链接库
- 打开Go的动态链接支持然后重新编译

## docker-init

我们都知道UNIX系统中，1号进程是init进程，也是所有孤儿进程的父进程。而使用docker时，如果不加 `--init` 参数，容器中的1号进程
就是所给的ENTRYPOINT，例如下面例子中的 `sh`。而加上 `--init` 之后，1号进程就会是 [tini](https://github.com/krallin/tini)：

```bash
jiajun@ubuntu:~$ docker run -it busybox sh
/ # ps aux
PID   USER     TIME  COMMAND
    1 root      0:00 sh
    6 root      0:00 ps aux
/ # exit
jiajun@ubuntu:~$ docker run -it --init busybox sh
/ # ps aux
PID   USER     TIME  COMMAND
    1 root      0:00 /dev/init -- sh
    6 root      0:00 sh
    7 root      0:00 ps aux
/ # exit
```

## docker-proxy

我猜测这个是用来做端口映射的，因为------名字里有proxy嘛，还能用来干啥，因此就验证一下：

```bash
jiajun@ubuntu:~$ docker run -d -p 10010:10010 busybox sleep 10000
be88279118ad7f8cfd3d418db00872aa4f3b1753278b67c28727f16d68f37ae5
jiajun@ubuntu:~$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS                      NAMES
be88279118ad        busybox             "sleep 10000"       2 seconds ago       Up 1 second         0.0.0.0:10010->10010/tcp   awesome_jackson
jiajun@ubuntu:~$ ps aux | grep docker
root        897  0.1  3.8 736592 78444 ?        Ssl  06:20   0:33 /usr/bin/dockerd -H fd://
root       1188  0.0  1.8 665876 37964 ?        Ssl  06:20   0:25 docker-containerd --config /var/run/docker/containerd/containerd.toml
root       5579  0.0  0.1 378868  3076 ?        Sl   14:57   0:00 /usr/bin/docker-proxy -proto tcp -host-ip 0.0.0.0 -host-port 10010 -container-ip 172.17.0.2 -container-port 10010
root       5585  0.0  0.1   7376  3808 ?        Sl   14:57   0:00 docker-containerd-shim -namespace moby -workdir /var/lib/docker/containerd/daemon/io.containerd.runtime.v1.linux/moby/be88279118ad7f8cfd3d418db00872aa4f3b1753278b67c28727f16d68f37ae5 -address /var/run/docker/containerd/docker-containerd.sock -containerd-binary /usr/bin/docker-containerd -runtime-root /var/run/docker/runtime-runc
jiajun     5666  0.0  0.0  13136  1076 pts/0    S+   14:57   0:00 grep --color=auto docker
```

可以看到这么一行 `/usr/bin/docker-proxy -proto tcp -host-ip 0.0.0.0 -host-port 10010 -container-ip 172.17.0.2 -container-port 10010`，其底层是使用iptables来完成的，参考：https://windsock.io/the-docker-proxy/。

---

- https://windsock.io/the-docker-proxy/
