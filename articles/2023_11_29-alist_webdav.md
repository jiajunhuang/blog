# Linux 自动挂载 alist 提供的webdav

首先需要安装 `davfs2`:

```bash
$ sudo apt install davfs2
```

然后把用户名密码写在 `/etc/davfs2/secrets` 里，例如：

```
http://127.0.0.1:5244/dav/ username password
```

下一步就是编辑自动挂载的文件。可以选择使用 `/etc/fstab` 但是我觉得使用 systemd 挂载会更好一些。首先需要确定挂载点，
例如 `/data/webdav`，然后编辑自动挂载文件，注意，挂载点是什么，自动挂载文件就要叫什么，比如你的挂载点是 `/data/webdav`，
那么你的文件就应该是 `/etc/systemd/system/data-webdav.mount` 和 `/etc/systemd/system/data-webdav.automount`。

`/etc/systemd/system/data-webdav.mount` 内容如下：

```bash
[Unit]
Description=Mount WebDAV
After=network-online.target
Wants=network-online.target

[Mount]
What=http://127.0.0.1:5244/dav/
Where=/data/webdav
Options=noauto,user,uid=你的用户名,gid=你的组
Type=davfs
TimeoutSec=60

[Install]
WantedBy=remote-fs.target
```

`/etc/systemd/system/data-webdav.automount` 内容如下：

```bash
[Unit]
Description=WebDAV automount
After=network-online.target
Wants=network-online.target

[Automount]
Where=/data/webdav
TimeoutIdleSec=300

[Install]
WantedBy=remote-fs.target
```

接下来直接启动即可：

```bash
$ sudo systemctl daemon-reload
$ sudo systemctl enable data-webdav.automount
$ sudo systemctl start data-webdav.automount
```
