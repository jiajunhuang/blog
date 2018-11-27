# etcd源码阅读（五）：mvcc

[MVCC](https://en.wikipedia.org/wiki/Multiversion_concurrency_control) 是数据库中常见的一种并发控制的方式，即保存数据的多个版本，在同一个事务里，
应用所见的版本是一致的。

但是，我还是很想吐槽etcd的mvcc实现，有点乱，在我看来，是过度抽象了。为了理解mvcc，我们需要预先了解这些东西（下文，mvcc如无特别说明，都是指代mvcc文件夹下，etcd的mvcc实现）：

- mvcc底层使用 [bolt](https://github.com/etcd-io/bbolt) 实现，bolt是一个基于B+树的KV存储。
- `kv.go` 这个文件定义了大量的接口，然后接口之间又各种组合，但是其实最后 etcdserver 使用的就是 `ConsistentWatchableKV` 这个接口。

此处，我们预先了解bolt的一些东西，但是暂时不去探究bolt的实现，bolt的实现粗略的瞄了一眼，如果要写的话，有的写了：

> https://github.com/etcd-io/bbolt

- bolt的顶级是一个DB，DB里有多个bucket。在物理上，bolt使用单个文件存储。
- bolt在某一刻只允许一个 read-write 事务，但是可以同时允许多个 read-only 事务。其实就是读写锁，写只能顺序来，读可以并发读。
- `DB.Update()` 是用来开启 read-write 事务的，`DB.View()` 则是用来开启 read-only 事务的。由于每次执行 `DB.Update()` 都会写入一次磁盘，可以使用 `DB.Batch()` 来进行批量操作。

store是上面所说的 `ConsistentWatchableKV` 的底层实现：

```go
type store struct {
	ReadView
	WriteView

	// consistentIndex caches the "consistent_index" key's value. Accessed
	// through atomics so must be 64-bit aligned.
	consistentIndex uint64

	// mu read locks for txns and write locks for non-txn store changes.
	mu sync.RWMutex

	ig ConsistentIndexGetter

	b       backend.Backend
	kvindex index

	le lease.Lessor

	// revMuLock protects currentRev and compactMainRev.
	// Locked at end of write txn and released after write txn unlock lock.
	// Locked before locking read txn and released after locking.
	revMu sync.RWMutex
	// currentRev is the revision of the last completed transaction.
	currentRev int64
	// compactMainRev is the main revision of the last compaction.
	compactMainRev int64

	// bytesBuf8 is a byte slice of length 8
	// to avoid a repetitive allocation in saveIndex.
	bytesBuf8 []byte

	fifoSched schedule.Scheduler

	stopc chan struct{}

	lg *zap.Logger
}
```

可以看到其中有几个很重要的东西：

- `mu sync.RWMutex` 读写锁用于并发控制
- `b       backend.Backend` 则是bolt
- `kvindex index` 是用 Google 的B树实现做索引，至于为什么，后面会讲到

接下来探究一下 `Put` 是怎么工作的，这样我们就可以粗略的了解 mvcc 是怎么工作的。

store 本身并没有实现 `Put` 方法， 但是却可以调用 `Put` 方法，因为在最上边，它嵌套了一个匿名的 `WriteView`，从而获得了
这个方法：

```go
type store struct {
	ReadView
	WriteView
```

而具体的实现则在 `NewStore` 这个函数里可以找到：

```go
	s.ReadView = &readView{s}
	s.WriteView = &writeView{s}
```

那我们就去看 `writeView` 怎么实现的：

```go
func (wv *writeView) Put(key, value []byte, lease lease.LeaseID) (rev int64) {
	tw := wv.kv.Write()
	defer tw.End()
	return tw.Put(key, value, lease)
}
```

看看 `wv.kv.Write` 返回的是个啥嘎达(注意，wv.kv是一个符合KV这个interface的东东)：

```go
type KV interface {
	ReadView
	WriteView

	// Read creates a read transaction.
	Read() TxnRead

	// Write creates a write transaction.
	Write() TxnWrite
```

那我们就要去看 `TxnWrite.Put` 是怎么实现的：

```go
// TxnWrite represents a transaction that can modify the store.
type TxnWrite interface {
	TxnRead
	WriteView
	// Changes gets the changes made since opening the write txn.
	Changes() []mvccpb.KeyValue
}
```

原来又是接口，`Put` 是在 `WriteView` 里定义的。所以呢，绕了一圈，我们又绕回来了，所以，etcd搞得这么复杂干啥呢。。。为了找出具体实现，我们得去
`NewStore` 里翻翻，具体传进去的是什么。原来传进去的就是自己啊，那就说明，`type store struct` 这玩意儿，肯定实现了 `Write` 这个方法。但是呢，
我找来找去，就是没有发现。最后我只能开启搜索大法，然后在 `kvstore_txn.go` 这个文件里找到了：

```go
// 哇好绕啊，又绕到这里来了，我是佩服的
func (s *store) Write() TxnWrite {
	s.mu.RLock()
	tx := s.b.BatchTx()
	tx.Lock()
	tw := &storeTxnWrite{
		storeTxnRead: storeTxnRead{s, tx, 0, 0},
		tx:           tx,
		beginRev:     s.currentRev,
		changes:      make([]mvccpb.KeyValue, 0, 4),
	}
	return newMetricsTxnWrite(tw)
}
```

我是服气的。

接下来就要去翻 `TxnWrite` 的实现，也就是 `storeTxnWrite` 的 `Put` 的实现了，然后你会发现，`Put` 调用了 `put`：

```go
// etcdctl put foo bar最后到了这里
func (tw *storeTxnWrite) put(key, value []byte, leaseID lease.LeaseID) {
	rev := tw.beginRev + 1
	c := rev
	oldLease := lease.NoLease

	// if the key exists before, use its previous created and
	// get its previous leaseID
	_, created, ver, err := tw.s.kvindex.Get(key, rev)
	if err == nil {
		c = created.main
		oldLease = tw.s.le.GetLease(lease.LeaseItem{Key: string(key)})
	}

	ibytes := newRevBytes() // revision的bytes
	idxRev := revision{main: rev, sub: int64(len(tw.changes))}
	revToBytes(idxRev, ibytes)

	ver = ver + 1
	kv := mvccpb.KeyValue{
		Key:            key,
		Value:          value,
		CreateRevision: c,
		ModRevision:    rev,
		Version:        ver,
		Lease:          int64(leaseID),
	}

	d, err := kv.Marshal() // kv的bytes
	if err != nil {
		if tw.storeTxnRead.s.lg != nil {
			tw.storeTxnRead.s.lg.Fatal(
				"failed to marshal mvccpb.KeyValue",
				zap.Error(err),
			)
		} else {
			plog.Fatalf("cannot marshal event: %v", err)
		}
	}

	tw.tx.UnsafeSeqPut(keyBucketName, ibytes, d) // 所以最后存储的，以revision为key，kv为value存储下来了
	tw.s.kvindex.Put(key, idxRev)
	tw.changes = append(tw.changes, kv)

	if oldLease != lease.NoLease {
		if tw.s.le == nil {
			panic("no lessor to detach lease")
		}
		err = tw.s.le.Detach(oldLease, []lease.LeaseItem{{Key: string(key)}})
		if err != nil {
			if tw.storeTxnRead.s.lg != nil {
				tw.storeTxnRead.s.lg.Fatal(
					"failed to detach old lease from a key",
					zap.Error(err),
				)
			} else {
				plog.Errorf("unexpected error from lease detach: %v", err)
			}
		}
	}
	if leaseID != lease.NoLease {
		if tw.s.le == nil {
			panic("no lessor to attach lease")
		}
		err = tw.s.le.Attach(leaseID, []lease.LeaseItem{{Key: string(key)}})
		if err != nil {
			panic("unexpected error from lease Attach")
		}
	}
}
```

这里可以看出来，bolt里存储的KV，实际上并不是用户给出的KV。Key是revision，而Value是用户给出的KV。所以，才需要那个b树做索引，
把用户的key换成revision，然后再到bolt里，把revision换成真正的KV。

之后我们再看mvcc实现里的其他特性，例如watch是怎么实现的，实际上这玩意儿存储的时候，是有序存储的。

好了。这一节就看到这里了。
