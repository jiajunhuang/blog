# Android自动展示和关闭进度条

客户端总是要有个进度条告诉用户，正在加载内容。可是我很懒，不想每次都自己来控制，那样子的话，得在调用api前设置
进度条显示出来，调用api之后不显示。每次都要这样，太麻烦了。

那么怎么偷懒呢？就是用一个trick，把每个布局文件里的progressbar的id都设置成 `progress_bar`，然后：

```kotlin
private fun showProgressBar() {
    runOnUiThread {
        findViewById<ProgressBar>(R.id.progress_bar)?.let {
            it.isVisible = true
        }
    }
}

private fun hideProgressBar() {
    runOnUiThread {
        findViewById<ProgressBar>(R.id.progress_bar)?.let {
            it.isVisible = false
        }
    }
}

// 调用gRPC函数的外层处理
fun callRPC(func: () -> Any): Any? {
    try {
        showProgressBar()
        return func()
    } catch (e: StatusRuntimeException) {
        Log.e(TAG, "callgRPC: failed with %s", e)
        showSnackBar(e.status.description.toString())
    } finally {
        hideProgressBar()
    }

    return null
}
```

duang，搞定。如果有些地方不想在调用api之前显示进度条，那可以改一改 `callRPC` 这个函数加入一个参数来控制。

不过，这篇就到这里吧。
