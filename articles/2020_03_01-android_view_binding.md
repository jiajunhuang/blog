# Android 使用view binding

在学习和使用flutter一段时间之后，发现flutter的生态和原生相比还是不在一个数量级。因此进行了原生安卓的学习。

Android开发中比较烦的一个事情是，当你要操作Activity中的控件时，你必须要写类似的代码：

```kotlin
val button1 = findViewById<Button>(R.id.button1)
val button2 = findViewById<Button>(R.id.button2)
val textview = findViewById<TextView>(R.id.textview)
...
```

可以说的上是要操作几个控件，就要写多少这样的代码。为了解决这个问题，Android jetpack中有一个组件叫做 `ViewBinding`。
它会根据布局描述文件自动生成一个类，比如 `activity_main.xml` 会生成一个 `ActivityMainBinding` 类，通过引用binding，
就可以直接使用Activity中的控件，比如：

```kotlin
package com.jiajunhuang.testapp

import androidx.appcompat.app.AppCompatActivity
import android.os.Bundle
import android.view.View
import com.jiajunhuang.blogapp.databinding.ActivityMainBinding

class MainActivity : AppCompatActivity() {
    private lateinit var binding: ActivityMainBinding

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityMainBinding.inflate(layoutInflater)
        setContentView(binding.root)

        binding.button.setOnClickListener(View.OnClickListener {

        })
    }
}
```

注意其中，原本 `val button1 = findViewById<Button>(R.id.button1)` 就可以直接变成 `binding.button`。

---

那么该如何做到呢？首先得在 `app/build.gradle` 的 `android` 块下增加：

```gradle
    viewBinding {
        enabled = true
    }
```

然后在 `dependencies` 块下增加：

```gradle
    // lifecycle
    def lifecycle_version = "2.2.0"
    def arch_version = "2.1.0"

    // ViewModel
    implementation "androidx.lifecycle:lifecycle-viewmodel-ktx:$lifecycle_version"
    // LiveData
    implementation "androidx.lifecycle:lifecycle-livedata-ktx:$lifecycle_version"
    // Lifecycles only (without ViewModel or LiveData)
    implementation "androidx.lifecycle:lifecycle-runtime-ktx:$lifecycle_version"

    // Saved state module for ViewModel
    implementation "androidx.lifecycle:lifecycle-viewmodel-savedstate:$lifecycle_version"
```

这个时候Android Studio会提醒要不要Sync，当然要Sync。完成之后，就可以如上使用了。

另外我尝试使用了一下 `DataBinding`，其实个人觉得还不是很好使用，虽然MVVM是一个很爽的体验，但是目前的实际操作体验不是
很好，比如Android Studio对xml的补全还不是很好，另外要在xml里声明变量以及所需要触发的回调。当布局发生大的改动之后，
其带来的优点是否真的能有大幅度的发挥呢？这一点值得探讨。
