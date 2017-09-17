# 再读vim help：vim小技巧

- `ZZ` 相当于 `<Esc>:wq`

- Normal模式下`a`相当于反向的`i`，i把字符插入到光标前，a插到光标后

- movement

    - `w` 移动到下一个单词的首字母
    - `b` 移动到上一个单词的首字母
    - `e` 移动到下一个单词的尾字母
    - `ge` 移动到上一个单词的首字母
    - `f` `fx` 移动到行内的第一个x

    > 不过有了 `vim-easymotion` 这个插件，这几个快捷键用的就少了

- `ctrl-g` 会在底部显示出文件的状态，包括文件名。这个拷贝文件名的时候就方便了

- scrolling

    - `ctrl-u` 向上滚动半屏
    - `ctrl-d` 向下滚动半屏
    - `ctrl-e` 向上滚动一行
    - `ctrl-y` 向下滚动一行
    - `ctrl-f` 向下滚动整屏
    - `ctrl-b` 向上滚动整屏
    - `zz` 把当前行放到屏幕中央
    - `zt` 把当前行放到屏幕顶部
    - `zb` 把当前行放到屏幕底部

- search

按 `/` 后接内容开始搜索，其中 `.*[]^%?~$` 需要加 `\` 转意，按 `?` 往前搜。
不过我装了 `Plug 'othree/eregex.vim'` 所以可以直接用perl的正则表达式。

- marks

    - `两个反向单引号(`，因为markdown中不好转意所以中文描述)` 记住了上一个跳转的位置
    - `ctrl-o` 跳转到一个更老的位置
    - `ctrl-i` 跳转到一个更新的位置

几个特殊的内置mark：

    - `'` 跳转前的位置
    - `"` 上次编辑文件的位置
    - `[` 上次更改开始的位置
    - `]` 上次更改结束的位置

- changing text

    - `c` 代表 change，所以 `cw`就是删掉当前光标到单词尾，并且处于插入模式。
    - `dd` 删除整行，对应的 `cc` 就是改变整行
    - `daw` 代表 delete a word， `aw` 会把整个单词块选上，这在光标处于单词中间
    但是要删除整个单词的时候很有用
    - `cis` 代表 change inner sentence，会删除整句话，`cas`类似，change a sentence

内置的快捷键：

    - `x` 相当于 `dl`
    - `X` 相当于 `dh`
    - `D` 相当于 `d$`
    - `C` 相当于 `c$`
    - `s` 相当于 `cl`，不过这个快捷键已经被我重新binding成了 `easymotion`的快捷键
    - `S` 相当于 `cc`
    - `.` 点号可以重复上一次的命令，加上 `vim-repeat` 这个插件就可以重复更多了，
    详见：https://github.com/tpope/vim-repeat

- visual mode

    - `v` 按字符移动
    - `V` 按行移动
    - `ctrl-v` 按选中的长方形移动，按 `o` 或者 `O` 对向移动

- plugin

我最开始学vim的时候，装插件是要靠手动一个一个解压到 `~/.vim` 下面的，好古老。
不过现在大把的vim插件管理器，对新手友好多了 :) 我用的是 `vim-plug`

- 窗口

    - `:split` 横向分窗口
    - `:vsplit` 竖着分
    - `:only` 仅保留当前窗口
    - `:close` 关闭窗口
    - `:new` 打开横的新的空的窗口
    - `:vnew` 竖着打开新的
    - `ctrl-w+`可以增大，把`+`换成`-`可以缩小，不过我几乎不用这两命令，屏幕大
    才是正道！
    - `:vertical` 后接 `new`, `help` 等竖屏拆分窗口，并且执行相应命令
    - `ctrl-w` 加上 `hjkl` 移动到相应窗口，不过我帮定了快捷键：

    ```vimrc
    nnoremap <C-h> <C-w>h
    nnoremap <C-j> <C-w>j
    nnoremap <C-k> <C-w>k
    nnoremap <C-l> <C-w>l
    ```
    - `ctrl-w` 加上 `HJKL` 调整窗口的布局，例如把三个横着的窗口中的一个摆到
    最左边，变成竖的就用 `ctrl-w-H`

    - tab 这个不常用，个人更喜欢用buffer，配合上 `ctrlp` 倍儿爽

- 宏录制

这个，还是得仔细读读manual啊，打开vim，输入 `:help usr_10` 然后回车吧

- 替换

`:%s/Professor/Teacher/c` 最后的c会一个一个让你确认，如果换成g就直接全局替换
替换可以指定区域，指定marks之间，指定某个单词的前面第几行或者后面第几行。

- `>>` `<<`

左移右移，配合visual mode，`.` 使用效果更佳

- `g ctrl-g` 会列出全文由多少个单词
