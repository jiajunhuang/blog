:Date: 12/12/2014

C语言与抽象思维
===============

Part I
------

本文是在读完《C程序设计的抽象思维》第九节\ ``效率与ADT``\ 之后的一些总结.同时我非常推荐这本书,
尽管译者笔误挺多.

往往我们学完一门语言的语法之后就不知道要干什么了,
这篇文章就带你用C语言实作一个简单地\ ``非所见即所得(WYSIWYG)``\ 编辑器.在\ ``UNIX/Linux``\ 下,
杰出的编辑器数不胜数, 为什么我们还要造轮子呢？
事实上,你会发现实现这个编辑器花费了大量的时间,
但是最终所得到的功能确非常的简陋.但是, 在这里更重要的是体现一种抽象思维,
让你理解抽象思维的重要性.

功能分析
--------

首先来说, 自顶向下的分析方法更容易让你从全局看清楚需求,
以便更好地制定方法策略. 我比较喜欢这种方法.

我们的编辑器是非所见即所得的,也就是说我们不动态的显示光标所在位置,我们需要做的就是维护一个缓冲区,
并且接受来自键盘的输入并且进行响应.

我们的编辑器需要一些什么功能呢？我们用下表来列出所有的功能：

+----------+-----------------------------------------+
| 命令     | 操作                                    |
+==========+=========================================+
| 'F'      | 将编辑器光标向右移动一个字符位置        |
+----------+-----------------------------------------+
| 'B'      | 将编辑器向左移动一个字符位置            |
+----------+-----------------------------------------+
| 'J'      | 跳到缓冲区的最左边                      |
+----------+-----------------------------------------+
| 'E'      | 跳到缓冲区的最右边                      |
+----------+-----------------------------------------+
| 'Ixxx'   | 将字符\ ``xxx``\ 插入到当前光标的位置   |
+----------+-----------------------------------------+
| 'D'      | 删除当前光标位置后面的一个字符          |
+----------+-----------------------------------------+

我们先来看一下运行效果示例：

.. code:: bash

    $ ./editor 
    * Iabcd
     a b c d
            ^
    * J
     a b c d
    ^
    * F
     a b c d
      ^
    * D
     a c d
      ^
    * Ib
     a b c d
        ^
    * Q

定义缓冲区抽象
--------------

我们的缓冲区必须要时刻知道当前光标位置,
能够进行增删查改,并且在程序结束之前不会丢失缓冲区内的字符.现在这一步我们只需要考虑我们的编辑器需要完成什么功能,
并不需要考虑怎样实现.

为了使接口尽可能的具有灵活性,
定义一个新的抽象数据结构来表示编辑器的缓冲区是合乎道理的.使用抽象数据类型的目的就是将行为与具体实现分离.我们可以用不同的实现完成相同的功能,
这一点我们稍后就能见识到.

定义缓冲区接口buffer.h
~~~~~~~~~~~~~~~~~~~~~~

我们有六个操作, 所以我们为这六个操作分别定义六个函数.
当然我们还要定义一个分配新的缓冲区的函数和一个释放缓冲区的函数.

.. code:: c

    /*
     * 在这里,bufferCDT是具体实现时候的缓冲区表示,为了让API不体现或者说不让用户接触到底层数据, 我们使用指针类型来表示缓冲区数据结构.
     * 问题一： `bufferADT`是什么？
     */
    typedef struct bufferCDT *bufferADT;

    想到了答案吗？ bufferADT是\ ``struct bufferCDT *``\ 的同义词,
    那么\ ``struct bufferCDT *p``\ 中的\ ``p``\ 是什么呢？\ ``p``\ 是指向\ ``bufferCDT``\ 结构体的指针,
    所以\ ``struct bufferCDT *``\ 就是指向\ ``bufferCDT``\ 这中结构体的指针类型,
    所以\ ``bufferADT``\ 也是.你可以用\ ``bufferADT buffer``\ 来定义数据,
    就跟你可以用\ ``int a=0;``\ 来定义数据一样.

接下来我们声明返回新的缓冲区的函数和销毁缓冲区的函数:

.. code:: c

    bufferADT NewBuffer(void);
    void FreeBuffer(bufferADT buffer);

下面我们声明六个操作函数和一个用于辅助我们可视化缓冲区的函数:

