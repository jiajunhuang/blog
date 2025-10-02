# TiDB 源码阅读（一）：服务监听、请求处理流程概览

对于能独立运行的，接受请求的服务端，阅读源码，都是先从 `main` 函数开始。我比较喜欢的思路是从一个请求的处理流程入手，
看看一条SQL运行，到底是如何被 TiDB 处理的。

## 主流程分析

`main` 函数的入口在 `cmd/tidb-server/main.go`:

```go
func main() {
    // ...初始化配置和flag
	fset := initFlagSet()
	config.InitializeConfig(*configPath, *configCheck, *configStrict, overrideConfig, fset)

    // ...

    // 设置signal处理函数，全局变量，CPU affinity等
	signal.SetupUSR1Handler()
	setGlobalVars()
	setupSEM()
	err = setCPUAffinity()
	cgmon.StartCgroupMonitor()
	err = setupTracing() // Should before createServer and after setup config.
	setupMetrics()

    // create server
	svr := createServer(storage, dom)
	if standbyController != nil {
		standbyController.EndStandby(nil)

		svr.StandbyController = standbyController
		svr.StandbyController.OnServerCreated(svr)
	}

	exited := make(chan struct{})

    // 启动服务，处理请求
	topsql.SetupTopSQL(keyspace.GetKeyspaceNameBytesBySettings(), svr)
	terror.MustNil(svr.Run(dom))
	<-exited
	syncLog()
}
```

跟进 `svr.Run(dom)`:

```go
// Run runs the server.
func (s *Server) Run(dom *domain.Domain) error {
    // ...
	// If error should be reported and exit the server it can be sent on this
	// channel. Otherwise, end with sending a nil error to signal "done"
	err := s.initTiDBListener()

	// Register error API is not thread-safe, the caller MUST NOT register errors after initialization.
	// To prevent misuse, set a flag to indicate that register new error will panic immediately.
	// For regression of issue like https://github.com/pingcap/tidb/issues/28190
	go s.startNetworkListener(s.listener, false, errChan)
	go s.startNetworkListener(s.socket, true, errChan)

	s.health.Store(true)
	err = <-errChan
	if err != nil {
		return err
	}
	return <-errChan
}
```

继续跟进 `s.startNetworkListener`:

```go
func (s *Server) startNetworkListener(listener net.Listener, isUnixSocket bool, errChan chan error) {
	for {
		conn, err := listener.Accept()

        // ...

        // 初始化代表客户端连接的结构体
		clientConn := s.newConn(conn)

        // 开始处理
		go s.onConn(clientConn)
	}
}
```

这里就是很经典的 Golang TCP 服务的样子，起一个 for 循环，然后 Accept 一个连接就开启一个 goroutine 去处理。

继续看 `s.onConn`:

```go
func (s *Server) onConn(conn *clientConn) {
    // ...

    if err := conn.handshake(ctx); err != nil {
        // ...
    }

    // ...

    conn.Run(ctx)

    // ...
}

// Run reads client query and writes query result to client in for loop, if there is a panic during query handling,
// it will be recovered and log the panic error.
// This function returns and the connection is closed if there is an IO error or there is a panic.
func (cc *clientConn) Run(ctx context.Context) {
	// Usually, client connection status changes between [dispatching] <=> [reading].
	// When some event happens, server may notify this client connection by setting
	// the status to special values, for example: kill or graceful shutdown.
	// The client connection would detect the events when it fails to change status
	// by CAS operation, it would then take some actions accordingly.
	for {
		// Close connection between txn when we are going to shutdown server.
		// Note the current implementation when shutting down, for an idle connection, the connection may block at readPacket()
		// consider provider a way to close the connection directly after sometime if we can not read any data.
		if cc.server.inShutdownMode.Load() {
			if !sessVars.InTxn() {
				return
			}
		}

		if !cc.CompareAndSwapStatus(connStatusDispatching, connStatusReading) ||
			// The judge below will not be hit by all means,
			// But keep it stayed as a reminder and for the code reference for connStatusWaitShutdown.
			cc.getStatus() == connStatusWaitShutdown {
			return
		}

		// close connection when idle time is more than wait_timeout
		// default 28800(8h), FIXME: should not block at here when we kill the connection.
		waitTimeout := cc.getWaitTimeout(ctx)
		cc.pkt.SetReadTimeout(time.Duration(waitTimeout) * time.Second)
		start := time.Now()
		data, err := cc.readPacket()

        // ...

		// It should be CAS before checking the `inShutdownMode` to avoid the following scenario:
		// 1. The connection checks the `inShutdownMode` and it's false.
		// 2. The server sets the `inShutdownMode` to true. The `DrainClients` process ignores this connection
		//   because the connection is in the `connStatusReading` status.
		// 3. The connection changes its status to `connStatusDispatching` and starts to execute the command.
		if !cc.CompareAndSwapStatus(connStatusReading, connStatusDispatching) {
			return
		}

        // ...

		startTime := time.Now()
		err = cc.dispatch(ctx, data)
		cc.ctx.GetSessionVars().ClearAlloc(&cc.chunkAlloc, err != nil)

        // ...
	}
}
```

