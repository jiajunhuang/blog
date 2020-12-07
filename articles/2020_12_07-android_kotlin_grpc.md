# Android调用gRPC的两个小工具函数

Android上调用gRPC，由于gRPC中的status错误，在gRPC Java/Kotlin生成的代码中，是以异常的形式展现出来的，所以，如果不想在
业务代码里到处都充斥着 `try {} catch(e: StatusRuntimeException) {}` 这样的代码，就得抽出一个工具函数来同一处理错误，
如下：

```kotlin
// 调用gRPC函数的外层处理
fun callRPC(func: () -> Any): Any? {
    try {
        return func()
    } catch (e: StatusRuntimeException) {
        Log.e(TAG, "callgRPC: failed with %s", e)
        showSnackBar(e.status.description.toString())
    }

    return null
}

```

这样，调用方的代码就可以写成：

```kotlin
val req = Blabla.build()
val respAny = callRPC {
    api.someAPI(req)
}
respAny ?: return@launch

val resp = respAny as BlablaResp
```

这是第一。

第二，在gRPC的响应里，可以约定，所有的响应第一个字段，都是 `string action = 1`，这样子，我们就可以结合上面的代码，再
来对具体的响应结果来进行处理，例如，因此业务代码会变成：

```kotlin
if (handleRespAction(resp.action)) {
    // 如果成功了，就执行这里的
}
```

而 `handleRespAction` 的，我们就可以对action进行约定。例如我的app里，我约定，action一定是一个URL，或者为空。为空时
表示成功，不为空时，表明需要做一定的操作，例如 `cashapp://jiajunhuang.com/alert` 就会弹出一个警告窗。
通过这样，我们可以对所有的请求，进行一个统一的处理，例如要显示一些消息，或者是弹出更新提醒等等。

这就是最近写Android，结合gRPC的两个小工具函数。
