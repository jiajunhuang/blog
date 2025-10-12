# TiDB 源码阅读（六）：TiDB Coprocessor 源码解析

TiDB 是存储和计算分离的设计，当 TiDB 的物理计划优化完成后，就需要将真正的取数请求发给 TiKV。而由于数据是分布在多个 TiKV
节点的，因此需要有一个框架来统筹计算，汇总结果，并将结果返回给 TiDB Server。这就是我们这篇文章要看的 coprocessor 模块。

## 一、背景与概念

### 1.1 什么是 Coprocessor？

在传统的数据库架构中，计算和存储通常是耦合在一起的。而在 TiDB 这样的分布式数据库中，存储层（TiKV）和计算层（TiDB Server）是分离的。
Coprocessor 就是实现**计算下推**的关键组件。

简单来说，Coprocessor 允许 TiDB 将部分计算逻辑（如过滤、聚合等）下推到 TiKV 节点执行，而不是将所有数据都拉取到 TiDB Server 再处理。这样做的好处是：

- **减少网络传输**：只传输过滤后的结果数据
- **并行计算**：多个 TiKV 节点可以并行处理各自的数据
- **提高性能**：利用 TiKV 的本地计算能力

### 1.2 一个例子

```sql
SELECT name, age FROM users WHERE age > 18;
```

在这个查询中：
1. TiDB 会构建一个 Coprocessor 请求，包含过滤条件 `age > 18`
2. 将请求发送到存储对应数据的多个 TiKV 节点
3. 每个 TiKV 节点在本地过滤数据，只返回符合条件的结果
4. TiDB 收集并合并各节点的结果

## 二、架构概览

TiDB Coprocessor 的整体架构可以分为以下几层：

```
┌─────────────────────────────────────────────┐
│           TiDB Server (SQL Layer)           │
│  ┌────────────────────────────────────┐    │
│  │        CopClient (发起请求)         │    │
│  └────────────────────────────────────┘    │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│         copIterator (任务调度器)             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
│  │ Worker 1 │  │ Worker 2 │  │ Worker N │ │
│  └──────────┘  └──────────┘  └──────────┘ │
└─────────────────────────────────────────────┘
         ↓             ↓             ↓
┌──────────┐    ┌──────────┐    ┌──────────┐
│  TiKV 1  │    │  TiKV 2  │    │  TiKV N  │
│ (Region) │    │ (Region) │    │ (Region) │
└──────────┘    └──────────┘    └──────────┘
```

## 三、核心数据结构

### 3.1 Store 和 CopClient

**Store** 是 Coprocessor 模块的入口，封装了底层的 TiKV 客户端：

```go
type Store struct {
    *kvStore
    coprCache       *coprCache    // coprocessor 缓存
    replicaReadSeed uint32        // 副本读随机种子
    numcpu          int           // CPU 核心数
}
```

**CopClient** 是真正发起 Coprocessor 请求的客户端：

```go
type CopClient struct {
    kv.RequestTypeSupportedChecker
    store           *Store
    replicaReadSeed uint32
}
```

它的核心方法是 `Send()`，负责构建请求并返回一个可迭代的响应。

### 3.2 copTask - 任务单元

`copTask` 代表一个需要发送到单个 Region 的 Coprocessor 任务：

```go
type copTask struct {
    taskID     uint64              // 任务 ID
    region     tikv.RegionVerID    // 目标 Region 信息
    bucketsVer uint64              // Bucket 版本
    ranges     *KeyRanges          // 要扫描的 Key 范围

    respChan   chan *copResponse   // 响应通道（用于 KeepOrder）
    storeAddr  string              // 目标 Store 地址
    cmdType    tikvrpc.CmdType     // 命令类型
    storeType  kv.StoreType        // Store 类型（TiKV/TiFlash）

    // 分页相关
    paging     bool
    pagingSize uint64

    // 批量任务
    batchTaskList map[uint64]*batchedCopTask

    // 其他元数据...
}
```

**关键点**：
- 一个 SQL 查询会被拆分成多个 `copTask`，每个对应一个 Region
- 每个 `copTask` 包含了该 Region 需要扫描的 Key 范围