这里可以看到，进入 onConn 函数之后，首先就是握手，然后将当前连接的状态标记为不同的状态，最终进入 `cc.dispatch` 函数开始处理请求。

这里就涉及到MySQL协议属于半双工协议的概念，后面我们会专门讲到这件事，此处暂时跳过。继续跟进 `cc.dispatch`:

```go
// dispatch handles client request based on command which is the first byte of the data.
// It also gets a token from server which is used to limit the concurrently handling clients.
// The most frequently used command is ComQuery.
func (cc *clientConn) dispatch(ctx context.Context, data []byte) error {
    // ...

	cc.lastPacket = data
	cmd := data[0]
	data = data[1:]
	if topsqlstate.TopSQLEnabled() {
		rawCtx := ctx
		defer pprof.SetGoroutineLabels(rawCtx)
		sqlID := cc.ctx.GetSessionVars().SQLCPUUsages.AllocNewSQLID()
		ctx = topsql.AttachAndRegisterProcessInfo(ctx, cc.connectionID, sqlID)
	}

    // ...

	switch cmd {
	case mysql.ComPing, mysql.ComStmtClose, mysql.ComStmtSendLongData, mysql.ComStmtReset,
		mysql.ComSetOption, mysql.ComChangeUser:
		cc.ctx.SetProcessInfo("", t, cmd, 0)
	case mysql.ComInitDB:
		cc.ctx.SetProcessInfo("use "+dataStr, t, cmd, 0)
	}

	switch cmd {
	case mysql.ComQuit:
		return io.EOF
	case mysql.ComInitDB:
		node, err := cc.useDB(ctx, dataStr)
		cc.onExtensionStmtEnd(node, false, err)
		if err != nil {
			return err
		}
		return cc.writeOK(ctx)
	case mysql.ComQuery: // Most frequently used command.
		// For issue 1989
		// Input payload may end with byte '\0', we didn't find related mysql document about it, but mysql
		// implementation accept that case. So trim the last '\0' here as if the payload an EOF string.
		// See http://dev.mysql.com/doc/internals/en/com-query.html
		if len(data) > 0 && data[len(data)-1] == 0 {
			data = data[:len(data)-1]
			dataStr = string(hack.String(data))
		}
		return cc.handleQuery(ctx, dataStr)
	case mysql.ComFieldList:
		return cc.handleFieldList(ctx, dataStr)
	// ComCreateDB, ComDropDB
	case mysql.ComRefresh:
		return cc.handleRefresh(ctx, data[0])
	case mysql.ComShutdown: // redirect to SQL
		if err := cc.handleQuery(ctx, "SHUTDOWN"); err != nil {
			return err
		}
		return cc.writeOK(ctx)
	case mysql.ComStatistics:
		return cc.writeStats(ctx)
	// ComProcessInfo, ComConnect, ComProcessKill, ComDebug
	case mysql.ComPing:
		if cc.server.health.Load() {
			return cc.writeOK(ctx)
		}
		return servererr.ErrServerShutdown
	case mysql.ComChangeUser:
		return cc.handleChangeUser(ctx, data)
	// ComBinlogDump, ComTableDump, ComConnectOut, ComRegisterSlave
	case mysql.ComStmtPrepare:
		// For issue 39132, same as ComQuery
		if len(data) > 0 && data[len(data)-1] == 0 {
			data = data[:len(data)-1]
			dataStr = string(hack.String(data))
		}
		return cc.HandleStmtPrepare(ctx, dataStr)
	case mysql.ComStmtExecute:
		return cc.handleStmtExecute(ctx, data)
	case mysql.ComStmtSendLongData:
		return cc.handleStmtSendLongData(data)
	case mysql.ComStmtClose:
		return cc.handleStmtClose(data)
	case mysql.ComStmtReset:
		return cc.handleStmtReset(ctx, data)
	case mysql.ComSetOption:
		return cc.handleSetOption(ctx, data)
	case mysql.ComStmtFetch:
		return cc.handleStmtFetch(ctx, data)
	// ComDaemon, ComBinlogDumpGtid
	case mysql.ComResetConnection:
		return cc.handleResetConnection(ctx)
	// ComEnd
	default:
		return mysql.NewErrf(mysql.ErrUnknown, "command %d not supported now", nil, cmd)
	}
}
```

