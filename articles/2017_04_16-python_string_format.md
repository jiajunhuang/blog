# Python字符串格式化

说来惭愧，用Python这么久了，还没有系统的看过字符串格式化的文档，因为常用的也就
那么几个。这篇文章就总结一下Python的字符串格式化。

## format() 方法

> https://docs.python.org/3.4/library/string.html#format-string-syntax

先来看一下文档：

```python
In [1]: format?
Docstring:
format(value[, format_spec]) -> string

Returns value.__format__(format_spec)
    format_spec defaults to ""
    Type:      builtin_function_or_method
```

format方法会替换value中 `{}` 包围的内容，其他字符直接拷贝过去不变。

具体允许的语法为：

```python
"{" [field_name] ["!" conversion] [":" format_spec] "}"
```

其中 `field_name` 可以为空，为数字，为变量名，分别对应用法：

```python
In [1]: "{}".format(1)
Out[1]: '1'

In [2]: "{0}{1}{2}".format(1, 2, 3)
Out[2]: '123'

In [3]: "{0}{1}{0}".format(1, 2)
Out[3]: '121'

In [4]: "{hello} world".format(hello="hello")
Out[4]: 'hello world'
```

并且 `field_name` 后可以接变量属性名，或者index，例如：

```python
In [1]: class MyObject(object):
    ...:     val = 1
    ...:

In [2]: myobj = MyObject()

In [3]: "{.val}".format(myobj)
Out[3]: '1'

In [4]: "{0.val}".format(myobj)
Out[4]: '1'

In [5]: "{myobj.val}".format(myobj=myobj)
Out[5]: '1'
In [6]: atuple = (1, 2)

In [7]: "{0[1]}{0[0]}".format(atuple)
Out[7]: '21'
```

但是不可以使用字典里的值，因为引用字典中的值需要使用 `adict["key"]` 的形式，
会把当前字符格式打乱掉。其中上面所说的接点号进行属性引用调用 `getattr`，接
index引用调用 `__getattr__`。

conversion有三个选项：

    - `s` 将调用 `str()`
    - `r` 将调用 `repr()`
    - `a` 将调用 `ascii()` (Python 3才有)

```python
In [1]: class MyObject(object):
   ...:     def __str__(self):
   ...:         return "__str__ been called"
   ...:     def __repr__(self):
   ...:         return "__repr__ been called"
   ...:

In [2]: myobj = MyObject()

In [3]: "{!r}".format(myobj)
Out[3]: '__repr__ been called'

In [4]: "{!s}".format(myobj)
Out[4]: '__str__ been called'

In [5]: "{!a}".format(myobj)
Out[5]: '__repr__ been called'
```

其中 `{!a}` 在调用完 `repr()` 之后把非ascii字符替换掉。

之后接 `:`，冒号之后是具体如何展现这个值，文档在：

https://docs.python.org/3.4/library/string.html#format-specification-mini-language

其格式为：`[[fill]align][sign][#][0][width][,][.precision][type]`

其中fill为填充空格的字符，默认不填则为空格，align可以是`<>=^`分别意味着向左靠齐，
向右靠齐，强制把正负号放在最左边，居中。

```python
In [1]: "{:a>#30.4f}".format(3.1415926)
Out[1]: 'aaaaaaaaaaaaaaaaaaaaaaaa3.1416'

In [2]: "{:a<#30.4f}".format(3.1415926)
Out[2]: '3.1416aaaaaaaaaaaaaaaaaaaaaaaa'

In [3]: "{:a^#30.4f}".format(3.1415926)
Out[3]: 'aaaaaaaaaaaa3.1416aaaaaaaaaaaa'

In [4]: "{:a=#30.4f}".format(3.1415926)
Out[4]: 'aaaaaaaaaaaaaaaaaaaaaaaa3.1416'

In [5]: "{:a=#30.4f}".format(-3.1415926)
Out[5]: '-aaaaaaaaaaaaaaaaaaaaaaa3.1416'
```

逗号代表是否把整数每三位插一个逗号表示：

```python
In [8]: "{:a=#30,.4f}".format(-300987.1415926)
Out[8]: '-aaaaaaaaaaaaaaaaa300,987.1416'
```

最后一位type可以是：

    - `b` 二进制
    - `c` 字符
    - `d` 十进制
    - `o` 八进制
    - `x` 小写十六进制
    - `X` 大写十六进制
    - `n` 数字，和 `d` 一致不过会根据locale不同而改变

最后如果想在这里面表示双括号自己怎么办？

```python
In [1]: "{{}}".format()
Out[1]: '{}'

In [2]: "{{".format()
Out[2]: '{'
```