### 3.3 copIterator - 任务调度器

`copIterator` 是整个 Coprocessor 执行的核心，负责：
- 任务分发
- 并发控制
- 结果收集

```go
type copIterator struct {
    store                *Store
    req                  *kv.Request         // 原始请求
    concurrency          int                 // 并发度
    smallTaskConcurrency int                 // 小任务额外并发度

    // 任务相关
    tasks []*copTask
    curr  int                                 // 当前处理到的任务索引

    // 通道相关
    respChan chan *copResponse                // 响应通道（无序）
    finishCh chan struct{}                    // 结束信号

    // 并发控制
    sendRate *util.RateLimit                  // 发送速率控制
    wg       sync.WaitGroup                   // 等待所有 worker 完成

    // 内存管理
    memTracker     *memory.Tracker
    actionOnExceed *rateLimitAction           // OOM 时的限流动作

    // 其他...
}
```

### 3.4 copIteratorWorker - 任务执行器

每个 worker 从任务通道取任务，发送到 TiKV 并处理响应：

```go
type copIteratorWorker struct {
    taskCh   <-chan *copTask              // 任务通道
    wg       *sync.WaitGroup
    store    *Store
    req      *kv.Request
    respChan chan<- *copResponse          // 结果通道
    finishCh <-chan struct{}

    vars       *tikv.Variables
    kvclient   *txnsnapshot.ClientHelper  // 底层 KV 客户端
    memTracker *memory.Tracker

    // 统计信息
    replicaReadSeed         uint32
    storeBatchedNum         *atomic.Uint64
    storeBatchedFallbackNum *atomic.Uint64
}
```

### 3.5 KeyRanges - Key 范围管理

`KeyRanges` 是一个优化的数据结构，用于高效管理 Key 范围：

```go
type KeyRanges struct {
    first *kv.KeyRange    // 头部额外范围
    mid   []kv.KeyRange   // 主体范围切片
    last  *kv.KeyRange    // 尾部额外范围
}
```

**设计亮点**：

- 通过 `first` 和 `last` 指针避免在头尾添加元素时重新分配大切片
- 提供 `Split()` 方法支持按 Key 切分范围

比如：场景：将 [a→z) 切分成 [a→m) 和 [m→z)

传统方案：

```
原始: [a→c) [c→f) [f→m) [m→s) [s→z)
               ↓ Split at 'm'
左边: [a→c) [c→f) [f→m)  ← 需要复制 3 个 KeyRange
右边:               [m→s) [s→z)  ← 需要复制 2 个 KeyRange
```

KeyRanges 方案:

```
原始: first=nil, mid=[a→c)[c→f)[f→z)], last=nil
                    ↓ Split at 'm'
左边: first=nil, mid=[a→c)[c→f)], last=&[f→m)   ← mid 共享底层数组！
右边: first=&[m→z)], mid=[], last=nil           ← 几乎零拷贝！
```

## 四、请求执行流程

### 4.1 整体流程图

```
CopClient.Send()
    ↓
BuildCopIterator()
    ↓
buildCopTasks()  ←─────┐
    ↓                   │
copIterator.open()      │ (Region 错误时重建)
    ↓                   │
启动 Workers            │
    ↓                   │
copIteratorTaskSender   │
    ↓                   │
worker.handleTask() ────┘
    ↓
copIterator.Next()
```

### 4.2 步骤一：构建 CopIterator

入口在 `CopClient.Send()` 方法：

```go
func (c *CopClient) Send(ctx context.Context, req *kv.Request,
    variables any, option *kv.ClientSendOption) kv.Response {

    // 1. 构建 copIterator
    it, errRes := c.BuildCopIterator(ctx, req, vars, option)
    if errRes != nil {
        return errRes
    }

    // 2. 启动 workers
    it.open(ctx, option.TryCopLiteWorker)
    return it
}
```

`BuildCopIterator()` 做了以下几件事：

