# 再读 Python Language Reference

https://docs.python.org/3.6/reference/ 重读笔记。

## Lexical analysis

- 在Python源文件中，空格将代码分割成token(s)。物理换行符（UNIX下的`\n`）对
Python程序没有实际意义，Python程序认定的是 `NEWLINE`。

- 在Python源文件中，一个Tab键在Python解释器解释的时候会被替换成8个空格的缩进，
而缩进对于Python有非常重要的意义。所以推荐在自己用的解释器里都把tab键替换成
4个空格（或者自己想要的空格数）。

```ipython
>>> def foo():
...     print("hello")  # 1 tab here
...     print("world")  # 4 spaces here
  File "<stdin>", line 3
    print("world")
                 ^
IndentationError: unindent does not match any outer indentation level
```

- `NEWLINE`

    - 显式join：使用`\`

    ```ipython
    >>> if True\
    ... and False:
    ...     print("False")
    ...
    >>>
    ```

    - 隐式join：在小括号，中括号，大括号中的行都会被合并成逻辑行（只有一个NEWLINE）

    ```ipython
    >>> (
        ... "hello"
        ... "world"
        ... )
    'helloworld'
    >>>
    ```

## Data model

Python中每个对象都有自己的id，类型和值。

- Python中的对象从对象的值的可变性上分为两类：

    - mutable，例如 dict, list
    - immutable，例如 str, tuple

- 从理论上来讲，副作用例如打开一个文件，当此文件没有引用时是会被Python垃圾回收
的，但是由于gc无法做出保证（例如Python开启了DEBUG模式，所有的对象都会被一个
双链表链着，这个时候引用计数就不为0），所以最好还是自己显式关闭。

## Execution model

- Python程序由 code blocks 组成，block是Python执行代码的最小单元，block分为：

    - module
    - function body
    - class definition
    - 交互模式下每次输入的命令都是一个block
    - 每一个script file
    - `$ python -c 'print("hello")'` 中的命令也是一个block
    - 传递给 `eval()` 和 `exec()` 的字符串也是一个block，但是这两个函数作用域规则有些特殊，会跳过LEGB中的E，也就是闭包。

每个block都在一个 execution frame 里执行。

- name binding

    - 给函数传参数
    - import语句
    - 类和函数定义
    - 赋值语句
    - for语句的后面一个 `for i in ...`
    - with语句的as后面 `with open(...) as ...`
    - except后面的as
    - `from ... import *` 绑定所有非下划线开头的名字到当前命名空间
    - del 也是，虽然它的作用是删除绑定

只要出现了上述的binding，Python就认为这个变量名所指向的变量在当前的block里，所以：

```ipython
In [1]: foo = "hello"

In [2]: def print_foo():
   ...:     print(foo)
   ...:     foo = "world"
   ...:

In [3]: print_foo()
---------------------------------------------------------------------------
UnboundLocalError                         Traceback (most recent call last)
<ipython-input-3-c80be14760b6> in <module>()
----> 1 print_foo()

<ipython-input-2-afc4cd9aff75> in print_foo()
      1 def print_foo():
----> 2     print(foo)
      3     foo = "world"
      4

UnboundLocalError: local variable 'foo' referenced before assignment

In [4]:
```

- "The global statement has the same scope as a name binding operation in
the same block. If the nearest enclosing scope for a free variable contains
a global statement, the free variable is treated as a global."

没看懂，来自：https://docs.python.org/3.6/reference/executionmodel.html

- 类定义的变量作用域局限于类内，而不会扩展至方法和表达式，generator里，所以：

```ipython
In [1]: class Foo:
   ...:     hello = "hello"
   ...:     def foo(self):
   ...:         print(hello)
   ...:

In [2]: Foo().foo()
---------------------------------------------------------------------------
NameError                                 Traceback (most recent call last)
<ipython-input-2-6c4f5adc4d1e> in <module>()
----> 1 Foo().foo()

<ipython-input-1-029db0d572c9> in foo(self)
      2     hello = "hello"
      3     def foo(self):
----> 4         print(hello)
      5

NameError: name 'hello' is not defined

In [3]: class Foo:
   ...:     a = 32
   ...:     b = [a + i for i in range(10)]
   ...:
---------------------------------------------------------------------------
NameError                                 Traceback (most recent call last)
<ipython-input-3-395876c62295> in <module>()
----> 1 class Foo:
      2     a = 32
      3     b = [a + i for i in range(10)]
      4

<ipython-input-3-395876c62295> in Foo()
      1 class Foo:
      2     a = 32
----> 3     b = [a + i for i in range(10)]
      4

<ipython-input-3-395876c62295> in <listcomp>(.0)
      1 class Foo:
      2     a = 32
----> 3     b = [a + i for i in range(10)]
      4

NameError: name 'a' is not defined

In [4]:
```

- 闭包内的变量（free variables）的值是在运行时确定的：

```ipython
In [1]: foo = "hello"

In [2]: def f():
   ...:     print(foo)
   ...:

In [3]: f()
hello

In [4]: foo = "world"

In [5]: f()
world

In [6]:
```

## The import system

- `import` 语句包含两个部分

    - `__import__` 搜寻指定名称的包
    - 然后将 `__import__` 返回值在语句所在的scope里binding成那个名字

- 导入的时候，会根据点号依次导入，而且会首先搜寻 `sys.modules`，例如：

```ipython
In [1]: import sys

In [2]: sys.modules['tornado']
---------------------------------------------------------------------------
KeyError                                  Traceback (most recent call last)
<ipython-input-2-36f3d84239fe> in <module>()
----> 1 sys.modules['tornado']

KeyError: 'tornado'

In [3]: sys.modules['tornado.web']
---------------------------------------------------------------------------
KeyError                                  Traceback (most recent call last)
<ipython-input-3-f1a34b2734fb> in <module>()
----> 1 sys.modules['tornado.web']

KeyError: 'tornado.web'

In [4]: import tornado.web

In [5]: tornado
Out[5]: <module 'tornado' from '/usr/local/lib/python3.5/dist-packages/tornado/__init__.py'>

In [6]: tornado.web
Out[6]: <module 'tornado.web' from '/usr/local/lib/python3.5/dist-packages/tornado/web.py'>

In [7]: sys.modules['tornado']
Out[7]: <module 'tornado' from '/usr/local/lib/python3.5/dist-packages/tornado/__init__.py'>

In [8]: sys.modules['tornado.web']
Out[8]: <module 'tornado.web' from '/usr/local/lib/python3.5/dist-packages/tornado/web.py'>

In [9]:
```
