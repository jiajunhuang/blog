# QT5使用GTK主题

我在使用arc这个GTK的主题，但是QT5之后，并不直接使用GTK的主题，因此显得格外的丑。解决方案：

```bash
$ sudo pacman -S qt5ct qt5-styleplugins
```

打开 `qt5tc` 之后，选择GTK。然后编辑 `~/.profile`：

```
export QT_QPA_PLATFORMTHEME=gtk2
```

然后退出，重新登录，就大功告成了！