```go
func (c *CopClient) BuildCopIterator(ctx context.Context, req *kv.Request,
    vars *tikv.Variables, option *kv.ClientSendOption) (*copIterator, kv.Response) {

    // 1. 创建 Backoffer (用于重试)
    bo := backoff.NewBackofferWithVars(ctx, copBuildTaskMaxBackoff, vars)

    // 2. 构建 copTask
    tasks, err := buildCopTasks(bo, ranges, buildOpt)

    // 3. 创建 copIterator
    it := &copIterator{
        store:       c.store,
        req:         req,
        concurrency: req.Concurrency,
        tasks:       tasks,
        // ... 其他初始化
    }

    // 4. 动态调整并发度
    if it.concurrency > len(tasks) {
        it.concurrency = len(tasks)
    }

    // 5. 创建响应通道
    if it.req.KeepOrder {
        it.sendRate = util.NewRateLimit(2 * it.concurrency)
        it.respChan = nil  // KeepOrder 模式使用 task.respChan
    } else {
        it.respChan = make(chan *copResponse)
        it.sendRate = util.NewRateLimit(it.concurrency)
    }

    return it, nil
}
```

### 4.3 步骤二：构建 copTask

`buildCopTasks()` 将 Key 范围拆分成多个任务：

```go
func buildCopTasks(bo *Backoffer, ranges *KeyRanges,
    opt *buildCopTaskOpt) ([]*copTask, error) {

    // 1. 通过 Region Cache 将 Key 范围按 Region 和 Bucket 切分
    locs, err := cache.SplitKeyRangesByBuckets(bo, ranges)

    // 2. 为每个 location 创建 copTask
    for _, loc := range locs {
        for i := 0; i < rLen; {
            // 限制单个 task 的 range 数量（默认 25000）
            nextI := min(i+rangesPerTaskLimit, rLen)

            task := &copTask{
                region:     loc.Location.Region,
                ranges:     loc.Ranges.Slice(i, nextI),
                cmdType:    tikvrpc.CmdCop,
                storeType:  req.StoreType,
                paging:     req.Paging.Enable,
                pagingSize: req.Paging.MinPagingSize,
                // ...
            }

            // KeepOrder 模式需要为每个 task 创建响应通道
            if req.KeepOrder {
                task.respChan = make(chan *copResponse, 2)
            }

            tasks = append(tasks, task)
        }
    }

    return tasks, nil
}
```

**关键机制**：
- **Region 定位**：通过 `SplitKeyRangesByBuckets()` 获取每个 Key 范围对应的 Region
- **范围限制**：每个 task 最多包含 25000 个范围，避免请求过大
- **Paging 支持**：如果启用分页，会设置初始的 `pagingSize`

### 4.4 步骤三：启动 Workers

`copIterator.open()` 启动并发的 worker 协程：

```go
func (it *copIterator) open(ctx context.Context, tryCopLiteWorker *atomic2.Uint32) {
    // 特殊优化：只有一个任务时使用轻量级 worker（避免启动 goroutine）
    if len(it.tasks) == 1 && tryCopLiteWorker != nil &&
        tryCopLiteWorker.CompareAndSwap(0, 1) {
        it.liteWorker = &liteCopIteratorWorker{
            ctx:    ctx,
            worker: newCopIteratorWorker(it, nil),
            tryCopLiteWorker: tryCopLiteWorker,
        }
        return
    }

    // 创建任务通道
    taskCh := make(chan *copTask, 1)
    it.wg.Add(it.concurrency + it.smallTaskConcurrency)

    // 如果有小任务，创建额外的小任务通道
    var smallTaskCh chan *copTask
    if it.smallTaskConcurrency > 0 {
        smallTaskCh = make(chan *copTask, 1)
    }

    // 启动 worker goroutines
    for i := range it.concurrency + it.smallTaskConcurrency {
        ch := taskCh
        if i >= it.concurrency && smallTaskCh != nil {
            ch = smallTaskCh
        }
        worker := newCopIteratorWorker(it, ch)
        go worker.run(ctx)
    }

    // 启动任务分发器
    taskSender := &copIteratorTaskSender{
        taskCh:      taskCh,
        smallTaskCh: smallTaskCh,
        wg:          &it.wg,
        tasks:       it.tasks,
        finishCh:    it.finishCh,
        sendRate:    it.sendRate,
        respChan:    it.respChan,
    }
    go taskSender.run(it.req.ConnID, it.req.RunawayChecker)
}
```

