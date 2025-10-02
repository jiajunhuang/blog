# TiDB 源码阅读（三）：插入数据

在这一篇文章中，我们主要来看看TiDB是怎么执行INSERT语句、如何编码数据写入数据的。

前面我们已经看过，一条SQL语句的大概执行过程，大概是解析AST、生成逻辑计划、优化逻辑计划、生成物理计划、优化物理计划、执行物理计划、返回数据。

对于简单的INSERT语句来说，没有这么多的步骤，比如执行 `INSERT INTO users(age, name) VALUES (10, "hello")`，逻辑计划优化和
物理计划优化都是走个过场，没有什么很多能优化的。但是，我们仍然要跟踪一下整个流程，看看你具体的执行。

## 源码分析

这一次，我们直接从 `handleQuery` 跟起，因为前面已经说过了，所有的文本DML，都是走的 `handleQuery` 函数：

```go
func (cc *clientConn) handleQuery(ctx context.Context, sql string) (err error) {
    // ...
    // 同样，也是先解析成AST树
    if stmts, err = cc.ctx.Parse(ctx, sql); err != nil {
        // ...
    }

    // ...
    // 然后调用 handleStmt
    		retryable, err = cc.handleStmt(ctx, stmt, parserWarns, i == len(stmts)-1)
}

// 最终调用到 ExecuteStmt

// ExecuteStmt implements QueryCtx interface.
func (tc *TiDBContext) ExecuteStmt(ctx context.Context, stmt ast.StmtNode) (resultset.ResultSet, error) {
	var rs sqlexec.RecordSet
	var err error
	if err = tc.checkSandBoxMode(stmt); err != nil {
		return nil, err
	}
	if s, ok := stmt.(*ast.NonTransactionalDMLStmt); ok {
		rs, err = session.HandleNonTransactionalDML(ctx, s, tc.Session)
	} else {
		rs, err = tc.Session.ExecuteStmt(ctx, stmt)
	}
	if err != nil {
		tc.Session.GetSessionVars().StmtCtx.AppendError(err)
		return nil, err
	}
	if rs == nil {
		return nil, nil
	}
	return resultset.New(rs, nil), nil
}

// 然后调用到 session 中的 ExecStmt

func (s *session) ExecuteStmt(ctx context.Context, stmtNode ast.StmtNode) (sqlexec.RecordSet, error) {
    // ...
    // Transform abstract syntax tree to a physical plan(stored in executor.ExecStmt).
	compiler := executor.Compiler{Ctx: s}
	stmt, err := compiler.Compile(ctx, stmtNode)
    // ...
}

// 然后到 runStmt
func runStmt(ctx context.Context, se *session, s sqlexec.Statement) (rs sqlexec.RecordSet, err error) {
    // ...
    	rs, err = s.Exec(ctx)
    // ...
}

// 然后到 pkg/executor/adapter.go

// Exec builds an Executor from a plan. If the Executor doesn't return result,
// like the INSERT, UPDATE statements, it executes in this function. If the Executor returns
// result, execution is done after this function returns, in the returned sqlexec.RecordSet Next method.
func (a *ExecStmt) Exec(ctx context.Context) (_ sqlexec.RecordSet, err error) {
    // ...
    e, err := a.buildExecutor()
    // ...
    if err = a.openExecutor(ctx, e); err != nil {
		terror.Log(exec.Close(e))
		return nil, err
	}
    // ...
    if handled, result, err := a.handleNoDelay(ctx, e, isPessimistic); handled || err != nil {
		return result, err
	}
    // ...
}

// buildExecutor build an executor from plan, prepared statement may need additional procedure.
func (a *ExecStmt) buildExecutor() (exec.Executor, error) {
    // ...

	ctx := a.Ctx
	stmtCtx := ctx.GetSessionVars().StmtCtx
	if _, ok := a.Plan.(*plannercore.Execute); !ok {
		if stmtCtx.Priority == mysql.NoPriority && a.LowerPriority {
			stmtCtx.Priority = kv.PriorityLow
		}
	}
	if _, ok := a.Plan.(*plannercore.Analyze); ok && ctx.GetSessionVars().InRestrictedSQL {
		ctx.GetSessionVars().StmtCtx.Priority = kv.PriorityLow
	}

	b := newExecutorBuilder(ctx, a.InfoSchema, a.Ti)
	e := b.build(a.Plan)
	if b.err != nil {
		return nil, errors.Trace(b.err)
	}

    // ...

	// ExecuteExec is not a real Executor, we only use it to build another Executor from a prepared statement.
	if executorExec, ok := e.(*ExecuteExec); ok {
		err := executorExec.Build(b)
		if err != nil {
			return nil, err
		}
		if executorExec.lowerPriority {
			ctx.GetSessionVars().StmtCtx.Priority = kv.PriorityLow
		}
		e = executorExec.stmtExec
	}
	a.isSelectForUpdate = b.hasLock && (!stmtCtx.InDeleteStmt && !stmtCtx.InUpdateStmt && !stmtCtx.InInsertStmt)
	return e, nil
}

// pkg/executor/adapter.go:877-921
func (a *ExecStmt) handleNoDelay(ctx context.Context, e exec.Executor, isPessimistic bool) (handled bool, rs sqlexec.RecordSet, err error) {
    // ... 前面省略 ...

    toCheck := e

    // 🔑 关键判断：检查执行器的 Schema
    if toCheck.Schema().Len() == 0 {  // Line 904
        // ✅ INSERT/UPDATE/DELETE 走这里
        // Schema 为空 = 没有返回结果
        handled = !isExplainAnalyze
        if isPessimistic {
            err := a.handlePessimisticDML(ctx, toCheck)
            return handled, nil, err
        }
        r, err := a.handleNoDelayExecutor(ctx, toCheck)
        return handled, r, err  // handled=true, r=nil
    } else if proj, ok := toCheck.(*ProjectionExec); ok && proj.calculateNoDelay {
        // DO 语句的特殊处理
        r, err := a.handleNoDelayExecutor(ctx, e)
        return true, r, err
    }

    // ✅ SELECT 走这里
    // Schema 不为空 = 有返回结果
    return false, nil, nil  // handled=false，继续执行
}

func (a *ExecStmt) handleNoDelayExecutor(ctx context.Context, e exec.Executor) (sqlexec.RecordSet, error) {
	sctx := a.Ctx
	r, ctx := tracing.StartRegionEx(ctx, "executor.handleNoDelayExecutor")
	defer r.End()

	var err error
	defer func() {
		terror.Log(exec.Close(e))
		a.logAudit()
	}()

	// Check if "tidb_snapshot" is set for the write executors.
	// In history read mode, we can not do write operations.
	// TODO: it's better to use a.ReadOnly to check if the statement is a write statement
	// instead of listing executor types here.
	switch e.(type) {
	case *DeleteExec, *InsertExec, *UpdateExec, *ReplaceExec, *LoadDataExec, *DDLExec, *ImportIntoExec:
		snapshotTS := sctx.GetSessionVars().SnapshotTS
		if snapshotTS != 0 {
			return nil, errors.New("can not execute write statement when 'tidb_snapshot' is set")
		}
		if sctx.GetSessionVars().UseLowResolutionTSO() {
			return nil, errors.New("can not execute write statement when 'tidb_low_resolution_tso' is set")
		}
	}

	err = a.next(ctx, e, exec.TryNewCacheChunk(e))
	if err != nil {
		return nil, err
	}
	err = a.handleStmtForeignKeyTrigger(ctx, e)
	return nil, err
}

// a.next 调用到 `exec.Next`:
// Next is a wrapper function on e.Next(), it handles some common codes.
func Next(ctx context.Context, e Executor, req *chunk.Chunk) (err error) {
    // ...
}
```

