# 自己写个搜索引擎

搜索引擎最核心的地方在于倒排索引，而倒排索引其实并不是一种具体的数据结构，确切的来说是一类。
这个实现中使用Golang中的 `map` 来做倒排索引，全部代码如下：

```go
package main

import (
	"io/ioutil"
	"log"
    "os"
)

// Index 倒排索引索引项
type Index struct {
	DocumentName string
	Offset       int
}

// ToParsingFile 待索引的文件
type ToParsingFile struct {
    FileName string
    Content string
}

// ReverseIndex 单个倒排索引具体
type ReverseIndex map[string][]Index
// Blacklist 倒排索引key的黑名单
type Blacklist map[string]bool
// Ignorelist 忽略字符列表
type Ignorelist map[rune]bool

func readFile(fileName string) (string, error) {
	contentBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Printf("open file failed with error: %v\n", err)
		return "", err
	}
	return string(contentBytes), nil
}

func parse(ignorelist Ignorelist, fileName, content string) ReverseIndex {
	reverseFile := make(ReverseIndex)

	lastIndex := 0

	for i, c := range(content) {
        if _, ok := ignorelist[c]; ok {
            newWord := string(content[lastIndex:i])
            reverseFile[newWord] = append(reverseFile[newWord], Index{fileName, lastIndex})
            lastIndex = i + 1
		}
	}

	return reverseFile
}

// 合并多个倒排索引并且剔除blacklist中的key
func mergeReverseIndex(blacklist Blacklist, args ...ReverseIndex) ReverseIndex {
    result := make(ReverseIndex)

    for _, arg := range(args) {
        for k, vs := range(arg) {
            // 忽略黑名单中的key
            if _, ok := blacklist[k]; ok {
                continue
            }

            // 忽略连续断字符组成的
            if len(k) == 0 {
                continue
            }

            // TODO: write myself a `extend` function for Golang slice
            for _, v := range(vs) {
                result[k] = append(result[k], v)
            }
        }
    }

    return result
}

func getDocumentNames(reverseFile ReverseIndex, keyWord string) map[string]bool {
    resultMap := make(map[string]bool)

    indexList, ok := reverseFile[keyWord]

    if !ok {
        return resultMap
    }

    for _, index := range(indexList) {
        resultMap[index.DocumentName] = true
    }

    return resultMap
}

func makeSubSetOf(map1, map2 map[string]bool) map[string]bool {
    var longer, shorter map[string]bool
    resultMap := make(map[string]bool)

    if len(map1) > len(map2) {
        longer = map1
        shorter = map2
    } else {
        longer = map2
        shorter = map1
    }

    for k := range(shorter) {
        if _, ok := longer[k]; ok {
            resultMap[k] = true
        }
    }

    return resultMap
}

// 搜索，目前的检索方案为：
// 检索出每个关键字所在的文章
// 对关键字做交集处理
// 取出所有关键字都存在的文章然后返回
func search(reverseFile ReverseIndex, keyWords ...string) []string {
    result := make([]string, 0)
    var finalDocumentNames map[string]bool

    for _, keyWord := range(keyWords) {
        documentNames := getDocumentNames(reverseFile, keyWord)

        if finalDocumentNames == nil {
            finalDocumentNames = documentNames
        }

        finalDocumentNames = makeSubSetOf(finalDocumentNames, documentNames)
    }

    for doc := range(finalDocumentNames) {
        result = append(result, doc)
    }

    return result
}

func main() {
    contents := make([]ToParsingFile, 0)

    files, err := ioutil.ReadDir("./docs")
    if err != nil {
        log.Panicf("read dir(./docs) failed, aborting...")
    }

    for _, file := range(files) {
        fileName := file.Name()
        content, err := readFile("./docs/" + fileName)

        if err != nil {
            log.Panicf("read file(%s) failed!", fileName)
        }

        contents = append(contents, ToParsingFile{fileName, content})
    }

    blacklist := Blacklist{"a": false, "is": false}
    ignorelist := Ignorelist{
        ' ': false, '\t': false, '\n': false, '\r': false, ',': false, '.': false,
        '"': false, ':': false, '`': false, '(': false, ')': false, '[': false, ']': false,
        '{': false, '}': false, '-': false, '_': false, '+': false, '=': false,
        '\\': false, '|': false, '\'': false, '<': false, '>': false, '$': false,
    }

    var reverseFiles []ReverseIndex
    for _, toParsingFile := range(contents) {
        reverseFiles = append(reverseFiles, parse(ignorelist, toParsingFile.FileName, toParsingFile.Content))
    }

    reverseFile := mergeReverseIndex(blacklist, reverseFiles...)

    log.Printf("we're going to create reverse index")
	for k, vs := range reverseFile {
		log.Printf("Key: %s -> ", k)
		for _, v := range vs {
			log.Printf("in doc %s with offset %d, ", v.DocumentName, v.Offset)
		}
	}
	log.Printf("looping done")

    keyWords := os.Args[1:]
    searchResults := search(reverseFile, keyWords...)

    log.Printf("========================= Search Engine =================\n")
    log.Printf("searing %v...\n", keyWords)
    for _, searchResult := range(searchResults) {
        log.Printf("in file: %s\n", searchResult)
    }
    log.Printf("done.\n")
}
```

执行日志：

```bash
$ go run reverse_file.go bigbang
2017/06/02 23:12:19 we're going to create reverse index
2017/06/02 23:12:19 Key: inherited ->
2017/06/02 23:12:19 in doc dev.rst with offset 297,
2017/06/02 23:12:19 Key: are ->
2017/06/02 23:12:19 in doc dev.rst with offset 16,
2017/06/02 23:12:19 Key: Installation ->
2017/06/02 23:12:19 in doc user.rst with offset 105,
2017/06/02 23:12:19 Key: about ->
2017/06/02 23:12:19 in doc user.rst with offset 78,
2017/06/02 23:12:19 Key: Using ->
2017/06/02 23:12:19 in doc user.rst with offset 266,
2017/06/02 23:12:19 Key: User ->
2017/06/02 23:12:19 in doc user.rst with offset 0,
2017/06/02 23:12:19 Key: of ->
2017/06/02 23:12:19 in doc dev.rst with offset 95,
2017/06/02 23:12:19 in doc user.rst with offset 36,
2017/06/02 23:12:19 in doc user.rst with offset 560,
2017/06/02 23:12:19 Key: method ->
2017/06/02 23:12:19 in doc dev.rst with offset 77,
2017/06/02 23:12:19 Key: by ->
2017/06/02 23:12:19 in doc user.rst with offset 337,
2017/06/02 23:12:19 Key: use ->
2017/06/02 23:12:19 in doc user.rst with offset 91,
2017/06/02 23:12:19 in doc user.rst with offset 309,
2017/06/02 23:12:19 Key: using ->
2017/06/02 23:12:19 in doc user.rst with offset 582,
2017/06/02 23:12:19 Key: set ->
2017/06/02 23:12:19 in doc user.rst with offset 468,
2017/06/02 23:12:19 in doc user.rst with offset 698,
2017/06/02 23:12:19 Key: for ->
2017/06/02 23:12:19 in doc dev.rst with offset 28,
2017/06/02 23:12:19 in doc dev.rst with offset 119,
2017/06/02 23:12:19 Key: loop ->
2017/06/02 23:12:19 in doc dev.rst with offset 231,
2017/06/02 23:12:19 in doc user.rst with offset 323,
2017/06/02 23:12:19 in doc user.rst with offset 381,
2017/06/02 23:12:19 in doc user.rst with offset 478,
2017/06/02 23:12:19 in doc user.rst with offset 567,
2017/06/02 23:12:19 in doc user.rst with offset 655,
2017/06/02 23:12:19 in doc user.rst with offset 679,
2017/06/02 23:12:19 in doc user.rst with offset 708,
2017/06/02 23:12:19 in doc user.rst with offset 713,
2017/06/02 23:12:19 Key: undoc ->
2017/06/02 23:12:19 in doc dev.rst with offset 279,
2017/06/02 23:12:19 Key: bigbang ->
2017/06/02 23:12:19 in doc dev.rst with offset 318,
2017/06/02 23:12:19 in doc user.rst with offset 720,
2017/06/02 23:12:19 Key: requires ->
2017/06/02 23:12:19 in doc user.rst with offset 168,
2017/06/02 23:12:19 Key: section ->
2017/06/02 23:12:19 in doc user.rst with offset 28,
2017/06/02 23:12:19 Key: templates ->
2017/06/02 23:12:19 in doc .gitignore with offset 16,
2017/06/02 23:12:19 Key: documentation ->
2017/06/02 23:12:19 in doc dev.rst with offset 102,
2017/06/02 23:12:19 in doc user.rst with offset 43,
2017/06/02 23:12:19 Key: or ->
2017/06/02 23:12:19 in doc dev.rst with offset 74,
2017/06/02 23:12:19 Key: event ->
2017/06/02 23:12:19 in doc dev.rst with offset 225,
2017/06/02 23:12:19 in doc user.rst with offset 317,
2017/06/02 23:12:19 in doc user.rst with offset 375,
2017/06/02 23:12:19 in doc user.rst with offset 472,
2017/06/02 23:12:19 in doc user.rst with offset 673,
2017/06/02 23:12:19 in doc user.rst with offset 702,
2017/06/02 23:12:19 Key: build ->
2017/06/02 23:12:19 in doc .gitignore with offset 1,
2017/06/02 23:12:19 Key: this ->
2017/06/02 23:12:19 in doc dev.rst with offset 85,
2017/06/02 23:12:19 Key: members ->
2017/06/02 23:12:19 in doc dev.rst with offset 186,
2017/06/02 23:12:19 in doc dev.rst with offset 267,
2017/06/02 23:12:19 in doc dev.rst with offset 285,
2017/06/02 23:12:19 in doc dev.rst with offset 307,
2017/06/02 23:12:19 Key: autofunction ->
2017/06/02 23:12:19 in doc dev.rst with offset 199,
2017/06/02 23:12:19 Key: Alternatively ->
2017/06/02 23:12:19 in doc user.rst with offset 518,
2017/06/02 23:12:19 Key: Use ->
2017/06/02 23:12:19 in doc user.rst with offset 190,
2017/06/02 23:12:19 Key: instance ->
2017/06/02 23:12:19 in doc user.rst with offset 551,
2017/06/02 23:12:19 Key: API ->
2017/06/02 23:12:19 in doc dev.rst with offset 0,
2017/06/02 23:12:19 Key: the ->
2017/06/02 23:12:19 in doc dev.rst with offset 98,
2017/06/02 23:12:19 in doc user.rst with offset 39,
2017/06/02 23:12:19 in doc user.rst with offset 313,
2017/06/02 23:12:19 in doc user.rst with offset 362,
2017/06/02 23:12:19 in doc user.rst with offset 563,
2017/06/02 23:12:19 Key: console ->
2017/06/02 23:12:19 in doc user.rst with offset 230,
2017/06/02 23:12:19 Key: autoclass ->
2017/06/02 23:12:19 in doc dev.rst with offset 148,
2017/06/02 23:12:19 in doc dev.rst with offset 240,
2017/06/02 23:12:19 Key: looking ->
2017/06/02 23:12:19 in doc dev.rst with offset 20,
2017/06/02 23:12:19 Key: provided ->
2017/06/02 23:12:19 in doc user.rst with offset 328,
2017/06/02 23:12:19 Key: If ->
2017/06/02 23:12:19 in doc dev.rst with offset 9,
2017/06/02 23:12:19 Key: 3 ->
2017/06/02 23:12:19 in doc user.rst with offset 184,
2017/06/02 23:12:19 Key: create ->
2017/06/02 23:12:19 in doc user.rst with offset 541,
2017/06/02 23:12:19 Key: manually ->
2017/06/02 23:12:19 in doc user.rst with offset 572,
2017/06/02 23:12:19 Key: an ->
2017/06/02 23:12:19 in doc user.rst with offset 548,
2017/06/02 23:12:19 Key: can ->
2017/06/02 23:12:19 in doc user.rst with offset 537,
2017/06/02 23:12:19 Key: function ->
2017/06/02 23:12:19 in doc dev.rst with offset 58,
2017/06/02 23:12:19 Key: new ->
2017/06/02 23:12:19 in doc dev.rst with offset 221,
2017/06/02 23:12:19 in doc user.rst with offset 669,
2017/06/02 23:12:19 Key: policy ->
2017/06/02 23:12:19 in doc user.rst with offset 386,
2017/06/02 23:12:19 in doc user.rst with offset 483,
2017/06/02 23:12:19 Key: It ->
2017/06/02 23:12:19 in doc user.rst with offset 165,
2017/06/02 23:12:19 Key: from ->
2017/06/02 23:12:19 in doc user.rst with offset 154,
2017/06/02 23:12:19 Key: available ->
2017/06/02 23:12:19 in doc user.rst with offset 144,
2017/06/02 23:12:19 Key: Guide ->
2017/06/02 23:12:19 in doc user.rst with offset 5,
2017/06/02 23:12:19 Key: uvloop ->
2017/06/02 23:12:19 in doc dev.rst with offset 130,
2017/06/02 23:12:19 in doc dev.rst with offset 160,
2017/06/02 23:12:19 in doc dev.rst with offset 214,
2017/06/02 23:12:19 in doc dev.rst with offset 252,
2017/06/02 23:12:19 in doc user.rst with offset 95,
2017/06/02 23:12:19 in doc user.rst with offset 133,
2017/06/02 23:12:19 in doc user.rst with offset 257,
2017/06/02 23:12:19 in doc user.rst with offset 272,
2017/06/02 23:12:19 in doc user.rst with offset 341,
2017/06/02 23:12:19 in doc user.rst with offset 367,
2017/06/02 23:12:19 in doc user.rst with offset 449,
2017/06/02 23:12:19 in doc user.rst with offset 490,
2017/06/02 23:12:19 in doc user.rst with offset 644,
2017/06/02 23:12:19 in doc user.rst with offset 662,
2017/06/02 23:12:19 Key: python ->
2017/06/02 23:12:19 in doc user.rst with offset 411,
2017/06/02 23:12:19 in doc user.rst with offset 606,
2017/06/02 23:12:19 Key: class ->
2017/06/02 23:12:19 in doc dev.rst with offset 68,
2017/06/02 23:12:19 Key: how ->
2017/06/02 23:12:19 in doc user.rst with offset 84,
2017/06/02 23:12:19 Key: it ->
2017/06/02 23:12:19 in doc user.rst with offset 209,
2017/06/02 23:12:19 Key: provides ->
2017/06/02 23:12:19 in doc user.rst with offset 57,
2017/06/02 23:12:19 Key: To ->
2017/06/02 23:12:19 in doc user.rst with offset 293,
2017/06/02 23:12:19 Key: 5 ->
2017/06/02 23:12:19 in doc user.rst with offset 186,
2017/06/02 23:12:19 Key: part ->
2017/06/02 23:12:19 in doc dev.rst with offset 90,
2017/06/02 23:12:19 Key: pip ->
2017/06/02 23:12:19 in doc user.rst with offset 194,
2017/06/02 23:12:19 in doc user.rst with offset 245,
2017/06/02 23:12:19 Key: asyncio ->
2017/06/02 23:12:19 in doc user.rst with offset 301,
2017/06/02 23:12:19 in doc user.rst with offset 430,
2017/06/02 23:12:19 in doc user.rst with offset 460,
2017/06/02 23:12:19 in doc user.rst with offset 625,
2017/06/02 23:12:19 in doc user.rst with offset 690,
2017/06/02 23:12:19 Key: PyPI ->
2017/06/02 23:12:19 in doc user.rst with offset 159,
2017/06/02 23:12:19 Key: static ->
2017/06/02 23:12:19 in doc .gitignore with offset 8,
2017/06/02 23:12:19 Key: specific ->
2017/06/02 23:12:19 in doc dev.rst with offset 49,
2017/06/02 23:12:19 Key: make ->
2017/06/02 23:12:19 in doc user.rst with offset 296,
2017/06/02 23:12:19 Key: Python ->
2017/06/02 23:12:19 in doc user.rst with offset 177,
2017/06/02 23:12:19 Key: code ->
2017/06/02 23:12:19 in doc user.rst with offset 217,
2017/06/02 23:12:19 in doc user.rst with offset 398,
2017/06/02 23:12:19 in doc user.rst with offset 593,
2017/06/02 23:12:19 Key: block ->
2017/06/02 23:12:19 in doc user.rst with offset 222,
2017/06/02 23:12:19 in doc user.rst with offset 403,
2017/06/02 23:12:19 in doc user.rst with offset 598,
2017/06/02 23:12:19 Key: you ->
2017/06/02 23:12:19 in doc dev.rst with offset 12,
2017/06/02 23:12:19 in doc dev.rst with offset 123,
2017/06/02 23:12:19 in doc user.rst with offset 350,
2017/06/02 23:12:19 in doc user.rst with offset 533,
2017/06/02 23:12:19 Key: EventLoopPolicy ->
2017/06/02 23:12:19 in doc dev.rst with offset 167,
2017/06/02 23:12:19 in doc user.rst with offset 497,
2017/06/02 23:12:19 Key: to ->
2017/06/02 23:12:19 in doc user.rst with offset 88,
2017/06/02 23:12:19 in doc user.rst with offset 198,
2017/06/02 23:12:19 Key: This ->
2017/06/02 23:12:19 in doc user.rst with offset 23,
2017/06/02 23:12:19 Key: on ->
2017/06/02 23:12:19 in doc dev.rst with offset 44,
2017/06/02 23:12:19 Key: import ->
2017/06/02 23:12:19 in doc user.rst with offset 423,
2017/06/02 23:12:19 in doc user.rst with offset 442,
2017/06/02 23:12:19 in doc user.rst with offset 618,
2017/06/02 23:12:19 in doc user.rst with offset 637,
2017/06/02 23:12:19 Key: install ->
2017/06/02 23:12:19 in doc user.rst with offset 201,
2017/06/02 23:12:19 in doc user.rst with offset 249,
2017/06/02 23:12:19 in doc user.rst with offset 354,
2017/06/02 23:12:19 Key: Loop ->
2017/06/02 23:12:19 in doc dev.rst with offset 259,
2017/06/02 23:12:19 Key: information ->
2017/06/02 23:12:19 in doc dev.rst with offset 32,
2017/06/02 23:12:19 in doc user.rst with offset 66,
2017/06/02 23:12:19 looping done
2017/06/02 23:12:19 ========================= Search Engine =================
2017/06/02 23:12:19 searing [bigbang]...
2017/06/02 23:12:19 in file: dev.rst
2017/06/02 23:12:19 in file: user.rst
2017/06/02 23:12:19 done.
```

TODO:

- 使用B树替代哈希表做倒排索引
- 使用文件持久化倒排索引
