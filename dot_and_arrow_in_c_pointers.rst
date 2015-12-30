:Date: 09/09/2014

C语言中点操作符(.)和箭头操作符(->)的不同之处
=============================================

自己写一个简单的uname, 要用到utsname结构体, 编译报错如下：

.. code:: c

    uname.c: In function ‘main’:
    uname.c:8:42: error: invalid type argument of ‘->’ (have ‘struct utsname’)
       printf("%s - %s - %s - %s - %s\n", name->sysname, name->nodename,\
                                              ^

查实一下, 是因为用错了操作符：

-  -> 的左侧必须是指针.
-  . 的左侧必须是结构体实体.

程序如下:

.. code:: c

    $ cat -n uname.c
    1  #include <sys/utsname.h>
    2  #include <stdio.h>
    3
    4  int main(void)
    5  {
    6    struct utsname name;
    7    printf("%d\n", uname(&name));
    8    printf("%s - %s - %s - %s - %s\n", name.sysname, name.nodename,\
    9        name.release, name.version, name.machine);
    10    return 0;
    11  }

2014-09-21:
~~~~~~~~~~~

(摘自《征服c指针》-前桥和弥)：

| p->hoge;
| 是
| (\*p).hoge;
| 的语法糖.
