---
layout: post
title: c语言中的命名空间(namespace)
tags: [c]
---

今天在看《C语言接口与实现》的时候， 里面讲到了命名空间， 大一学C++的时候有命名空间（老师只是告诉我们记住这个东西就可以， 根本不将干啥用的， 自己也没去探究）这个我记得， 但是我没想到C里面也有命名空间这一说。我查阅了相关资料， 得到两篇比较好的相关文章：

* [概念了解点我](http://www.findfunaax.com/notes/file/134)

* [详细说明点我](http://ejrh.wordpress.com/2012/01/24/namespaces-in-c/) (自备梯子)

如同上面所描述， 第一篇文章只是说了个大概， 没有详细说明， 是中文的； 第二篇有详细说明，但是是英文的， 所以我决定把第二篇翻译出来：

## Namespaces in C

这篇文章讲了为什么命名空间在编程中是如此的有用。也讨论了一些在C中很明显的模拟出实例的方法， 包括用结构体(struct)“具体化”的一种方式.

命名空间是在一个系统中对象名称的集合； 它提供了一种可以区分在其他命名空间中有相似命名的方法。命名空间在大型程序中很有用， 特别是可以用来避免在程序库(libraries)和独立开发的模块(modules)中的符号(命名)冲突的问题。

它也有助于按语义组织代码。例如， 在代码中当一个函数的命名空间可见时， 明显能帮助读代码的人了解更多的信息。这(namespaces)也可以应用到外部文档和其他源代码文件中函数的实现中去。

但不幸的是， C在语言层面上对命名空间的支持并不好(C has little namespacing functionality at the language level)。 在标识符(函数和变量)中, 有一个命名空间， 受到作用域规则限制。在两个编译单元中一定会共享的标识符在所有编译单元中都会共享：当你在用“extern”描述一个标识符的时候， 你在使用链接层级的命名空间。

### 命名规范(Naming conventions for namespaces )

通常解决C对命名空间支持不够的问题的方案是，在每个模块中(的函数前)添加一个前缀, 前缀暗示了这个app， 库， 或者是模块的来源。

例如， 在Subversion中有这样的名字`svn_fs_initialize` - 这表明这个函数属于 `svn_fs` 函数库。 Subversion描述了命名空间规范的多个层级。`svn_fs_initialize` 是公共API的一个例子。在单个Subversion库中需要对其他库可见的命名在命名空间之后用双下划线连接, 例如， `svn_fs_base__dag_get_node`. 最后， 在单个模块中使用的非公共函数(non-exported functions)不加前缀。

特别说明命名中使用双下划线在实际应用编程中是不鼓励的；因为这是“实现”(通常错误的把它当做编译器和标准库)的保留用法。然而Subversion是一系列的函数库， 并且在其中的双下划线是明确定义的(well-defined)并且总是有snv前缀开头的， 所以它们这种情况下还是比较安全的。我编程的时候一般是（把双下划线）用来分隔类名和方法名的， 例如 `galaxy__add_star`, 在我以前关于[封装](http://ejrh.wordpress.com/2011/04/29/encapsulation-in-c/)的博文和下面的例子中也是这样。尽管清晰的代码会避免像这样使用双下划线。

前缀独一无二和冗长之间是有一个权衡的。一个很短的命名空间前缀就有和其他库命名冲突的风险。一个很长的（命名）又是的代码变多， 并且阻塞编辑器自动补全的特性----你需要打出整个命名空间直到在你的库中有一些能够匹配的补全出现。

### 用struct来实例化命名空间 (Reified namespaces with structs)

我们能用比这几个字符更明显的的方式区分命名空间和命名吗？不是像"namespace_object", 依靠人的大脑来人为区分："_"的前后两部分组成了一个命名， 而是用语言的某种标志来区分它们吗？ 要知道， 我说的就是像C++（"::"）和python(".")这样做。
C++是静态语言， 它在编译的时候就决定了每一个命名的作用域。“Namespace::Object”表明Object是在Namespace这个命名空间的。

另一方面， python是动态语言， 它没有明确的命名空间（然而它是有作用域的）。当对象被创建时， 它们会自动分配得到一些其它对象的属性；它的父对象就变成了命名空间。其实这就是python的模块： 它们是有属性的对象，比如“os”模块， 就有它的属性“os.getcwd"。

C没有"::"操作符。但是它有".": 点操作符用来访问struct(成员)。如果我们定义一个struct， 并且用适当的值初始化并且把结构体实例化， 我们就可以把实例当做命名空间来用。就意义而言， 现在这个命名空间变成了一个实例， 不再是每个名称前的那一串字符， 现在是程序中的一个分隔符。

```c
/* galaxy.h */
#ifndef GALAXY_H
#define GALAXY_H
 
typedef struct STAR { ... } STAR;
 
typedef struct GALAXY { ... } GALAXY;
 
extern GALAXY *galaxy__create(void);
extern void galaxy__create(GALAXY *galaxy);
extern void galaxy__add_star(GALAXY *galaxy, STAR *star);
 
static const struct
{
    GALAXY *(* create)(void);
    void (* destroy)(GALAXY *galaxy);
    void (* add_star)(GALAXY *galaxy, STAR *star);
} galaxy = {
    galaxy__create,
    galaxy__destroy,
    galaxy__add_star
};
 
#endif /* GALAXY_H */
```

模块的用户可以通过命名空间使用它的函数。很明显`galaxy.add_star`是命名空间`galaxy`的成员。只要在程序中没有其他的被命名为`galaxy`的模块被定义， 那么`add_star`就没有命名冲突的风险。完整的函数名`galaxy__add_star`遵守了传统的命名规范。

```c
/* main.c */
#include "galaxy.h"
 
int main(void)
{
    /* Every call to a galaxy-related function is prefixed by the "galaxy" namespace. */
    GALAXY *g = galaxy.create();
    galaxy.add_star(g, s);
    galaxy.destroy(g);
}
```

在galaxy模块中， 可以直接用`galaxy__add_star`而避免使用命名空间。 这会对程序性能有一定的影响（下面将会讨论）。

```c
/* galaxy.c */
GALAXY *galaxy__load(const char *filename)
{
    GALAXY *g = galaxy__create();
    ...
    while (!feof(f))
    {
        STAR *s = read_star(f);
        galaxy__add_star(s);
    }
    ...
    return g;
}
```

### 权衡

这种方法有一些缺陷。

第一个缺点就是编写模块的人需要写很多头文件样板。其次， 模块使用者每次使用都要用命名空间 -- C里面没有像C++，Java， Python那样简单的可以用import导入命名空间的方法。他们也可以用全名来调用， 但是这样名字很长。

命名空间也有限制。类型(structs, enums, typedefs)就不能放到命名空间， 因为没有变量可以代表它们。所以我们不能有这样一种类型"galaxy.GALAXY"。

在命名空间中使用可变的全局变量需要慎重。如上所述， 命名空间在每一个包含了头文件的模块中都是不同的全局变量。如果使用全局变量(例如 `galaxy.enable_tracing`), 每个模块都会有它自己的一份私有副本；在`main.c`中使用`galaxy.enable_tracing`不会对`galaxy.c`产生任何影响。

对于一个支持可变全局变量的命名空间结构， 命名空间需要在仅一个地方定义并且在所有要用到它的地方进行声明。这样才不会变成一个静态对象。所以上面的代码应该改成：

```c
/* galaxy.h */
#ifndef GALAXY_H
#define GALAXY_H
 
typedef struct STAR { ... } STAR;
 
typedef struct GALAXY { ... } GALAXY;
 
struct galaxy_struct
{
    GALAXY *(* create)(void);
    void (* destroy)(GALAXY *galaxy);
    void (* add_star)(GALAXY *galaxy, STAR *star);
    int enable_tracing;
};
 
extern struct galaxy_struct galaxy;
 
#endif /* GALAXY_H */
 
/* galaxy.c */
#include "galaxy.h"
 
/* The single instance of the namespace object. */
struct galaxy_struct galaxy =
{
    galaxy__create,
    galaxy__destroy,
    galaxy__add_star,
    0
} galaxy;
```

这样同时也把一些多余(即在这里其实不需要用的变量)的变量一道私有的实现中，这是很有用的。这就意味着一个库中本来很长的名字变短了， 因为不在需要"_"和前缀了(假设它们是在同一个模块中作为命名空间实例定义的)。

### 考虑一下效率

使用命名空间的第一种实现方式， 即在位头文件中的全局变量， 意味着再每个包含这个头文件的编译单元中都会有一份他自己的私有拷贝。关键字const意味着编译器在理论上可以把所有的结构体成员函数内联起来。 实际上是这个样子吗？

```c
#include
 
void f1(int x)
{
    printf("f1 %d\n", x);
}
 
void ns_f2(int x)
{
    printf("f2 %\n", x);
}
 
static const struct {
    int filler;
    void (* f2)(int x);
} ns = { ns_f2 };
 
struct nsx_struct {
    int filler;
    void (* f3)(int x);
};
 
extern struct nsx_struct nsx;
 
int main(void)
{
    f1(1);
    ns.f2(2);
    nsx.f3(2);
}
```

不开启优化选项， GCC把它们编译成：

```c
; Call f1, directly
    movl    $1, (%esp)
    call    _f1
; Call ns.f2 via the local, static namespace
    movl    _ns+4, %eax
    movl    $2, (%esp)
    call    *%eax
; Call nsx.f3 via the external namespace
    movl    _nsx+4, %eax
    movl    $2, (%esp)
    call    *%eax
```

So, calling a function in a namespace requires looking up the function in the struct, including applying its offset, and calling it via a register.(这一句不会翻译， 没学过汇编， 不敢瞎扯。 大概是说在命名空间中调用一个函数需要在结构体中查找成员函数， 包括实现偏移， 偏移之后从寄存器中调用函数。)

使用1级或更高级别的优化以后， 汇编代码变成:

```c
; Call f1, directly
    movl    $1, (%esp)
    call    _f1
; Call ns.f2 via the local, static namespace
    movl    $2, (%esp)
    call    *_ns+4
; Call nsx.f3 via the external namespace
    movl    $2, (%esp)
    call    *_nsx+4
```

实际情况中(用这个编译器)， 本地命名空间和外联命名空间调用会编译成相同的指令。实验表明如果f2是ns中的第一个元素， 那么对f2的调用可以直接被优化到f1的调用。 However, adding filler prevents this optimisation in GCC. 我怀疑这是因为很难在链接级别解释跳转偏移量。

延伸阅读。 在StackOverflow的[这个问题](http://stackoverflow.com/questions/389827/namespaces-in-c)有关于以上技术的更多讨论和附加建议。

-------------------------------------

##翻译之后的感想：

这篇文章我从上午开始翻译， 一直到现在(晚上7点半)才翻译完成(中午睡了个觉)，没想到第一次翻译就遇到这么个难缠的主，自己读原文很快，但是翻译的时候总是要反复的读， 怕翻译的不够native， 但是因为自身技术基础不够， 我觉得翻译的不好， 鉴于我希望大家能帮我一把， 我还没有完全校对完（好吧。。。我已经烦了）直接把译文发布了出来， 希望大家能够指点， 我定虚心受教～
