# Linux低电量自动关机

2019.09.12更新：

可以直接使用自带的UPower，编辑 `/etc/UPower/UPower.conf`：

```
PercentageLow=40
PercentageCritical=30
PercentageAction=20
```

然后重启一下，这样低电量的时候就会自动关机，笔记本就相当于有一个UPS。

---

最近住所电不稳定，经常突然之间就断电了，虽然我的是笔记本，但是也不一定能挨到来电，这样下去过不了多久笔记本里的SSD就要挂比
的节奏啊。

所以写了一个简单的脚本，当电量低了之后，就关机，再配合 `crontab` 或者 `systemd timers` 定时检查。

`check_shutdown.timer`:

```bash
$ cat /etc/systemd/system/check_shutdown.timer 
[Unit]
Description=Check if battery is low every 10 minutes

[Timer]
OnCalendar=*:0/10
Persistent=true

[Install]
WantedBy=timers.target
```

`check_shutdown.service`:

```bash
$ cat /etc/systemd/system/check_shutdown.service 
[Service]
ExecStart=
ExecStart=/home/jiajun/.xmonad/scripts/shutdown.py
```

`check_shutdown.py`:

```python
#!/home/jiajun/.py3k/bin/python

import psutil
import logging
import os
import datetime

bat = psutil.sensors_battery()
logging.warn("%s: battery status: %s", datetime.datetime.now(), bat)

if bat.percent < 15:
    logging.warn("gonna shutdown")
    os.system("sudo shutdown -h now")
```
