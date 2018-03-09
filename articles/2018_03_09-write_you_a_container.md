# 自己写一个容器

见： https://github.com/jiajunhuang/cup

容器技术并不是什么新技术，Docker能如此风行估计是让容器技术降低了使用难度。大概说一下这个demo用到的技术：

- namespace，通过使用Linux下的namespace来对进程进行隔离(挂载信息，pid，用户，网络等)
- chroot，通过chroot来限制进程的rootfs
- reexec，通过这个来模拟fork，见[这篇文章](https://jiajunhuang.com/articles/2018_03_08-golang_fork.md.html)

而资源限制则可以通过cgroups来配置.

参考：

- http://man7.org/linux/man-pages/man7/namespaces.7.html
- http://man7.org/linux/man-pages/man7/cgroup_namespaces.7.html
- http://man7.org/linux/man-pages/man7/network_namespaces.7.html
- http://man7.org/linux/man-pages/man7/pid_namespaces.7.html
- http://man7.org/linux/man-pages/man7/user_namespaces.7.html
