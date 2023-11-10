# 使用开源软KVM - synergy-core

前面我写过使用barrier的文章，Linux + Windows的时候，Linux做服务端，Windows做客户端，工作的还是挺不错的，大部分时候都能
正常工作，偶尔剪贴板会有点小问题，比如贴出来的是乱码，第二次贴才正常，但是无伤大雅。

然而，当我用 macOS 做服务端，Linux做客户端的时候，barrier套装就不那么好用了，bug比较多，而barrier 已经2年没有新的提交了，
这就让我不得不去找替代品，barrier 的活跃提交者fork了一份，变成了 [input-leap](https://github.com/input-leap/input-leap)，
但是找不到能够下载的二进制，尝试了自己编译，也是一堆的报错，查了资料和 Apple M1 芯片有关系。

然后我又转向了使用他们的老祖宗：synergy-core，仔细看了以后才发现，这玩意儿是开源的，收费的是GUI，命令行程序开源且免费，
Windows 下可能不好使用，但是 Linux 和 macOS 用命令行程序，还是可以的。

## 安装

- macOS 可以直接 brew 安装：`brew install synergy-core`
- Linux/Windows 可以去这里下载安装包然后安装：https://github.com/DEAKSoftware/Synergy-Binaries

## 配置文件

具体的配置语法可以参考官方文档：https://github.com/symless/synergy-core/wiki/Text-Config 。我是直接把 barrier 生成的
抄了过来：

```conf
section: screens
        macos:
                halfDuplexCapsLock = false
                halfDuplexNumLock = false
                halfDuplexScrollLock = false
                xtestIsXineramaUnaware = false
                preserveFocus = false
                switchCorners = none
                switchCornerSize = 0
        linux:
                halfDuplexCapsLock = false
                halfDuplexNumLock = false
                halfDuplexScrollLock = false
                xtestIsXineramaUnaware = false
                preserveFocus = false
                switchCorners = none
                switchCornerSize = 0
end

section: aliases
end

section: links
        macos:
                right = linux
        linux:
                left = macos
end

section: options
        relativeMouseMoves = false
        win32KeepForeground = false
        clipboardSharing = true
        switchCorners = none
        switchCornerSize = 0
end
```

这样，就实现了左侧是 macOS，右边是 Linux 的屏幕布局。

## 开机自启

对于 Linux，由于我是当作客户端，因此编辑文件 `/etc/systemd/system/synergyc.service`:

```systemd
[Unit]
Description=Synergy KVM Client
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=jiajun
ExecStart=/usr/bin/unbuffer /usr/bin/synergyc -f --display :0 <macOS IP地址>:24800
Environment=XAUTHORITY=/var/run/lightdm/root/:0
Restart=always
RestartSec=3

[Install]
WantedBy=graphical.target
```

对于 macOS，无法设置为全局开机自启，但是可以设置为用户登录后启动，编辑文件 `~/bin/startsynergys.command`:

```bash
# Hides tray icon. set to 0 to have it show again.
defaults write /Applications/Synergy.app/Contents/Info LSUIElement 1

#/Applications/Synergy.app/Contents/MacOS/synergys --enable-crypto --config ~/Library/Synergy/synergy.conf -n osx -l /var/log/synergy.log
/opt/homebrew/bin/synergys --config /etc/synergy.conf -l /opt/homebrew/var/logs/synergys.log
```

记得 `touch /opt/homebrew/var/logs/synergys.log` 文件，并且确保文件是属于当前用户的，否则日志会写不进去。

然后去设置，通用，启动项里，点加号，把 `~/bin/startsynergys.command` 加进去，再去 `Privacy & Security -> Accessibility` 中把 `Terminal` 加上，并且启用。

接下来就可以使用了。

## 问题

使用 synergy-core 时，如果 macOS 当前的输入法为拼音时，Linux这边就无法输入特殊符号，例如问好，逗号，斜杠等，英文输入法时则一切正常。
