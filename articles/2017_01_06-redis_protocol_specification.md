# Redis通信协议阅读

最近要搞搞redis的通信协议，先阅读一下官方文档，记录一下。

## 应答模型

通常redis收到请求之后会立刻响应，但是有两种情况属于例外：

- 使用pipeling的时候

- 使用pub/sub的时候

## 协议

每次请求的第一个字节代表着本次请求内容类型：

- `+` 代表简单字符串，例如 `OK`, `PING`。

- `-` 代表出错，目前 `-`后字符一直到空格为止代表错误类型，但这并不是此协议的一部分，只是一种偏好。

- `:` 代表Integers，例如 `:1000\r\n`

- `$` 代表Bulk Strings，例如 `foobar` 将会encode成 `$6\r\nfoobar\r\n`，而空字符串将会encode成 `$0\r\n\r\n`

    - NULL将会表示成 `$-1\r\n`

- `*` 代表Array，第一个字符是 `*`，其后紧接array的数据成员个数，然后接一个 `\r\n`，接下来就是各成员具体表示。例如返回一个array，其成员为 `foo`, `bar`，将会被encode成：

`*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n`

    - 空数组将会表示成 `*-1\r\n`

每次请求都需要以 `\r\n` 结尾。
