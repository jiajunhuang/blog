# Using xmonad with xfce4

I've been using XMonad as my desktop environment for 13 years. It's effecient and productive, but with a little bit ugly
because actually, XMonad is not a desktop environment, it's just a window manager.

I want to try xmonad work with a desktop environment, and I want to try xfce4 as my desktop environment.

# Installation

```bash
$ sudo apt install xfce4-panel xfce4-power-manager xfce4-settings xfce4-terminal
```

# Configuration

You have to change your xmonad to spawn xfce4 after login:

```haskell
myStartupHook = do
    setWMName "LG3D"
    spawnOnce "xcompmgr"
    spawnOnce "xfce4-panel"
    spawnOnce "xfce4-power-manager"
    spawnOnce "xfsettingsd"
```

Full configuration is on github: https://github.com/jiajunhuang/dotxmonad/blob/master/xmonad.hs

# Compile and re-login

```bash
$ xmonad --recompile
$ xmonad --restart
```

Then logout and login again.
