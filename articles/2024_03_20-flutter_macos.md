# flutter macOS 构建出错

今天尝试用flutter构建一个 macOS 应用，结果一直报错：

```bash
/macos/Pods/Pods.xcodeproj: warning: The macOS deployment target 'MACOSX_DEPLOYMENT_TARGET' is set to 10.12, but the range of supported deployment target versions is 10.13 to 14.4.99. (in target 'AppAuth-AppAuthCore_Privacy' from project 'Pods')
```

找了很久都没有找到原因，直到我搜索 `MACOSX_DEPLOYMENT_TARGET`，发现 `macos/Pods/Pods.xcodeproj/project.pbxproj` 里有许多
这个变量，但是版本确实是很低，我尝试手动改写一下，把版本统一到 10.15，结果发现还真的编译成功了，只是 pub get 以后再次编译的时候，
又回报错，原因是这个变量又被覆盖了。

最后找到的解决办法，编辑 `macos/Podfile` 的 `post_install` 部分，从

```ruby
post_install do |installer|
  installer.pods_project.targets.each do |target|
    flutter_additional_macos_build_settings(target)
  end
end
```

改成

```ruby
post_install do |installer|
  installer.pods_project.targets.each do |target|
    flutter_additional_macos_build_settings(target)
    target.build_configurations.each do |config|
      config.build_settings['MACOSX_DEPLOYMENT_TARGET'] = '10.15'
    end
  end
end
```

再次生成，就不会报错了，这里的原理，其实就是强制改写每一个依赖的最低构建版本。

坑呀！又踩了一个。

---

ref: https://stackoverflow.com/questions/72891223/what-should-i-do-to-change-the-macosx-deployment-target-when-build-flutter-macos
