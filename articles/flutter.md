# Flutter组件(widget)库手册

Flutter的组建分为三种：

- `StatefulWidget`: 这种类型的组件会保存状态，比如，一个按钮按了几次，这个几次就是一个状态
- `StatelessWidget`: 这种类型的组件不会保存状态
- `InheritedWidget`: 这种类型的组件可以用来进行一些子代和祖先的操作，比如 `Theme.of()` 就是这样的操作

Flutter自带两套主题，一套是Google的Material，一套是Apple的Cupertino。我们以Material为主。

一个简单的Material的例子是这样的：

```dart
import 'package:flutter/material.dart';

void main() => runApp(MyApp());

class MyApp extends StatelessWidget {
  Widget build(BuildContext context) {
    return MaterialApp(
      title: "Flutter demo",
      home: Scaffold(
        appBar: AppBar(title: Text("AppBar")),
        body: Center(child: Text("Hello world")),
      ),
    );
  }
}
```

效果如下：

![Flutter Hello World](/articles/img/flutter_hello_world.png)

## 组件库

- `MaterialApp` 使用Material Design的App就用这个，我们来看看它的参数：

```dart
const MaterialApp({
    Key key,
    this.navigatorKey,
    this.home, // 默认情况下，home就是 '/' 路由
    this.routes = const <String, WidgetBuilder>{}, // 路由
    this.initialRoute,
    this.onGenerateRoute,
    this.onUnknownRoute,
    this.navigatorObservers = const <NavigatorObserver>[],
    this.builder,
    this.title = '',
    this.onGenerateTitle,
    this.color,
    this.theme,
    this.darkTheme,
    this.themeMode = ThemeMode.system,
    this.locale,
    this.localizationsDelegates,
    this.localeListResolutionCallback,
    this.localeResolutionCallback,
    this.supportedLocales = const <Locale>[Locale('en', 'US')],
    this.debugShowMaterialGrid = false,
    this.showPerformanceOverlay = false,
    this.checkerboardRasterCacheImages = false,
    this.checkerboardOffscreenLayers = false,
    this.showSemanticsDebugger = false,
    this.debugShowCheckedModeBanner = true,
})
```

    - For the / route, the home property, if non-null, is used.
    - Otherwise, the routes table is used, if it has an entry for the route.
    - Otherwise, onGenerateRoute is called, if provided. It should return a non-null value for any valid route not handled by home and routes.
    - Finally if all else fails onUnknownRoute is called.

- `Scaffold` 是一个UI的架子，一般我们会把它的实例放在 `MaterialApp` 的 `home` 上。

示例图：![A screenshot of the Scaffold widget with a body and floating action button](https://flutter.github.io/assets-for-api-docs/assets/material/scaffold.png)

我们来看看它的属性：

```dart
const Scaffold({
    Key key,
    this.appBar, // 标题
    this.body, // body就是我们要显示的其他的widget了
    this.floatingActionButton, // 右下角的浮动按钮
    this.floatingActionButtonLocation,
    this.floatingActionButtonAnimator,
    this.persistentFooterButtons,
    this.drawer,
    this.endDrawer,
    this.bottomNavigationBar,
    this.bottomSheet,
    this.backgroundColor,
    this.resizeToAvoidBottomPadding,
    this.resizeToAvoidBottomInset,
    this.primary = true,
    this.drawerDragStartBehavior = DragStartBehavior.start,
    this.extendBody = false,
    this.drawerScrimColor,
    this.drawerEdgeDragWidth,
})
```

- `AppBar` 是标题栏。上面可以放标题和一些图标，`title` 属性是标题。看看AppBar的主要属性：

![AppBar 属性](https://flutter.github.io/assets-for-api-docs/assets/material/app_bar.png)

```dart
AppBar({
    Key key,
    this.leading,
    this.automaticallyImplyLeading = true,
    this.title,
    this.actions,
    this.flexibleSpace,
    this.bottom,
    this.elevation,
    this.shape,
    this.backgroundColor,
    this.brightness,
    this.iconTheme,
    this.actionsIconTheme,
    this.textTheme,
    this.primary = true,
    this.centerTitle,
    this.titleSpacing = NavigationToolbar.kMiddleSpacing,
    this.toolbarOpacity = 1.0,
    this.bottomOpacity = 1.0,
})
```

- `Center` 把它的子组件放在中间。

- `Column` 和 `Row` 分别是列(纵向的叫列)和行(横向的叫行)，他们接受一个 `children` 参数，里面摆放子元素。

- `Text` 是显示文字的组件，常见用法：`Text("Hello World")`，看看它的主要参数：

```dart
const Text(
    this.data, {
    Key key,
    this.style, // 样式
    this.strutStyle,
    this.textAlign,
    this.textDirection,
    this.locale,
    this.softWrap,
    this.overflow,
    this.textScaleFactor,
    this.maxLines,
    this.semanticsLabel,
    this.textWidthBasis,
})
```

- `SingleChildScrollView` 是一个容器，它可以接受一个子组件，然后当超过屏幕长度或者宽度时，可以滑动。对应一组子组建的，
应当使用 `ListView`(性能更高)。我个人实验结果，他们的区别似乎在于，`ListView` 不能做其他组件的child，而
`SingleChildScrollView` 可以。

```dart
const SingleChildScrollView({
    Key key,
    this.scrollDirection = Axis.vertical,  // 元素摆放的方向
    this.reverse = false,
    this.padding,
    bool primary,
    this.physics,
    this.controller,
    this.child, // 比如，放一个Column或者Row在这
    this.dragStartBehavior = DragStartBehavior.start,
})
```

- `Builder` 有的时候我们需要使用context，但是没法直接使用父容器的context，这个时候就需要用Builder。Builder接受一个函数，通过这个函数会把context传进去。

```dart
drawer: Builder(
    builder: (context) => getDrawer(
        context, _controller.future, this.setLoadingStatus)),
```

- `Drawer` 这个就是App打开时的左侧边栏，例如：

```dart
Widget getDrawer(BuildContext context) {
  return Drawer(
      child: ListView(
    padding: EdgeInsets.zero,
    children: <Widget>[
      ListTile(
          leading: Icon(Icons.home),
          title: Text("主页"),
          onTap: () => __loadUrl("/")),
    ],
  );
}
```

- `WillPopScope` 用来支持这么一种情况，当你按下返回键，App将会采取什么措施。通过提供 `onWillPop` 来进行定义的操作。比如：

```dart
onWillPop: () async {
    var controller = await _controller.future;
    bool canGoBack = await controller.canGoBack();
    if (canGoBack) {
        controller.goBack();
        return true;
    }

    return false;
});
```

- `InkWell` 用来让一个东西可点击。比如标题栏。

```dart
title: InkWell(
    child: Text(homeTitle),
    onTap: () async {
        var done = await _controller.future;
        this.setLoadingStatus(true);
        done.loadUrl(homePage);
    }),
```
