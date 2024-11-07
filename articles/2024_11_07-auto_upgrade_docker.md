# 自动升级Docker容器

我有一些自托管软件，都是以Docker容器的形式运行的。之前都是隔段时间手动升级，不够方便。因此花了点时间写了一个自动升级
脚本，一劳永逸。

```python
import logging
import os
import subprocess

logging.basicConfig(level=logging.INFO)


def upgrade_docker(bash_script_path):
    result = subprocess.run(['bash', bash_script_path], capture_output=True, text=True)
    logging.info(
        "execute %s, stdout %s, stderr %s, success? %s",
        bash_script_path, result.stdout, result.stderr, result.returncode == 0,
    )


def upgrade_all():
    # iter all bash scripts in the 'docker' directory
    for file in os.listdir('docker'):
        if file.endswith('.sh'):
            upgrade_docker(f'docker/{file}')

    # finally, run docker system prune -f
    subprocess.run(['docker', 'system', 'prune', '-f'])


if __name__ == '__main__':
    upgrade_all()
```

> 代码很简单，就是遍历 `docker` 目录下的所有 `.sh` 文件，然后执行它们。最后再清理一下无用的镜像。

然后在同级目录下创建一个 `docker` 目录，里面放置所有的升级脚本，比如 `docker/nextcloud.sh`:

```bash
#!/bin/bash

docker rm -f alist

docker pull xhofe/alist-aria2:latest

docker run -d --restart=always \
    -v /var/lib/data/alist:/opt/alist/data \
    -p 5244:5244 -e PUID=0 -e PGID=0 -e UMASK=022 \
    --name="alist" xhofe/alist-aria2:latest
```

> 套路都是一样的，先删除旧容器，再拉取新镜像，最后重新运行容器。

配合一个 `cronjob`:

```bash
0 0 * * 1 /usr/bin/python3 /path/to/upgrade_docker.py
```

每周一凌晨自动升级一次，再也不用担心忘记升级了。
