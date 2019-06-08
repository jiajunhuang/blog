# 从XMonad迁移到i3

XMonad用了很多年了，但是ArchLinux上，Haskell的包更新太频繁了，而XMonad的配置语言使用Haskell，每次更新之后，如果没有
编译一下，可能下次就进不去了，而且ArchLinux上Haskell的包分得太小了，再加上我使用Haskell就是为了写配置文件，但是配置文件
不常更新，因此有点忘记了。

所以对比了一下i3和awesome，最后选择了i3。

首先安装i3：

```bash
$ sudo pacman -S i3-wm i3status i3lock dmenu
```

然后卸载xmonad(此处因为使用i3lock，我把slock一起卸载了)：

```bash
sudo pacman -Rsn xmonad xmonad-contrib xmobar slock
```

另外在ArchLinux上清理一下无用的包：

```bash
$ sudo pacman -Rns $(pacman -Qtdq)
```

接下来就是配置了，我仔细研究了一下i3的配置，然后把它配置的快捷键改成和XMonad差不多，因此目前迁移还挺愉快的，主要配置：

```conf
set $mod Mod1
font pango:Consolas 8

# start a terminal
bindsym $mod+Shift+Return exec i3-sensible-terminal
bindsym $mod+Return exec i3-sensible-terminal

# kill focused window
bindsym $mod+Shift+c kill

# start dmenu (a program launcher)
bindsym $mod+p exec dmenu_run

# Start i3bar to display a workspace bar (plus the system information i3status
# finds out, if available)
bar {
    status_command i3status -c ~/.i3/i3status.conf
    position top
    tray_output primary
}

# default orientation
default_orientation horizontal

# automatically starting applications on i3 startup
exec xcompmgr
exec feh --bg-scale ~/.i3/background.jpg
exec fcitx
exec volumeicon
exec nm-applet
exec redshift

# focus do not follow
focus_follows_mouse no

# executing applications
bindsym $mod+c exec firefox
bindsym $mod+z exec zathura

# hide title bar by default
for_window [class="^.*"] border pixel 1

# i3lock
bindsym Mod4+l exec i3lock
```

详见： https://github.com/jiajunhuang/doti3/
