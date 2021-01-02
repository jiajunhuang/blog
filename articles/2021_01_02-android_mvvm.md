# Android手动挡MVVM

Android官方的自动档MVVM方案 Jetpack Compose 还没正式发布，目前只有手动挡的MVVM方案，那就是下面几者的组合：

- view binding
- view model
- livedata
- (可选)data binding 

我一般就用前三者，因为不喜欢在XML里写一堆的代码。

来对比一下，MVC和手动挡MVVM。首先我们来看看Android开发经典的使用方式，也就是MVC的模式来开发，首先我们要在
XML里画好界面，然后我们需要在各个业务操作逻辑里，写上一大堆的 `findViewById`，然后转型成对应的类型。

kotlin-android-extension 解决了这个痛点，但是官方已经放弃维护了，推荐使用view binding。使用这套MVVM方案之后，
代码就会变成这样子：

```java
// ...
super.onCreate(savedInstanceState)
binding = ActivityInvestmentListBinding.inflate(layoutInflater)

setContentView(binding.root)

setUpTopBar(binding.topBar.topBar, getString(R.string.title_investment_list))

// view model
model.isRefreshing.observe(this, {
    binding.progressBar.isVisible = it
    binding.swipeRefresh.isRefreshing = it
})
// ...

// 初始化数据
model.isRefreshing.value = true
```

很明显的变化是，设置好了observer以及数据变更之后需要进行的操作之后，逻辑代码就只需要操作model里数据，
而不需要去更新UI。这样子，业务代码只和数据打交道，数据变更之后，会统一去变更UI，极大的降低了复杂页面的
开发难度和维护难度，降低了BUG出现率。

此时，不得不感叹一句，flutter真香！
