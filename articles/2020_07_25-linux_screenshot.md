# Linux窗口管理器下的截图

Linux使用窗口管理器时，也想做到截图然后保存到剪贴板，之后就可以到处贴比如贴到虚拟机里的Windows，不过如果直接使用截图
工具的话，是没有办法保存到剪贴板的，不过使用xclip可以做到。我使用的是XMonad，加这么一行：

```haskell
((mod4Mask, xK_a), spawn "sleep 0.2; scrot -s -e 'xclip -selection clipboard -t \"image/png\" < $f && rm $f'")
```

这样就可以使用 `Win + a` 快捷键来进行截图了，注意前面的 `sleep 0.2`，如果没有这一行，scrot也没有办法成功调用。

这个办法同样也适用于其它窗口管理器，例如i3, openbox, awesome等。
