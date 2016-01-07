:Date: 01/07/2016

分析日志的Bash脚本
===================

日志格式::

    [W 151127 17:59:45 web:1908] 404 GET /404.html (183.136.190.62) 0.46ms

.. code:: bash

    awk '$5 == 200 {print $9,$7}' im.out | grep -v '/404' | sort -u -k 2 | sort -g
