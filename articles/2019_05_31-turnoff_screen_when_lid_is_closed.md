# Linux下当笔记本合上盖子之后只使用扩展显示器

鉴于我的笔记本是7年前的本子，那会儿分辨率还是 `1366x768`，辣眼睛，所以接了一个扩展显示器，笔记本呢，就丢在角落里，
连几根线出来就可以了。但是有个问题就是，每次打开的时候，`XMonad` 都以为是两个屏幕，因此它会开两个 `workspace`，所以
要实现这么一个目的，就是打开图形界面登录的时候，检测一下是不是接了扩展显示器，而且笔记本的盖子是关着的。实现方式就是
添加这个文件 `~/.xprofile`：

```bash
#!/bin/bash

# close screen if lid is close
if grep "closed" /proc/acpi/button/lid/LID0/state >> /dev/zero; then
    if [ ! -z "$DISPLAY" ] && [[ $(xrandr -d :0 -q | grep ' connected ' | wc -l) = 2 ]]; then
        DISPLAY=:0 xrandr --output LVDS-0 --off
        DISPLAY=:0 feh --bg-scale ~/.xmonad/background.jpg
    fi
fi
```

注意，要给可执行权限。注销登录，然后重新登录，大功告成。

---

参考资料：

- https://jiajunhuang.com/articles/2017_09_19-xmonad.md.html
- https://wiki.archlinux.org/index.php/Xprofile
