# Lua Manual 阅读笔记

- 关键字:

```
and break do else elseif end false for function if in local
nil not or repeat return then true until while
```

- Lua 中约定: 以下划线开头, 后接大写字母的为保留的全局变量. 例如 `_VERSION`

```lua
> print(_VERSION)
Lua 5.1
```

- tokens:

```
+ - * / % ^ # == ~= <= >= < > =
(  ) {  } [  ] ; : , . .. ...
```

- 字符串可以以单引号或者双引号包围,中间可以包含c风格的转义字符串如 `\t` 等.

    - `\a` 响铃
    - `\b` 空格
    - `\f` form feed
    - `\n` 新的一行
    - `\r` 回车
    - `\t` 横向tab
    - `\v` 纵向tab
    - `\\` `\`本身
    - `\"` 双引号本身
    - `\'` 单引号本身
    - \ 加 回车代表字符串里新的一行

- 不转义字符可以用 `[[` 的两个 `[` 中间加上n个 `=` 来表示, 0个就是0级, n个就是
n级. 如果开括号后直接跟一个换行符, 则忽略这个换行符:

```lua
> print('hello\n123')
hello
123
> print([[hello
>> 123]])
hello
123
> print([[
>> hello
>> 123]])
hello
123
```

- 字符串也可以用 `\` 加上其ascii值来表示, 十进制,最多三位. 如 `\97`

- `--` 开头的是注释, 后面可以接双括号, 不过我还是喜欢每行 `--` 因为vim帮我干了.

- lua中的值有八种类型:

    - nil
    - boolean
    - number
    - string
    - function
    - userdata 用来存储裸的内存块,只能通过c API来操作,不能在lua中直接操作
    - thread 协程中的"线程"的概念,相当于Golang中的 `G`
    - table 是lua中的一个数据结构,既可以当数组用,又可以当哈希表用,还有 `a.name` 语法糖,相当于 `a["name"]`

`type` 函数返回数据的类型:

```lua
> print(type(1))
number
> print(type("hello"))
string
```

- lua会在操作字符串和数字的时候自动进行类型转换,例如:

```lua
> print(1 + "2")
3
```

> 坑啊...和js一样...我很讨厌这种特性.

- Lua有三种变量

    - 全局变量
    - 本地变量
    - table项(table fields)

以 `var ::= Name` 的形式赋值给变量, 默认全局变量, 除非加了 `local` 关键字, 
本地变量可以被其作用域内的函数访问到. 第一次赋值前, 变量的默认值为 `nil`.
全局变量存储在 `_env` 里.

> 默认全局变量,又是一个坑.

- 赋值, 赋值的时候, 如果右边比左边长, 则多余的值会被丢掉. 相反, 则用nil填充.
赋值语句首先计算出所有变量,然后才进行赋值,例如:

```lua
i = 3
i, a[i] = i+1, 20
```
执行完之后, `a[3]` 的值是20, `a[4]` 不受影响.

- `false` 和 `nil` 是false,其他的都是true, 所以0和空字符串也是true.

- for有两种形式, 一种像c:

```lua
local i = 0

for i = 0, 10, 2 do
    print(i)
end
```

```bash
$ luajit test.lua 
0
2
4
6
8
10
```

只能用来做算术循环.

另一种像 Python:

```lua
local names = {"hello", "world"}

for k, v in ipairs(names) do
    print(k, v)
end
```

```bash
$ luajit test.lua 
1	hello
2	world
```

- 函数和变长参数都可以产生多值. 如果作为表达式, 那么结果会被忽略, 例如函数
调用 `f()` 那么其结果会被忽略. 如果作为表达式的非最后一个元素, 那么多值结果会被
省略,只剩下第一个, 如果作为最后一个元素,那么就会保留所有结果.

```lua
function f()
    return 1, 2, 3
end

local a
local b
local c
f()
print(a, b, c)

a, b, c = f(), 4, 5
print(a, b, c)

a, b, c = f()
print(a, b, c)

a, b, c = 9, 8, f()
print(a, b, c)
```

```bash
$ luajit t.lua 
nil	nil	nil
1	4	5
1	2	3
9	8	1
```

- `== ~= < > <= >=` 总是产生布尔值. 数字和字符串比较值,对象(tables, userdata, threads,
functions)比较引用值.

- `and, or, not`其中 `and` 和 `or` 支持条件短路. 所以如果第一个参数是 `false` 或者
`nil` 就返回第一个参数,否则返回第二个参数.

- `..` 连接字符串

```lua
> print("hello" .. 1)
hello1
```

- length. 通过 `#` 号取出长度, 对字符串来说是bytes的个数,对table来说,是最后一个
非空值的index. 例如 `a[1], a[2], a[3]` 为1, 而 `a[4]` 为nil, `a[5]` 为2, 那么
长度为3.

```lua
local a = {1, 2, 3, nil, 4}

print(#a)
```

```bash
$ luajit t.lua 
3
```

> 坑啊,table支持空洞额...