这就是INSERT的整个调用流程。

### 总体调用流程图

```
handleQuery
    ↓
ExecuteStmt (解析+编译)
    ↓
ExecStmt.Exec (构建执行器)
    ↓
handleNoDelayExecutor
    ↓
exec.Next (调用执行器)
    ↓
InsertExec.Next
    ↓
insertRows (准备数据)
    ↓
InsertExec.exec (批量写入)
    ↓
addRecord (逐行处理)
    ↓
TableCommon.AddRecord (表层)
    ↓
├─ 分配RowID
├─ 编码行数据 (tablecodec)
├─ 生成Key (t{tableID}_r{rowID})
├─ memBuffer.Set(key, value) [写入MemBuffer]
└─ addIndices (写入索引KV)
    ↓
[等待COMMIT命令]
    ↓
txn.Commit
    ↓
├─ Prewrite (加锁+写数据)
└─ Commit (删锁+数据可见)
    ↓
TiKV持久化 (Raft + RocksDB)
```

## InsertExec

由于最终其实要调用到 `InsertExec.Exec` 方法，我们得看看它的实现(`pkg/executor/insert.go`)：

```go
// InsertExec represents an insert executor.
type InsertExec struct {
	*InsertValues
	OnDuplicate    []*expression.Assignment
	evalBuffer4Dup chunk.MutRow
	curInsertVals  chunk.MutRow
	row4Update     []types.Datum

	Priority mysql.PriorityEnum
}

// Next implements the Executor Next interface.
func (e *InsertExec) Next(ctx context.Context, req *chunk.Chunk) error {
	req.Reset()
	if e.collectRuntimeStatsEnabled() {
		ctx = context.WithValue(ctx, autoid.AllocatorRuntimeStatsCtxKey, e.stats.AllocatorRuntimeStats)
	}

	if !e.EmptyChildren() && e.Children(0) != nil {
		return insertRowsFromSelect(ctx, e)
	}
	err := insertRows(ctx, e)
	if err != nil {
		terr, ok := errors.Cause(err).(*terror.Error)
		if ok && len(e.OnDuplicate) == 0 && terr.Code() == errno.ErrAutoincReadFailed {
			ec := e.Ctx().GetSessionVars().StmtCtx.ErrCtx()
			return ec.HandleError(err)
		}
		return err
	}
	return nil
}

func insertRows(ctx context.Context, base insertCommon) (err error) {
    // ...
    for i, list := range e.Lists {
		e.rowCount++
		var row []types.Datum
		row, err = evalRowFunc(ctx, list, i)
		if err != nil {
			return err
		}
    }
    // ...
	err = base.exec(ctx, rows)
    // ...
}

func (e *InsertExec) exec(ctx context.Context, rows [][]types.Datum) error {
    // ...
    				err = e.addRecord(ctx, row, dupKeyCheck)
    // ...
}

func (e *InsertValues) addRecordWithAutoIDHint(
	ctx context.Context, row []types.Datum, reserveAutoIDCount int, dupKeyCheck table.DupKeyCheckMode,
) (err error) {
	vars := e.Ctx().GetSessionVars()
	txn, err := e.Ctx().Txn(true)
	if err != nil {
		return err
	}
	pessimisticLazyCheck := getPessimisticLazyCheckMode(vars)
	if reserveAutoIDCount > 0 {
		_, err = e.Table.AddRecord(e.Ctx().GetTableCtx(), txn, row, table.WithCtx(ctx), table.WithReserveAutoIDHint(reserveAutoIDCount), dupKeyCheck, pessimisticLazyCheck)
	} else {
		_, err = e.Table.AddRecord(e.Ctx().GetTableCtx(), txn, row, table.WithCtx(ctx), dupKeyCheck, pessimisticLazyCheck)
	}
    // ...
}

// AddRecord implements table.Table AddRecord interface.
func (t *TableCommon) AddRecord(sctx table.MutateContext, txn kv.Transaction, r []types.Datum, opts ...table.AddRecordOption) (recordID kv.Handle, err error) {
	// TODO: optimize the allocation (and calculation) of opt.
	opt := table.NewAddRecordOpt(opts...)
	return t.addRecord(sctx, txn, r, opt)
}

func (t *TableCommon) addRecord(sctx table.MutateContext, txn kv.Transaction, r []types.Datum, opt *table.AddRecordOpt) (recordID kv.Handle, err error) {
    // ...
    		recordID, err = AllocHandle(ctx, sctx, t) // 分配行号
    // ...
    	key := t.RecordKey(recordID) // 编码生成key
    // ...
    	err = encodeRowBuffer.WriteMemBufferEncoded(sctx.GetRowEncodingConfig(), tc.Location(), ec, memBuffer, key, recordID, flags...) // 编码value并写入
    // ...

    // 编码写入索引
    // Insert new entries into indices.
	h, err := t.addIndices(sctx, recordID, r, txn, opt.GetCreateIdxOpt())
	if err != nil {
		return h, err
	}
}
```

