---
layout: post
title: C语言与抽象思维(一)
---

本文是在读完《C程序设计的抽象思维》第九节`效率与ADT`之后的一些总结。同时我非常推荐这本书， 尽管译者笔误挺多。

往往我们学完一门语言的语法之后就不知道要干什么了， 这篇文章就带你用C语言实作一个简单地`非所见即所得(WYSIWYG)`编辑器。在`UNIX/Linux`下， 杰出的编辑器数不胜数， 为什么我们还要造轮子呢？ 事实上，你会发现实现这个编辑器花费了大量的时间， 但是最终所得到的功能确非常的简陋。但是， 在这里更重要的是体现一种抽象思维， 让你理解抽象思维的重要性。

## 功能分析

首先来说， 自顶向下的分析方法更容易让你从全局看清楚需求， 以便更好地制定方法策略。 我比较喜欢这种方法。

我们的编辑器是非所见即所得的，也就是说我们不动态的显示光标所在位置，我们需要做的就是维护一个缓冲区， 并且接受来自键盘的输入并且进行响应。

我们的编辑器需要一些什么功能呢？我们用下表来列出所有的功能：

| 命令 | 操作 |
|------|:--------: |
| 'F' | 将编辑器光标向右移动一个字符位置 |
| 'B' | 将编辑器向左移动一个字符位置 | 
| 'J' | 跳到缓冲区的最左边 | 
| 'E' | 跳到缓冲区的最右边 |
| 'Ixxx' | 将字符`xxx`插入到当前光标的位置 |
| 'D' | 删除当前光标位置后面的一个字符 |

我们先来看一下运行效果示例：

```bash
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
```

## 定义缓冲区抽象

我们的缓冲区必须要时刻知道当前光标位置， 能够进行增删查改，并且在程序结束之前不会丢失缓冲区内的字符。现在这一步我们只需要考虑我们的编辑器需要完成什么功能， 并不需要考虑怎样实现。

为了使接口尽可能的具有灵活性， 定义一个新的抽象数据结构来表示编辑器的缓冲区是合乎道理的。使用抽象数据类型的目的就是将行为与具体实现分离。我们可以用不同的实现完成相同的功能， 这一点我们稍后就能见识到。

### 定义缓冲区接口buffer.h

我们有六个操作， 所以我们为这六个操作分别定义六个函数。 当然我们还要定义一个分配新的缓冲区的函数和一个释放缓冲区的函数。

```c
/*
 * 在这里，bufferCDT是具体实现时候的缓冲区表示，为了让API不体现或者说不让用户接触到底层数据， 我们使用指针类型来表示缓冲区数据结构。
 * 问题一： `bufferADT`是什么？
 */
typedef struct bufferCDT *bufferADT;
```

> 想到了答案吗？ bufferADT是`struct bufferCDT *`的同义词， 那么`struct bufferCDT *p`中的`p`是什么呢？`p`是指向`bufferCDT`结构体的指针, 所以`struct bufferCDT *`就是指向`bufferCDT`这中结构体的指针类型， 所以`bufferADT`也是。你可以用`bufferADT buffer`来定义数据， 就跟你可以用`int a=0;`来定义数据一样。

接下来我们声明返回新的缓冲区的函数和销毁缓冲区的函数:

```c
bufferADT NewBuffer(void);
void FreeBuffer(bufferADT buffer);
```

下面我们声明六个操作函数和一个用于辅助我们可视化缓冲区的函数:

```c
void MoveCursorForward(bufferADT buffer);
void MoveCursorBackward(bufferADT buffer);

void MoveCursorToStart(bufferADT buffer);
void MoveCursorToEnd (bufferADT buffer);

void InsertCharacter(bufferADT buffer, char ch);
void DeleteCharacter(bufferADT buffer);

void DisplayBuffer(bufferADT buffer);
```

好了， 既然我们已经把完成功能的函数声明好了， 那么我们就直接在抽象思维上把编辑器给写了吧, 直接贴代码:

```c
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
```

看着这个代码你可以脑补出一开始我们的编辑器示例吗？

> 可能你会觉得抽象思维体现在哪里？这不是实打实的代码吗？你应该仔细观察，上面的这一个代码没有牵扯到任何的一个具体实现，我们只是定义了缓冲区操作该有什么函数， 然后就拿这些函数写了一个编辑器出来， 我们并不关心具体是怎么实现的， 我们之关心函数能并且要完成哪些功能。这就是我们的抽象。当然， 抽象的后果就是， 你现在复制粘贴代码是运行不了的， 哈哈哈

未完待续。。。