.. code:: c

    void MoveCursorForward(bufferADT buffer);
    void MoveCursorBackward(bufferADT buffer);

    void MoveCursorToStart(bufferADT buffer);
    void MoveCursorToEnd (bufferADT buffer);

    void InsertCharacter(bufferADT buffer, char ch);
    void DeleteCharacter(bufferADT buffer);

    void DisplayBuffer(bufferADT buffer);

好了, 既然我们已经把完成功能的函数声明好了,
那么我们就直接在抽象思维上把编辑器给写了吧, 直接贴代码:

.. code:: c

    /*
     * File: editor.c
     *
     * This program implements a simple character editor, which is used to test 
     * the buffer abstraction. The editor reads and executes simple commands 
     * entered by the user.
     */

    #include <stdio.h>
    #include <ctype.h>
    #include "boolean.h"
    #include "buffer.h"
    #include "genlib.h"
    #include "simpio.h"

    /* Private function prototypes */

    static void ExecuteCommand(bufferADT buffer, string line);
    static void HelpCommand(void);

    /* Main program */

    int main(void)
    {
      bufferADT buffer;

      buffer = NewBuffer();
      while(TRUE) {
        printf("* ");
        ExecuteCommand(buffer, GetLine());
        DisplayBuffer(buffer);
      }
      FreeBuffer(buffer);
    }

    /*
     * Function: ExecuteCommand
     * Usage: ExecuteCommand(buffer, line);
     *
     * This function parses the user command in the string line and execute it on
     * the buffer.
     */

    static void ExecuteCommand(bufferADT buffer, string line)
    {
      int i;

      switch(toupper(line[0])) {
        case 'I':
          for(i=1; line[i] != '\0'; i++) {
            InsertCharacter(buffer, line[i]);
          }
          break;
        case 'D':
          DeleteCharacter(buffer); break;
        case 'F':
          MoveCursorForward(buffer); break;
        case 'B':
          MoveCursorBackward(buffer); break;
        case 'J':
          MoveCursorToStart(buffer); break;
        case 'E':
          MoveCursorToEnd(buffer); break;
        case 'H':
          HelpCommand(); break;
        case 'Q':
          exit(0);
        default:
          printf(" Illegal command\n"); break;
      }
    }

    /*
     * Function: HelpCommand
     * Usage: HelpCommand();
     *
     * This function lists the acailabel editor commands.
     */

    static void HelpCommand(void)
    {
      printf(" Use the following commands to edit the buffer: \n");
      printf(" I ... Inserts text up to the end of the line.\n");
      printf(" F     Moves forward a character\n");
      printf(" B     Moves backward a character\n");
      printf(" J     Jumps to the beginning of the buffer\n");
      printf(" E     Jumps to the end of the character\n");
      printf(" D     Delete the next character\n");
      printf(" H     Generates a help message\n");
      printf(" Q     Quits the program\n");
    }

看着这个代码你可以脑补出一开始我们的编辑器示例吗？

    可能你会觉得抽象思维体现在哪里？这不是实打实的代码吗？你应该仔细观察,上面的这一个代码没有牵扯到任何的一个具体实现,我们只是定义了缓冲区操作该有什么函数,
    然后就拿这些函数写了一个编辑器出来, 我们并不关心具体是怎么实现的,
    我们之关心函数能并且要完成哪些功能.这就是我们的抽象.当然,
    抽象的后果就是, 你现在复制粘贴代码是运行不了的, 哈哈哈

Part II
-------


上一次我们说到C语言结合抽象思维完成一个非所见即所得的编辑器,
并且我们已经定义了这个编辑器应有的行为,
基本上抽象也已经完成.这一节讲的更多是实现上的事情.光有设计思路是不够的,
到最后我们得作出一点什么东西才行.

数组实现
--------

字符串缓冲区有什么特点呢？首先我们需要记录光标位置,
其次要能对字符进行增删,
很自然的我们可以想到用数组来进行表示.数组表示可以轻易的记录当前光标的位置,
只需要记录下标值就可以.并且缓冲区中的字符十一个有序的同类序列,
这和数组的表示相吻合.但由于C语言中为数组申请空间时必须知道数组大小,
所以我们需要一个值记录现在已经使用了多少个字符.于是我们把结构体\ ``bufferCDT``\ 定义如下：