这里就可以看到，已经在根据不同的命令，做了不同的处理了。

此处有3个类型值得注意：`ComQuery` 和 `ComStmtPrepare` + `ComStmtExecute`。第一个对应我们常见的DML语句比如
`SELECT`, `UPDATE`, `DELETE` 语句的文字版(就是发一条执行一条，不带预编译)，第2/3个其实对应的就是预编译语句
的准备阶段和执行阶段，待会儿我们会看看代码，就知道为什么说预编译语句性能更好，以及它是怎么实现的了。

> 附加知识：DML 和 DDL，分别是啥呢？其实我一直都记不清楚，DML 是 Data Manipulation Language，数据操作语言比如 SELECT/UPDATE等；
> DDL 是 Data Definition Language，就是创建表，删除表等等管理性质的语句。但是由于这个概念不常用，我经常有点搞混。

先来看看 `handleQuery`:

```go
// handleQuery executes the sql query string and writes result set or result ok to the client.
// As the execution time of this function represents the performance of TiDB, we do time log and metrics here.
// Some special queries like `load data` that does not return result, which is handled in handleFileTransInConn.
func (cc *clientConn) handleQuery(ctx context.Context, sql string) (err error) {
    // ...

	if stmts, err = cc.ctx.Parse(ctx, sql); err != nil {
		cc.onExtensionSQLParseFailed(sql, err)

		// If an error happened, we'll need to remove the warnings in previous execution because the `ResetContextOfStmt` will not be called.
		// Ref https://github.com/pingcap/tidb/issues/59132
		sc.SetWarnings(sc.GetWarnings()[warnCountBeforeParse:])
		return err
	}

	if len(stmts) == 0 {
		return cc.writeOK(ctx)
	}

    // ...

	var pointPlans []base.Plan
	cc.ctx.GetSessionVars().InMultiStmts = false
    // ...
	for i, stmt := range stmts {
        // ...

		retryable, err = cc.handleStmt(ctx, stmt, parserWarns, i == len(stmts)-1)
		if err != nil {
			action, txnErr := sessiontxn.GetTxnManager(&cc.ctx).OnStmtErrorForNextAction(ctx, sessiontxn.StmtErrAfterQuery, err)
			if txnErr != nil {
				err = txnErr
				break
			}

			if retryable && action == sessiontxn.StmtActionRetryReady {
				cc.ctx.GetSessionVars().RetryInfo.Retrying = true
				_, err = cc.handleStmt(ctx, stmt, parserWarns, i == len(stmts)-1)
				cc.ctx.GetSessionVars().RetryInfo.Retrying = false
				if err != nil {
					break
				}
				continue
			}

            // ...

			_, err = cc.handleStmt(ctx, stmt, warns, i == len(stmts)-1)
		}
	}

	return err
}

func (cc *clientConn) handleStmt(
	ctx context.Context, stmt ast.StmtNode,
	warns []contextutil.SQLWarn, lastStmt bool,
) (bool, error) {
    // ...

	rs, err := cc.ctx.ExecuteStmt(ctx, stmt)

    // ...
}

// ExecuteStmt implements QueryCtx interface.
func (tc *TiDBContext) ExecuteStmt(ctx context.Context, stmt ast.StmtNode) (resultset.ResultSet, error) {
	var rs sqlexec.RecordSet
	if s, ok := stmt.(*ast.NonTransactionalDMLStmt); ok {
		rs, err = session.HandleNonTransactionalDML(ctx, s, tc.Session)
	} else {
		rs, err = tc.Session.ExecuteStmt(ctx, stmt)
	}
	if rs == nil {
		return nil, nil
	}
	return resultset.New(rs, nil), nil
}

func (s *session) ExecuteStmt(ctx context.Context, stmtNode ast.StmtNode) (sqlexec.RecordSet, error) {
    // ...
    // Transform abstract syntax tree to a physical plan(stored in executor.ExecStmt).
	compiler := executor.Compiler{Ctx: s}
	stmt, err := compiler.Compile(ctx, stmtNode)
    // ...

	var recordSet sqlexec.RecordSet
	if stmt.PsStmt != nil { // point plan short path
        // 点查，就是直接根据 CLUSTERED INDEX(一般也就是 Primary Key)能直接查到的
		recordSet, err = stmt.PointGet(ctx)
		s.setLastTxnInfoBeforeTxnEnd()
		s.txn.changeToInvalid()
	} else {
		recordSet, err = runStmt(ctx, s, stmt)
	}

    // ...

	return recordSet, nil
}

func runStmt(ctx context.Context, se *session, s sqlexec.Statement) (rs sqlexec.RecordSet, err error) {
    // ...
    rs, err = s.Exec(ctx)
    // ...
}
```

