# 自建DERP服务器提升Tailscale连接速度(使用Nginx转发)

官方文档里，DERP服务器默认是直接占用443端口的，但是我的服务器上已经有了Nginx服务，因此好一顿折腾，终于成功了。

## 安装DERP

我没有使用 Docker 镜像，而是直接二进制+systemd的方式安装的(首先你得分配一个域名指向这个机器)：

```bash
# go install tailscale.com/cmd/derper@latest
# cp ~/go/bin/derper /usr/local/bin/
```

然后编辑`/etc/systemd/system/derper.service`：

```ini
[Unit]
Description=My Derper Service
After=network.target

[Service]
ExecStart=/usr/local/bin/derper -hostname=<域名，后面还会用> -a :30001 -http-port 30001 -stun-port 3478 -verify-clients
Restart=on-failure
User=root

[Install]
WantedBy=multi-user.target
```

然后`systemctl enable derper && systemctl start derper`。

## 配置Nginx

增加如下Nginx配置文件：

```nginx
server {
    listen 80;
    listen 443 ssl;
    server_name <域名>;

    access_log <Nginx 日志路径>;
    error_log <Nginx 错误日志路径>;

    ssl_certificate <Let's Encrypt 证书路径>;
    ssl_certificate_key <Let's Encrypt 证书私钥路径>;

    location / {
        client_max_body_size 1G;

        # websockets
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        # other settings
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_pass http://127.0.0.1:30001;
    }
}
```

请注意，这里 WebSocket 的配置是必须的，否则 Tailscale 无法正常工作，我掉到这个坑好久才爬出来。

## 配置 ACL 规则

打开 tailscale 的管理页面，添加 ACL 规则，允许你的域名访问 DERP 服务器。

在 `ssh` 同级，添加：

```js
// Define private derp
"derpMap": {
    "omitDefaultRegions": false,
    "regions": {
        "901": {
            "regionID":   901,
            "regionCode": "MyHK",
            "regionName": "My HongKong",
            "nodes": [
                {
                    "name":     "myhk",
                    "regionID": 901,
                    "hostName": "<域名>",
                    "DERPPort": 443,
                    "IPv4":     "<机器IP>",
                    "IPv6":     "none", // 如果你的服务器没有 IPv6
                    //"InsecureForTests": true,
                    "STUNPort": 3478,
                },
            ],
        },
    },
},

// ssh...
```

然后重启 tailscaled 服务，`systemctl restart tailscaled`。

就可以使用自有 DERP 服务器了，速度飞起。
