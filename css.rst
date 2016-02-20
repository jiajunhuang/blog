:Date: 02/20/2016

学习css
========

CSS语法
--------

CSS规则由两个主要的部分构成：选择器，以及一条或多条声明。

.. code:: css

    selector {declaration1; declaration2; ... declarationN }

如:

.. code:: css

    h1 {color:red; font-size:14px;}

选择器可以分组，用逗号将需要分组的选择器分开。

.. code:: css

    h1,h2,h3,h4,h5,h6 {
        color: green;
    }

派生选择器允许你根据文档的上下文关系来确定某个标签的样式。

.. code:: css

    li strong {
        font-style: italic;
        font-weight: normal;
    }

会使下面的strong变成斜体字，而不是通常的粗体字。

.. code:: html

    <li><strong>我是斜体字。这是因为 strong 元素位于 li 元素内。</strong></li>

id 选择器可以为标有特定 id 的 HTML 元素指定特定的样式。id 选择器以 "#" 来定义。

.. code:: css

    #red {color:red;}
    #green {color:green;}

产生下面的影响:

.. code:: html

    <p id="red">这个段落是红色。</p>
    <p id="green">这个段落是绿色。</p>

派生选择器:

.. code:: css

    #sidebar p {
        font-style: italic;
        text-align: right;
        margin-top: 0.5em;
    }

上面的样式只会应用于出现在 id 是 sidebar 的元素内的段落。

类选择器以一个点号显示:

.. code:: css

    .center {text-align: center}

会对下面产生效果:

.. code:: html

    <h1 class="center">
        This heading will be center-aligned
    </h1>


class 也可以被用作派生选择器。

插入CSS有三种方式:

- 外部样式表

.. code:: html

    <head>
        <link rel="stylesheet" type="text/css" href="mystyle.css" />
    </head>

- 内部样式表

.. code:: html

    <head>
        <style type="text/css">
            hr {color: sienna;}
            p {margin-left: 20px;}
            body {background-image: url("images/back40.gif");}
        </style>
    </head>

- 多重样式，如果某些属性在不同的样式表中被同样的选择器定义,
  那么属性值将从更具体的样式表中被继承过来。

Bootstrap CSS
==============

Bootstrap是很出名的前端开源框架，我直接上v4版本。
