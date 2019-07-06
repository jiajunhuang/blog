# 后端工程师学前端（二）：CSS基础知识(规则与选择器)

> 这一篇文章主要出自阅读《CSS权威指南》之后的笔记

## 基础介绍

CSS的全称是Cascading Style Sheet，即层叠样式表。层叠就说明样式可以组合起来使用。CSS说白了，其实就是一种DSL。
浏览器提供了各种各样调整样式的规则，而CSS就是把这些规则集合在一起来进行描述。

元素是HTML中的基本结构，例如p，div，span等等。元素按是否直接显示内容可以分为两种：

- 置换元素：浏览器展示的不是元素的内容，而是元素所指向的内容，例如 `<img src="xxx.png" >` ，浏览器并不展示 `<img src="xxx.png" >` 这样的文字，而是展示 `src="xxx.png"` 所指向的图片
- 非置换元素：浏览器直接展示元素的内容，例如 `<p>hello world</p>` 那么浏览器会展示 `hello world`

如果按显示的方式，可以这样分为两种：

- 块级元素：块级元素默认生成一个填满父级元素内容区域的框，旁边不能有其他元素。也就是说，块级元素在前后都会断行。例如 `<p>`, `<div>`
- 行内元素：行内元素在一行文本内生成元素框，不打断所在的行。例如 `<strong>`, `<em>`

虽然元素有默认的值来指示自己是块级元素还是行内元素，但是我们可以通过设置 `display` 这个值来强制让他们表现为块级元素/行内元素。`display` 默认取值 `inline`，可以为 `block`, `inline` 和 `run-in`。

HTML要求，行内元素可以放在块级元素中，反之则不行，但是CSS中没有这种要求。

## CSS与HTML的结合

我们如何把CSS应用到HTML上呢？常见的有三种方式：

- 在HTML元素中加一个属性：`style="xxx"` 来书写CSS，例如 `<button style="border-radius:9999px">this is</button>`
- 在HTML中使用 `<style>` 标签来书写CSS，例如：

```html
<!DOCTYPE html>
<html>
    <style>
button {
    border-radius: 99999px;
}
    </style>
    <body>
        <button>this is</button>
    </body>
</html>
```

- 在HTML中使用 `<link>` 来引用CSS，例如 `<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css" >`

此外，CSS中也可以使用 `@import` 指令导入其它CSS，例如 `@import url(styles.css)`。

## CSS规则的结构

举个例子：

```css
h1 {
    color: red;
    background: yellow;
}
```

其中的 h1 就叫选择符，括号和里面的内容就是声明块，而其中的每一个分号隔开来的，就是一个声明，例如 `color: red;`，而一个声明中
包含一个属性和一个值，例如上述例子中，`color` 就是属性，`red` 就是值。

CSS 的属性中，有一个概念是厂商前缀，例如：

- `-epub-`：ePub格式
- `-moz-`：基于Mozilla的浏览器
- `-ms-`：微软的浏览器
- `-o-`：基于Opera的浏览器
- `-webkit-`：基于WebKit的浏览器

## 媒体查询和特性查询

媒体查询是CSS中用于查询一些浏览器信息用的，例如屏幕宽度等，使用 `@media` 来进行查询；特性查询是CSS查询是否支持某特性，
使用 `@supports` 来进行查询。

## CSS 选择符

CSS有一套强大的选择符，用于匹配HTML中的元素。以下是常见的几种：

- 元素选择符：元素选择符直接使用HTML元素进行选择，例如：

```css
h2 {
    color: red;
}

p {
    color: yellow;
}
```

- 群组选择符，群组选择符就是一次选中多个元素，他们之间使用逗号分隔。例如：

```css
h2, p {
    color: yellow;
}
```

- 通用选择符：`*` 是通用选择符，他会选中所有元素

- 类选择符：类选择符的形式是 `.class`，以点开头，后面接类名
- ID选择符：ID选择符的形式是 `#id`，以 `#` 开头，后面接id
- 属性选择符：选择具有某个属性的元素，例如：`h1[class] {color: silver;}` 就会选中所有有 `class` 这个属性的元素。
- 后代选择符：例如 `h1 em {color: gray;}` 就会选中 `h1` 中的每个 `em` 元素。
- 伪类选择符：这个用于选择一些非元素的属性，例如：`a:link:hover {color: red;}` 就会使得当鼠标放在上面的时候，颜色变红

---

总结：

这一篇中看了一些CSS的基本概念，例如什么是CSS，如何使用CSS，CSS选择符。这是使用CSS的前置内容。