**并发控制亮点**：
- **动态并发**：根据任务数量和任务大小动态调整 worker 数量
- **小任务优化**：为小任务（行数少）分配额外的并发度
- **Lite Worker**：单任务时避免 goroutine 开销

### 4.5 步骤四：Worker 处理任务

Worker 的核心逻辑在 `handleTask()` 方法：

```go
func (worker *copIteratorWorker) handleTask(ctx context.Context,
    task *copTask, respCh chan<- *copResponse) {

    remainTasks := []*copTask{task}
    backoffermap := make(map[uint64]*Backoffer)

    // 循环处理任务（可能因为错误产生新的子任务）
    for len(remainTasks) > 0 {
        curTask := remainTasks[0]

        // 为每个 Region 独立使用 Backoffer
        bo := chooseBackoffer(ctx, backoffermap, curTask, worker)

        // 处理单次任务
        result, err := worker.handleTaskOnce(bo, curTask)
        if err != nil {
            // 发送错误响应
            resp := &copResponse{err: errors.Trace(err)}
            worker.sendToRespCh(resp, respCh)
            return
        }

        // 发送成功响应
        if result != nil {
            if result.resp != nil {
                worker.sendToRespCh(result.resp, respCh)
            }
            for _, resp := range result.batchRespList {
                worker.sendToRespCh(resp, respCh)
            }
        }

        // 处理剩余任务（Region 错误或锁错误会产生）
        if result != nil && len(result.remains) > 0 {
            remainTasks = append(result.remains, remainTasks[1:]...)
        } else {
            remainTasks = remainTasks[1:]
        }
    }
}
```

`handleTaskOnce()` 执行实际的 RPC 调用：

```go
func (worker *copIteratorWorker) handleTaskOnce(bo *Backoffer,
    task *copTask) (*copTaskResult, error) {

    // 1. 构建 Coprocessor 请求
    copReq := coprocessor.Request{
        Tp:        worker.req.Tp,
        StartTs:   worker.req.StartTs,
        Data:      worker.req.Data,
        Ranges:    task.ranges.ToPBRanges(),
        PagingSize: task.pagingSize,
        // ...
    }

    // 2. 构建 TiKV RPC 请求
    req := tikvrpc.NewReplicaReadRequest(task.cmdType, &copReq,
        replicaReadType, &worker.replicaReadSeed, context)

    // 3. 发送请求到 TiKV
    resp, rpcCtx, storeAddr, err := worker.kvclient.SendReqCtx(
        bo.TiKVBackoffer(), req, task.region, timeout,
        getEndPointType(task.storeType), task.storeAddr)

    if err != nil {
        return nil, errors.Trace(err)
    }

    copResp := resp.Resp.(*coprocessor.Response)

    // 4. 处理响应
    if worker.req.Paging.Enable {
        return worker.handleCopPagingResult(bo, rpcCtx,
            &copResponse{pbResp: copResp}, task, costTime)
    } else {
        return worker.handleCopResponse(bo, rpcCtx,
            &copResponse{pbResp: copResp}, task, costTime)
    }
}
```

### 4.6 步骤五：处理响应

`handleCopResponse()` 处理各种错误情况：

