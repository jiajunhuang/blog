# 再读 Python Language Reference

https://docs.python.org/3.6/reference/ 重读笔记。

## Lexical analysis

- 在Python源文件中，空格将代码分割成token(s)。物理换行符（UNIX下的`\n`）对
Python程序没有实际意义，Python程序认定的是 `NEWLINE`。

- 在Python源文件中，一个Tab键在Python解释器解释的时候会被替换成8个空格的缩进，
而缩进对于Python有非常重要的意义。所以推荐在自己用的解释器里都把tab键替换成
4个空格（或者自己想要的空格数）。

```python
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

    ```python
    >>> if True\
    ... and False:
    ...     print("False")
    ...
    >>>
    ```

    - 隐式join：在小括号，中括号，大括号中的行都会被合并成逻辑行（只有一个NEWLINE）

    ```python
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
