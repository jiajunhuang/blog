# podman 无故退出

最近把 docker 删了，换成了 podman，但是遇到一件非常诡异的事情：容器启动以后，加上了 `--restart=always`，加上了 `-d`，
但是总会发现，容器会无缘无故退出，而且是一堆容器一起退出，`inspect` 的时候又发现 `ExitCode` 为 0，相当诡异。

最终发现，原因是：我是以普通用户的身份来运行的，也就是 rootless 模式，默认情况下，systemd 会在用户退出的时候，把会话中
的进程都杀掉以节省资源：验证：

```bash
$ loginctl list-users
```

如果输出 NO，说明没有打开 systemd 的 LINGER 的功能。

> systemd 的 LINGER 功能用于控制用户进程在用户注销（退出登录）后是否继续运行。它的核心作用是管理用户级服务的生命周期，即使用户已经退出登录，也能保持某些服务或进程在后台运行。
> LINGER 的作用
> 默认情况下，当用户注销（如关闭 SSH 连接、退出图形会话）时，systemd 会终止该用户的所有进程（包括用户级服务）。这是为了防止未使用的进程占用资源。
> 当为用户启用 LINGER（loginctl enable-linger <username>）后，即使用户注销，其用户级服务（通过 systemd --user 管理的服务）仍会继续运行。这些服务将由系统级的 systemd 实例（PID 1）接管，而不是依赖于用户会话。

## 启用 systemd linger

```bash
$ loginctl enable-linger
```

然后就好了，退出登录以后，容器就不会退出了。
