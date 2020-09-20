# Linux常用命令(四)：xargs

xargs命令可以从标准输入读取参数，然后组建成新的命令去执行，举个例子：

```bash
$ echo 'this that' | xargs
this that

```

这里就是把 echo 打印到标准输出的内容，通过管道传到 xargs 的标准输入，因此 xargs 拼接出命令 `xargs echo this that`，
那么为啥会有一个echo出来呢？因为xargs的格式是 `xargs <选项> <命令> <默认给的参数>`，如果默认没有填写命令，那么就是echo。
继续照着上面的例子讲解，正常情况下，xargs构建命令，是直接把标准输入读到的命令放在最后，比如：

```bash
$ echo this that whoops well | xargs echo duang
duang this that whoops well

```

接下来我们来看看xargs常见的命令行选项：

- `-d` 设置分隔符，默认情况下，xargs以换行符和空格切割命令，然后拼接成命令
- `-I` 设置替换字符串，`-I` 后面的第一个字符串，会把再往后的字符串替换，比如：

```bash
$ echo this that well | xargs -I % echo % end
this that well end

$ echo this that well | xargs -I % sh -c 'echo %; echo %'
this that well
this that well

```

可见，正常情况下，参数会被放在xargs最后，但是通过 `-I` 可以把从标准输入读进来的参数放在其它位置。

- `-p` 参数设置之后，会变成交互式的，xargs会对每一次执行，都询问是否确认执行：

```bash
$ echo this that well | xargs -p touch
touch this that well ?...y
$ echo this that well | xargs ls
that  this  well
$ echo this that well | xargs rm
$ echo this that well | xargs ls
ls: cannot access 'this': No such file or directory
ls: cannot access 'that': No such file or directory
ls: cannot access 'well': No such file or directory
```

- `-P` 设置最大并发数，默认情况下是1，如果设置成了0，那么就是有多少就fork多少进程去执行。
- `-t` 打印将要执行的命令

## 总结

这一篇中我们看了xargs的用法及其常见命令行参数，还有一部分没有列出的，如果有兴趣的话，可以去查看manual。
