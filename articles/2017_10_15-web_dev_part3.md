# Web开发系列(三)：什么是HTML,CSS,JS？

这篇博客要写的内容比较无聊，属于概念性的东西，我们介绍一下web开发中无法避开的三个东西，HTML，CSS和JS。

HTML全称是Hypertext Markup Language，也是一种格式，或者说一种约定，大概张这样子：

```html
<!DOCTYPE html>
<html>
<body>

<h1>My First Heading</h1>
<p>My first paragraph.</p>

</body>
</html>
```

而HTML5则是在传统的HTML上增加了一些标签，好像是最新版吧，好久没去了解这个了。
顺便需要说一句，上面所说的标签就是例子中 `<html>`，`</html>`这样的，通过写这些规定的标签，浏览器会展现出不同的样子。
例如 `<p>` 是正文，`<h1>` 是一级标题，通常会有加大字体，`<a>`是超链接，一般都会是蓝色字体，外带下划线。

那么为什么不同的标签在浏览器中会有不同的样子呢？我们可以更改或者制定这些样子吗？这就涉及到我们需要了解的第二个知识了，CSS。
Cascading Style Sheets是CSS的全称，其实就是用来告诉浏览器这个标签要怎么展示出来，是不是要加特效。其实CSS也可以写在HTML里，
例如：

```html
<!DOCTYPE html>
<html>
<body>

<h1>My First Heading</h1>
<p style="font-weight: bold">My first paragraph.</p>

</body>
</html>
```

也可以写成HTML:

```html
<!DOCTYPE html>
<html>
<head>
    <link rel="stylesheet" type="text/css" href="theme.css">
</head>
<body>

<h1>My First Heading</h1>
<p>My first paragraph.</p>

</body>
</html>
```

外加css：

```css
p {
    font-weight: bold;
}
```

记得要在HTML里链接css的路径。

那么JS呢？JS是JavaScript的缩写，为什么我们需要这样一个东西呢？因为浏览器端除了渲染出HTML之外，如果能
执行一些脚本，那么将会减轻服务端的很多压力，尽管服务端为了安全仍然要做各种校验。比如一个输入框用来
输入电话号码，如果在浏览器端能够检测号码格式是否合法，如果格式不合法就不提交请求，那么服务器端便可以减少
很多不必要的请求。另外，如果能通过脚本动态和服务器通信，从而根据通信内容动态更改内容，那就最好了，此种技术
叫做AJAX(asynchronous JavaScript and XML)。

讲完，手工 :)
