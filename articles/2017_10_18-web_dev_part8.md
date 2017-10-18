# Web开发系列(八)：单点故障，负载均衡

试想我们有一个内容服务器，假设是监听在 `192.168.1.1:8000` 上，我们所有的请求都打到这上面，那么这个进程或者机器挂了怎么办？

因此有一个新的概念，叫做 [单点故障](https://en.wikipedia.org/wiki/Single_point_of_failure)。即，只要我们这唯一的，仅有的
内容服务器挂了，我们的网站就挂了。所以我们需要一个新的概念，叫做 [负载均衡](https://en.wikipedia.org/wiki/Load_balancing_(computing)) 。

当然，负载均衡的目的远不止是这一个，包括但不限于：冗余，降低响应时间，降低内容服务器负载，健康检查等。

在现实生产环境中，对HTTP或HTTPS请求而言，我们一般使用nginx来做负载均衡，即使用例如这样的配置：

```
upstream backend {
    server backend1.example.com       weight=5;
    server backend2.example.com:8080;
    server unix:/tmp/backend3;

    server backup1.example.com:8080   backup;
    server backup2.example.com:8080   backup;
}

server {
    location / {
        proxy_pass http://backend;
    }
}
```

而Nginx的负载均衡有以下几种算法：

- round-robin 几乎是个负载均衡器都支持，即，一个一个轮着来
- least-connected 找活跃连接数最少的进行请求
- ip-hash 根据ip地址进行哈希，然后分配到对应的内容服务器上

这三种算法各有千秋，第一种足够公平，大家都有份（当然也可以配置权重和backup），第二种能比较好的均衡负载，第三种则能
保证只要是同一个用户的请求，一定会到同一台服务器上，比较适用于这样一种情况：每台内容服务器连接一个私有的Redis，Redis中
保存了session，如果用第一种或者第二种算法，那么用户会出现随机要求重新登录的情况，第三种则不会。

**但是**，我们引入了Nginx作为内容服务器的负载均衡，同时Nginx自身也成为了新的单点故障，即，只要Nginx挂了，网站也就挂了。
那么，有没有什么好办法呢？有！

- 其一，我们可以用 [上篇](https://jiajunhuang.com/articles/articles/2017_10_19-web_dev_part7.md.html) 提到的，利用DNS做
负载均衡，这样不同的地方的用户访问到不同的服务器上，但是这在实际开发中比较困难，例如，如果有数据存放在Redis中，多个请求
可能需要共享信息怎么办？无疑不同的机房里，访问该Redis的速度会不一样并且可能比较慢。
- 其二，keepalived + nginx。目前我们公司线上使用的是阿里云的SLB，其实该方案根据云栖社区的 [文章](https://yq.aliyun.com/articles/1803)
是和keepalived + nginx的组合类似的，原理就在于做了一层TCP级别的负载均衡，使用VRRP协议，将一组机器（及其IP）组成一个虚拟
路由器，其中这个组的成员使用主从方式，主来承担流量分发任务，当主挂了之后，从挺身而出，承担主的任务，主恢复之后，便恢复
原来的主从模式。这需要路由器的配合，即支持VRRP。

> 那其实看起来还是有问题，路由器挂了怎么办？这也是单点故障，网线断了怎么办？这也是单点故障，但是别忘了，网络是拓扑结构，只要
网络布的好，虽然请求会有点绕，但是最终是能到达目的地的。

参考资料：

- https://yq.aliyun.com/articles/1803
- http://nginx.org/en/docs/http/load_balancing.html
- https://en.wikipedia.org/wiki/Virtual_Router_Redundancy_Protocol
