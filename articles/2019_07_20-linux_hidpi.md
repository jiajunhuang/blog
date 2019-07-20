# Linux高分屏支持

整了个高分屏的本子，由于分辨率实在是太高了，默认情况下字太小了无法看清楚。因此编辑 `.Xresources` ，内容如下：

```
Xft.dpi: 260
```

编辑 `/etc/lightdm/lightdm-gtk-greeter.conf`：

```
[greeter]
xft-dpi=260
```

然后 `sudo systemctl restart lightdm` 就OK了。