.. code:: c

    #define MaxBuffer 100

    struct bufferCDT {
      char text[MaxBuffer];
      int length; // 目前已经使用的长度
      int cursor; // 光标的位置
    };

接下来就只需要把对字符串进行操作的几个函数实现就可以了,
但因为我们需要把函数和数据分离, 也就是说, 可以同时并发调用这个函数,
但各个不同调用函数的人之间数据不会相互影响.所以我们的函数应该定义为这样子：\ ``void InsertCharacter(bufferADT buffer, char ch)``
每次都把bufferADT的实体传进去,
那么函数进行操作的时候就会有单独的一块空间,
多次调用相同函数不会相互影响.

我想,
在当前光标下删除字符、插入字符你一定可以自己动手完成的！什么？你不确定？那好,
我给你一个参考思路：对于删除字符,
首先要检查当前光标位置是否在有效范围之内,如果是,
那么直接把光标之后的所有字符向前移动一位,然后\ ``buffer->cursor--;``;
对于插入字符, 需要先检查目前有效范围是不是超过最大范围,
如果没有,那么把光标后的所有字符向后移动一位,然后把字符插入进字符串,
最后\ ``buffer->length++; buffer->cursor++;``.

光标移动？那更简单了, 我相信你可以的！

``DisplayBuffer``\ 的实现：

.. code:: c

    void DisplayBuffer(bufferADT buffer)
    {
      int i;

      for(i = 0; i < buffer->length; i++) {
        printf(" %c", buffer->text[i]);
      }
      printf(" \n");
      for(i = 0; i < buffer->cursor; i++) {
        printf("  ");
      }
      printf("^\n");
    }

栈实现
------

栈？编辑器？反正我一开始是没想到可以用栈表示.但思路其实很简单：分别用两个栈,
一个表示光标之前的字符, 一个表示光标之后的字符.原来是这样！

    这让我想到火影忍者里的一个小段子：鸣人在修炼风遁螺旋手里剑的时候非常努力,但由于需要大量查克拉,并且很多分身在同时修炼,九尾的查克拉很容易溢出.后来下雨了,鸣人累得趴下了,向卡卡西抱怨,风遁螺旋手里剑就像走路的时候,一边要看左边,同时还要看右边,这怎么做的到啊！卡卡西说,
    哦, 这很简单啊,
    于是就使用了一个影分身,一个负责看左边,一个负责看右边.这里也是一样的,
    一开始我在想,用栈怎么表示缓冲区？同时记录一个索引位置吗？这样子很不方便啊！翻到这一页的时候才发现,可以用两个栈,一个表示光标前,一个表示光标后...

数据结构该怎么定义呢？ 如上面所说：用两个栈！

.. code:: c

    struct bufferCDT {
      stackADT before;
      stackADT after;
    };

用栈其实很方便, 完成移动光标的操作只需要一个Push,
一个Pop就可以了.完成删除只需要Pop并丢弃该字符、插入只需要Push就可以.

别看我,\ `我可没有代码 <#资料>`__, 你一定可以自己写出来的！

总结
----

我们看了两种实现, 该总结一下了, 不知道大家有没有发现,
我们用了两种表现方式, 但是代码的接口却完全没动！这就是抽象的好处.
抽象可以让逻辑和实现分开, 只要实现提供能完成功能的函数, 实现随便改,
而逻辑动都不要动！感觉到了吗？为了验证我们的总结, 我们再说一种实现 ----
链表实现.

链表实现
--------

链表有什么好处呢？首先, 只要内存扛得住,
编辑器缓冲区可以无限长！其次,相比栈和数组表示,
把光标移动到缓冲区的首部和尾部时消耗特别小, 再者,
打字出错是经常发生地事情,
如果在缓冲区内插入完数据后发现,在最前面漏了一个字符！
如果我们用的是字符表示的话, 电脑会说“jerk!你想累死我是吧！”,
因为数组需要把一大堆字符全部往后移动, 然后才能插入！ 栈表示？
电脑也会骂你的！ 栈也需要一大把的Pop和Push！

*链表实现实际上也很方便！但是有一个小坑, 如果你认真自己思考的话,
很快你就会发现的,
当然,你发现以后要解决那就更简单了.链表表示由于需要画大量的图,Linux下我也一直没找到一个顺手的画图工具,
就先不写了！容我偷偷懒*

    链表表示你就当作是习题吧.习题二：用双向链表表示一下！
