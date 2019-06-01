# 从Chrome切换到Firefox

Chrome一家独大之后，就喜欢开始搞幺蛾子，商业公司毕竟是商业公司。再加上Chrome实在是越来越卡，让我的老本子打开都要白屏
加载好几秒钟，试用了一下Firefox Quantum，发现Firefox已经不是当年的Firefox，响应非常快，常见的插件也都有，于是就直接
切换到Firefox了。

- 首先安装好firefox，注意是安装firefox，不是火狐。火狐浏览器是谋智中国打包的中国版，看起来也不是什么好幺蛾子。一般Linux
直接从官方源里安装就好了：

```bash
$ sudo pacman -S firefox
```

- 然后从Chrome导入书签，在Firefox的菜单里选择 `Bookmarks`，然后选择 `Show All Bookmarks`，或者直接按 `Ctrl-Shift-O`，
打开书签管理器，然后选择 `Import And Backup`，在下拉框中选择 `Import Data From Another Browser`，点击之后下一步下一步即可。
导入之后Firefox会把这些书签和历史导入到文件夹 `From Chrome`，自己整理整理就好了。

- 比较麻烦的问题在于密码的导出。我之前是使用了Chrome自带的密码管理器，但是Google并没有提供方法导出，因此我只能挑几个常用
的网站把密码让Firefox记住。说到这个我就想起了，浏览器厂商这一招真是精明，因为很有可能你就不会卸载Chrome转到别的浏览器了。
说起来我应该使用一个第三方密码管理器，但是想想还是算了，第一是，很多年才会切换一次浏览器；第二是，把密码交给别人我也不是
特别放心，自己搭其实也不安全，Mozilla应该相对安全；第三是，实际上很多密码根本不常用，就当清洗一遍账号了吧。

现在我打开浏览器，又是飞速了 :)

---

参考资料：

- https://support.mozilla.org/en-US/kb/import-bookmarks-google-chrome