`pkg/session/session.go` 中的 `runStmt` 其实也没有真正的触发对tikv的请求，也就是没有开始查询数据，这个函数更多的
也是包装一下这些变量。那么在哪里真正触发请求的呢？这时候我们就要往回看看 sqlexec.RecordSet 是在哪里被消费：

```go
// The first return value indicates whether the call of handleStmt has no side effect and can be retried.
// Currently, the first return value is used to fall back to TiKV when TiFlash is down.
func (cc *clientConn) handleStmt(
	ctx context.Context, stmt ast.StmtNode,
	warns []contextutil.SQLWarn, lastStmt bool,
) (bool, error) {
    // ...

	// if stmt is load data stmt, store the channel that reads from the conn
	// into the ctx for executor to use

	rs, err := cc.ctx.ExecuteStmt(ctx, stmt)

    // ...

	if rs != nil {
        // ...
		if retryable, err := cc.writeResultSet(ctx, rs, false, status, 0); err != nil {
			return retryable, err
		}
		return false, nil
	}

    // ...

	return false, err
}
```

看起来是在 `writeResultSet` 函数中触发：

```go
// writeResultSet writes data into a result set and uses rs.Next to get row data back.
// If binary is true, the data would be encoded in BINARY format.
// serverStatus, a flag bit represents server information.
// fetchSize, the desired number of rows to be fetched each time when client uses cursor.
// retryable indicates whether the call of writeResultSet has no side effect and can be retried to correct error. The call
// has side effect in cursor mode or once data has been sent to client. Currently retryable is used to fallback to TiKV when
// TiFlash is down.
func (cc *clientConn) writeResultSet(ctx context.Context, rs resultset.ResultSet, binary bool, serverStatus uint16, fetchSize int) (retryable bool, runErr error) {
    // ...

	if retryable, err := cc.writeChunks(ctx, rs, binary, serverStatus); err != nil {
		return retryable, err
	}

	return false, cc.flush(ctx)
}

// writeChunks writes data from a Chunk, which filled data by a ResultSet, into a connection.
// binary specifies the way to dump data. It throws any error while dumping data.
// serverStatus, a flag bit represents server information
// The first return value indicates whether error occurs at the first call of ResultSet.Next.
func (cc *clientConn) writeChunks(ctx context.Context, rs resultset.ResultSet, binary bool, serverStatus uint16) (bool, error) {
	data := cc.alloc.AllocWithLen(4, 1024)
	req := rs.NewChunk(cc.chunkAlloc)
	gotColumnInfo := false
	firstNext := true
	validNextCount := 0
	var start time.Time
	var stmtDetail *execdetails.StmtExecDetails
	stmtDetailRaw := ctx.Value(execdetails.StmtExecDetailKey)
	if stmtDetailRaw != nil {
		//nolint:forcetypeassert
		stmtDetail = stmtDetailRaw.(*execdetails.StmtExecDetails)
	}
	for {
        // ...
		// Here server.tidbResultSet implements Next method.
		err := rs.Next(ctx, req)
		if err != nil {
			return firstNext, err
		}
		if !gotColumnInfo {
			// We need to call Next before we get columns.
			// Otherwise, we will get incorrect columns info.
			columns := rs.Columns()
			if stmtDetail != nil {
				start = time.Now()
			}
			if err = cc.writeColumnInfo(columns); err != nil {
				return false, err
			}
			if cc.capability&mysql.ClientDeprecateEOF == 0 {
				// metadata only needs EOF marker for old clients without ClientDeprecateEOF
				if err = cc.writeEOF(ctx, serverStatus); err != nil {
					return false, err
				}
			}
			if stmtDetail != nil {
				stmtDetail.WriteSQLRespDuration += time.Since(start)
			}
			gotColumnInfo = true
		}
		rowCount := req.NumRows()
		if rowCount == 0 {
			break
		}
		validNextCount++
		firstNext = false
		reg := trace.StartRegion(ctx, "WriteClientConn")
		if stmtDetail != nil {
			start = time.Now()
		}
		for i := range rowCount {
			data = data[0:4]
			if binary {
				data, err = column.DumpBinaryRow(data, rs.Columns(), req.GetRow(i), cc.rsEncoder)
			} else {
				data, err = column.DumpTextRow(data, rs.Columns(), req.GetRow(i), cc.rsEncoder)
			}
			if err != nil {
				reg.End()
				return false, err
			}
			if err = cc.writePacket(data); err != nil {
				reg.End()
				return false, err
			}
		}
		reg.End()
		if stmtDetail != nil {
			stmtDetail.WriteSQLRespDuration += time.Since(start)
		}
	}
	if err := rs.Finish(); err != nil {
		return false, err
	}

	if stmtDetail != nil {
		start = time.Now()
	}

	err := cc.writeEOF(ctx, serverStatus)
	if stmtDetail != nil {
		stmtDetail.WriteSQLRespDuration += time.Since(start)
	}
	return false, err
}
```