```go
func (worker *copIteratorWorker) handleCopResponse(bo *Backoffer,
    rpcCtx *tikv.RPCContext, resp *copResponse, task *copTask) (*copTaskResult, error) {

    // 1. 处理 Region 错误（Region 分裂、合并、迁移等）
    if regionErr := resp.pbResp.GetRegionError(); regionErr != nil {
        // Backoff 后重建任务
        if err := bo.Backoff(tikv.BoRegionMiss(), errors.New(errStr)); err != nil {
            return nil, errors.Trace(err)
        }

        // 重新构建 copTask
        remains, err := buildCopTasks(bo, task.ranges, buildOpt)
        if err != nil {
            return nil, err
        }
        return &copTaskResult{remains: remains}, nil
    }

    // 2. 处理锁错误
    if lockErr := resp.pbResp.GetLocked(); lockErr != nil {
        if err := worker.handleLockErr(bo, lockErr, task); err != nil {
            return nil, err
        }
        task.meetLockFallback = true
        return &copTaskResult{remains: []*copTask{task}}, nil
    }

    // 3. 处理其他错误
    if otherErr := resp.pbResp.GetOtherError(); otherErr != "" {
        err := errors.Errorf("other error: %s", otherErr)
        return nil, errors.Trace(err)
    }

    // 4. 正常响应：设置 startKey，收集执行信息
    resp.startKey = task.ranges.At(0).StartKey
    if err := worker.handleCollectExecutionInfo(bo, rpcCtx, resp); err != nil {
        return nil, err
    }

    // 5. 检查内存使用
    worker.checkRespOOM(resp)

    return &copTaskResult{resp: resp}, nil
}
```

### 4.7 步骤六：获取结果

上层调用 `copIterator.Next()` 逐个获取结果：

```go
func (it *copIterator) Next(ctx context.Context) (kv.ResultSubset, error) {
    var resp *copResponse

    // 1. Lite Worker 路径（单任务优化）
    if it.liteWorker != nil {
        resp = it.liteWorker.liteSendReq(ctx, it)
        // ...
    }
    // 2. 无序模式：从共享通道获取
    else if it.respChan != nil {
        resp, ok, closed = it.recvFromRespCh(ctx, it.respChan)
        if !ok || closed {
            return nil, errors.Trace(ctx.Err())
        }
        // finCopResp 是结束标记，递归获取下一个
        if resp == finCopResp {
            it.sendRate.PutToken()  // 归还令牌
            return it.Next(ctx)
        }
    }
    // 3. 有序模式：按任务顺序从各自通道获取
    else {
        for {
            if it.curr >= len(it.tasks) {
                return nil, nil  // 所有任务完成
            }
            task := it.tasks[it.curr]
            resp, ok, closed = it.recvFromRespCh(ctx, task.respChan)
            if closed {
                return nil, errors.Trace(ctx.Err())
            }
            if ok {
                break
            }
            // 当前任务完成，移到下一个
            it.sendRate.PutToken()
            it.tasks[it.curr] = nil
            it.curr++
        }
    }

    if resp.err != nil {
        return nil, errors.Trace(resp.err)
    }

    return resp, nil
}
```

**KeepOrder vs 无序模式**：
- **KeepOrder**：每个 task 有独立的 `respChan`，按顺序读取
- **无序**：所有 task 共享一个 `respChan`，谁先到谁先处理

## 五、关键机制

### 5.1 并发控制

TiDB Coprocessor 的并发控制非常精细：

#### 1. 基础并发度
```go
it.concurrency = req.Concurrency
if it.concurrency > len(tasks) {
    it.concurrency = len(tasks)
}
```

#### 2. 小任务额外并发
对于行数很少的"小任务"，额外分配并发度以提高吞吐：

```go
func smallTaskConcurrency(tasks []*copTask, numcpu int) (int, int) {
    res := 0
    for _, task := range tasks {
        if isSmallTask(task) {  // RowCountHint <= 32
            res++
        }
    }
    if res == 0 {
        return 0, 0
    }
    // 使用公式计算额外并发度
    extraConc := int(float64(res) / (1 + 0.5*math.Sqrt(2*math.Log(float64(res)))))

    // 限制不超过 smallConcPerCore * numcpu
    smallTaskConcurrencyLimit := 20 * numcpu
    if extraConc > smallTaskConcurrencyLimit {
        extraConc = smallTaskConcurrencyLimit
    }
    return res, extraConc
}
```

#### 3. 流量控制（Rate Limit）
使用令牌机制控制在途任务数量：