### Key 和 Index Key 的编码

这里我直接把AI整理出来的贴出来，整理的很好！

#### 1. 表数据 Record Key（第822行）

##### 格式结构
```
t{tableID}_r{rowID}
```

##### 组成部分
- `t` - 表前缀（1字节）
- `{tableID}` - 编码后的表ID（使用 `codec.EncodeInt` 编码，8字节）
- `_r` - Record 分隔符（2字节）
- `{rowID}` - 编码后的行ID/Handle（使用 `codec.EncodeInt` 编码，对于int handle是8字节）

##### 具体例子
假设表ID为 `100`，行ID为 `1`:

```
Key (十六进制): 74 80 00 00 00 00 00 00 64 5f 72 80 00 00 00 00 00 00 01
                │  └────────tableID=100──────┘ │  │  └──────rowID=1────────┘
                t                              _  r

Key (可读形式): t\x80\x00\x00\x00\x00\x00\x00\x64_r\x80\x00\x00\x00\x00\x00\x00\x01
```

##### Record Value（表数据的值）

使用 **rowcodec** 编码格式存储：
- 包含所有列的 ID 和对应的值
- 格式：`[列1ID: 值1, 列2ID: 值2, ...]`

例如表有两列 `id INT`, `name VARCHAR(50)`:
```
Value: [1: 1, 2: "张三"]
编码后: <rowcodec格式的二进制数据>
```

---

#### 2. 索引 Index Key（第884行）

##### 格式结构

**非唯一索引：**
```
t{tableID}_i{indexID}{indexValues}{handle}
```

**唯一索引：**
```
t{tableID}_i{indexID}{indexValues}
```

