# logrotate read only filesystem问题

最近配置了OpenResty的 logrotate 规则，但是遇到一个奇怪的问题，自己直接
在命令行执行是ok的，但是systemd timer执行却是不行。

最终发现原因如下：

systemd service文件中，对 logrotate 加了文件读写保护:

```
[Unit]
Description=Rotate log files
Documentation=man:logrotate(8) man:logrotate.conf(5)                
ConditionACPower=true

[Service]
Type=oneshot
ExecStart=/usr/sbin/logrotate /etc/logrotate.conf                   

# performance options
Nice=19
IOSchedulingClass=best-effort
IOSchedulingPriority=7

# hardening options
#  details: https://www.freedesktop.org/software/systemd/man/systemd.exec.html
#  no ProtectHome for userdir logs                                  
#  no PrivateNetwork for mail deliviery                             
#  no ProtectKernelTunables for working SELinux with systemd older than 235
#  no MemoryDenyWriteExecute for gzip on i686                       
PrivateDevices=true
PrivateTmp=true
ProtectControlGroups=true
ProtectKernelModules=true
ProtectSystem=full
RestrictRealtime=true
```

问题就出在 `ProtectSystem=full` 上，查看上方的链接，发现如下描述：

```
ProtectSystem=
Takes a boolean argument or the special values "full" or "strict". If true, mounts the /usr/ and the boot loader
directories (/boot and /efi) read-only for processes invoked by this unit. If set to "full", the /etc/ directory
is mounted read-only, too. If set to "strict" the entire file system hierarchy is mounted read-only, except for
the API file system subtrees /dev/, /proc/ and /sys/ (protect these directories using PrivateDevices=,
ProtectKernelTunables=, ProtectControlGroups=). This setting ensures that any modification of the vendor-supplied
operating system (and optionally its configuration, and local mounts) is prohibited for the service. It is
recommended to enable this setting for all long-running services, unless they are involved with system updates
or need to modify the operating system in other ways. If this option is used, ReadWritePaths= may be used to
exclude specific directories from being made read-only. This setting is implied if DynamicUser= is set.
This setting cannot ensure protection in all cases. In general it has the same limitations as ReadOnlyPaths=,
see below. Defaults to off.

```

如果 `ProtectSystem=full` 那么会把 `/usr/`, `/boot`, `/efi`, `/etc` 挂载为只读，如果是 `ProtectSystem=strict` 那么整个文件系统都会挂载为只读。

然而OpenResty刚好就安装在 `/usr/local/openresty` 下，这可真是巧了。

所以，在 `/lib/systemd/system/logrotate.service` 最后加一行，改成这样：

```
[Unit]
Description=Rotate log files
Documentation=man:logrotate(8) man:logrotate.conf(5)
ConditionACPower=true

[Service]
Type=oneshot
ExecStart=/usr/sbin/logrotate /etc/logrotate.conf

# performance options
Nice=19
IOSchedulingClass=best-effort
IOSchedulingPriority=7

# hardening options
#  details: https://www.freedesktop.org/software/systemd/man/systemd.exec.html
#  no ProtectHome for userdir logs
#  no PrivateNetwork for mail deliviery
#  no ProtectKernelTunables for working SELinux with systemd older than 235
#  no MemoryDenyWriteExecute for gzip on i686
PrivateDevices=true
PrivateTmp=true
ProtectControlGroups=true
ProtectKernelModules=true
ProtectSystem=full
RestrictRealtime=true
ReadWritePaths=/usr/local/openresty/nginx/logs

```

`systemctl daemon-reload && systemctl restart logrotate` 搞定！
