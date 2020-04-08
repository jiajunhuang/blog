# 解决k8s cron无法读取环境变量的问题

我们使用 `cron` 放一个deployment里，而不是使用k8s自带的cron方案，原因有几个：

- cron经过了历史验证，并且满足要求（虽然精度只到分钟，但是够用了）
- 减少程序员知识负担，大家都对cron和k8s中的deployment比较熟悉
- 多个异步任务的时候cron很好用

但是有一个问题，起初k8s中的cron没法拿到环境变量，但是我直接 `exec` 进入容器却是可以拿到环境变量的，
我们的环境变量是在k8s deployment的yaml描述文件中写入的。

最后的解决方案是，在启动脚本加上如下一行：

```bash
printenv | grep -v "no_proxy" >> /etc/environment
```

然后就解决了问题，那么，这是为什么呢？原因是debian镜像的cron的配置中，限定了cron读取环境变量的位置：

```
$ cat /etc/pam.d/cron
# The PAM configuration file for the cron daemon

@include common-auth

# Sets the loginuid process attribute
session    required     pam_loginuid.so

# Read environment variables from pam_env's default files, /etc/environment
# and /etc/security/pam_env.conf.
session       required   pam_env.so

# In addition, read system locale information
session       required   pam_env.so envfile=/etc/default/locale

@include common-account
@include common-session-noninteractive 

# Sets up user limits, please define limits for cron tasks
# through /etc/security/limits.conf
session    required   pam_limits.so
```

可以看到，cron只会从 `/etc/environment` 和 `/etc/security/pam_env.conf` 读取环境变量。

> alpine 没有这个问题，因为alpine镜像没有带PAM。

---

参考资料：

- https://askubuntu.com/questions/700107/why-do-variables-set-in-my-etc-environment-show-up-in-my-cron-environment
- https://stackoverflow.com/questions/27771781/how-can-i-access-docker-set-environment-variables-from-a-cron-job
