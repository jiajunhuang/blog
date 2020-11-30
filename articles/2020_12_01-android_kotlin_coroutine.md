# Android上结合kotlin使用coroutine

最近入了Android坑，目前还处于疯狂学习的状态，所以很久都没有写博客了。今天记录一个小代码片段，在Android上使用coroutine
的小例子。

由于我自己是做一个记账软件来学习的，我用了gRPC，最开始我是使用线程来做网络请求的：

```kotlin
thread {
    // 网络请求代码

    runOnUiThread {
        // 更新UI的代码
    }
}
```

今天把这一套全部重写成用coroutine。

首先coroutine得有个调度器，英文叫做 "Dispatchers"，有这么几个：

- `Dispatchers.Main` 这里面的coroutine跑在主线程上，在Android里也就是UI线程，所以如果在这里面的coroutine也执行大量耗时代码的话，也是会卡UI的
- `Dispatchers.IO` 用来跑大IO的
- `Dispatchers.Default` 用来跑高CPU消耗的
- `Dispatchers.Unconfined` 不绑定在任何特定执行线程上

然后，为了多个coroutine之间可以分组啊，就像进程里可以放很多线程那样，又搞了一个概念，叫做 scope，默认有一个全局scope，叫做 `GlobalScope`，全局的，
就和全局变量一样，在Android上，这个里面跑的coroutine，生命周期和app一样久，不推荐在这里起coroutine。

推荐的方式是每个Activity里起一个scope，然后再launch。

所以我就这样写基类：

```kotlin
abstract class BaseActivity : AppCompatActivity(), CoroutineScope {
    /*
    默认的coroutine scope是Main，也就是UI线程(主线程)。如果要做IO，比如网络请求，记得
    包裹在 launch(Dispatchers.IO) {} 里，如果要大量计算，包裹在 launch(Dispatcher.Default) {} 里
    或者直接写 launch。 UI操作则用 withContext(Dispatchers.Main) {} 切回来
     */
    private val job = SupervisorJob()
    override val coroutineContext: CoroutineContext
        get() = Dispatchers.Main + job

    override fun onDestroy() {
        super.onDestroy()
        coroutineContext.cancelChildren()
    }
```

这样子之后，就可以直接launch，起coroutine了：

```kotlin
launch {
    val req = CreateFeedbackReq.newBuilder().build()
    val respAny = callRPC {
        api.createFeedback(req)
    }
    respAny?:return@launch

    val resp = respAny as CreateFeedbackResp
    if (handleRespAction(resp.action)) {
        withContext(Dispatchers.Main) {
            showSnackBar(R.string.thank_you_for_feedback)
            delay(1000)
            finish()
        }
    }
}
```

如上，默认情况下，root coroutine就是当前所在activity，而他们默认会在 `Dispatchers.Main` 上执行，如果想要coroutine在
别的 dispatcher 上执行，就用 `withContext`，然后里面如果又想更新UI的话，就用 `withContext(Dispatchers.Main)`。

那为啥 `launch` 不传参数的话，就是直接用的 `Dispatchers.Main` 呢？因为其实 `CoroutineScope` 是一个接口，而
`coroutineContext` 是里面的一个变量：

```kotlin
public interface CoroutineScope {
    /**
     * The context of this scope.
     * Context is encapsulated by the scope and used for implementation of coroutine builders that are extensions on the scope.
     * Accessing this property in general code is not recommended for any purposes except accessing the [Job] instance for advanced usages.
     *
     * By convention, should contain an instance of a [job][Job] to enforce structured concurrency.
     */
    public val coroutineContext: CoroutineContext
}
```

我们再来看看 `launch` 的实现：

```kotlin
public fun CoroutineScope.launch(
    context: CoroutineContext = EmptyCoroutineContext,
    start: CoroutineStart = CoroutineStart.DEFAULT,
    block: suspend CoroutineScope.() -> Unit
): Job {
    val newContext = newCoroutineContext(context)
    val coroutine = if (start.isLazy)
        LazyStandaloneCoroutine(newContext, block) else
        StandaloneCoroutine(newContext, active = true)
    coroutine.start(start, coroutine, block)
    return coroutine
}

@ExperimentalCoroutinesApi
public actual fun CoroutineScope.newCoroutineContext(context: CoroutineContext): CoroutineContext {
    val combined = coroutineContext + context
    val debug = if (DEBUG) combined + CoroutineId(COROUTINE_ID.incrementAndGet()) else combined
    return if (combined !== Dispatchers.Default && combined[ContinuationInterceptor] == null)
        debug + Dispatchers.Default else debug
}
```

可以看到，默认情况下，会把当前的 `coroutineContext` 放在前面。

Kotlin的coroutine很好用，不过我感觉还是有点复杂，我也还在学习。

---

ref:

- https://kotlin.github.io/kotlinx.coroutines/kotlinx-coroutines-core/kotlinx.coroutines/-dispatchers/index.html
