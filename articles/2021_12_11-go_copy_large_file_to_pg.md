# 用Go导入大型CSV到PostgreSQL

最近我想试试 `PostgreSQL`，素闻美名，一直没有尝试过。从网上下载了一个超大的CSV，解压后达18G，一般的文件编辑器
直接打不开，简单的方案是直接用 PostgreSQL 提供的 `\copy` 命令，或者 `COPY` 语句，但是这个文件无法使用，因为
其中有几行都是坏数据。

如果是MySQL的话，可以使用 `LOAD FILE IGNORE...` 来忽略错误，但是PostgreSQL没有这个选项，所以我只能选择用Go自己
来导入。

> 吐个槽，MySQL用IGNORE之后，连数据错误也会忽略，导致我导入数据之后，才发现 int 不够表示CSV里的数据字段，导入的
> 很多数据直接变成了 2 ** 31 -1 也就是 2147483647 了，白等了一个小时。

对于大型文件，如果没有足够的内存，也确实是很难处理，我们采取的基本策略就是分块处理，为了提高吞吐量，我们要做
批量提交和并发处理，为了处理异常数据，我们要能dump出有问题的那一块数据，以便处理之后我们再次导入，
由于每一个块是相对较小的，dump出来之后，我们是可以直接用文本编辑器处理问题行的，此外由于涉及到
string和bytes 的转换，我们需要避免频繁申请内存，可以使用上 [黑科技](https://github.com/valyala/fasthttp/blob/master/bytesconv.go)：

```go
// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func b2s(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// s2b converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func s2b(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}
```

以及 `sync.Pool`。

最后代码如下：

```go
package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const size uint64 = 10000

var (
	tokens          = make(chan bool, 50)
	stringSlicePool = sync.Pool{
		New: func() interface{} {
			cache := make([]string, size)
			return cache[:0]
		},
	}
)

// 用wrapper避免参数是 []string 时，是值拷贝的问题
type wrapper struct {
	lines []string
}

func dumpData(w *wrapper) {
	file, err := ioutil.TempFile("./dumps/", "damage")
	if err != nil {
		log.Printf("failed to open temp file")
		return
	}
	defer file.Close()

	for _, line := range w.lines {
		file.WriteString(line)
	}
}

func newWrapper() *wrapper {
	lines := stringSlicePool.Get().([]string)
	return &wrapper{lines: lines}
}

func deleteWrapper(w *wrapper) {
	w.lines = w.lines[:0]
	stringSlicePool.Put(w.lines)
}

func writeData(wg *sync.WaitGroup, db *sqlx.DB, w *wrapper) {
	wg.Add(1)

	token := <-tokens // 并发控制

	tx := db.MustBegin()
	stmt, err := tx.Prepare(pq.CopyIn("表名", "字段1", "字段2" /*字段3...*/))
	if err != nil {
		log.Printf("failed to prepare: %s", err)
		goto done
	}

	if len(w.lines) == 0 {
		goto done
	}

	for _, line := range w.lines {
		data := strings.Split(line, "\t") // 此处是一个频繁内存申请的点
		if len(data) < 2 {
			log.Printf("ignore %s", line)
			continue
		}

		stmt.Exec(data[len(data)-2], data[len(data)-1][:len(data[1])-1])
	}
	stmt.Close()
	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit: %s", err)
		dumpData(w)
		goto done
	}

	log.Printf("saving %d lines", len(w.lines))

done:
	tokens <- token
	deleteWrapper(w)
	wg.Done()
}

func main() {
	var wg sync.WaitGroup

	db, err := sqlx.Connect("postgres", "user=postgres dbname=数据库名 sslmode=disable password=密码")
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%v, %s", db, err)

	file, err := os.Open("./to_import.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	for i := 0; i < cap(tokens); i++ {
		tokens <- true
	}

	reader := bufio.NewReader(file)
	cache := newWrapper()

	// optionally, resize scanner's capacity for lines over 64K, see next example
	var i uint64 = 0
	reader.ReadString('\n')
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		cache.lines = append(cache.lines, line)
		i += 1

		if i%size == 0 {
			oldCache := cache
			go writeData(&wg, db, oldCache)
			cache = newWrapper()
		}
	}

	log.Printf("wait wg...")
	wg.Wait()
	log.Printf("done...")
}
```
