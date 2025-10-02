# TiDB 源码阅读（二）：MySQL协议概览

今天我们来结合TiDB源码，一起看看MySQL通信协议大概的样子，不会深究到每一个字节，有个大概了解即可。

> 部分内容是AI生成，写的挺好的我就懒得敲键盘了。

## 概述

MySQL 通信协议是 MySQL 客户端（如命令行工具、JDBC 驱动、ORM 框架等）与服务器（mysqld）进行交互时所遵循的一套规则。
这是一套**基于 TCP/IP 的二进制协议**，设计目标是轻量、高效。

协议的核心是**请求-响应模型**：客户端发送一个命令包，服务器返回一个或多个响应包。

## 协议分层与连接生命周期

一次完整的 MySQL 会话可以分为以下几个阶段：

### 1. 连接建立阶段

1.  **TCP 三次握手**：客户端与服务器的 `3306` 端口建立 TCP 连接。
2.  **握手初始化**：
    *   服务器主动发送一个 **`Greeting` 包**（或 `Handshake Initialization` 包）给客户端。这个包包含了：
        *   协议版本
        *   服务器版本（例如：`8.0.33`）
        *   服务器线程 ID
        *   随机生成的 **Auth Seed**（用于后续的密码加密）
        *   服务器能力标志（如是否支持 SSL、是否使用 4.1 认证协议等）
3.  **客户端认证**：
    *   客户端收到 `Greeting` 包后，发送一个 **`Login Request` 包**。这个包包含了：
        *   客户端能力标志
        *   用户名
        *   加密后的密码（使用 Auth Seed 和用户密码通过算法生成）
        *   默认数据库名称（可选）
        *   字符集信息
4.  **认证结果**：
    *   服务器验证客户端提供的凭据。如果成功，返回一个 **`OK Packet`**；如果失败，返回一个 **`ERR Packet`**，并关闭连接。

> **注意**：从 MySQL 8.0.4 开始，默认的认证插件从 `mysql_native_password` 切换到了 `caching_sha2_password`。这两者使用的密码加密算法不同，这也是很多老版本客户端/驱动连接新版本服务器时出现 `authentication protocol` 错误的原因。

### 2. 命令执行阶段

认证成功后，连接进入命令执行阶段。这个阶段的通信模式固定：

**`COM_QUERY` 命令（文本协议）**
这是最常用的命令，用于执行 SQL 语句（如 `SELECT`, `INSERT`, `UPDATE` 等）。

1.  **客户端发送 `COM_QUERY` 包**：包内容就是纯文本的 SQL 语句。
2.  **服务器响应**：响应类型取决于 SQL 语句的类型。
    *   **对于 `SELECT` / `SHOW` 等返回数据的语句**：
        *   第一步：发送一个 **`Column Definition` 包**，描述结果集的列结构（列名、类型、长度等）。有多少列，就会发送多少个这样的包。
        *   第二步：发送一个 **`EOF Packet`**（在旧协议中）或直接开始发送数据行，来标识列定义部分的结束。
        *   第三步：发送多个 **`Row Data` 包**，每个包包含一行的数据（二进制格式，按照列定义解析）。
        *   第四步：发送一个 **`EOF Packet`** 或 **`OK Packet`** 来标识结果集的结束，并包含附加信息如 `affected rows`, `last insert id`, `warnings` 等。
    *   **对于 `INSERT` / `UPDATE` / `DELETE` 等不返回数据的语句**：
        *   服务器直接返回一个 **`OK Packet`**，其中包含了受影响的行数、上次插入的 ID 等信息。
    *   **对于出错的语句**：
        *   服务器返回一个 **`ERR Packet`**，其中包含错误码和错误信息。

**`COM_STMT_PREPARE` 和 `COM_STMT_EXECUTE` 命令（二进制协议/预处理语句）**
这是更高效的命令执行方式，主要用于防止 SQL 注入和提高重复查询的性能。

1.  **准备阶段**：客户端发送 `COM_STMT_PREPARE` 包，包含带占位符（`?`）的 SQL 语句。
    *   服务器解析 SQL，生成一个预处理语句，并返回一个 **`STMT_PREPARE_OK`** 包，其中包含了**语句 ID**、参数个数、结果集列数等信息。
2.  **参数绑定**：客户端使用语句 ID，发送 `COM_STMT_SEND_LONG_DATA`（用于发送 BLOB 等长数据）和 `COM_STMT_EXECUTE` 包。在 `EXECUTE` 包中，客户端以二进制形式发送每个占位符参数的值。
3.  **执行阶段**：服务器收到 `COM_STMT_EXECUTE` 后，使用绑定的参数执行预处理好的语句，返回结果的过程与 `COM_QUERY` 类似（列定义、数据行、EOF/OK）。
4.  **销毁阶段**：执行完毕后，客户端可以发送 `COM_STMT_CLOSE` 来释放服务器端的资源。

**文本协议 vs. 二进制协议**
*   **文本协议**：SQL 和结果集都以字符串形式传输，客户端需要做字符串到特定数据类型的解析和转换。
*   **二进制协议**：参数和数据以原生二进制格式传输，效率更高。并且通过预处理，可以避免 SQL 注入，因为数据和指令是分开发送的。

