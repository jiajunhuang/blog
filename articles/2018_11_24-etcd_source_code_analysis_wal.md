# etcd源码阅读与分析（三）：wal

今天来看看WAL(Write-Ahead Logging)。这是数据库中保证数据持久化的常用技术，即每次真正操作数据之前，先往磁盘上追加一条日志，由于日志
是追加的，也就是顺序写，而不是随机写，所以写入性能还是很高的。这样做的目的是，如果在写入磁盘之前发生崩溃，那么数据肯定是没有写入
的，如果在写入后发生崩溃，那么还是可以从WAL里恢复出来。

首先看一下 `wal` 里有什么：

```bash
$ tree
.
├── decoder.go
├── doc.go
├── encoder.go
├── file_pipeline.go
├── file_pipeline_test.go
├── metrics.go
├── record_test.go
├── repair.go
├── repair_test.go
├── util.go
├── wal.go
├── wal_bench_test.go
├── wal_test.go
└── walpb
    ├── record.go
    ├── record.pb.go
    └── record.proto

1 directory, 16 files
```

我们先阅读 `doc.go`，可以知道这些东西：

- WAL这个抽象的结构体是由一堆的文件组成的
- 每个WAL文件的头部有一部分数据，是metadata
- 使用 `w.Save` 保存数据
- 使用完成之后，使用 `w.Close` 关闭
- WAL中的每一条记录，都有一个循环冗余校验码（CRC）
- WAL是只能打开来用于读，或者写，但是不能既读又写

我们看看 `Save` 的实现：

```go
func (w *WAL) Save(st raftpb.HardState, ents []raftpb.Entry) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// short cut, do not call sync
	if raft.IsEmptyHardState(st) && len(ents) == 0 {
		return nil
	}

	mustSync := raft.MustSync(st, w.state, len(ents))

	// TODO(xiangli): no more reference operator
	for i := range ents {
		if err := w.saveEntry(&ents[i]); err != nil {
			return err
		}
	}
	if err := w.saveState(&st); err != nil {
		return err
	}

	curOff, err := w.tail().Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	if curOff < SegmentSizeBytes {
		if mustSync {
			return w.sync()
		}
		return nil
	}

	return w.cut()
}
```

可以看出来，`Save` 做的事情，就是写入一条记录，然后调用 `w.sync`，而 `w.sync` 做的事情就是：

```go
func (w *WAL) sync() error {
	if w.encoder != nil {
		if err := w.encoder.flush(); err != nil {
			return err
		}
	}
	start := time.Now()
	err := fileutil.Fdatasync(w.tail().File)

	took := time.Since(start)
	if took > warnSyncDuration {
		if w.lg != nil {
			w.lg.Warn(
				"slow fdatasync",
				zap.Duration("took", took),
				zap.Duration("expected-duration", warnSyncDuration),
			)
		} else {
			plog.Warningf("sync duration of %v, expected less than %v", took, warnSyncDuration)
		}
	}
	walFsyncSec.Observe(took.Seconds())

	return err
```

调用了 `fileutil.Fdatasync`，而 `fileutil.Fdatasync` 就是调用了 `fsync` 这个系统调用保证数据会被写到磁盘。

而快照也是类似的，写入一条记录，然后同步。

```go
func (w *WAL) SaveSnapshot(e walpb.Snapshot) error {
	b := pbutil.MustMarshal(&e)

	w.mu.Lock()
	defer w.mu.Unlock()

	rec := &walpb.Record{Type: snapshotType, Data: b}
	if err := w.encoder.encode(rec); err != nil {
		return err
	}
	// update enti only when snapshot is ahead of last index
	if w.enti < e.Index {
		w.enti = e.Index
	}
	return w.sync()
}
```

WAL更多的是对多个WAL文件进行管理，WAL文件的命名规则是 `$seq-$index.wal`。第一个文件会是 `0000000000000000-0000000000000000.wal`，
此后，如果文件大小到了64M，就进行一次cut，比如，第一次cut的时候，raft的index是20，那么文件名就会变成 `0000000000000001-0000000000000021.wal`。

WAL就看到这。

---

- WAL: https://en.wikipedia.org/wiki/Write-ahead_logging

---

### etcd源码阅读与分析系列文章

- [etcd源码阅读与分析（一）：raftexample](https://jiajunhuang.com/articles/2018_11_20-etcd_source_code_analysis_raftexample.md.html)
- [etcd源码阅读与分析（二）：raft](https://jiajunhuang.com/articles/2018_11_22-etcd_source_code_analysis_raft.md.html)
- [etcd源码阅读与分析（三）：wal](https://jiajunhuang.com/articles/2018_11_24-etcd_source_code_analysis_wal.md.html)
- [etcd源码阅读与分析（四）：lease](https://jiajunhuang.com/articles/2018_11_27-etcd_source_code_analysis_lease.md.html)
- [etcd源码阅读与分析（五）：mvcc](https://jiajunhuang.com/articles/2018_11_28-etcd_source_code_analysis_mvvc.md.html)
