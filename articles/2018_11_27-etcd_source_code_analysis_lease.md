# etcd源码阅读与分析（四）：lease

lease是租约，类似于Redis中的TTL(Time To Live)。可以看一下怎么使用lease：

```go
cli, err := clientv3.New(clientv3.Config{
    Endpoints:   endpoints,
    DialTimeout: dialTimeout,
})
if err != nil {
    log.Fatal(err)
}
defer cli.Close()

// minimum lease TTL is 5-second
resp, err := cli.Grant(context.TODO(), 5)
if err != nil {
    log.Fatal(err)
}

// after 5 seconds, the key 'foo' will be removed
_, err = cli.Put(context.TODO(), "foo", "bar", clientv3.WithLease(resp.ID))
if err != nil {
    log.Fatal(err)
}
```

可以看出来，我们就是拿一个lease的ID作为凭证。那么，lease是怎么实现的呢？

```go
type Lease struct {
	ID           LeaseID
	ttl          int64 // time to live of the lease in seconds
	remainingTTL int64 // remaining time to live in seconds, if zero valued it is considered unset and the full ttl should be used
	// expiryMu protects concurrent accesses to expiry
	expiryMu sync.RWMutex
	// expiry is time when lease should expire. no expiration when expiry.IsZero() is true
	expiry time.Time

	// mu protects concurrent accesses to itemSet
	mu      sync.RWMutex
	itemSet map[LeaseItem]struct{}
	revokec chan struct{}
}
```

可以看出来，Lease在创建的时候，就会分配一个ID和设定好TTL。

接下来看看lessor这个接口的定义：

```go
// Lessor owns leases. It can grant, revoke, renew and modify leases for lessee.
type Lessor interface {
	// SetRangeDeleter lets the lessor create TxnDeletes to the store.
	// Lessor deletes the items in the revoked or expired lease by creating
	// new TxnDeletes.
	SetRangeDeleter(rd RangeDeleter)

	SetCheckpointer(cp Checkpointer)

	// Grant grants a lease that expires at least after TTL seconds.
	Grant(id LeaseID, ttl int64) (*Lease, error)
	// Revoke revokes a lease with given ID. The item attached to the
	// given lease will be removed. If the ID does not exist, an error
	// will be returned.
	Revoke(id LeaseID) error

	// Checkpoint applies the remainingTTL of a lease. The remainingTTL is used in Promote to set
	// the expiry of leases to less than the full TTL when possible.
	Checkpoint(id LeaseID, remainingTTL int64) error

	// Attach attaches given leaseItem to the lease with given LeaseID.
	// If the lease does not exist, an error will be returned.
	Attach(id LeaseID, items []LeaseItem) error

	// GetLease returns LeaseID for given item.
	// If no lease found, NoLease value will be returned.
	GetLease(item LeaseItem) LeaseID

	// Detach detaches given leaseItem from the lease with given LeaseID.
	// If the lease does not exist, an error will be returned.
	Detach(id LeaseID, items []LeaseItem) error

	// Promote promotes the lessor to be the primary lessor. Primary lessor manages
	// the expiration and renew of leases.
	// Newly promoted lessor renew the TTL of all lease to extend + previous TTL.
	Promote(extend time.Duration)

	// Demote demotes the lessor from being the primary lessor.
	Demote()

	// Renew renews a lease with given ID. It returns the renewed TTL. If the ID does not exist,
	// an error will be returned.
	Renew(id LeaseID) (int64, error)

	// Lookup gives the lease at a given lease id, if any
	Lookup(id LeaseID) *Lease

	// Leases lists all leases.
	Leases() []*Lease

	// ExpiredLeasesC returns a chan that is used to receive expired leases.
	ExpiredLeasesC() <-chan []*Lease

	// Recover recovers the lessor state from the given backend and RangeDeleter.
	Recover(b backend.Backend, rd RangeDeleter)

	// Stop stops the lessor for managing leases. The behavior of calling Stop multiple
	// times is undefined.
	Stop()
}
```

可以看出来作为一个lessor，也就是管理lease的东东，需要实现这些接口。我们具体关注怎么完成Grant，此外，expiration是怎么做的。
所以我们找到具体的实现来看看：

```go
// lessor implements Lessor interface.
// TODO: use clockwork for testability.
type lessor struct {
	mu sync.RWMutex

	// demotec is set when the lessor is the primary.
	// demotec will be closed if the lessor is demoted.
	demotec chan struct{}

	leaseMap            map[LeaseID]*Lease
	leaseHeap           LeaseQueue
	leaseCheckpointHeap LeaseQueue
	itemMap             map[LeaseItem]LeaseID

	// When a lease expires, the lessor will delete the
	// leased range (or key) by the RangeDeleter.
	rd RangeDeleter

	// When a lease's deadline should be persisted to preserve the remaining TTL across leader
	// elections and restarts, the lessor will checkpoint the lease by the Checkpointer.
	cp Checkpointer

	// backend to persist leases. We only persist lease ID and expiry for now.
	// The leased items can be recovered by iterating all the keys in kv.
	b backend.Backend

	// minLeaseTTL is the minimum lease TTL that can be granted for a lease. Any
	// requests for shorter TTLs are extended to the minimum TTL.
	minLeaseTTL int64

	expiredC chan []*Lease
	// stopC is a channel whose closure indicates that the lessor should be stopped.
	stopC chan struct{}
	// doneC is a channel whose closure indicates that the lessor is stopped.
	doneC chan struct{}

	lg *zap.Logger

	// Wait duration between lease checkpoints.
	checkpointInterval time.Duration
}
```

