# etcd源码阅读（二）：raft

今天讲的是 `raft` 这个文件夹下的内容。我觉得etcd的代码写得不够好，当然，也有可能是因为我外行，不过这只是我的感受，不喜勿喷。

首先要看一下 `doc.go` 这个文件，里面写了很多注释，有利于理解，此外看 `raft` 文件夹下的代码，结合上一篇所说的raftexample一起来
理解，效果更佳。

首先，先把 `doc.go` 里的内容大概说一下：

- 新建一个raft集群，使用 `raft.StartNode`，传入一个Config和其他节点的ID
- 从已有的数据恢复raft集群，使用 `raft.RestartNode`，传入一个Config即可
- raft这个包只实现了raft协议，其余的例如数据持久化，处理数据等等，需要调用这个包的代码来做，调用者要做的事情包括：

    - 调用 `Node.Ready()` 接受目前产生的更新，然后：
        - 把HardState，Entries和Snapshot持久化到硬盘里
        - 把信息发送到To所指定的节点里去
        - 把快照和已经提交的Entry应用到状态机里去
        - 调用 `Node.Advance()` 通知Node之前调用 `Node.Ready()` 所接受的数据已经处理完毕
    - 所有持久化的操作都必须使用满足 `Storage` 这个接口的实现来进行持久化
    - 当接收到其他节点发来的消息时，调用 `Node.Step` 这个函数
    - 每隔一段时间需要主动调用一次 `Node.Tick()`

    把上面的几个步骤集中起来，差不多是这么些代码：

    ```go
    for {
        select {
        case <-s.Ticker:
        n.Tick()
        case rd := <-s.Node.Ready():
        saveToStorage(rd.State, rd.Entries, rd.Snapshot)
        send(rd.Messages)
        if !raft.IsEmptySnap(rd.Snapshot) {
            processSnapshot(rd.Snapshot)
        }
        for _, entry := range rd.CommittedEntries {
            process(entry)
            if entry.Type == raftpb.EntryConfChange {
            var cc raftpb.ConfChange
            cc.Unmarshal(entry.Data)
            s.Node.ApplyConfChange(cc)
            }
        }
        s.Node.Advance()
        case <-s.done:
        return
        }
    }
    ```

- 要处理接收到的请求，就去调用 `Node.Propose`，如果请求被提交了，就会出现在 `CommittedEntries` 里，并且状态是 `raftpb.EntryNormal`
- MessageType 有很多种类型，详见 `doc.go`

看完 `doc.go` 之后，我们再来看一眼 raftexample 中，`rc.serveChannels` 里的那一段代码：

```go
case rd := <-rc.node.Ready():
    rc.wal.Save(rd.HardState, rd.Entries)
    if !raft.IsEmptySnap(rd.Snapshot) {
        rc.saveSnap(rd.Snapshot)
        rc.raftStorage.ApplySnapshot(rd.Snapshot)
        rc.publishSnapshot(rd.Snapshot)
    }
    rc.raftStorage.Append(rd.Entries)
    rc.transport.Send(rd.Messages)
    if ok := rc.publishEntries(rc.entriesToApply(rd.CommittedEntries)); !ok {
        rc.stop()
        return
    }
    rc.maybeTriggerSnapshot()
    rc.node.Advance()
```

是不是和 `doc.go` 里说的一毛一样？

好了，接下来开始说 `raft` 的实现。为啥我觉得代码写的不好呢？因为实现上，有两个结构体，一个是 `node`，一个是 `raft`。我看下
来的感受是：

- `node` 负责节点相关的一些东西，`raft` 负责raft协议相关的东西，可能他们是想这么分开来，但实际上，`node.run` 的时候就需要
传入一个 `raft` ，也就是说其实运行的时候 `node` 和 `raft` 是捆绑在一起的。直接把它们放在一起，代码复杂度可以下降很多。而且他们本身
在raft协议里就是在一起的。
- `raft` 结构体里，表示各个节点的ID，使用的是一个 uint64，然后需要由调用者去自己根据ID追踪具体是什么URL，既然注定了要跨网络，何不把
网络操作封装在接口里，然后raft库本身来通过接口完成操作？这样可以进一步降低理解成本。

