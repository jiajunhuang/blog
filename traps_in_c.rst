:Date: 11/04/2014

C注意事项
=========

2014-11-04:
~~~~~~~~~~~

-  C标准中main函数只有两种形态:

.. code:: c

    int main(void);
    int main(int argc, char *argv[]);

-  不可以通过指向字符串常量的指针修改字符串:
   很明显, 字符串常量, 不可以修改, 但是很多时候还是容易犯错误

.. code:: c

    #include <stdio.h>

    int main(void)
    {
      char *pstr = "hello, world";
      // *(pstr+2) = 'd'; 编译会报错！

      char str[] = "hello, world";
      str[3] = 'd'; // OK
      char *p_str = str;
      *(p_str + 3) = 'f'; // OK

      return (0);
    }

2014-12-01:
~~~~~~~~~~~

-  C语言中void类型怎么返回？使用\ ``return``:

.. code:: c

    void func(void)
    {
      if (...) {
        return;
      }
      ...
    }

-  C语言和面向对象思想并不冲突 ;) 这两者是可以在一起的

2014-12-14:
~~~~~~~~~~~

-  typedef, struct, do...while后切记要加分号;
