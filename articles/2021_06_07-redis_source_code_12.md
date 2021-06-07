# Redis源码阅读：pub/sub实现

今天我们来看看Redis是怎么实现 PUB/SUB 这个功能的，其实比较简单，但是我们还是从命令入手。PUB/SUB 命令用法如下：

```bash
SUBSCRIBE foo bar
PUBLISH second Hello
PSUBSCRIBE news.*
```

PUB/SUB 支持模糊匹配，我们先从 `SUBSCRIBE` 入手：

```c
    {"subscribe",subscribeCommand,-2,
     "pub-sub no-script ok-loading ok-stale",
     0,NULL,0,0,0,0,0,0},


/* SUBSCRIBE channel [channel ...] */
void subscribeCommand(client *c) {
    int j;
    if ((c->flags & CLIENT_DENY_BLOCKING) && !(c->flags & CLIENT_MULTI)) {
        /**
         * A client that has CLIENT_DENY_BLOCKING flag on
         * expect a reply per command and so can not execute subscribe.
         *
         * Notice that we have a special treatment for multi because of
         * backword compatibility
         */
        addReplyError(c, "SUBSCRIBE isn't allowed for a DENY BLOCKING client");
        return;
    }

    for (j = 1; j < c->argc; j++)
        pubsubSubscribeChannel(c,c->argv[j]);
    c->flags |= CLIENT_PUBSUB;
}

/* Subscribe a client to a channel. Returns 1 if the operation succeeded, or
 * 0 if the client was already subscribed to that channel. */
int pubsubSubscribeChannel(client *c, robj *channel) {
    dictEntry *de;
    list *clients = NULL;
    int retval = 0;

    /* Add the channel to the client -> channels hash table */
    // client存储了自己订阅了哪些channel
    if (dictAdd(c->pubsub_channels,channel,NULL) == DICT_OK) {
        retval = 1;
        incrRefCount(channel);
        /* Add the client to the channel -> list of clients hash table */
        // 服务端存储了channel和客户端列表的对应关系
        de = dictFind(server.pubsub_channels,channel);
        if (de == NULL) {
            clients = listCreate();
            dictAdd(server.pubsub_channels,channel,clients);
            incrRefCount(channel);
        } else {
            clients = dictGetVal(de);
        }
        listAddNodeTail(clients,c);
    }
    /* Notify the client */
    addReplyPubsubSubscribed(c,channel);
    return retval;
}

    {"psubscribe",psubscribeCommand,-2,
     "pub-sub no-script ok-loading ok-stale",
     0,NULL,0,0,0,0,0,0},


/* PSUBSCRIBE pattern [pattern ...] */
void psubscribeCommand(client *c) {
    int j;
    if ((c->flags & CLIENT_DENY_BLOCKING) && !(c->flags & CLIENT_MULTI)) {
        /**
         * A client that has CLIENT_DENY_BLOCKING flag on
         * expect a reply per command and so can not execute subscribe.
         *
         * Notice that we have a special treatment for multi because of
         * backword compatibility
         */
        addReplyError(c, "PSUBSCRIBE isn't allowed for a DENY BLOCKING client");
        return;
    }

    for (j = 1; j < c->argc; j++)
        pubsubSubscribePattern(c,c->argv[j]);
    c->flags |= CLIENT_PUBSUB;
}

/* Subscribe a client to a pattern. Returns 1 if the operation succeeded, or 0 if the client was already subscribed to that pattern. */
int pubsubSubscribePattern(client *c, robj *pattern) {
    dictEntry *de;
    list *clients;
    int retval = 0;

    // 客户端存储了当前订阅的模糊匹配的规则
    if (listSearchKey(c->pubsub_patterns,pattern) == NULL) {
        retval = 1;
        listAddNodeTail(c->pubsub_patterns,pattern);
        incrRefCount(pattern);
        /* Add the client to the pattern -> list of clients hash table */
        // 服务端同样存储了模糊匹配规则对应客户端列表的关系
        de = dictFind(server.pubsub_patterns,pattern);
        if (de == NULL) {
            clients = listCreate();
            dictAdd(server.pubsub_patterns,pattern,clients);
            incrRefCount(pattern);
        } else {
            clients = dictGetVal(de);
        }
        listAddNodeTail(clients,c);
    }
    /* Notify the client */
    addReplyPubsubPatSubscribed(c,pattern);
    return retval;
}
```

以上就是 `SUBSCRIBE` 和 `PSUBSCRIBE` 两个命令的实现。接下来我们看看当发布一条消息时，会发生什么：

```c
    {"publish",publishCommand,3,
     "pub-sub ok-loading ok-stale fast may-replicate",
     0,NULL,0,0,0,0,0,0},


/* PUBLISH <channel> <message> */
void publishCommand(client *c) {
    int receivers = pubsubPublishMessage(c->argv[1],c->argv[2]);
    if (server.cluster_enabled)
        clusterPropagatePublish(c->argv[1],c->argv[2]);
    else
        forceCommandPropagation(c,PROPAGATE_REPL);
    addReplyLongLong(c,receivers);
}

/* Publish a message */
int pubsubPublishMessage(robj *channel, robj *message) {
    int receivers = 0;
    dictEntry *de;
    dictIterator *di;
    listNode *ln;
    listIter li;

    /* Send to clients listening for that channel */
    // 找到精确匹配的客户端，然后逐个下发消息
    de = dictFind(server.pubsub_channels,channel);
    if (de) {
        list *list = dictGetVal(de);
        listNode *ln;
        listIter li;

        listRewind(list,&li);
        while ((ln = listNext(&li)) != NULL) {
            client *c = ln->value;
            addReplyPubsubMessage(c,channel,message);
            receivers++;
        }
    }
    /* Send to clients listening to matching channels */
    // 找到模糊匹配的客户端，然后逐个下发消息
    di = dictGetIterator(server.pubsub_patterns);
    if (di) {
        channel = getDecodedObject(channel);
        while((de = dictNext(di)) != NULL) {
            robj *pattern = dictGetKey(de);
            list *clients = dictGetVal(de);
            if (!stringmatchlen((char*)pattern->ptr,
                                sdslen(pattern->ptr),
                                (char*)channel->ptr,
                                sdslen(channel->ptr),0)) continue;

            listRewind(clients,&li);
            while ((ln = listNext(&li)) != NULL) {
                client *c = listNodeValue(ln);
                addReplyPubsubPatMessage(c,pattern,channel,message);
                receivers++;
            }
        }
        decrRefCount(channel);
        dictReleaseIterator(di);
    }
    return receivers;
}
```

以上其实就是 PUB/SUB 的实现，挺简单的，Redis分别使用 `pubsub_channels` 和 `pubsub_patterns` 来存储精确匹配和模糊匹配
的key和客户端的对应关系，当下发消息时，就找到对应的客户端，然后逐个下发消息。

---

ref:

- https://redis.io/topics/pubsub