```go
func (sender *copIteratorTaskSender) run(connID uint64,
    checker resourcegroup.RunawayChecker) {

    for _, t := range sender.tasks {
        // 获取令牌（阻塞直到有可用令牌）
        exit := sender.sendRate.GetToken(sender.finishCh)
        if exit {
            break
        }

        // 发送任务
        exit = sender.sendToTaskCh(t, taskCh)
        if exit {
            break
        }
    }

    close(sender.taskCh)
    sender.wg.Wait()

    if sender.respChan != nil {
        close(sender.respChan)
    }
}
```

令牌容量：
- **KeepOrder**：`2 * concurrency`（允许更多在途任务）
- **无序**：`concurrency`

### 5.2 错误处理与重试

#### 1. Region 错误

Region 错误是分布式系统中常见的情况（分裂、合并、迁移等）：

```go
if regionErr := resp.pbResp.GetRegionError(); regionErr != nil {
    // 1. Backoff 等待
    if err := bo.Backoff(tikv.BoRegionMiss(), errors.New(errStr)); err != nil {
        return nil, errors.Trace(err)
    }

    // 2. 重新构建任务（会查询最新的 Region 信息）
    remains, err := buildCopTasks(bo, task.ranges, buildOpt)
    if err != nil {
        return nil, err
    }

    // 3. 返回新任务继续执行
    return &copTaskResult{remains: remains}, nil
}
```

#### 2. 锁错误

遇到未提交的事务锁时，需要解锁后重试：

```go
func (worker *copIteratorWorker) handleLockErr(bo *Backoffer,
    lockErr *kvrpcpb.LockInfo, task *copTask) error {

    if lockErr == nil {
        return nil
    }

    // 记录锁信息
    resolveLockDetail := worker.getLockResolverDetails()

    // 尝试解锁
    resolveLocksOpts := txnlock.ResolveLocksOptions{
        CallerStartTS: worker.req.StartTs,
        Locks:         []*txnlock.Lock{txnlock.NewLock(lockErr)},
        Detail:        resolveLockDetail,
    }
    resolveLocksRes, err := worker.kvclient.ResolveLocksWithOpts(
        bo.TiKVBackoffer(), resolveLocksOpts)

    if err != nil {
        return errors.Trace(err)
    }

    // 如果锁还未过期，等待一段时间
    msBeforeExpired := resolveLocksRes.TTL
    if msBeforeExpired > 0 {
        if err := bo.BackoffWithMaxSleepTxnLockFast(
            int(msBeforeExpired), errors.New(lockErr.String())); err != nil {
            return errors.Trace(err)
        }
    }

    return nil
}
```

#### 3. Backoffer 机制

每个 Region 使用独立的 Backoffer，避免一个 Region 的问题影响其他：

```go
func chooseBackoffer(ctx context.Context, backoffermap map[uint64]*Backoffer,
    task *copTask, worker *copIteratorWorker) *Backoffer {

    bo, ok := backoffermap[task.region.GetID()]
    if ok {
        return bo
    }

    // 为新 Region 创建独立的 Backoffer
    boMaxSleep := CopNextMaxBackoff  // 20000ms
    newbo := backoff.NewBackofferWithVars(ctx, boMaxSleep, worker.vars)
    backoffermap[task.region.GetID()] = newbo
    return newbo
}
```

### 5.3 Paging 分页请求

为了避免单次请求返回数据过多，TiDB 支持分页协议：

```go
func (worker *copIteratorWorker) handleCopPagingResult(bo *Backoffer,
    rpcCtx *tikv.RPCContext, resp *copResponse, task *copTask) (*copTaskResult, error) {

    // 1. 先处理响应
    result, err := worker.handleCopResponse(bo, rpcCtx, resp, task, costTime)
    if err != nil {
        return nil, errors.Trace(err)
    }

    // 2. 检查是否有剩余数据
    pagingRange := resp.pbResp.Range
    if pagingRange == nil {
        // TiKV 不支持分页或已返回全部数据
        return result, nil
    }

    // 3. 计算剩余范围
    task.ranges = worker.calculateRemain(task.ranges, pagingRange, worker.req.Desc)
    if task.ranges.Len() == 0 {
        return result, nil
    }

    // 4. 增长分页大小（指数增长）
    task.pagingSize = paging.GrowPagingSize(task.pagingSize,
        worker.req.Paging.MaxPagingSize)

    // 5. 将剩余任务加入待处理列表
    result.remains = []*copTask{task}
    return result, nil
}
```