然后接下来我们结合 `raftexample` 的代码来理解 `raft` 文件夹下的代码，注意，我们目前暂时不去看 `etcdserver` 下的代码，那里是真正
跑etcd的代码，我瞄了一眼，等我们先多看几个底层的东西，再去看那里。[上一篇文章](https://jiajunhuang.com/articles/2018_11_20-etcd_source_code_analysis_raftexample.md.html)
我们说到，`raftexample` 里调用顺序是：

- `main` 函数
- `newRaftNode`
    - 新建一个 `raftNode` 并且调用 `raftNode.startRaft`，`raftNode.startRaft` 做的事情就是：
        - 检查WAL（Write Ahead Log）是否存在
        - 添加raft集群的其他节点
        - `rc.transport` 里添加其他节点
        - `rc.serveRaft()`
        - `rc.serveChannels()`
- `newKVStore`
- `serveHttpKVAPI`

`serveChannels` 做的事情呢，就是不断的接受 `rc.proposeC` 里的信息，而 `rc.proposeC` 信息的来源呢，就是 `serveHttpKVAPI` 里的HTTP接口接收到
请求，然后给塞进去的。`serveChannels` 接收到 `rc.proposeC` 里的信息呢，就调用 `rc.node.Propose`，这玩意儿呢，就是 `raft` 文件夹里，`node.go`
的 `Propose` 函数，因为他是个接口，而真正的实现就是 `raft/node.go` 里的 `Propose` 方法。

而 `raft/node.go` 里的 `Propose` 方法呢，最后就会调用 `func (n *node) stepWithWaitOption(ctx context.Context, m pb.Message, wait bool) error`，
它做的事情就是，把消息放到 `n.propc` 这个channel里，如果需要等待，那么就等待：

```go
select {
case rsp := <-pm.result: // 要等待的话，如果result不为空就返回，否则不返回（那就会执行到下面，返回nil）
    if rsp != nil {
        return rsp
    }
```

那么哪里会处理 `n.propc` 的消息呢？就在 `func (n *node) run(r *raft)` 这个函数里：

```go
select {
// TODO: maybe buffer the config propose if there exists one (the way
// described in raft dissertation)
// Currently it is dropped in Step silently.
case pm := <-propc: // proposal 是有结果的消息，应该是用来等待是否成功处理的
    m := pm.m
    m.From = r.id
    err := r.Step(m) // 注意，Step 是一个函数，这个函数用来处理消息。但是不同的身份有不同的Step实现，点进去看一下default里的代码，就调用了。参见 raft.go->becomeFollower, raft.go->becomeCandidate等等里的stepXXX函数
    if pm.result != nil {
        pm.result <- err
        close(pm.result)
    }
case m := <-n.recvc: // 收到消息，这里的消息应该是不等待结果的
    // filter out response message from unknown From.
    if pr := r.getProgress(m.From); pr != nil || !IsResponseMsg(m.Type) {
        r.Step(m)
    }
```

然后呢，你就发现，他收到消息之后，就会调用 `r.Step`，实现在 `raft.go` 里：

```go
// Step 就是传说中的状态机了
func (r *raft) Step(m pb.Message) error {
	// Handle the message term, which may result in our stepping down to a follower.
	switch {
```

这里呢，就是raft状态机，就是那一坨，如果消息的Term比自己的大，就主动变Follower那一坨规则。当然了，状态机我还没有仔细研究
每一个状态，毕竟，一开始就太深入细节，不方便理解，读代码的时候还是要注意不能一叶障目。瞄几眼，发现很多地方呢，其实会调用
`r.send`，然后呢，这嘎达，长这样：

```go
// send persists state to stable storage and then sends to its mailbox.
// TODO: send先持久化，然后发送到mailbox，那么问题来了，mailbox是什么？
func (r *raft) send(m pb.Message) {
	m.From = r.id
	if m.Type == pb.MsgVote || m.Type == pb.MsgVoteResp || m.Type == pb.MsgPreVote || m.Type == pb.MsgPreVoteResp {
		if m.Term == 0 {
			// All {pre-,}campaign messages need to have the term set when
			// sending.
			// - MsgVote: m.Term is the term the node is campaigning for,
			//   non-zero as we increment the term when campaigning.
			// - MsgVoteResp: m.Term is the new r.Term if the MsgVote was
			//   granted, non-zero for the same reason MsgVote is
			// - MsgPreVote: m.Term is the term the node will campaign,
			//   non-zero as we use m.Term to indicate the next term we'll be
			//   campaigning for
			// - MsgPreVoteResp: m.Term is the term received in the original
			//   MsgPreVote if the pre-vote was granted, non-zero for the
			//   same reasons MsgPreVote is
			panic(fmt.Sprintf("term should be set when sending %s", m.Type))
		}
	} else {
		if m.Term != 0 {
			panic(fmt.Sprintf("term should not be set when sending %s (was %d)", m.Type, m.Term))
		}
		// do not attach term to MsgProp, MsgReadIndex
		// proposals are a way to forward to the leader and
		// should be treated as local message.
		// MsgReadIndex is also forwarded to leader.
		if m.Type != pb.MsgProp && m.Type != pb.MsgReadIndex {
			m.Term = r.Term
		}
	}
	// TODO: 哪有持久化？？？
	r.msgs = append(r.msgs, m)
}
```

就是把消息追加到 `r.msgs` 这个 slice 里。绕这么大一圈，你说这实现闲的蛋疼么。那么现在的新问题是，哪里消费了 `r.msgs` ？
毕竟，`r.msgs` 目前还属于在内存的消息，注释里说要持久化，也没看到哪里持久化了。于是我就搜索了一下 `r.msgs`：

```bash
$ ack -Q 'r.msgs'
rawnode.go
218:	if len(r.msgs) > 0 || len(r.raftLog.unstableEntries()) > 0 || r.raftLog.hasNextEnts() {

raft_test.go
52:	msgs := r.msgs
53:	r.msgs = make([]pb.Message, 0)
738:		if len(r.msgs) != 1 {
739:			t.Errorf("%s,%s: %d response messages, want 1: %+v", vt, st, len(r.msgs), r.msgs)
741:			resp := r.msgs[0]

raft.go
470:	r.msgs = append(r.msgs, m)

node.go
427:			r.msgs = nil // 不是并发安全的啊
609:		Messages:         r.msgs,
```

我看 `node.go` 427行最可疑，就点进去看了一下，还真是！这叫基于瞎猜的代码阅读法。。。其实很多人读代码一开始都是东看看西看看，
自顶向下嘛，但是难免会有思路断开的时候，这种时候呢，继续多看看，自然到后边就会连起来。但是很多写源码分析的人不会写出来，
我这可是说了大实话了。

```go
case readyc <- rd: // Ready是各种准备好的变更
    if rd.SoftState != nil {
        prevSoftSt = rd.SoftState
    }
    if len(rd.Entries) > 0 {
        prevLastUnstablei = rd.Entries[len(rd.Entries)-1].Index
        prevLastUnstablet = rd.Entries[len(rd.Entries)-1].Term
        havePrevLastUnstablei = true
    }
    if !IsEmptyHardState(rd.HardState) {
        prevHardSt = rd.HardState
    }
    if !IsEmptySnap(rd.Snapshot) {
        prevSnapi = rd.Snapshot.Metadata.Index
    }
    if index := rd.appliedCursor(); index != 0 {
        applyingToI = index
    }

    r.msgs = nil // 不是并发安全的啊
    r.readStates = nil
    r.reduceUncommittedSize(rd.CommittedEntries)
    advancec = n.advancec
```

看下 `rd` 是啥，原来是`rd = newReady(r, prevSoftSt, prevHardSt)`。进去看看 `newReady` 干了什么：

```go
func newReady(r *raft, prevSoftSt *SoftState, prevHardSt pb.HardState) Ready {
	rd := Ready{
		Entries:          r.raftLog.unstableEntries(),
		CommittedEntries: r.raftLog.nextEnts(),
		Messages:         r.msgs,
	}
	if softSt := r.softState(); !softSt.equal(prevSoftSt) {
		rd.SoftState = softSt
	}
	if hardSt := r.hardState(); !isHardStateEqual(hardSt, prevHardSt) {
		rd.HardState = hardSt
	}
	if r.raftLog.unstable.snapshot != nil {
		rd.Snapshot = *r.raftLog.unstable.snapshot
	}
	if len(r.readStates) != 0 {
		rd.ReadStates = r.readStates
	}
	rd.MustSync = MustSync(r.hardState(), prevHardSt, len(rd.Entries))
	return rd
}
```

原来就是把 `r.msgs` 塞到 `r.msgs`，然后置空 `r.msgs`。好了，大概晓得了。

要注意到，`r.msgs = nil` 出现的这段代码，在我们上面说过的 `node.run` 这个函数里。上面我们说了，`run` 函数还有一个分支是
`case pm := <-propc`，我把几个分支抽出来看看：

```go
for {
    // 略略略，准备 rd

    select {
    // TODO: maybe buffer the config propose if there exists one (the way
    // described in raft dissertation)
    // Currently it is dropped in Step silently.
    case pm := <-propc: // proposal 是有结果的消息，应该是用来等待是否成功处理的
        m := pm.m
        m.From = r.id
        err := r.Step(m) // 注意，Step 是一个函数，这个函数用来处理消息。但是不同的身份有不同的Step实现，点进去看一下default里的代码，就调用了。参见 raft.go->becomeFollower, raft.go->becomeCandidate等等里的stepXXX函数
        if pm.result != nil {
            pm.result <- err
            close(pm.result)
        }
    case m := <-n.recvc: // 收到消息，这里的消息应该是不等待结果的
        // filter out response message from unknown From.
        if pr := r.getProgress(m.From); pr != nil || !IsResponseMsg(m.Type) {
            r.Step(m)
        }
    case cc := <-n.confc: // 配置变更
        // 略略略
    case <-n.tickc: // 心跳和选举的timeout，参见doc.go
        r.tick()
    case readyc <- rd: // Ready是各种准备好的变更
        // 略略略
    case <-advancec: // 这个是用来确认Ready已经处理完的
        // 略略略
    case c := <-n.status: // TODO: 好像也是状态变更？？？
        c <- getStatus(r)
    case <-n.stop: // 那就是stop咯
        close(n.done)
        return
    }
}
```

原来，是这样的。真的是有点绕啊。好了，这一篇分析就到这里了，其他的细节等我继续更新吧 :)
