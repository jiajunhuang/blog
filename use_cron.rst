
用crontab完成自动化
====================

在我看来, 计算机就是为了帮人类完成一些繁琐的事情而生的, 而linux更是如此,
所以我们要充分利用自动化.

crontab是一个定时执行任务的daemon,
如果是执行\ **非图形化**\ 的shell脚本或者其他, 直接 ``crontab -e``
然后添加自己所需要执行的命令就可以了

但是我写了一个判断每小时提醒一次 注意休息 的用到了\ **图形**\ 的脚本,
这是用到了notify-send这个GUI程序的,
在\ `ubuntu的help <https://help.ubuntu.com/community/CronHowto#GUI%20Applications>`__\ 里有相关说明,
在命令前加上

.. code:: bash

    env DISPLAY=:0

比如:

.. code:: bash

    env DISPLAY=:0 notify-send 'Test' 'This is a test'

即可, 当然, 这是显示在当前工作屏幕上, 如果想指定显示器, 就用

.. code:: bash

    env DISPLAY=:0.0

这是显示在默认的屏幕上, 如果是笔记本的话一般就是笔记本上的屏幕了

--------------

2014-10-04 增加:

下面给出我的crontab做示范:

``bash # m h  dom mon dow   command */5 * * * * env DISPLAY=:0 /home/jiajun/.xmonad/tools/battery.sh 0 * * * * env DISPLAY=:0 /home/jiajun/.xmonad/tools/goodforeyes.sh``
