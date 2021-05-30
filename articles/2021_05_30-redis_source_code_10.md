# Redis源码阅读：bitmap 位图的运算

Redis提供了位图，位图也就是从bit的角度来看数据，可以把某一个bit设置为0或者1，举个简单的例子，我们希望记录用户某个月
的每一天是否登录，那么就只需要给一个32bit的值，如果第n天登录了，就把第n个bit设置为1。位图很节省内存，毕竟是从bit的
角度来看待和存储数据的，但是缺点也很明显，位图所需要存储的数据的大小取决于上限。

接下来我们看看Redis是怎么实现位图的，位运算比较难以理解，所以我们要用一个具体例子来进行辅助，我们将会阅读 `SETBIT`
和 `GETBIT` 两个命令来看源码。

```c
    {"setbit",setbitCommand,4,
     "write use-memory @bitmap",
     0,NULL,1,1,1,0,0,0},

    {"getbit",getbitCommand,3,
     "read-only fast @bitmap",
     0,NULL,1,1,1,0,0,0},


/* SETBIT key offset bitvalue */
void setbitCommand(client *c) {
    robj *o;
    char *err = "bit is not an integer or out of range";
    size_t bitoffset;
    ssize_t byte, bit;
    int byteval, bitval;
    long on;

    if (getBitOffsetFromArgument(c,c->argv[2],&bitoffset,0,0) != C_OK)
        return;

    if (getLongFromObjectOrReply(c,c->argv[3],&on,err) != C_OK)
        return;

    /* Bits can only be set or cleared... */
    if (on & ~1) {
        addReplyError(c,err);
        return;
    }

    if ((o = lookupStringForBitCommand(c,bitoffset)) == NULL) return;

    /* Get current values */
    // byte是指需要多少个字节来保存数据，向右位移3位，其实就是除以8，也就是从bit数得到byte数，因为一个byte等于8 bit
    // 当然，你看下面的运算的话，就可以看到，实际长度应该是 byte + 1，比如maxbit等于31，31 >> 3 = 3
    // 这实际上是保存对应bit 所在的byte的索引值，而长度应该是再加上1才对。下同。
    byte = bitoffset >> 3;
    // 得到对应byte的值，比如offset如果是31，那么byte就是3。如果第一次设置这个值，那么o->ptr应该是 00000000 00000000 00000000 00000001，byteval就是 00000001
    byteval = ((uint8_t*)o->ptr)[byte];

    // 得到对应bit的值，比如上面的运算，那么 应该是先计算括号里的值，
    // bitoffset & 0x7 就是 31 & 0x7 其实就是
    // 00011111 & 00000111 = 00000111 也就是0x7的值，
    // 然后计算 7 - 0x7，也就是 00000111 - 00000111 = 00000000
    // 这里其实就是计算出到底位于第几个bit
    // 我算了一下，如果bitoffset为30，那么值就是1，如果是29，那么值就是2:
    // >>> 7 - 31 & 0x7
    // 0
    // >>> 7 - 30 & 0x7
    // 1
    // >>> 7 - 29 & 0x7
    // 2

    bit = 7 - (bitoffset & 0x7);

    // 先计算 1 << 0，也就是1，二进制就是 00000001
    // 然后计算 byteval & 1，也就是 00000001 & 00000001 = 00000001
    // 得到只有那一个bit的时候的值
    bitval = byteval & (1 << bit);

    /* Update byte with new bit value and return original value */
    // 更新值，先找到对应的bit，然后取反，然后再和原值做&，
    // 按上面的例子， 1 << 0 也就是 00000001，取反以后是 11111110，byteval 是 00000001
    // 做 & 操作，得到 00000000
    // 这里其实就是把对应bit位上的值消掉，置为0，其它位不变。
    byteval &= ~(1 << bit);
    // 先执行 on & 0x1，如果我们是设置为1，那么这里就还是1，然后便宜bit位，得到 00000001
    // 然后 byteval 取或运算，也就是 00000000 | 00000001 = 00000001
    // 这里其实就是设置对应bit位的值
    byteval |= ((on & 0x1) << bit);
    // 然后把对应的byte设置成这个值，也就是 00000001，由于byte=3，也就是32bit中最后8个bit
    // 所以最后 o->ptr 所在值，其实就是 00000000 00000000 00000000 00000001
    ((uint8_t*)o->ptr)[byte] = byteval;
    signalModifiedKey(c,c->db,c->argv[1]);
    notifyKeyspaceEvent(NOTIFY_STRING,"setbit",c->argv[1],c->db->id);
    server.dirty++;
    addReply(c, bitval ? shared.cone : shared.czero);
}

/* GETBIT key offset */
void getbitCommand(client *c) {
    robj *o;
    char llbuf[32];
    size_t bitoffset;
    size_t byte, bit;
    size_t bitval = 0;

    if (getBitOffsetFromArgument(c,c->argv[2],&bitoffset,0,0) != C_OK)
        return;

    if ((o = lookupKeyReadOrReply(c,c->argv[1],shared.czero)) == NULL ||
        checkType(c,o,OBJ_STRING)) return;

    byte = bitoffset >> 3;
    bit = 7 - (bitoffset & 0x7);
    if (sdsEncodedObject(o)) {
        if (byte < sdslen(o->ptr))
            bitval = ((uint8_t*)o->ptr)[byte] & (1 << bit); // 取出对应值
    } else {
        if (byte < (size_t)ll2string(llbuf,sizeof(llbuf),(long)o->ptr))
            bitval = llbuf[byte] & (1 << bit);
    }

    addReply(c, bitval ? shared.cone : shared.czero);
}

/* This is an helper function for commands implementations that need to write
 * bits to a string object. The command creates or pad with zeroes the string
 * so that the 'maxbit' bit can be addressed. The object is finally
 * returned. Otherwise if the key holds a wrong type NULL is returned and
 * an error is sent to the client. */
robj *lookupStringForBitCommand(client *c, size_t maxbit) {
    size_t byte = maxbit >> 3;
    robj *o = lookupKeyWrite(c->db,c->argv[1]);
    if (checkType(c,o,OBJ_STRING)) return NULL;

    if (o == NULL) { // 如果没有这个对象，就创建并且保存
        o = createObject(OBJ_STRING,sdsnewlen(NULL, byte+1));
        dbAdd(c->db,c->argv[1],o);
    } else {
        // 如果有的话，看是否需要增加长度
        o = dbUnshareStringValue(c->db,c->argv[1],o);
        o->ptr = sdsgrowzero(o->ptr,byte+1);
    }
    return o;
}
```

## 总结

这一篇比较简单，我们仔细看了一下Redis是如何实现位图的，位运算比较难以理解，
我们以一个具体的例子进行了讲述。