`rs.Next` 中触发了请求去查询，然后不断的迭代数据并且返回。继续追踪 `Next` 函数的调用和实现，可以追踪到
`pkg/executor/internal/exec/executor.go` 中的 `Next` 函数，然后看到 `Executor` 接口的定义：

```go
// Executor is the physical implementation of an algebra operator.
//
// In TiDB, all algebra operators are implemented as iterators, i.e., they
// support a simple Open-Next-Close protocol. See this paper for more details:
//
// "Volcano-An Extensible and Parallel Query Evaluation System"
//
// Different from Volcano's execution model, a "Next" function call in TiDB will
// return a batch of rows, other than a single row in Volcano.
// NOTE: Executors must call "chk.Reset()" before appending their results to it.
type Executor interface {
	NewChunk() *chunk.Chunk
	NewChunkWithCapacity(fields []*types.FieldType, capacity int, maxCachesize int) *chunk.Chunk

	RuntimeStats() *execdetails.BasicRuntimeStats

	HandleSQLKillerSignal() error
	RegisterSQLAndPlanInExecForTopSQL()

	AllChildren() []Executor
	SetAllChildren([]Executor)
	Open(context.Context) error
	Next(ctx context.Context, req *chunk.Chunk) error

	// `Close()` may be called at any time after `Open()` and it may be called with `Next()` at the same time
	Close() error
	Schema() *expression.Schema
	RetFieldTypes() []*types.FieldType
	InitCap() int
	MaxChunkSize() int

	// Detach detaches the current executor from the session context without considering its children.
	//
	// It has to make sure, no matter whether it returns true or false, both the original executor and the returning executor
	// should be able to be used correctly.
	Detach() (Executor, bool)
}
```

而在这里，我们看到了 `Volcano Execution Model` 这个概念，叫做火山模型，后续的文章中我们会讲讲，不过此处先跳过。
这个接口，如果查一下的话，会发现有很多 Executor 都实现了它，这就是物理执行引擎的实现。到这里为止，我们就可以大概
了解一下 TiDB 的整个执行过程：

