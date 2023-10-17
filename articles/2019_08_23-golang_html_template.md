# Golang的template(模板引擎)简明教程

模板语言，在前后端分离的时代，大概可以归类到上古时代的技术了。不过前后端分离并不是银弹(而且也只是
把模板从后端移到前端而已)，它也有很多问题：

    - SEO
    - 需维护两套程序
    - ...

模板语言仍然有它的用处，例如我的博客部署在一台512M的小机器上，起两套程序不仅写起来麻烦，而且内存占用
大，部署也不方便。

Go语言的模板与大多数语言一样，是用 `{{` 和 `}}` 来做标识，`{{  }}` 里可以是表达式，也可以是变量，不过
Go语言的模板比较简陋，没办法像Python中 `jinja2` 那样使用继承的写法，而只能是用拼接的写法。

举个例子，看看这个博客的模板(https://github.com/jiajunhuang/blog/tree/master/templates)：

```html
{{ block "header.html" . }}{{ end }}

{{ .content }}

{{ block "footer.html" . }}{{ end }}
```

这是渲染这篇博客的模板，很明显，它由三部分组成：

- `header.html` 这是顶栏
- `{{ .content }}` 这是内容
- `footer.html` 这是右侧边栏和底部

它的写法，就是把一个完整的HTML页面拆成几个部分，然后进行拼装。有点像这样的意思：

```go
function renderHeader() {
    renderNavbar()
    renderPartOfBody()
}

function renderContent() {
    renderContentBody()
}

function renderFooter() {
    renderPartOfBody()
    renderSidebar()
    renderBottom()
}

renderHeader()
renderContent()
renderFooter()
```

接下来我们来看看具体的语法：

- 使用 `{{ define "layout" }}{{ end }}` 来定义一个块，这个块的名字就是 "layout"，如果不使用 `define`，那么名字就是文件名
- 使用 `{{ block "template filename" . }} {{ end }}` 来调用一个模板，就像上面的例子中，调用一个函数那样，其中 `.` 也可以是变量名等等，就是引用变量的上下文，如果我们传入 `.`，那么子模板里可以访问的变量就和当前可以访问的上下文一样
- `{{ range $i := .items }} {{ end }}` 相当于Go语言里的 `for i := range items`
- `{{ if .items }}` 相当于Go语言里的 `if items`，完整的是 `{{if pipeline}} T1 {{else if pipeline}} T0 {{end}}`
- `{{ .variable }}` 渲染 `variable` 的值
- `{{/* a comment */}}` 是注释
- `{{- /* a comment with white space trimmed from preceding and following text */ -}}` 可以跨行的注释

当然，还有个比较坑的地方在于，Go会自动把输出进行转义，渲染的时候如果想要不转义，就使用 `template.HTML("blablabla")`，这里的 `template` 就是导入的包 `html/template`。

例如，这个博客输出RSS的Controller是这样的：

```go
// RSSHandler RSS
func RSSHandler(c *gin.Context) {
	c.Header("Content-Type", "application/xml")
	c.HTML(
		http.StatusOK, "rss.html", gin.H{
			"rssHeader": template.HTML(`<?xml version="1.0" encoding="UTF-8"?>`),
			"articles":  articles,
		},
	)
}
```

模板是这样的：

```xml
{{ .rssHeader }}
<rss version="2.0">
    <channel>
        <title>Jiajun的技术笔记</title>
        <link>https://jiajunhuang.com</link>
        <description>Jiajun的技术笔记</description>

        {{ range $article := .articles }}
            <item>
                <title>{{ .Title }}</title>
                <link>https://jiajunhuang.com/{{ .DirName }}/{{ .Filename }}.html</link>
                <description>{{ .Title }}</description>
            </item>
        {{ end }}

    </channel>
</rss>
```

这样才成功的避免了XML的第一个尖括号 `<` 被转义。

---

参考资料：

- https://golang.org/pkg/text/template/
