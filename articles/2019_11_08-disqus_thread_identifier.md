# disqus获取评论时忽略query string

我的博客使用disqus，但是默认情况下有一个问题，那就是如果有query string，那么评论只会在对应的query string下才会显示，
而我想要忽略query string，以去除query string的URL为唯一标识来加载评论。

方法就是修改js，原本在disqus代码上有这么一段注释：

```html
/**
*  RECOMMENDED CONFIGURATION VARIABLES: EDIT AND UNCOMMENT THE SECTION BELOW TO INSERT DYNAMIC VALUES FROM YOUR PLATFORM OR CMS.
 *  LEARN WHY DEFINING THESE VARIABLES IS IMPORTANT: https://disqus.com/admin/universalcode/#configuration-variables*/
/*
    var disqus_config = function () {
    this.page.url = PAGE_URL;  // Replace PAGE_URL with your page's canonical URL variable
    this.page.identifier = PAGE_IDENTIFIER; // Replace PAGE_IDENTIFIER with your page's unique identifier variable
    };
 */
```

告诉我们这是自定义变量，我们写入这么一段代码：

```js
var disqus_config = function () {
    var PAGE_IDENTIFIER = window.location.pathname.split(/[?#]/)[0];
    var PAGE_URL = "https://jiajunhuang.com" + PAGE_IDENTIFIER;

    this.page.url = PAGE_URL;  // Replace PAGE_URL with your page's canonical URL variable
    this.page.identifier = PAGE_IDENTIFIER; // Replace PAGE_IDENTIFIER with your page's unique identifier variable
};
```

作用就是忽略query string，当然，以前如果是在带query string的情况下评论的话，现在就不会显示出来了。

---

参考资料：

- [Disqus官方指南](https://help.disqus.com/en/articles/1717084-javascript-configuration-variables)