1. 首先建立连接
2. 当请求来到时，onConn 函数准备并且记录好本次连接的状态
3. 然后开始握手
4. 开始处理请求，如果是普通SQL请求，那么就是在 `handleQuery` 函数中处理
5. 在 `handleQuery` 函数中，开始解析SQL语句，解析的第一步就是将文字版的SQL语句，转换成 AST 语法树
6. 接下来就是将 AST 语法树，转换成逻辑计划，然后进行逻辑优化
7. 最后就是将逻辑计划，转换成物理计划，然后进行物理优化
8. 构建好物理计划后，从 `writeResultSet` 函数中，开始迭代数据，迭代数据的开始就会触发对应的请求获取数据
9. 最后就是将数据返回给客户端

TiDB 的代码量实在是太大了，上述的代码无法将所有涉及到的点都贴出来，我也是反反复复看了好几遍，然后才梳理出这个流程，因此
如果你也在看的话，不妨也多看几遍代码，就会慢慢理解逻辑了。

## 预编译语句为什么性能更好

我们常常听说，预编译语句的性能更好，为什么呢？还得从源码看起：

```go
func (cc *clientConn) HandleStmtPrepare(ctx context.Context, sql string) error {
	stmt, columns, params, err := cc.ctx.Prepare(sql)
	if err != nil {
		return err
	}
	data := make([]byte, 4, 128)

	// status ok
	data = append(data, 0)
	// stmt id
	data = dump.Uint32(data, uint32(stmt.ID()))
	// number columns
	data = dump.Uint16(data, uint16(len(columns)))
	// number params
	data = dump.Uint16(data, uint16(len(params)))
	// filter [00]
	data = append(data, 0)
	// warning count
	data = append(data, 0, 0) // TODO support warning count

	if err := cc.writePacket(data); err != nil {
		return err
	}

    // ...

	return cc.flush(ctx)
}

// Prepare implements QueryCtx Prepare method.
func (tc *TiDBContext) Prepare(sql string) (statement PreparedStatement, columns, params []*column.Info, err error) {
	stmtID, paramCount, fields, err := tc.Session.PrepareStmt(sql)
	if err != nil {
		return
	}
	stmt := &TiDBStatement{
		sql:         sql,
		id:          stmtID,
		numParams:   paramCount,
		boundParams: make([][]byte, paramCount),
		ctx:         tc,
	}
	statement = stmt
	columns = make([]*column.Info, len(fields))
	for i := range fields {
		columns[i] = column.ConvertColumnInfo(fields[i])
	}
	params = make([]*column.Info, paramCount)
	for i := range params {
		params[i] = &column.Info{
			Type: mysql.TypeBlob,
		}
	}
	tc.stmts[int(stmtID)] = stmt
	return
}
```

可以看到，预编译语句，顾名思义就是先发给服务器解析好AST树，之后每一次请求都套用这个语句，变化的只有数据，因此每一次执行，
都不用重复的做编译和优化的动作：

```go
func (cc *clientConn) handleStmtExecute(ctx context.Context, data []byte) (err error) {
    // ...
	stmtID := binary.LittleEndian.Uint32(data[0:4])
	pos += 4

	stmt := cc.ctx.GetStatement(int(stmtID))
	if stmt == nil {
		return mysql.NewErr(mysql.ErrUnknownStmtHandler,
			strconv.FormatUint(uint64(stmtID), 10), "stmt_execute")
	}

	var (
		nullBitmaps []byte
		paramTypes  []byte
		paramValues []byte
	)
	cc.initInputEncoder(ctx)
	numParams := stmt.NumParams()
	args := make([]param.BinaryParam, numParams)

	sessVars := cc.ctx.GetSessionVars()
	// expiredTaskID is the task ID of the previous statement. When executing a stmt,
	// the StmtCtx will be reinit and the TaskID will change. We can compare the StmtCtx.TaskID
	// with the previous one to determine whether StmtCtx has been inited for the current stmt.
	expiredTaskID := sessVars.StmtCtx.TaskID
	err = cc.executePlanCacheStmt(ctx, stmt, args, useCursor)
	cc.onExtensionBinaryExecuteEnd(stmt, args, sessVars.StmtCtx.TaskID != expiredTaskID, err)
	return err
}
```

这里可以看到，执行的时候，会拿着服务器分配的 stmtID 来执行。

## 总结

这就是第一篇，请求流程概览的内容了，下一篇，我想再深入看看代码中握手的流程，同时看看MySQL的通信协议大概是啥样。
