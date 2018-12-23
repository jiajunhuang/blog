# Docker组件介绍（一）：runc和containerd

> TL;DR: 主要介绍了Docker的各个组件：`runc`, `containerd`, `shim`, `docker-init`, `docker-proxy`。

最近在研究Docker，为 [我自己的容器编排系统](https://github.com/jiajunhuang/huang) 做知识储备工作。初次接触到Docker是在大二的时候，
现在Docker已经改的面目全非了，当初单纯的Docker现在被拆的七零八散，但尽管如此，Docker仍然是容器界垄断地位，提容器必提Docker。Rkt之后
我也会研究一下，不过这篇文章，主要还是看看Docker现在的组成。首先，安装Docker，在我的 Ubuntu 18.04 里执行 `sudo apt install docker.io`
即可，然后就可以看到 `/usr/bin/` 下Docker相关的组件：

```bash
$ ls /usr/bin/docker*
/usr/bin/docker             /usr/bin/docker-containerd-ctr   /usr/bin/dockerd      /usr/bin/docker-proxy
/usr/bin/docker-containerd  /usr/bin/docker-containerd-shim  /usr/bin/docker-init  /usr/bin/docker-runc
```

是不是很多？首先 `/usr/bin/docker` 和 `/usr/bin/dockerd` 就是命令行客户端和daemon，Docker的架构是 `client-server` 模式的：

![docker architecture](./img/docker_arch.jpeg)

这两个，我们就不看了。我们重点看看 `docker-containerd`, `docker-containerd-ctr`, `docker-containerd-shim`, `docker-init`, `docker-proxy`, `docker-runc` 是干什么的，其实最简单的方式，就是加个命令行参数 `--help` 看看他们的简介：

```bash
$ docker-containerd --help
NAME:
   containerd -
                    __        _                     __
  _________  ____  / /_____ _(_)___  ___  _________/ /
 / ___/ __ \/ __ \/ __/ __ `/ / __ \/ _ \/ ___/ __  /
/ /__/ /_/ / / / / /_/ /_/ / / / / /  __/ /  / /_/ /
\___/\____/_/ /_/\__/\__,_/_/_/ /_/\___/_/   \__,_/

high performance container runtime

$ docker-containerd-ctr --help
NAME:
   ctr -
        __
  _____/ /______
 / ___/ __/ ___/
/ /__/ /_/ /
\___/\__/_/

containerd CLI

$ docker-containerd-shim --help
Usage of docker-containerd-shim:
...

$ docker-init --help
docker-init: invalid option -- '-'
docker-init (tini version 0.18.0)
...
$ docker-proxy --help
Usage of docker-proxy:
  -container-ip string
  ...

$ docker-runc --help
NAME:
   runc - Open Container Initiative runtime

runc is a command line client for running applications packaged according to
the Open Container Initiative (OCI) format and is a compliant implementation of the
Open Container Initiative specification.

runc integrates well with existing process supervisors to provide a production
container runtime environment for applications. It can be used with your
existing process monitoring tools and the container will be spawned as a
direct child of the process supervisor.

Containers are configured using bundles. A bundle for a container is a directory
that includes a specification file named "config.json" and a root filesystem.
The root filesystem contains the contents of the container.

To start a new instance of a container:

    # runc run [ -b bundle ] <container-id>

Where "<container-id>" is your name for the instance of the container that you
are starting. The name you provide for the container instance must be unique on
your host. Providing the bundle directory using "-b" is optional. The default
value for "bundle" is the current directory.
```

可以看出来，`docker-init`, `docker-containerd-shim` 和 `docker-proxy` 没有在帮助里告诉我们是干啥的，其他的都有。

- docker-containerd: 高性能容器运行时
- docker-containerd-ctr: docker-containerd的命令行客户端
- docker-runc: 运行容器的命令行工具

如果去搜索一番，就会发现：`docker-containerd` 就是 [containerd](https://github.com/containerd/containerd)，而 `docker-runc`
就是 [runc](https://github.com/opencontainers/runc)。containerd是真正管控容器的daemon，执行容器的时候用的是runc。为什么
要分的七零八散呢？我估计其中主要的原因是防止Docker垄断，因此把容器标准独立出来，就有了 [runtime-spec](https://github.com/opencontainers/runtime-spec)，然后有了runc，然后有了containerd(此处发展历史没有考究，并不关心)。

看一张图(来自 containerd 官网)：

![containerd architecture](./img/containerd_architecture.png)

可以看出来，Docker本身其实已经被剥离干净了，只剩下Docker自身的一些特色功能了，真正容器的管控都在containerd里实现。
所以接下来介绍的顺序是 `runc`, `containerd`, `shim`, `docker-init`, `docker-proxy`。今天是第一篇，介绍 `runc` 和 `containerd`。

## runc

runc是标准化的产物，为了防止一家商业公司主导容器化标准，因此又了opencontainers组织，因此，创建容器，其实最终通过runc就可以了。下面我们来看看例子：

```bash
# 如果要自己编译，就要先执行 go get github.com/opencontainers/runc，然后cd到对应目录
jiajun@ubuntu:~/go/src/github.com/opencontainers/runc$ pwd
/home/jiajun/go/src/github.com/opencontainers/runc
jiajun@ubuntu:~/go/src/github.com/opencontainers/runc$ make
go build -buildmode=pie  -ldflags "-X main.gitCommit="f5b99917df9fbe1d9a4114966fb088dd6860e85a" -X main.version=1.0.0-rc6+dev " -tags "seccomp" -o runc .
```

如果要自己编译，就要先克隆代码，然后执行 `sudo make install`。然后我们用runc创建一个容器试试：

```bash
jiajun@ubuntu:~$ mkdir -p mycontainer/rootfs
jiajun@ubuntu:~$ cd mycontainer/
jiajun@ubuntu:~/mycontainer$ docker export $(docker create busybox) | tar -C rootfs -xf -
jiajun@ubuntu:~/mycontainer$ ls
rootfs
jiajun@ubuntu:~/mycontainer$ runc spec
jiajun@ubuntu:~/mycontainer$ runc run mycontainer
rootless container requires user namespaces
jiajun@ubuntu:~/mycontainer$ sudo runc run mycontainer
/ # exit
```

可以看出来，默认情况下，runc是要root用户才能执行的，对比了一下执行 `runc spec` 和 `runc spec --rootless` 生成的 `config.json`，
原因是 `runc spec` 生成的 `config.json` 默认会挂载 `/sys` 下的东西，此外使用了user这个命名空间把容器里的root用户和容器外的非root用户对应起来：

```json
"namespaces": [
    {
        "type": "pid"
    },
    {
        "type": "ipc"
    },
    {
        "type": "uts"
    },
    {
        "type": "mount"
    },
    {
        "type": "user"
    }
],
```

如果想要mycontainer在后台执行，那么需要修改 `config.json`, 把 `terminal: true` 改成 `terminal: false`，此外，还要修改
`args` 的值，使得容器执行的命令不含需要终端的命名：

```json
{
	"ociVersion": "1.0.1-dev",
	"process": {
		"terminal": false,
		"user": {
			"uid": 0,
			"gid": 0
		},
		"args": [
			"sleep", "50"
		],
```

修改后执行：

```bash
jiajun@ubuntu:~/mycontainer$ runc run -d mycontainer
jiajun@ubuntu:~/mycontainer$ runc list
ID            PID         STATUS      BUNDLE                     CREATED                          OWNER
mycontainer   14035       running     /home/jiajun/mycontainer   2018-12-23T07:12:24.721222193Z   jiajun
```

如果我们这个时候使用 `docker-runc` 执行一下 `list`，就会发现也能列出一样的结果：

```bash
jiajun@ubuntu:~/mycontainer$ runc list
ID            PID         STATUS      BUNDLE                     CREATED                          OWNER
mycontainer   0           stopped     /home/jiajun/mycontainer   2018-12-23T07:12:24.721222193Z   jiajun
jiajun@ubuntu:~/mycontainer$ docker-runc list
ID            PID         STATUS      BUNDLE                     CREATED                          OWNER
mycontainer   0           stopped     /home/jiajun/mycontainer   2018-12-23T07:12:24.721222193Z   jiajun
```

说明，runc在某个地方存储了执行的容器的信息，读完帮助之后，发现有一个参数是用来配置这个的：

```bash
jiajun@ubuntu:~/mycontainer$ runc --help
...
--root value        root directory for storage of container state (this should be located in tmpfs) (default: "/run/user/1000/runc")
...
jiajun@ubuntu:~/mycontainer$ tree /run/user/1000/runc/
/run/user/1000/runc/
└── mycontainer
    └── state.json

1 directory, 1 file
```

如果重启一下，就会发现这些信息都没了，因为 `/run` 下面的文件不是持久化的，都是在内存里的：

```bash
jiajun@ubuntu:~/mycontainer$ df -h
Filesystem      Size  Used Avail Use% Mounted on
udev            956M     0  956M   0% /dev
tmpfs           197M  1.3M  196M   1% /run
/dev/sda2        20G  6.2G   13G  33% /
tmpfs           985M  8.0K  985M   1% /dev/shm
tmpfs           5.0M     0  5.0M   0% /run/lock
tmpfs           985M     0  985M   0% /sys/fs/cgroup
/dev/loop0       89M   89M     0 100% /snap/core/5897
/dev/loop1       79M   79M     0 100% /snap/go/3095
/dev/loop2       90M   90M     0 100% /snap/core/6034
/dev/loop3       90M   90M     0 100% /snap/core/6130
tmpfs           197M  8.0K  197M   1% /run/user/1000
```

## containerd

接下来我们看看 `containerd` 是干啥的。我们用 `docker` 来运行一个容器：

```bash
jiajun@ubuntu:~$ docker run -d busybox sleep 100
61ff791b4e5b3d9a98e739a49f2e4d9371bcda2f704aed702af51dac24c83c7a
jiajun@ubuntu:~$ ps axjf | grep -A 3 dockerd
     1    955    955    955 ?            -1 Ssl      0   0:28 /usr/bin/dockerd -H fd://
   955   1175   1175   1175 ?            -1 Ssl      0   0:36  \_ docker-containerd --config /var/run/docker/containerd/containerd.toml
  1175  15001  15001   1175 ?            -1 Sl       0   0:00      \_ docker-containerd-shim -namespace moby -workdir /var/lib/docker/containerd/daemon/io.containerd.runtime.v1.linux/moby/eaa0ec8c3459e2ed4fda7bd532a996bc6d4e9342e2a47d6fa0b8bc9dfc1e7553 -address /var/run/docker/containerd/docker-containerd.sock -containerd-binary /usr/bin/docker-containerd -runtime-root /var/run/docker/runtime-runc
 15001  15028  15028  15028 ?            -1 Ss       0   0:00      |   \_ sleep 100
```

也就是说，`dockerd` 有个子进程，是 `containerd`，然后 `containerd` 有子进程，就是上面busybox容器里的 `sleep`。

我们来看看 `containerd` 的架构图：

![containerd architecture](./img/containerd_architecture.png)

从 [官方仓库](https://github.com/containerd/containerd) 的描述可以看出来，其实 `containerd` 就包含了我们常用的 `docker` 的命令：

- 增删查改容器
- 增删查改镜像

也就是说，如果我们要对容器进行操控，直接使用 `containerd` 其实就够了，那我们试试看：

```bash
jiajun@ubuntu:~$ sudo docker-containerd-ctr --address=/var/run/docker/containerd/docker-containerd.sock images pull docker.io/library/busybox:latest
docker.io/library/busybox:latest:                                                 resolved       |++++++++++++++++++++++++++++++++++++++|
index-sha256:2a03a6059f21e150ae84b0973863609494aad70f0a80eaeb64bddd8d92465812:    done           |++++++++++++++++++++++++++++++++++++++|
manifest-sha256:915f390a8912e16d4beb8689720a17348f3f6d1a7b659697df850ab625ea29d5: done           |++++++++++++++++++++++++++++++++++++++|
layer-sha256:90e01955edcd85dac7985b72a8374545eac617ccdddcc992b732e43cd42534af:    done           |++++++++++++++++++++++++++++++++++++++|
config-sha256:59788edf1f3e78cd0ebe6ce1446e9d10788225db3dedcfd1a59f764bad2b2690:   done           |++++++++++++++++++++++++++++++++++++++|
elapsed: 9.1 s                                                                    total:  627.4  (68.9 KiB/s)
unpacking sha256:2a03a6059f21e150ae84b0973863609494aad70f0a80eaeb64bddd8d92465812...
done
jiajun@ubuntu:~$ sudo docker-containerd-ctr --address=/var/run/docker/containerd/docker-containerd.sock images list
REF                              TYPE                                                      DIGEST                                                                  SIZE      PLATFORMS                                                                                             LABELS
docker.io/library/busybox:latest application/vnd.docker.distribution.manifest.list.v2+json sha256:2a03a6059f21e150ae84b0973863609494aad70f0a80eaeb64bddd8d92465812 715.5 KiB linux/386,linux/amd64,linux/arm/v5,linux/arm/v6,linux/arm/v7,linux/arm64/v8,linux/ppc64le,linux/s390x -
```

这里需要说明两点：

- 加参数 `--address=/var/run/docker/containerd/docker-containerd.sock` 是因为 `containerd` 默认不是监听这个路径，而我没有单独起一个containerd，而是使用了 `docker-containerd`，通过 `ps aux | grep docker` 发现它使用了 `/var/run/docker/containerd/containerd.toml` 这个配置文件，而监听路径就写在里面
- 使用sudo的原因是，jiajun这个用户没有root权限，无法连上对应的socket

拉取了镜像，还需要运行一下才行：

```bash
jiajun@ubuntu:~$ sudo docker-containerd-ctr --address=/var/run/docker/containerd/docker-containerd.sock run -t docker.io/library/busybox:latest my_busybox_demo sh
/ # ls
bin   dev   etc   home  proc  root  run   sys   tmp   usr   var
/ # exit
jiajun@ubuntu:~$ sudo docker-containerd-ctr --address=/var/run/docker/containerd/docker-containerd.sock run -d docker.io/library/busybox:latest my_busybox_sleep_demo sleep 10000
jiajun@ubuntu:~$ sudo docker-containerd-ctr --address=/var/run/docker/containerd/docker-containerd.sock container list
CONTAINER                IMAGE                               RUNTIME
my_busybox_demo          docker.io/library/busybox:latest    io.containerd.runtime.v1.linux
my_busybox_sleep_demo    docker.io/library/busybox:latest    io.containerd.runtime.v1.linux
```

这篇文章就到这里，我们介绍了 `runc` 和 `containerd`。接下来的文章会介绍：

- shim
- docker-init
- docker-proxy

---

- https://hackernoon.com/docker-containerd-standalone-runtimes-heres-what-you-should-know-b834ef155426
- http://alexander.holbreich.org/docker-components-explained/
- https://github.com/opencontainers/runc
- https://github.com/containerd/containerd
- https://segmentfault.com/a/1190000011294361
