# Android滑动时隐藏FAB

我在BaseClass里加入了如下代码，实现了在滑动 `RecyclerView` 的时候，隐藏浮动按钮(FAB)：

```kotlin
    fun setUpFABHide(recyclerView: RecyclerView, fab: FloatingActionButton) {
        recyclerView.addOnScrollListener(object: RecyclerView.OnScrollListener() {
            override fun onScrollStateChanged(recyclerView: RecyclerView, newState: Int) {
                if (newState == RecyclerView.SCROLL_STATE_IDLE) {
                    fab.show()
                }

                super.onScrollStateChanged(recyclerView, newState)
            }

            override fun onScrolled(recyclerView: RecyclerView, dx: Int, dy: Int) {
                if (dy < 0 || dy > 0 && fab.isShown) {
                    fab.hide()
                }
            }
        })
    }

```

这样子在滑动的时候就会隐藏FAB，然后滑动停止时，FAB又会显示出来。

---

参考资料：

- https://stackoverflow.com/questions/31617398/floatingactionbutton-hide-on-list-scroll
