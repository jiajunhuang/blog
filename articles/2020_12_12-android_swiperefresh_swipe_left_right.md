# Android SwipeRefreshLayout左右滑动冲突的解决

我有一个页面是需要下拉刷新的，所以在 `<RecyclerView>` 的外层包了 `<SwipeRefreshLayout>`，但是同时我希望 `<RecyclerView>`
里的组件，可以支持左右滑动，比如，右滑删除，左滑编辑。

答案是通过 `ItemTouchHelper` 来实现，诸如：

```kotlin
ItemTouchHelper(createItemTouchCallback(adapter)).attachToRecyclerView(main_recycler_view)

private fun createItemTouchCallback(adapter: BKRecyclerViewAdapter): ItemTouchHelper.SimpleCallback {
    // ...
}
```

实现完成之后，的确可以左右滑动，但是会带来一个小问题，那就是 `SwipeRefreshLayout` 特别灵敏，左右滑动的时候，稍微带有
一点向下的方向，就会触发下拉刷新，导致左右滑动被取消，那么解决方案是什么呢？方案就是，检查左右滑动的量，然后当水平
偏移达到一定程度的时候，禁用下拉刷新，当复原，或者滑动完成之后，重新启用下拉刷新：

```kotlin
override fun onChildDraw(
    c: Canvas,
    recyclerView: RecyclerView,
    viewHolder: RecyclerView.ViewHolder,
    dX: Float,
    dY: Float,
    actionState: Int,
    isCurrentlyActive: Boolean
) {
    val textMargin = resources.getDimension(R.dimen.mdtp_ampm_label_size).roundToInt()
    val width = viewHolder.itemView.width
    val absDX = abs(dX)
    main_swipe_refresh.isEnabled = !(5 <= absDX && absDX <= viewHolder.itemView.width - 5)  // 重点在这里
    // ...
}

override fun onSwiped(viewHolder: RecyclerView.ViewHolder, direction: Int) {
    main_swipe_refresh.isEnabled = true
    // ...
}
```

上面的，超过5像素就禁用下拉刷新，是我自己调出来的参数，大家可以自己在设备上调一调。
