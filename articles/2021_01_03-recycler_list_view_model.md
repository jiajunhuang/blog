# 怎么使用ViewModel 和 RecyclerView

当他们分开使用的时候，很简单，但是怎么把ViewModel和RecyclerView结合在一起呢？

代码如下：

```kotlin
model.assetItemList.observe(this, {
    binding.nothingHint.nothingHint.isVisible = it.isEmpty()

    val adapter = AssetDebtItemAdapter(it, currency)
    binding.recyclerView.adapter = adapter
    adapter.notifyDataSetChanged()
})
```
