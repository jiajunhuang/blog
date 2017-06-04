# Redis 命令总览

| 命令 | 示例 | 一句话解释 |
| ------------- |:-------------:| -----:|
| APPEND  | APPEND key value | key是一个string，value会被追加到后面 |
| AUTH | AUTH password | 密码认证 |
| BGREWRITEAOF | BGREWRITEAOF | 起一个新的进程重写AOF文件 |
| BGSAVE | BGSAVE | 起一个新的进程将数据存储到磁盘上然后退出 |
| BITCOUNT | BITCOUNT key [start end] | 计数给定key存储的字符串里bit为1的数量 |
| BITFIELD | BITFIELD key [GET type offset] [SET type offset value] [INCRBY type offset increment] [OVERFLOW `WRAP|SAT|FAIL]` | 对给定offset开始的给定有符号/无符号定长类型进行操作 |
| BITOP | BITOP operation destkey key [key...] | 对给定的key进行 `AND, OR, XOR, NOT` 操作并将结果存储到 `destkey` 上去 |
| BITPOS | BITPOS key bit [start] [end] | 返回第一个设置成0或1的bit的位置 |
| BLPOP | BLPOP key [key...] timeout | 阻塞版 LPOP |
| BRPOP | BRPOP key [key...] timeout | 阻塞版 RPOP |
| BRPOPLPUSH | BRPOPLPUSH source destination timeout | 阻塞版 RPOPLPUSH |
| CLIENT KILL | CLIENT KILL [ip:port] [ID client-id] `[TYPE normal|master|slave|pubsub]` [ADDR ip:port] [SKIPME yes/no] | 杀掉某个连接 |
| CLIENT LIST | CLIENT LIST | 列出所有连接 |
| CLIENT GETNAME | CLIENT GETNAME | 获取由CLIENT SETNAME设置的当前连接的名字 |
| CLIENT PAUSE | CLIENT PAUSE timeout | 暂停所有连接timeout milliseconds |
| CLIENT REPLY | CLIENT REPLY `ON|OFF|SKIP` | 客户端暂停/继续/忽略接受来自服务器的回复 |
| CLIENT SETNAME | CLIENT SETNAME connection-name | 客户端设置连接的名称 |
| CLUSTER ADDSLOTS | CLUSTER ADDSLOTS slot [slot...] | 集群功能，占位 |
| CLUSTER COUNT-FAILURE-REPRTS | CLUSTER COUNT-FAILURE-REPORTS node-id | 统计某个节点的失败报告 |
| CLUSTER COUNTKEYSINSLOT | CLUSTER COUNTKEYSINSLOT slot | 统计某个slot里key的数量 |
| CLUSTER DELSLOTS | CLUSTER DELSLOTS slot [slot ...] | 删除某个slot |
| CLUSTER FAILOVER | CLUSTER FAILOVER `[FORCE|TAKEOVER]` | 集群中某个slave取代master |
| CLUSTER FORGET | CLUSTER FORGET node-id | 移出集群内某个节点 |
| CLUSTER GETKEYSINSLOT | CLUSTER GETKEYSINSLOT slot count | 获取某个slot里的所有key |
| CLUSTER INFO | CLUSTER INFO | 返回集群信息 |
| CLUSTER KEYSLOT | CLUSTER KEYSLOT key | 找到key在哪个slot |
| CLUSTER MEET | CLUSTER MEET ip port | 连接到集群内另一个节点 |
| CLUSTER NODES | CLUSTER NODES | 打印出集群内所有节点 |
| CLUSTER REPLICATE | CLUSTER REPLICATE node-id | 将某个节点配置为指定的master的slave |
| CLUSTER RESET | CLUSTER RESET `[HARD|SOFT]` | 删除所有信息，对keys不为空的master无效 |
| CLUSTER SAVECONFIG | CLUSTER SAVECONFIG | 保存 `nodes.conf` 到磁盘 |
| CLUSTER SET-CONFIG-EPOCH | CLUSTER SET-CONFIG-EPOCH config-epoch | 设置逻辑时钟 |
| CLUSTER SETSLOT | CLUSTER SETSLOT slot `IMPORTING|MIGRATING|STABLE|NODE` [node-id] | 设置slot的状态 |
| CLUSTER SLAVES | CLUSTER SLAVES node-id | 列出给定节点的slave |
| CLUSTER SLOTS | CLUSTER SLOTS | 返回集群里slots的信息 |
| COMMAND | COMMAND | 返回redis的所有命令 |
| COMMAND COUNT | COMMAND COUNT | 返回redis所有命令数量 |
| COMMAND GETKEYS | COMMAND GETKEYS | 返回命令中为key的，以列表形式返回 |
| COMMAND INFO | COMMAND INFO command-name [command-name...] | 以详细信息返回command的信息 |
| CONFIG GET | CONFIG GET parameter | 返回配置中的某些 |
| CONFIG REWRITE | CONFIG REWRITE | 重写 `redis.conf` |
| CONFIG SET | CONFIG SET parameter value | 设置 |
| CONFIG RESETSTAT | CONFIG RESETSTAT | 重置统计 |
| DBSIZE | DBSIZE | 返回当前选择的数据库里key的数量 |
| DEBUG OBJECT 和 DEBUG SEGFAULT | DEBUG OBJECT key | 调试用的 |
| DECR | DECR key | 将key对应的value 减一 |
| DECRBY | DECRBY key decrement | 将key对应的value减n |
| DEL | DEL key [key...] | 删除key（和对应的value) |
| DISCARD | DISCARD | 删除已排队的命令 |
| DUMP | DUMP key | 返回key对应的值序列化后的值 |
| ECHO | ECHO message | echo |
| EVAL | EVAL script numkeys key [key ...] arg [arg ...] | 执行lua脚本 |
| EVALSHA | EVALSHA sha1 numkeys key [key ...] arg [arg ...] | 根据脚本的sha1值执行脚本 |
| EXEC | EXEC | 执行事务里的命令 |
| EXISTS | EXISTS key [key...] | 判断key是否存在 |
| EXPIRE | EXPIRE key seconds | key多少秒后失效 |
| EXPIREAT | EXPIREAT key timestamp | 指定时间戳之后key将会失效 |
| FLUSHALL | FLUSHALL [ASYNC] | 删除数据库里的所有的key |
| FLUSHDB | FLUSHDB [ASYNC] | 删除当前db里的所有key |
| GEOADD | GEOADD key lonitude latitude member [longitude latitude member ...] | 存储地理位置(latitude, longitude, name) |
| GEOHASH | GEOHASH key member [member ...] | 返回 [Geohash](https://en.wikipedia.org/wiki/Geohash) |
| GEOPOS | GEOPOS key member [member ...] | 返回指定位置的经纬度 |
| GEODIST | GEODIST key member1 member2 [unit] | 返回地理位置距离 |
| GEORADIUS | GEORADIUS key longitude latitude radius `m|km|ft|mi` [WITHCOORD] [WITHDIST] [WITHHASH] [COUNT count] [ASC|DESC] [STORE key] [STOREDIST key] | 返回最大距离 |
| GEORADIUSBYMEMBER | GEORADIUSBYMEMBER key member radius `m|km|ft|mi` [WITHCOORD] [WITHDIST] [WITHHASH] [COUNT count] [ASC|DESC] [STORE key] [STOREDIST key] | 没用过 |
| GET | GET key | 返回key对应的value |
| GETBIT | GETBIT key offset | 返回offset上对应的bit的值 |
| GETRANGE | GETRANGE key start end | 返回闭区间[start, end]上对应的值 |
| GETSET | GETSET key value | 把key设置成value然后返回老的value，原子操作 |
| HDEL | HDEL key field [field ...] | 移出hash里的对应的field |
| HEXISTS | HEXISTS key field | 检查hash里对应的field是否存在 |
| HGET | HGET key field | 获取hash里对应field的值 |
| HGETALL | HGETALL key | 获取hash里所有的值 |
| HINCRBY | HINCRBY key field increment | hash里增加 |
| HINCRBYFLOAT | HINCRBYFLOAT key field increment | hash里增加小数 |
| HKEYS | HKEYS key | 返回hash里所有的key |
| HLEN | HLEN key | 返回hash里key的数量 |
| HMGET | HMGET key field [field ...] | 批量获取hash里的值 |
| HMSET | HMSET key field value [field value ...] | 批量设置hash里的值 |
| HSET | HSET key field value | 设置hash里的值 |
| HSETNX | HSETNX key field value | 如果hash里不存在就设置，要不然无动作 |
| HSTRLEN | HSTRLEN key field | 返回hash里key对应的value的长度 |
| HVALS | HVALS key | 返回hash里key对应的所有值 |
| INCR | INCR key | key对应的value+1 |
| INCRBY | INCRBY key increment | key对应的value +n |
| INCRBYFLOAT | INCRBYFLOAT key increment | key对应的value加浮点数 |
| INFO | INFO [section] | 打印节点信息 |
| KEYS | KEYS pattern | 打印符合pattern的key |
| LASTSAVE | LASTSAVE | 返回最近一次保存数据库的时间戳 |
| LINDEX | LINDEX key index | 返回list里index对应的值 |
| LINSERT | LINSERT key BEFORE|AFTER pivot value | list里插入值 |
| LLEN | LLEN key | 返回list的长度 |
| LPOP | LPOP key | 弹出list左边的值 |
| LPUSH | LPUSH key value [value ...] | 从左边插入到list |
| LPUSHX | LPUSHX key value | 如果key对应的list存在就从左边插入，否则啥也不做 |
| LRANGE | LRANGE key start stop | 返回list对应范围 |
| LREM | LREM key count value | 移出n个对应value的值 |
| LSET | LSET key index value | 设置列表里对应index的值 |
| LTRIM | LTRIM key start stop | 把list里start到stop范围之外的全删了 |
| MGET | MGET key [key ...] | 批量获取 |
| MIGRATE | MIGRATE host port `key|""` destination-db timeout [COPY] [REPLACE] [KEYS key [key ...]] | 迁移 |
| MONITOR | MONITOR | 实时打印在干啥 |
| MOVE | MOVE key db | 把key移到其他db里 |
| MSET | MSET key value [key value ...] | 批量设置 |
| MSETNX | MSETNX key value [key value ...] | 批量设置如果全都不存在的话 |
| MULTI | MULTI | 事务开始 |
| OBJECT | OBJECT subcommand [arguments [arguments ...]] | 对redis里的对象进行操作 |
| PERSIST | PERSIST key | 持久化指定的key |
| PEXPIRE | PEXPIRE key milliseconds | 和EXPIRE一样，只是单位是 milliseconds |
| PEXPIREAT | PEXPIREAT key milliseconds-timestamp | 和上面一样，只是使用时间戳 |
| PFADD | PFADD key element [element ...] | 把对应的element加到hyperloglog里 |
| PFCOUNT | PFCOUNT key [key ...] | 统计对应的hyperloglog |
| PFMERGE | PFMERGE destkey sourcekey [sourcekey ...] | 合并hyperloglog |
| PING | PING [message] | 你懂的, ping |
| PSETEX | PSETEX key milliseconds value | 在hyperloglog里设置超时 |
| PSUBSCRIBE | PSUBSCRIBE pattern [pattern ...] | 订阅 |
| PUBSUB | PUBSUB subcommand [argument [argument ...]] | 订阅 https://redis.io/commands/pubsub |
| PTTL | PTTL key | 和TTL一样但是单位是milliseconds |
| PUBLISH | PUBLISH channel message | 发消息 |
| PUNSUBSCRIBE | PUNSUBSCRIBE [pattern [pattern ...]] | 取消消息 |
| QUIT | QUIT | 退出 |
| RANDOMKEY | RANDOMKEY | 返回选定db里的随机一个key |
| READONLY | READONLY | 对集群里的节点使用 |
| READWRITE | READWRITE | 对集群里的节点使用 |
| RENAME | RENAME key newkey | 重命名key |
| RENAMENX | RENAMENX key newkey | 如果newkey不存在，就重命名 |
| RESTORE | RESTORE key ttl serialized-value [REPLACE] | 恢复 |
| ROLE | ROLE | 集群里，打印节点当前状态，master还是slave还是sentinel |
| RPOP | RPOP key | list里右边弹出 |
| RPOPLPUSH | RPOPLPUSH source destination | 右出左进 |
| RPUSH | RPUSH key value [value ...] | list里右边进 |
| RPUSHX | RPUSHX key value | 如果key对应的list存在，就RPUSH，要不然啥也不干 |
| SADD | SADD key member [member ...] | 集合，加入 |
| SAVE | SAVE | 阻塞进行快照 |
| SCARD | SCARD key | 返回key对应的set里东西的数量 |
| SCRIPT DEBUG | SCRIPT DEBUG `YES|SYNC|NO` | 是否调试lua脚本 |
| SCRIPT EXISTS | SCRIPT EXISTS sha1 [sha1 ...] | 是不是存在脚本 |
| SCRIPT FLUSH | SCRIPT FLUSH | 删除缓存的脚本 |
| SCRIPT KILL | SCRIPT KILL | 杀掉脚本 |
| SCRIPT LOAD | SCRIPT LOAD script | 缓存脚本 |
| SDIFFSTORE | SDIFFSTORE destination key [key ...] | 和SDIFF一样，但是把结果存在destination里，返回都不在参数里的集合里的值 |
| SELECT | SELECT index | 选择数据库 |
| SET | SET key value [EX seconds] [PX milliseconds] `[NX|XX]` | 设置值 |
| SETBIT | SETBIT key offset value | 设置对应的bit的值 |
| SETEX | SETEX key seconds value | 设置值和超时，原子操作 |
| SETNX | SETNX key value | 如果不存在就设置 |
| SETRANGE | SETRANGE key offset value | 某个范围内设置成某个值 |
| SHUTDOWN | SHUTDOWN `[NOSAVE|SAVE]` | 关闭server |
| SINTER | SINTER key [key ...] | 做交集操作 |
| SINTERSTORE | SINTERSTORE destination key [key ...] | 交集，然后存到destination |
| SISMEMBER | SISMEMBER key member | 判断是否是集合的成员 |
| SLAVEOF | SLAVEOF host port | 判断是否是对应的slave |
| SLOWLOG | SLOWLOG subcommand [argument] | 慢日志 |
| SMEMBERS | SMEMBERS key | 打印集合里的key |
| SMOVE | SMOVE source destination member | 从一个集合移动到另一个集合 |
| SORT | SORT key [BY pattern] [LIMIT offset count] [GET pattern [GET pattern ...]] `[ASC|DESC]` [ALPHA] [STORE destination] | 对list进行排序 |
| SPOP | SPOP key [count] | 弹出set里的值 |
| SRANDMEMBER | SRANDMEMBER key [count] | 随机挑两个出来 |
| SREM | SREM key member [member ...] | 批量在set里删除 |
| STRLEN | STRLEN key | 返回字符串长度 |
| SUBSCRIBE | SUBSCRIBE channel [channel ...] | 订阅 |
| SUNION | SUNION key [key ...] | 并集操作 |
| SUNIONSTORE | SUNIONSTORE destination key [key ...] | 并集操作然后存着 |
| SWAPDB | SWAPDB index index | 把两个db互换一下 |
| SYNC | SYNC | 数据落地 |
| TIME | TIME | 打印时间 |
| TOUCH | TOUCH key [key ...] | 更新key的最后访问时间 |
| TTL | TTL key | 存活时间 |
| TYPE | TYPE key | 打印key的类型 |
| UNSUBSCRIBE | UNSUBSCRIBE [channel [channel ...]] | 取消订阅 |
| UNLINK | UNLINK key [key ...] | 非阻塞版DEL |
| UNWATCH | UNWATCH | 放弃事务里watched 的key |
| WAIT | WAIT numslaves timeout | 集群里numslaves个节点或者超时之后，否则阻塞 |
| WATCH | WATCH key [key ...] | 事务里watch |
| ZADD | ZADD key [NX|XX] [CH] [INCR] score member [score member ...] | 有序集 |
| ZCARD | ZCARD key | 统计有序集里key的数量 |
| ZCOUNT | ZCOUNT key min max | 统计有序集里min和max之间的key的数量 |
| ZINCRBY | ZINCRBY key increment member | 有序集里增加 |
| ZINTERSTORE | ZINTERSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] `[AGGREGATE SUM|MIN|MAX]` | |
| ZLEXCOUNT | ZLEXCOUNT key min max | |
| ZRANGE | ZRANGE key start stop [WITHSCORES] | |
| ZRANGEBYLEX | ZRANGEBYLEX key min max [LIMIT offset count] | |
| ZREVRANGEBYLEX | ZREVRANGEBYLEX key max min [LIMIT offset count] | |
| ZRANGEBYSCORE | ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count] | |
| ZRANK | ZRANK key member | |
| ZREM | ZREM key member [member ...] | |
| ZREMRANGEBYLEX | ZREMRANGEBYLEX key min max | |
| ZREMRANGEBYRANK | ZREMRANGEBYRANK key start stop | |
| ZREMRANGEBYSCORE | ZREMRANGEBYSCORE key min max | |
| ZREVRANGE | ZREVRANGE key start stop [WITHSCORES] | |
| ZREVRANGEBYSCORE | ZREVRANGEBYSCORE key max min [WITHSCORES] [LIMIT offset count] | |
| ZREVRANK | ZREVRANK key member | |
| ZSCORE | ZSCORE key member | |
| ZUNIONSTORE | ZUNIONSTORE destination numkeys key [key ...] [WEIGHTS weight [weight ...]] `[AGGREGATE SUM|MIN|MAX]` | |
| SCAN | SCAN cursor [MATCH pattern] [COUNT count] | |
| SSCAN | SSCAN key cursor [MATCH pattern] [COUNT count] | |
| HSCAN | HSCAN key cursor [MATCH pattern] [COUNT count] | |
| ZSCAN | ZSCAN key cursor [MATCH pattern] [COUNT count] | |
