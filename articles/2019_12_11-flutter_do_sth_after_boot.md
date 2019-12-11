# Flutter应用启动后检查更新

开发一个App时，有这么一种需求，用户打开App时，做一些操作，例如检查更新。Flutter的UI是单线程的，不过好在有 `async` 等，
但是有一个问题，那就是当检查到有更新的安装包时，怎么在UI上显示出来。

我们需要在 `initState()` 时执行一个异步操作：

```dart
  @override
  void initState() {
    super.initState();
    checkRelease();  // 这里是一个异步操作
  }
```

由于我是新手，所以我想使用SnackBar，但是如果直接使用 `Scaffold.of(context).showSnackBar()` 是不行的，会报错：

```
Scaffold.of() called with a context that does not contain a Scaffold. No Scaffold ancestor could be found starting from the context that was passed to Scaffold.of(). This usually happens when the context provided is from the same StatefulWidget as that whose build function actually creates the Scaffold widget being sought. There are several ways to avoid this problem. The simplest is to use a Builder to get a context that is “under” the Scaffold. For an example of this, please see the documentation for Scaffold.of(): https://docs.flutter.io/flutter/material/Scaffold/of.html
```

原因是此处的context还不包含Scaffold，因为Scaffold在 `Widget build(BuildContext context)` 里才生成。解决方案就是使用一个
全局变量来保存 `Scaffold`，当 `Scaffold` 初始化完成之后，我们把这个值保存到全局变量，而异步任务则使用全局变量来做相关
的显示操作。

```dart
class _BlogState extends State<Blog> {
  final Completer<WebViewController> _controller =
      Completer<WebViewController>();
  final GlobalKey<ScaffoldState> _scaffoldKey = new GlobalKey<ScaffoldState>();
  bool isLoading = true;

  @override
  void initState() {
    super.initState();
    Report().reportDeviceInfo();
    Future.delayed(
        Duration(seconds: 5), // 此处使用一个Timer来等待_scaffoldKey完成初始化
        () => checkRelease(
            _scaffoldKey)); // wait 5 seconds to wait it init _scaffoldKey
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      key: _scaffoldKey,
      ... // 其余代码略
  }
}
```

注意这么几点：

- 使用 `final GlobalKey<ScaffoldState> _scaffoldKey = new GlobalKey<ScaffoldState>()` 保存全局Scaffold
- 使用 `Future.delayed` 来确保 `_scaffoldKey` 已经被初始化
- 在需要显示 `SnackBar` 的地方这样使用： `scaffoldKey.currentState.showSnackBar(SnackBar())`
---

参考资料：

- [Medium上一篇介绍这种做法的文章](https://medium.com/@ksheremet/flutter-showing-snackbar-within-the-widget-that-builds-a-scaffold-3a817635aeb2)
