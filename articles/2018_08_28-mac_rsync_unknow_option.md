# macOS ansible 遇到 rsync: --chown=www-data: unknown option

用Ansible同步数据的时候，遇到了错误：

```bash
FAILED! => {"changed": false, "cmd": "/usr/bin/rsync --delay-updates -F --compress --delete-after --archive --rsh=/usr/bin/ssh xxx(略)", "msg": "rsync: --chown=www-data: unknown option\nrsync error: syntax or usage error (code 1) at  [client=2.6.9]\n", "rc": 1}
```

原来是本地rsync太老了。安装一个最新的就可以了：

```bash
$ brew install rsync
```