### 3. 连接关闭阶段

*   客户端可以发送 **`COM_QUIT`** 命令，优雅地关闭连接。
*   服务器或客户端任何一方也可以直接关闭底层的 TCP 连接。

由此我们可以看出来，MySQL通信协议是半双工协议，此处我们需要了解一下什么是半双工：

半双工

比喻： 像对讲机或一条单车道的桥梁。

定义： 通信的双方都可以发送和接收数据，但不能同时进行。在同一时刻，只能有一方在说话，另一方必须听着。当一方说完后，需要说一个“完毕”（并释放信道），另一方才能开始说话。

关键特点： 轮流传输。存在明确的“发言权”切换。

全双工

比喻： 像电话通话或一条双车道的桥梁。

定义： 通信的双方都可以同时发送和接收数据。你说你的，我听我的，我也可以同时说我的，你听你的。两个方向的数据流互不干扰。

关键特点： 并发传输。发送和接收是完全独立的。

一开始我以为只有MySQL是这样，但是搜索一番发现PG也是半双工，想想也是，我们用PG、MySQL都要调节连接池参数，半双工就是原因
所在，因为建立一个连接以后，这个连接上只能处于一种状态，要么发送，要么接收数据，当并发上来以后，如果不使用连接池，就会
阻塞和等待。

---

## 数据包结构

所有在网络上传输的 MySQL 数据包都有统一的头部结构：

| 字段 | 长度（字节） | 描述 |
| :--- | :--- | :--- |
| **Payload Length** | 3 | 数据包体（Payload）的长度（小端序）。 |
| **Sequence ID** | 1 | 数据包序列号。从 0 开始，每发送一个新的请求/响应就加 1，用于检测数据包是否丢失或乱序。 |
| **Payload** | `Payload Length` | 实际的数据，可能是命令、响应数据、错误信息等。 |

---

## 源码分析

我们此处主要看看握手的代码：

```go
// handshake works like TCP handshake, but in a higher level, it first writes initial packet to client,
// during handshake, client and server negotiate compatible features and do authentication.
// After handshake, client can send sql query to server.
func (cc *clientConn) handshake(ctx context.Context) error {
    // 服务端主动发送Greeting包
	if err := cc.writeInitialHandshake(ctx); err != nil {
		if errors.Cause(err) == io.EOF {
			logutil.Logger(ctx).Debug("Could not send handshake due to connection has be closed by client-side")
		} else {
			logutil.Logger(ctx).Debug("Write init handshake to client fail", zap.Error(errors.SuspendStack(err)))
		}
		return err
	}
    // 客户端发送Login请求，服务端读取并校验
	if err := cc.readOptionalSSLRequestAndHandshakeResponse(ctx); err != nil {
		err1 := cc.writeError(ctx, err)
		if err1 != nil {
			logutil.Logger(ctx).Debug("writeError failed", zap.Error(err1))
		}
		return err
	}

	// MySQL supports an "init_connect" query, which can be run on initial connection.
	// The query must return a non-error or the client is disconnected.
	if err := cc.initConnect(ctx); err != nil {
		logutil.Logger(ctx).Warn("init_connect failed", zap.Error(err))
		initErr := servererr.ErrNewAbortingConnection.FastGenByArgs(cc.connectionID, "unconnected", cc.user, cc.peerHost, "init_connect command failed")
		if err1 := cc.writeError(ctx, initErr); err1 != nil {
			terror.Log(err1)
		}
		return initErr
	}

    // 服务端验证通过，返回OK
	data := cc.alloc.AllocWithLen(4, 32)
	data = append(data, mysql.OKHeader)
	data = append(data, 0, 0)
	if cc.capability&mysql.ClientProtocol41 > 0 {
		data = dump.Uint16(data, mysql.ServerStatusAutocommit)
		data = append(data, 0, 0)
	}

	err := cc.writePacket(data)
	cc.pkt.SetSequence(0)
	if err != nil {
		err = errors.SuspendStack(err)
		logutil.Logger(ctx).Debug("write response to client failed", zap.Error(err))
		return err
	}

	err = cc.flush(ctx)
	if err != nil {
		err = errors.SuspendStack(err)
		logutil.Logger(ctx).Debug("flush response to client failed", zap.Error(err))
		return err
	}

    // ...

	return err
}
```

另外上一篇我们看到 `dispatch` 函数中，根据第一个字节来区分不同类型的数据包，然后调用不同的函数来处理，此处就不再展开代码了。

对返回结果写入的代码位于 `pkg/server/conn.go` 中的
`func (cc *clientConn) writeChunksWithFetchSize(ctx context.Context, rs resultset.CursorResultSet, serverStatus uint16, fetchSize int) error`
函数中。

## 总结

这篇文章中，我们概览了一下MySQL协议，并且看了一下TiDB中是如何实现的。同时了解了半双工协议这个概念，也知道了正是因为
半双工通信，我们平时使用数据库的时候，才需要涉及到连接池的管理。
