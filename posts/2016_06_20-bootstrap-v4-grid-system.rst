学习使用Bootstrap4的栅格系统
============================

响应式布局，在n年前被炒的火热火热的，因为前端的代码只需要写一次，却可以在各种
分辨率的屏幕上显示，确实是个好东西。不过前端的东西总是变的比摩尔定律还快，现
在很多公司又开始倾向于桌面设备和移动设备分开写两套前端代码了。

不管前端风向那么多，今天一起来看一下Bootstrap的响应式栅格布局，这可以说是后端
程序员值得掌握的一个好东西。不过呢，由于各大管理系统用的都是Bootstrap的那套图标，
所以图标我们就忽略吧，看多了都有视觉疲劳了。

容器
-------

容器(Containers)是Bootstrap中的最基本的布局元素，如果你想用Bootstrap的栅格布局，
那就一定要用到它。Bootstrap中的容器是会根据屏幕的宽度来变化的，如果css中选择的
class为 ``class="container"`` 那么在几个屏幕宽度的临界值之间，容器的宽度是不变
的。下表是临界值的说明表格：

=============  ==================================  ===========
临界值         设备种类和例子                      对应的简称
-------------  ----------------------------------  -----------
34em以下       extra small 例如竖屏的iPhone6s      xs
34em ~ 48em    small 例如横屏的iPhone6s            sm
48em ~ 62em    medium 例如iPad pro 9.7             md
62em ~ 75em    large 例如常见的1366x768笔记本      lg
75em以上       extra large 例如24英寸外接显示器    xl
=============  ==================================  ===========

其中em是相对浏览器默认字体大小。

另外还有一个class为 ``class="containers-fluid"`` 这个container就是一直占着
100%屏宽的容器。

栅格
------

对于Bootstrap布局来说，有三个重要的概念，分别是容器，行(拼音hang，后同)，列。
容器我们上面 已经讲完了，接下来我们讲行(hang)，行(hang)呢，功能上有点像一个
更小的容器，在这 里面我们可以放置更多列，注意一点，只有列能作为行的直系子元素。
over，讲完了。不理解？没关系，先记住。

列(column)是布局的关键点啊，就是有一个长方形，横着摆的，切11刀，分成等分的
12列，因为切的时候是等分，所以与具体的宽度是没有关系的。对于列，有这些个class:

- ``col-xs-*``

- ``col-sm-*``

- ``col-md-*``

- ``col-lg-*``

- ``col-xl-*``

上面的 ``*`` 是可以被替换的，可以换成 1-12 中的任意数字，举个例子，如果你想
把这个长方形切成三份咋写？

.. code:: html

    <body>
        <div class="container">
            <div class="col-md-4" style="background: yellow">
                hello world
            </div>
            <div class="col-md-4" style="background: red">
                hello world
            </div>
            <div class="col-md-4" style="background: green">
                hello world
            </div>
        </div>
    </body>

那我想2:1比例切成两份呢？

.. code:: html

    <body>
        <div class="container" style="background: yellow">
            <div class="col-md-8">
                hello world
            </div>
            <div class="col-md-4" style="background: red">
                hello world
            </div>
        </div>
    </body>

就是这么简单。打开chrome，加载这个文件（顺便说，这种方式网页里的在线资源无法
加载，这是chrome的设置，想避免的话，还是搞个web服务器吧），预览上面两种效果，
怎样，第一种是不是三等分？第二种是不是2:1切分？

好了，现在我们可以按F12打开chrome的控制台，然后手贱的点点控制台左上角的第二个
图标，来，虽是穷屌丝，但是最少这个时候我们可以选择iPhone6s！顺便刷新一下吧。

不对啊。。。。怎么不是排成排了。原来是列(column)的class只会对它以及比它更大的
产生效果，这话好想有点拗口。不过意思就是，我们选的class是 ``col-md-*`` 所以
只有当屏幕宽度大于上面定义的md才会生效。soga，怪不得我们选择6S的时候变化了，
比他更小的时候就会变成堆叠式的布局。

更多的class
-------------

如果我想在iPad上三等分，但是在iPhone6S上2:1:1显示呢？怎么同时匹配这两种设备？

.. code:: html

    <body>
        <div class="container">
            <div class="col-md-4 col-sm-6" style="background: yellow">
                hello world
            </div>
            <div class="col-md-4 col-sm-3" style="background: red">
                hello world
            </div>
            <div class="col-md-4 col-sm-3" style="background: green">
                hello world
            </div>
        </div>
    </body>

好，继续在chrome控制台选择iPhone6s，刷新，看效果，然后选择最上面的横屏按钮，
看效果。

特效，导演我要求加特效
---------------------------

有的时候，我想在大屏上 2:8:2 分屏，左边吧，放点站点的介绍，右边吧，放个子
tag标签云啊什么的，但是呢，在手机上我就简单点，只显示主体部分，这咋办？
下面导演给你加特效，隆重介绍 ``hidden-*-up`` 和 ``hidden-*-down`` 其中
``*`` 可以替换为 ``sm, md, lg, xl`` 。比如下面的代码，你选6s然后横屏竖屏
看看效果？

.. code:: html

    <body>
        <div class="container">
            <div class="col-sm-2 hidden-xs-down" style="background: yellow">
                hello world
            </div>
            <div class="col-sm-8" style="background: red">
                hello world
            </div>
            <div class="col-sm-2 hidden-xs-down" style="background: green">
                hello world
            </div>
        </div>
    </body>

收工，做个总结
------------------

好了，讲了这么多，其实Bootstrap的栅格布局用起来还是相当好用的。啊？就这点总结啊？
当然啦，无形装逼，最为致命。
