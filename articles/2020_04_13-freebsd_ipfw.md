# FreeBSD ipfw使用教程

FreeBSD，古老的UNIX系统，最近在研究它的ipfw防火墙，鉴于国内相关资料较少，我就记录下来，以飨读者。

首先在FreeBSD 12中，ipfw已经默认编译进内核了，所以中文资料包括很多英文资料里，还需要编译的，就不用看了，那是过时的。

注意ipfw有一个比较坑的地方，那就是它默认会有一条规则，规则号为65536，是不可以删除的，这条规则会把所有流量都切断，
所以还没配置好之前，千万不要随意启动ipfw，否则就会面临无法连上远程FreeBSD的尴尬问题了。

```bash
$ sudo ipfw list
Password:
65535 deny ip from any to any
```

ipfw的规则是这样的，ipfw有一个规则编号，按照规则编号一次进行处理，所以由于最后一条是deny所有流量，就会产生刚才所说的
那个问题。

我们来看看ipfw的规则长啥样：

```man
RULE FORMAT
     The format of firewall rules is the following:

           [rule_number] [set set_number] [prob match_probability] action
           [log [logamount number]] [altq queue] [{tag | untag} number] body

     where the body of the rule specifies which information is used for
     filtering packets, among the following:

        Layer-2 header fields                 When available
        IPv4 and IPv6 Protocol                SCTP, TCP, UDP, ICMP, etc.
        Source and dest. addresses and ports
        Direction                             See Section PACKET FLOW
        Transmit and receive interface        By name or address
        Misc. IP header fields                Version, type of service,
                                              datagram length, identification,
                                              fragment flag (non-zero IP
                                              offset), Time To Live
        IP options
        IPv6 Extension headers                Fragmentation, Hop-by-Hop
                                              options, Routing Headers, Source
                                              routing rthdr0, Mobile IPv6
                                              rthdr2, IPSec options.
        IPv6 Flow-ID
        Misc. TCP header fields               TCP flags (SYN, FIN, ACK, RST,
                                              etc.), sequence number,
                                              acknowledgment number, window
        TCP options
        ICMP types                            for ICMP packets
        ICMP6 types                           for ICMP6 packets
        User/group ID                         When the packet can be
                                              associated with a local socket.
        Divert status                         Whether a packet came from a
                                              divert socket (e.g., natd(8)).
        Fib annotation state                  Whether a packet has been tagged
                                              for using a specific FIB
                                              (routing table) in future
                                              forwarding decisions.

```

- `rule_number` 是从1到65536的一个数字，这些规则会按照这个数值从小到大依次检查，如果规则号相同，就会按照在文件中的顺序检查
- `set`, `tag`, `untag`, `altq` 和 `prob` 不管，`prob` 是用来随机丢包用的
- `log` 表示是否打日志
- `action` 是我们要对流量采取的行动，比如 `deny`, `allow`等
- `body` 就是我们的具体规则，它的语法是 `[proto from src to dst] [options]`

我们来看一个具体的规则：

`ipfw add 100 allow ip from not 1.2.3.4 to any`，其中 `ip` 这里是协议，可选值是 `ip`, `tcp`, `udp`，而 src 和 dst 可以是
一个ip地址，一个网络号，也可以是 `any`，比如通常我们不想把本机出去的流量给掐掉，那么就加上这么一句：

```bash
ipfw -q add 110 allow all from any to any out
```

注意，src 和 dst 都可以加一个具体的端口号，比如我们要允许别的机器可以访问22:

```bash
ipfw -q add 130 allow tcp from any to any 22 in
```

## 配置ipfw

看完了规则的语法要求，我们来看看该怎么配置ipfw，执行以下命令：

```bash
$ sudo sysrc firewall_enable="YES"  # 允许防火墙开机自启
$ sudo sysrc firewall_type="open"  # 让系统把流量通过，这样就可以使用防火墙
$ sudo sysrc firewall_script="/etc/ipfw.rules"  # 制定ipfw规则的路径，我们待会儿在这里编辑规则
$ sudo sysrc firewall_logging="YES"  # 这样ipfw就可以打日志
$ sudo sysrc firewall_logif="YES"  # 把日志打到 `ipfw0` 这个设备里
```

然后编辑 `/etc/ipfw.rules`：

```
IPF="ipfw -q add"
ipfw -q -f flush

#loopback 
$IPF 10 allow all from any to any via lo0
$IPF 20 deny all from any to 127.0.0.0/8
$IPF 30 deny all from 127.0.0.0/8 to any
$IPF 40 deny tcp from any to any frag

# statefull
$IPF 50 check-state
$IPF 60 allow tcp from any to any established
$IPF 70 allow all from any to any out keep-state
$IPF 80 allow icmp from any to any

# open port for ssh
$IPF 110 allow all from any to any out
$IPF 130 allow tcp from any to any 22 in

# deny and log everything 
$IPF 500 deny log all from any to any
```

最后，启动ipfw：

```bash
$ sudo service ipfw start
```

---

参考资料：

- https://www.freebsd.org/doc/handbook/firewalls-ipfw.html
- man ipfw
