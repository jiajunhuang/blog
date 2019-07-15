# 后端工程师学前端(五): SASS

此前我们学习了基本的CSS和HTML的知识，因此我们已经可以开始构建基本的页面，
但是现实工程中有一个问题，即CSS使用的越来越多，代码维护难度也上升。

因此2015-2018年左右，开始有了轰轰烈烈的前端工程化。

> 这一篇是阅读SASS官方指南之后的笔记

现在(2019年)，基本上要用的工具都定型了，我们作为渔翁，可以直接上手开始使用。今天我们介绍一个CSS预处理器： SASS。

## SASS

CSS的本质是DSL，它是一套规则，因此书写的时候就是一条条的规则，不像编程语言那样可以继承。而SASS的作用就是解决这个问题，
SASS编译之后产生css。

> SASS有两种语法：SCSS和SASS。SASS只是CSS的语法糖；SCSS是超集。官方推荐SCSS，不过我觉得使用SASS就够了。

首先看一下如何使用：

```bash
$ sudo pacman -S sassc  # 安装sass编译器，我安装的是C语言实现版本
$ cat > hello.sass
$font-stack: Helvetica, sans-serif
$primary-color: #333

body
    font: 100% $font-stack
    color: $primary-color
$ sassc -t compressed hello.sass hello.min.css
$ cat hello.min.css
body{font:100% Helvetica,sans-serif;color:#333}
```

接下来看看SASS的语法：

- 支持 `//` 和 `/* */` 两种注释

- 使用美元符号来声明变量：

```sass
$font-stack: Helvetica, sans-serif
$primary-color: #333

body
    font: 100% $font-stack
    color: $primary-color
```

- 通过缩进来嵌套：

```sass
nav
  ul
    margin: 0
    padding: 0
    list-style: none

  li
    display: inline-block

  a
    display: block
    padding: 6px 12px
    text-decoration: none
```

- 通过使用下划线开头的文件名来将一个sass文件作为可导入的模块；通过使用 `@import` 来导入

```sass
// _reset.sass
// _reset.sass
html,
body,
ul,
ol
  margin:  0
  padding: 0

// base.sass
@import reset
body
  font: 100% Helvetica, sans-serif
  background-color: #efefef
```

- 通过使用 `%` 定义可复用组件，然后使用 `@extend` 继承：

```sass
/* This CSS will print because %message-shared is extended. */
%message-shared
  border: 1px solid #ccc
  padding: 10px
  color: #333


// This CSS won't print because %equal-heights is never extended.
%equal-heights
  display: flex
  flex-wrap: wrap


.message
  @extend %message-shared


.success
  @extend %message-shared
  border-color: green


.error
  @extend %message-shared
  border-color: red


.warning
  @extend %message-shared
  border-color: yellow
```

- 使用 `+, -, *, /, %` 进行运算：

```sass
.container
  width: 100%


article[role="main"]
  float: left
  width: 600px / 960px * 100%


aside[role="complementary"]
  float: right
  width: 300px / 960px * 100%
```

---

参考资料：

- https://sass-lang.com/guide