##### 组成部分
- `t` - 表前缀（1字节）
- `{tableID}` - 编码后的表ID（8字节）
- `_i` - Index 分隔符（2字节）
- `{indexID}` - 编码后的索引ID（8字节）
- `{indexValues}` - 编码后的索引列值（使用 `codec.EncodeKey` 编码）
- `{handle}` - 对于非唯一索引，需要附加行handle以保证唯一性

##### 具体例子

假设：
- 表ID: `100`
- 索引ID: `1`
- 索引列值: `"Beijing"`（字符串）
- 行ID: `5`

**非唯一索引 Key:**
```
Key (十六进制): 74 80 00 00 00 00 00 00 64 5f 69 80 00 00 00 00 00 00 01 [Beijing编码] 03 80 00 00 00 00 00 00 05
                │  └────────tableID=100──────┘ │  │  └────indexID=1─────┘  └─索引值──┘  │  └─────handle=5─────┘
                t                              _  i                                    flag

组成:
- t                    : 表前缀
- 100 (encoded)        : 表ID
- _i                   : 索引分隔符
- 1 (encoded)          : 索引ID
- "Beijing" (encoded)  : 索引列值
- flag + 5 (encoded)   : handle（IntHandleFlag=0x03 + rowID）
```

**唯一索引 Key:**
```
Key (十六进制): 74 80 00 00 00 00 00 00 64 5f 69 80 00 00 00 00 00 00 01 [Beijing编码]
                │  └────────tableID=100──────┘ │  │  └────indexID=1─────┘  └─索引值──┘
                t                              _  i

（唯一索引不需要附加handle，因为索引值本身就是唯一的）
```

##### Index Value（索引的值）

**唯一索引的 Value:**
- 存储对应的行handle（rowID）
- 格式：8字节的 BigEndian uint64

```
Value (唯一索引): 00 00 00 00 00 00 00 05
                  └──────handle=5──────┘
```

**非唯一索引的 Value:**
- 通常为空或标记字节（因为handle已经在Key中了）
- 格式：可能是 `0x30` (字符'0') 或其他标记

```
Value (非唯一索引): 30
                    │
                    标记字节 '0'
```

**新版本的 Index Value (v5.0+):**
对于支持 CommonHandle、Global Index 或需要 RestoredData 的情况：
```
Value 格式:
[tailLen 1字节][可选的版本信息][CommonHandle信息][PartitionID信息][RestoredData]

例如:
00 7D 01 ...
│  │  │
│  │  └─ 版本号=1
│  └─ IndexVersionFlag (125)
└─ tailLen (尾部长度)
```

---

#### 3. 完整示例

假设执行 SQL：
```sql
CREATE TABLE users (
  id INT PRIMARY KEY,
  name VARCHAR(50),
  age INT,
  INDEX idx_name (name)
);

INSERT INTO users VALUES (1, 'Alice', 25);
```

假设表ID为100，索引ID为1，生成的KV对为：

##### 表数据
```
Key:   t\x80\x00\x00\x00\x00\x00\x00\x64_r\x80\x00\x00\x00\x00\x00\x00\x01
       (tableID=100, rowID=1)

Value: <rowcodec编码: {1: 1, 2: "Alice", 3: 25}>
```

##### 索引数据
```
Key:   t\x80\x00\x00\x00\x00\x00\x00\x64_i\x80\x00\x00\x00\x00\x00\x00\x01["Alice"编码]\x03\x80\x00\x00\x00\x00\x00\x00\x01
       (tableID=100, indexID=1, indexValue="Alice", handle=1)

Value: 0x30
       (非唯一索引的标记)
```

---

#### 4. 关键编码函数

- **Record Key**: `tablecodec.EncodeRecordKey(recordPrefix, handle)`
- **Index Key**: `tablecodec.GenIndexKey(loc, tblInfo, idxInfo, phyTblID, indexedValues, h, buf)`
- **Index Value**: `tablecodec.GenIndexValuePortal(...)`或`tablecodec.GenIndexValueForClusteredIndexVersion1(...)`

---

#### 5. 代码位置总结

| 位置 | 功能 | 代码 |
|------|------|------|
| tables.go:822 | 生成Record Key | `key := t.RecordKey(recordID)` |
| tables.go:858 | 写入Record KV | `encodeRowBuffer.WriteMemBufferEncoded(...)` |
| tables.go:884 | 添加索引 | `h, err := t.addIndices(...)` |
| tablecodec.go:1113 | Record前缀 | `GenTableRecordPrefix(tableID)` |
| tablecodec.go:1201 | 生成Index Key | `GenIndexKey(...)` |
| tablecodec.go:1584 | 生成Index Value | `GenIndexValueForClusteredIndexVersion1(...)` |

## 总结

这篇文章我们总结了一下INSERT语句的调用链，以及数据是如何编码保存到tikv中的。
