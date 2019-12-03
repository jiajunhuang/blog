# flutter webview加载时显示进度

最近在学习flutter，在webview加载页面时，如果网速不好的话，就会一直白屏，用户看到这个场景可能会比较迷惑，因此我们得加个进度条：

```dart
class _BlogState extends State<Blog> {
  final Completer<WebViewController> _controller =
      Completer<WebViewController>();
  bool isLoading = true;  // 设置状态

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text(homeTitle),
        actions: <Widget>[
          BlogMenu(_controller.future),
        ],
      ),
      body: Stack(
        children: [
          WebView(
            initialUrl: homePage,
            javascriptMode: JavascriptMode.unrestricted,
            onWebViewCreated: (WebViewController webViewController) {
              _controller.complete(webViewController);
            },
            navigationDelegate: (NavigationRequest request) {
              var url = request.url;
              print("visit $url");
              setState(() {
                isLoading = true; // 开始访问页面，更新状态
              });

              return NavigationDecision.navigate;
            },
            onPageFinished: (String url) {
              setState(() {
                isLoading = false; // 页面加载完成，更新状态
              });
            },
          ),
          isLoading
              ? Container(
                  child: Center(
                    child: CircularProgressIndicator(),
                  ),
                )
              : Container(),
        ],
      ),
      floatingActionButton: getFloatingButton(),
    );
  }
```

我把 `WebView` 放在一个 `StatefulWidget` 里，通过一个 `bool isLoading` 来指示当前是否正在加载页面，当开始访问链接时，
将 `isLoading` 设置为 `true`，访问完成时，设置为 `false`。使用一个 `Stack` 将 `WebView` 和 `CircularProgressIndicator`
放在一起，当 `isLoading` 为 `true` 时，显示 `CircularProgressIndicator`，否则显示一个空的 `Container`，这样就可以
实现加载时，有一个圈圈在转，而加载完成时，圈圈消失的效果了。
