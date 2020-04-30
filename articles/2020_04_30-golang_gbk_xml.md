# Go语言解析GBK编码的xml

最近接短信提供商，因为要做审计功能，所以就要把短信的trace id等信息存储下来，但是捏，提供商返回的是GBK格式的XML，而Go
xml库默认只支持UTF-8。那咋办呢？下面是两个方案，地一个比较trick，但是还挺好玩的，第二个比较正式：

## 把xml从GBK转换成UTF-8

```go
package main

import (
	"bytes"
	"io/ioutil"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}

	str := string(d)
	str = strings.ReplaceAll(str, `<?xml version="1.0" encoding="gbk" ?>`, `<?xml version="1.0" encoding="UTF-8" ?>`)
	str = strings.ReplaceAll(str, `<?xml version="1.0" encoding="GBK" ?>`, `<?xml version="1.0" encoding="UTF-8" ?>`)

	return []byte(str), nil
}
```

请注意后面的那两行 `strings.ReplaceAll`，如果不加上这两个，Go就会报错：`xml: encoding "gbk" declared but Decoder.CharsetReader is nil`。
这上面的原理是啥呢？就是把xml从GBK编码转换为UTF-8,然后把XML里的编码声明也一起替换掉，所以说比较trick，但是还挺好玩的 doge。

好，接下来我们来看正经一点的。

## 让xml支持解码GBK格式

同样首先我们要转换编码，但是这次我们传入一个Reader：

```go
xmlBytes, err := GbkToUtf8(resp.Bytes())
if err != nil {
    log.Printf("failed to transform gbk to utf8 but I don't care: %s", err)
}
decoder := xml.NewDecoder(bytes.NewReader(xmlBytes))
decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
    return transform.NewReader(input, simplifiedchinese.GBK.NewEncoder()), nil
}
err = decoder.Decode(&guoduResp)
```

这是因为xml提供的 `Decoder` 支持自定义一个处理编码声明的函数，也就是我们上面覆盖的 `decoder.CharsetReader`：

```
// A Decoder represents an XML parser reading a particular input stream.
// The parser assumes that its input is encoded in UTF-8.
type Decoder struct {
    ...
	// CharsetReader, if non-nil, defines a function to generate
	// charset-conversion readers, converting from the provided
	// non-UTF-8 charset into UTF-8. If CharsetReader is nil or
	// returns an error, parsing stops with an error. One of the
	// CharsetReader's result values must be non-nil.
	CharsetReader func(charset string, input io.Reader) (io.Reader, error)
    ...
}
```

这样子我们也可以解析GBK格式的xml。

---

参考资料：

- https://mengqi.info/html/2015/201507071345-using-golang-to-convert-text-between-gbk-and-utf-8.html
