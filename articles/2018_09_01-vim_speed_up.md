# Vim打开很慢，怎么找出最慢的插件？怎么解决？

很久很久以前，YouCompleteMe还只是一般般的卡，现在，YCM（简称，后同）简直是巨卡，巨卡。仔细分析一下自己用Vim的地方：

- 写博客
- 写代码
- 编辑一些临时文件


那其实就很明显了，只有在写代码的时候才会用上YCM，而且只有Go和Python的时候用得上。得益于 `vim-plug` 的强大功能，支持
惰性加载。于是我就改成了这样：

```vim
Plug 'Shougo/neco-syntax'
Plug 'hynek/vim-python-pep8-indent', { 'for': 'python' }
Plug 'itchyny/vim-haskell-indent', { 'for': 'haskell' }
Plug 'stephpy/vim-yaml', { 'for': 'yaml' }
Plug 'uber/prototool', { 'rtp':'vim/prototool', 'for': 'proto' }
Plug 'Valloric/YouCompleteMe', { 'for': ['python', 'go'] }
Plug 'plasticboy/vim-markdown' | Plug 'godlygeek/tabular', { 'for': 'markdown' }
Plug 'vim-jp/vim-go-extra', { 'for': 'go' }
```

这样，这些插件就只有在对应的 filetype 被打开的时候才会加载。

下面分享一些找出最慢的插件的方式：

## vim --startuptime

```bash
$ vim --startuptime vim.log
```

会记录下每一步，所花费的时间。Neovim也支持这个选项。

## 二分查找

每次注释一半的插件，用 `log(N)` 的次数就可以找出来谁最慢啦！
