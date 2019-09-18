# ssh时自动运行tmux

tmux，终端复用神器，之前我一直用byobu，它是tmux的封装，我看了一下源代码，其实就是一堆的bash脚本+python脚本。因为一些
byobu的bug，我选择使用原生tmux，但是有一个问题，就是以前执行tmux的时候，是在 `~/.bashrc` 里加上：

```bash
# start tmux
if [[ -z "$TMUX"  ]] && [ "$SSH_CONNECTION" != ""  ]; then
   tmux attach || tmux new
fi
```

意思就是，当检测到当前是ssh连接并且当前没有使用tmux时，就执行 `tmux attach || tmux new`，这样就会优先选择连接到上次的
会话，如果没有，那就创建一个新的会话，这样的确也能运行，能在ssh登录时自动运行tmux，但是有一个比较麻烦的缺点，那就是
退出时，由于它是使用一个子进程来执行 `tmux attach || tmux new`，因此即使退出，我们还是会回到一个没有tmux的连接，也就是说，
我们需要退出两次。解决方案就是写一个脚本，实现和 `tmux attach || tmux new` 一样的功能，但是使用bash内置的exec替换当前进程
的代码，我用python来实现的：

```python

#!/usr/bin/env python3

import os
import sys
import subprocess


def get_sessions():
    sessions = []

    output = subprocess.Popen(["tmux", "list-sessions"], stdout=subprocess.PIPE).communicate()[0]
    if sys.stdout.encoding is None:
        output = output.decode("UTF-8")
    else:
        output = output.decode(sys.stdout.encoding)
    if output:
        for s in output.splitlines():
            # Ignore hidden sessions (named sessions that start with a "_")
            if s and not s.startswith("_"):
                sessions.append(s.strip())
    return sessions


sessions = get_sessions()
if sessions:
    session_name = sessions[-1].split(":")[0]
    os.execvp("tmux", ["tmux", "attach", "-t", session_name])
else:
    os.execvp("tmux", ["tmux", "new"])
```

然后把 `~/.bashrc` 改成这样的：

```bash
# start tmux
if [[ -z "$TMUX"  ]] && [ "$SSH_CONNECTION" != ""  ]; then
    exec ~/.xmonad/bash/tmux.py
fi
```

> 注意，要把上面的 `~/.xmonad/bash/tmux.py` 替换成你保存 `tmux.py` 这个脚本的路径，而且记得要给 `tmux.py`
> 这个脚本加可执行权限 `chmod +x tmux.py`

完美！