可以看到，里边有一个 `leaseMap`, 有 `leaseHeap`，为什么要有两个呢？堆的特性不知道大家还记得吗？此处的 `leaseHeap` 实现是
一个小堆，比较的关键是Lease失效的时间：

```go
type LeaseQueue []*LeaseWithTime

func (pq LeaseQueue) Len() int { return len(pq) }

func (pq LeaseQueue) Less(i, j int) bool {
	return pq[i].time < pq[j].time
}

func (pq LeaseQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *LeaseQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*LeaseWithTime)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *LeaseQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}
```

所以，怎么保证lease失效呢？我们每次从小堆里判断堆顶元素是否失效，失效就 `Pop` 就可以了。那为什么又要有 `leaseMap` 呢？因为
这样可以加速查找，毕竟，哈希表的时间复杂度是 `O(1)`。

那么，什么时候会进行lease的失效管理呢？我们看新建lessort的地方：

```go
func NewLessor(lg *zap.Logger, b backend.Backend, cfg LessorConfig) Lessor {
	return newLessor(lg, b, cfg)
}

func newLessor(lg *zap.Logger, b backend.Backend, cfg LessorConfig) *lessor {
	checkpointInterval := cfg.CheckpointInterval
	if checkpointInterval == 0 {
		checkpointInterval = 5 * time.Minute
	}
	l := &lessor{
		leaseMap:            make(map[LeaseID]*Lease),
		itemMap:             make(map[LeaseItem]LeaseID),
		leaseHeap:           make(LeaseQueue, 0),
		leaseCheckpointHeap: make(LeaseQueue, 0),
		b:                   b,
		minLeaseTTL:         cfg.MinLeaseTTL,
		checkpointInterval:  checkpointInterval,
		// expiredC is a small buffered chan to avoid unnecessary blocking.
		expiredC: make(chan []*Lease, 16),
		stopC:    make(chan struct{}),
		doneC:    make(chan struct{}),
		lg:       lg,
	}
	l.initAndRecover()

	go l.runLoop()

	return l
}
```

倒数第二行，`go l.runLoop()`，跟进去看：

```go
func (le *lessor) runLoop() {
	defer close(le.doneC)

	for {
		le.revokeExpiredLeases()
		le.checkpointScheduledLeases()

		select {
		case <-time.After(500 * time.Millisecond):
		case <-le.stopC:
			return
		}
	}
}
```

每500毫秒会进行一次循环，检查失效的lease然后传递到 `expiredC` 这个channel里。但是 `type lessor struct` 这个结构体并没有
对其进行处理，估计是调用者负责处理，所以我搜索了一下是不是有地方处理：

```bash
$ ack -Q 'ExpiredLeasesC()'
lease/lessor.go
126:	ExpiredLeasesC() <-chan []*Lease
559:func (le *lessor) ExpiredLeasesC() <-chan []*Lease {
906:func (fl *FakeLessor) ExpiredLeasesC() <-chan []*Lease { return nil }

lease/lessor_test.go
380:	case el := <-le.ExpiredLeasesC():
433:	case el := <-le.ExpiredLeasesC():

etcdserver/server.go
1013:		expiredLeaseC = s.lessor.ExpiredLeasesC()
```

果然，`etcdserver/server.go` 作为调用者对它进行处理。

最后，需要提到的是，lease也会进行持久化的，并且新建lessort的时候会优先看是否能从已有持久化的文件中恢复。

---

### etcd源码阅读与分析系列文章

- [etcd源码阅读与分析（一）：raftexample](https://jiajunhuang.com/articles/2018_11_20-etcd_source_code_analysis_raftexample.md.html)
- [etcd源码阅读与分析（二）：raft](https://jiajunhuang.com/articles/2018_11_22-etcd_source_code_analysis_raft.md.html)
- [etcd源码阅读与分析（三）：wal](https://jiajunhuang.com/articles/2018_11_24-etcd_source_code_analysis_wal.md.html)
- [etcd源码阅读与分析（四）：lease](https://jiajunhuang.com/articles/2018_11_27-etcd_source_code_analysis_lease.md.html)
- [etcd源码阅读与分析（五）：mvcc](https://jiajunhuang.com/articles/2018_11_28-etcd_source_code_analysis_mvvc.md.html)
