# 虚拟机里的Ubuntu sudo时卡住

> https://serverfault.com/questions/38114/why-does-sudo-command-take-long-to-execute

今天装了一个虚拟机，其他都很快，但是执行sudo的时候却非常的卡，要等好几秒钟，搜了一下，是这么个解决法，把主机名加到 `/etc/hosts` 里，原因是安装过程中，hostname没有追加到 `/etc/hosts` 里。

那么为什么sudo要读取 `/etc/hosts` 呢？原因在这里：https://superuser.com/questions/429790/sudo-command-trying-to-search-for-hostname
