# Memory leak in net/http

> English is not my mother languange, help me to improve this if you'd like, thanks!

## What happened?

Recently we've get stuck with a problem that we're serving file download service in Go, but the service is killed by server because of
[OOM](https://en.wikipedia.org/wiki/Out_of_memory), so we decide to dig out the reason.

Code in server is like this:

```go
package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"time"

	_ "net/http/pprof"
)

func main() {
	go func() { // here, this goroutine is for debug
		for {
			println("gonna gc")
			runtime.GC()
			time.Sleep(time.Second * 30)
		}
	}()

	http.Handle("/download", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		data, err := ioutil.ReadFile("/Users/jiajun/Images/ubuntu-18.04.1-live-server-amd64.iso")
		if err != nil {
			log.Panicf("failed to read file: %s", err)
		}
		w.Header().Set("Connection", "close")
		w.Write(data)
	}))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

and client code:

```go
package main

import (
	"log"
	"net/http"
)

func main() {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://127.0.0.1:8080/download", nil)
	if err != nil {
		log.Panicf("error: %s", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Connection", "close")
	resp, err := client.Do(req)
	defer resp.Body.Close()

	println("file received")
	select {} // client will block here
}
```

run the server and client, we then use `go tool pprof` to inspect the memory usage:

```bash
$ go run server.go &
$ go run client.go &
$ go tool pprof http://localhost:8080/debug/pprof/heap
```

and we got this:

![memory leak](./img/net_http_mem_leak_1.svg)

although we call `runtime.GC` periodically, the memory is still hold by Golang runtime, until we close the client:

![memory leak2](./img/net_http_mem_leak_2.svg)

## Reason

Although we've set `Connection: close` header, but as [RFC2616](https://tools.ietf.org/html/rfc2616#page-117) says:

> HTTP/1.1 defines the "close" connection option for the sender to signal that the connection will be closed after completion of the response.

it does not speficy who is responsible for close the connection, so Golang wish the client to close the connection.

```go
func (w *response) Write(data []byte) (n int, err error) {
	return w.write(len(data), data, "")
}

func (w *response) write(lenData int, dataB []byte, dataS string) (n int, err error) {
	if w.conn.hijacked() {
		if lenData > 0 {
			caller := relevantCaller()
			w.conn.server.logf("http: response.Write on hijacked connection from %s (%s:%d)", caller.Function, path.Base(caller.File), caller.Line)
		}
		return 0, ErrHijacked
	}
	if !w.wroteHeader {
		w.WriteHeader(StatusOK)
	}
	if lenData == 0 {
		return 0, nil
	}
	if !w.bodyAllowed() {
		return 0, ErrBodyNotAllowed
	}

	w.written += int64(lenData) // ignoring errors, for errorKludge
	if w.contentLength != -1 && w.written > w.contentLength {
		return 0, ErrContentLength
	}
	if dataB != nil {
		return w.w.Write(dataB)
	} else {
		return w.w.WriteString(dataS)
	}
}
```

and the writer use a byte slice to hold the data:

```go
// Write writes the contents of p into the buffer.
// It returns the number of bytes written.
// If nn < len(p), it also returns an error explaining
// why the write is short.
func (b *Writer) Write(p []byte) (nn int, err error) {
	for len(p) > b.Available() && b.err == nil {
		var n int
		if b.Buffered() == 0 {
			// Large write, empty buffer.
			// Write directly from p to avoid copy.
			n, b.err = b.wr.Write(p)
		} else {
			n = copy(b.buf[b.n:], p)
			b.n += n
			b.Flush()
		}
		nn += n
		p = p[n:]
	}
	if b.err != nil {
		return nn, b.err
	}
	n := copy(b.buf[b.n:], p)
	b.n += n
	nn += n
	return nn, nil
}
```

## How to solve it

- Do not use Golang `net/http` to serve files
- I've create a [PR](https://github.com/golang/go/pull/28936) to make sure server will close the connection after it write
all the data, but I'm not sure will it be merged.

-----

- https://en.wikipedia.org/wiki/Out_of_memory
- https://tools.ietf.org/html/rfc2616
- https://github.com/golang/go/pull/28936
