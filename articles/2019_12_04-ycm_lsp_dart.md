# Vim YouCompleteMe使用LSP(以dart为例)

YCM(YouCompleteMe)是Vim下大名鼎鼎的补全插件，现在YCM也支持LSP了，因此可以使用YCM来补全支持LSP的代码，再加上YCM自带的
语义补全，写起代码来如有神助。

其实配置很简单，在 `vimrc` 中添加如下配置(以dart为例)：

```vim
let g:ycm_language_server = [
  \   {
  \     'name': 'dart',
  \     'cmdline': ['dart', '/opt/dart-sdk/bin/snapshots/analysis_server.dart.snapshot', '--lsp'],
  \     'filetypes': [ 'dart' ],
  \   },
  \ ]
```

然后就可以进行补全了 :-)
