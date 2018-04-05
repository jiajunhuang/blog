# 折腾Kubernetes

## 安装

哇，在国内安装k8s可就费劲了，因为你懂的网络的原因，各种镜像拉不下来。。。主要是参考以下资料自己
组建一个k8s集群：

- https://kubernetes.io/docs/setup/independent/install-kubeadm/
- https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/

主要步骤有：

- 禁用swap `sudo swapoff -a` 并且在 `/etc/fstab` 中注释相关swap分区

- 安装Docker，给Docker加代理：

```bash
cat > /etc/systemd/system/docker.service.d/https-proxy.conf << EOF
[Service]
Environment="ALL_PROXY=socks5://192.168.1.21:1083"
EOF
```

- 安装 `kubeadm`, `kubelet` 和 `kubectl`

- 选一个节点当master，给master节点上k8s的配置加上 `--cgroup-driver=cgroupsfs`，前提是 `docker info | grep -i cgroup` 输出是 `cgroupfs`

- 开始执行 `kubeadm init` 等待执行，看它报什么错并且解决对应的错误

- 按照提示，在master节点上，你的普通用户执行：

```bash
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

- 安装网络插件

- 在slave节点上执行 master节点安装成功之后提示的那一行 `kubeadm join ...`

## k8s的概念

k8s的概念还是挺多的，相当于在分布式环境中（多节点），给容器这个东西抽象出一个新的操作系统。我们来
分表了解以下这些个概念：

- Pod: Pod是一个节点上 一个或多个 容器的组合，他们放在一起，叫做一个Pod。他们会共享一个namespace。
- Label: Label就是标签，用于Pod上的一些KV对，例如 `app=nginx`
- Service: Pod会在多个slave之间创建或者销毁，所以肯定是不会有固定的ip了，那如果想要访问Pod里的服务怎么办？这时候
就需要创建service了，service是用来充当一个转发者的角色，自动将打向自己的流量分发到Pod里。
- Namespace: Namespace就是namespace，用来隔离的
- Job: 一次性工作，例如平时我们在UNIX系统中的cronjob
- ReplicaControllers: 例如当设置为 `replica=3` 时，它会负责保证无论何时，集群内的某个Pod都会是三份。它是靠标签来区分Pod的
- ReplicaSet: 和 `ReplicaControllers` 一样的功能，但是标签选择器更加丰富
- DaemonSet: 这个和ReplicaControllers不一样，DaemonSet中的Pod每个节点上都分发一个
- Deployment: 这个和ReplicaSet又差不多，但是区别是会保证镜像的新旧，例如如果标签为 `:latest`，则它里面的镜像永远都是最新的
- Statefulsets: 这种适合用来放数据库这类有状态的应用，比如因为数据库会写磁盘

还有许多其他的概念。。。慢慢熟悉吧，k8s真的有点大
