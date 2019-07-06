# 后端工程师学前端（三）：CSS进阶(特指度、单位和字体族)

> 这一篇文章主要出自阅读《CSS权威指南》之后的笔记

## 特指度

CSS中，可能有多个规则针对同一个元素起作用，那么怎么确定他们的覆盖顺序呢？我们依靠特指度来确定。一个特指度由四个部分
组成：

- 选择符中的每个ID属性值加 `0,1,0,0`
- 选择符中的每个类属性值、属性选择或伪类加 `0,0,1,0`
- 选择符中的每个元素和伪元素加 `0,0,0,1`
- 行内声明的特指度为 `1,0,0,0`
- 连接符和通用选择符不增加特指度

举个例子：

- `h1 {color: red;}` 的特指度为 `0,0,0,1`
- `p em {color: red;}` 的特指度为 `0,0,0,2`
- `.grape {color: red;}` 的特指度为 `0,0,1,0`
- `div#sidebar {color: red;}` 的特指度为 `0,1,0,1`

特指度高的胜出。总而言之，就是具体度越高，则特指度越高。

## 层叠

层叠的权重从高到底：

- 读者提供的样式中以 `!important` 标记的声明
- 创作人员编写的样式中以 `!important` 标记的声明
- 创作人员编写的常规声明
- 读者提供的常规声明
- 浏览器的默认声明

## 常见单位

- 像素：像素是长度的绝对单位
- 百分比：相对于父元素的百分比
- rem：相对于html元素的 font-size 的大小
- 颜色：rgb和16进制值

## 字体族

字体是CSS中的一个重要组成部分，CSS中将英文字体分为如下五个分类：

- 衬线字体(serif)(对应中文中的白体)：这些字体，每个字母宽度各异，字母的笔画末尾有装饰，因此叫做衬线字体。例如 <p style="font-size:20px;font-family:Times,serif">AaBbCcDdEeFf</p> 的字母中，字母的边角处都有装饰。衬线体一般用于正文印刷，中文对应的一般称为宋体。
- 无衬线字体(sans-serif)(对应中文中的黑体)：无衬线字体字母宽度各异，字母的笔画末尾没有装饰，例如 <p style="font-size:20px;font-family:'Gill Sans', sans-serif;">AaBbCcDdEeFf</p> 。无衬线体简明精干，识别度高，适用于标题，广告。
- 等宽字体(monospace)：等宽字体中的每个字母的宽度是一样的，例如 <p style="font-size:20px;font-family:Consolas,monospace;">AaBbCcDdEeFf</p> 。等宽字体多用于平面设计，代码、虚拟终端也经常使用等宽字体。
- 草书字体：草书字体是模仿人类笔迹的字体
- 奇幻字体：不归上述字体族的字体就被归类到这里

使用 `@font-face` 来使用自定义的字体：

```css
@font-face {
  font-family: "Open Sans";
  src: url("/fonts/OpenSans-Regular-webfont.woff2") format("woff2"),
       url("/fonts/OpenSans-Regular-webfont.woff") format("woff");
}
<p style="font-family:'Open Sans'">Hello World</p>
```

---

- https://developer.mozilla.org/en-US/docs/Learn/CSS/Introduction_to_CSS/Values_and_units
- https://zh.wikipedia.org/zh/%E5%AD%97%E4%BD%93%E5%AE%B6%E6%97%8F
