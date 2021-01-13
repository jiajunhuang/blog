# Kotlin/Java 列表Protobuf序列化

本来想保存一些Protobuf生成的类的实例到本地，方法自然就是把一个列表的数据序列化成bytes，然后存起来。不过
搞了半天都没办法，于是就自己整了一个 "poor man's serialization for list of protobuf object"。

方法很简单，首先遍历 `List<Cashapp.Bookkeeping>` ，依次调用 `toByteArray()` 得到 `[]byte`，然后进行
Base64编码，就得到了字符串。然后把多个字符串用某个字符串拼接，例如 `;` 或者 `\n`，最后写入，如果是写文件的话，
也可以直接bytes写入。

那么反序列化，就按着上面步骤，反过来执行即可。

代码：

```kotlin
private fun saveBKList(bkList: List<Cashapp.Bookkeeping>) {
    val encodedBKList = ArrayList<String>()

    for (i in bkList) {
        encodedBKList.add(Base64.encodeToString(i.toByteArray(), Base64.DEFAULT))
    }

    kv.encode("latest_bk_list", encodedBKList.joinToString("\n"))
}
```

不过最后我还是没有执行这一步，因为虽然把列表缓存下来了，打开App时可以直接显示最近一次的列表，但是此时用户还没登录，
感觉怪怪的。