**分页大小动态增长**：
- 初始：`MinPagingSize`（如 128 行）
- 每次翻倍增长
- 最大：`MaxPagingSize`（如 8192 行）

### 5.4 Store Batch 批量请求

对于小任务，可以将多个 Region 的请求批量发送到同一个 Store：

```go
type batchStoreTaskBuilder struct {
    bo          *Backoffer
    req         *kv.Request
    cache       *RegionCache
    taskID      uint64
    limit       int                          // 每批最多任务数
    store2Idx   map[storeReplicaKey]int      // Store -> Task 索引
    tasks       []*copTask
    replicaRead kv.ReplicaReadType
}

func (b *batchStoreTaskBuilder) handle(task *copTask) error {
    b.taskID++
    task.taskID = b.taskID

    // 只批量小任务
    if b.limit <= 0 || !isSmallTask(task) {
        b.tasks = append(b.tasks, task)
        return nil
    }

    // 构建批量任务
    batchedTask, err := b.cache.BuildBatchTask(b.bo, b.req, task, b.replicaRead)
    if err != nil {
        return err
    }

    key := storeReplicaKey{
        storeID:     batchedTask.storeID,
        replicaRead: batchedTask.loadBasedReplicaRetry,
    }

    // 查找或创建 Store 的批量任务
    if idx, ok := b.store2Idx[key]; !ok || len(b.tasks[idx].batchTaskList) >= b.limit {
        // 新建批量任务
        b.tasks = append(b.tasks, batchedTask.task)
        b.store2Idx[key] = len(b.tasks) - 1
    } else {
        // 添加到现有批量任务
        if b.tasks[idx].batchTaskList == nil {
            b.tasks[idx].batchTaskList = make(map[uint64]*batchedCopTask, b.limit)
        }
        b.tasks[idx].batchTaskList[task.taskID] = batchedTask
    }

    return nil
}
```

## 六、举个例子

`SELECT name FROM users ORDER BY id DESC LIMIT 20` 的执行过程：

```
假设数据分布在 3 个 Region：
Region 1: id [1,     30000)   
Region 2: id [30000, 60000)   
Region 3: id [60000, 100001)  ← 包含最大的 ID

因为是 DESC（降序），需要从大到小返回：

1️⃣ 构建 3 个 copTask
   ┌────────┐  ┌────────┐  ┌────────┐
   │Region1 │  │Region2 │  │Region3 │
   └────────┘  └────────┘  └────────┘

2️⃣ Desc=true → 反转任务顺序！
   ┌────────┐  ┌────────┐  ┌────────┐
   │Region3 │  │Region2 │  │Region1 │  
   └────────┘  └────────┘  └────────┘
      ↑ 先处理（最大ID）

3️⃣ KeepOrder=true → 每个 task 独立通道
   ┌────────┐      ┌──────────┐
   │Region3 │─────→│ respCh 3 │─┐
   └────────┘      └──────────┘ │
   ┌────────┐      ┌──────────┐ │
   │Region2 │─────→│ respCh 2 │─┼→ Next() 按顺序读
   └────────┘      └──────────┘ │
   ┌────────┐      ┌──────────┐ │
   │Region1 │─────→│ respCh 1 │─┘
   └────────┘      └──────────┘

4️⃣ 执行流程
   Worker → 处理 Region3
         ↓
   TiKV3 → 倒序扫描 (id=100000→99999...)
         ↓
   返回 128 行 (paging)
         ↓
   TiDB → 取前 20 行
         ↓
   Close() → 停止！Region2/Region1 不再处理
```

## 七、总结

这篇文章中，我们讲了 TiDB 的 Coprocessor 的实现，大概了解了一下 TiDB 与 TiKV 之间是如何交互的，最后以一个实际例子来看
了一下请求的过程。
