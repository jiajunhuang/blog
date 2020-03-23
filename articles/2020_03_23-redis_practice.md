# Redis使用中的几点注意事项

- 如非必要，一定要设置TTL。如果不是业务所需，必须持久存储，那么请一定要设置好TTL，否则随着时间流逝，Redis里会塞满垃圾。
此外还要注意使用框架时，确定好框架是否会设置ttl，比如最近遇到的一个坑就是Python RQ没有默认为job设置ttl，因此几年过去，
现在Redis内存不够用了，分析之后才发现，里面有诸多的垃圾，比如一些不用的业务数据，还有很早以前的job的数据等等，全部都堆在
Redis中，成为了持久的垃圾。

- 不要设置过长的key。比如spring框架就会有这样的key：`spring:session:sessions:1c88a003-63a4-48a0-979d-9b3be4ed9c0c`，其中
很大一部分都是无用的数据，占用了过多的内存。

- 客户端使用连接池，以复用连接，提升性能。

- 使用 pipeline 来执行多个动作，避免减少多次网络来回的开销。

- 如果使用了Lua，那么一定要注意Lua脚本不能占用太长时间。

---

附我最近分析Redis中内存占用的脚本：

```python
import logging
import sys

import redis


logging.basicConfig(level=logging.INFO)


def get_type_and_subcount(client, key):
    _type = client.type(key).decode()
    sub_count = 0

    if _type == "set":
        sub_count = client.scard(key)
    elif _type == "list":
        sub_count = client.llen(key)
    elif _type == "hash":
        sub_count = client.hlen(key)
    elif _type == "string":
        sub_count = client.strlen(key)
    elif _type == "zset":
        sub_count = client.zcard(key)
    else:
        logging.error("bad key %s with type %s", key, _type)

    return _type, sub_count


BYTES_TO_GB = 1024 * 1024 * 1024


def analytic_db(db):
    logging.info("we're now parse db %s", db)
    redis_client = redis.Redis(host="127.0.0.1", db=db)

    total_count = 0  # 总数
    key_bytes_count = 0  # 总bytes
    big_key_count = 0  # >1KB 总数
    big_key_bytes_count = 0  # >1KB 总bytes
    big_big_key_count = 0  # > 100KB 总数
    big_big_key_bytes_count = 0  # > 100KB 总bytes
    no_ttl_big_key_count = 0  # 没有设置ttl的>1KB 总数
    no_ttl_big_key_bytes_count = 0  # 没有设置ttl的 >1KB 总bytes

    for key in redis_client.scan_iter():
        bytes_num = redis_client.memory_usage(key)
        total_count += 1
        key_bytes_count += bytes_num

        ttl = redis_client.ttl(key)

        if bytes_num > 1024:  # 1K
            big_key_count += 1
            big_key_bytes_count += bytes_num

            key_type, sub_count = get_type_and_subcount(redis_client, key)

            if ttl == -1:
                no_ttl_big_key_count += 1
                no_ttl_big_key_bytes_count += bytes_num

            if bytes_num > 102400:  # 100K
                big_big_key_count += 1
                big_big_key_bytes_count += bytes_num
                logging.warning(
                    "big key found %s, bytes: %s, type is %s, sub_count %s, ttl is %s",
                    key, bytes_num, key_type, sub_count, ttl,
                )

    logging.info(
        "db %s, %s keys(%s GB), %s keys are > 1KB (%s GB), %s keys are > 100KB (%sGB), %s no ttl big keys > 100KB(%sGB)",
        db, total_count, str.format("{:+.2f}", key_bytes_count / BYTES_TO_GB),
        big_key_count, str.format("{:+.2f}", big_key_bytes_count / BYTES_TO_GB),
        big_big_key_count, str.format("{:+.2f}", big_big_key_bytes_count / BYTES_TO_GB),
        no_ttl_big_key_count, str.format("{:+.2f}", no_ttl_big_key_bytes_count / BYTES_TO_GB),
    )


if __name__ == "__main__":
    analytic_db(sys.argv[1])
```

---

参考资料：

- https://docs.microsoft.com/en-us/azure/azure-cache-for-redis/cache-best-practices
- https://redislabs.com/redis-best-practices/introduction/
