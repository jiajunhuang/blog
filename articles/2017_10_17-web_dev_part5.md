# Web开发系列(五)：form, json, xml

## 旧石器时代：表单

很久以前的web开发是怎么一个工作模式呢？那时候的前端不像现在的前端，那时候的前端基本上就是熟练CSS和HTML就可以了，JS要求其实
也不高，开发流程是这样的，前端先写好静态页面，然后丢给后端来套模板，那么问题来了，如果发现页面上的样式要改怎么办？

- 方法一：前端到后端的电脑上来改
- 方法二：前端改好之后，后端重新套模板，然后各种渲染

如果前端水平高，一次性写出无bug代码，那还好说，如果前端水平不行，哈哈哈哈哈哈哈哈，就等着哭吧。

那怎么处理用户数据呢？以前流行jQuery，一般是在button上绑定事件，用户提交的时候，jQuery用AJAX向后端发送一个form表单，当然
发送之前前端可能会对数据进行校验，但是无所谓，后端还是要校验，后端一般拿到数据之后，除了校验还要做类型转换。为什么呢？

参考 [这篇博客](https://jiajunhuang.com/articles/2017_10_14-web_dev_part2.md.html) ，我们的form表单传给后端的时候是没有
类型信息的，全都是字符串，所以我们要把取出来的参数做强制转型，此外还要检验参数是否存在等。

所以代码长得有点像这样：

```python
from flask import Flask, request, abort
app = Flask(__name__)


@app.route("/")
def hello():
    user_id = request.args.get("user_id")
    passwd = request.args.get("passwd")

    if not (user_id and passwd):
        abort(401)

    user_id = int(user_id)
    # 略略略


if __name__ == "__main__":
    app.run()
```

因此就有了一个东西，叫做表单校验工具，是啊，谁乐意老是写这么多重复的校验代码，说实话，之前没用过第三方的表单校验，都是
公司内部造的轮子，但是Python界还是有一个非常出名的表单校验工具：[WTForms](https://wtforms.readthedocs.io/en/latest/)，
不过说实话，我没用过，不太好作评价。但是表单校验工具长得都差不多：

- 定义好这个参数是否必须
- 定义好这个参数的类型
- 定义好这个参数的要求，例如范围，长度，可否为空等

然后根据定义好的这些用响应的validators进行校验，如果失败就返回错误或者抛出对应的异常。此外，表单校验工具一般都会把form表单
的数据根据所定义好的类型转换成对应的数据。

## 新的曙光，前后端分离

一直一直一直到了近些年，大前端时代，前后端分离这种架构慢慢流行起来。流行的原因是什么呢？因为移动互联网，没人愿意给前端渲染
一遍模板，又给app开发一套代码，麻烦得要命，所以我们只提供API，前端自己渲染数据。近年来更是有RESTFul啊这些概念，以前写接口
那都是xml来作为传输格式的。

JSON，全名是JavaScript Object Notation，就是把js里的集中基本数据类型，拿出来表示数据，其实这货在Python里可以很好的映射，
例如JSON的array对应Python的List，object对应dict，blablabla。一个简单的JSON长得像这样：

```json
{
    "glossary": {
        "title": "example glossary",
		"GlossDiv": {
            "title": "S",
			"GlossList": {
                "GlossEntry": {
                    "ID": "SGML",
					"SortAs": "SGML",
					"GlossTerm": "Standard Generalized Markup Language",
					"Acronym": "SGML",
					"Abbrev": "ISO 8879:1986",
					"GlossDef": {
                        "para": "A meta-markup language, used to create markup languages such as DocBook.",
						"GlossSeeAlso": ["GML", "XML"]
                    },
					"GlossSee": "markup"
                }
            }
        }
    }
}
```

像Python这样的动态类型语言写这些可爽了，直接 `json.loads` 然后开干，但是像Go这样的强类型静态语言就不行了，要先定义好
JSON的结构，然后慢慢解析，虽然Go也可以强行解析成一个 `interface{}`，但是interface走天下也不是太好。

## xml

如果你调用过微信的接口你就会发现，微信接口竟然还用xml。。。xml也是一种数据表现形式，长得跟html有点像，各种tag和嵌套，用来
表示复杂的关系时还可以用用，不过这年头用xml的确实不多了。给个例子你就知道为什么没人愿意用了，上面的JSON的例子，用xml表示
是这样：

```xml
<!DOCTYPE glossary PUBLIC "-//OASIS//DTD DocBook V3.1//EN">
 <glossary><title>example glossary</title>
  <GlossDiv><title>S</title>
   <GlossList>
    <GlossEntry ID="SGML" SortAs="SGML">
     <GlossTerm>Standard Generalized Markup Language</GlossTerm>
     <Acronym>SGML</Acronym>
     <Abbrev>ISO 8879:1986</Abbrev>
     <GlossDef>
      <para>A meta-markup language, used to create markup
languages such as DocBook.</para>
      <GlossSeeAlso OtherTerm="GML">
      <GlossSeeAlso OtherTerm="XML">
     </GlossDef>
     <GlossSee OtherTerm="markup">
    </GlossEntry>
   </GlossList>
  </GlossDiv>
 </glossary>
```

emm...看到都不想碰

## 对比：模板渲染和前后端分离

最后我们来对比一下模板渲染和前后端分离。模板渲染有什么好处呢？一般来说，模板渲染的性能会高一些，因为模板渲染的原理是，
渲染器把模板文件“编译”成对应的程序，这个程序里的代码就是模板里想表述的代码，然后各种字符串拼接，所以模板实际上最后变成了
程序代码。其实你要是看到很多Java程序员写模板，他们还有很多for循环里put一个标签之类的，那就是在手写模板。

而前后端分离性能理论上应该要稍差，因为首先html下发到浏览器之后，浏览器才去获取首页的数据，然后渲染，当然，现在也有用
nodejs在服务器端请求api之后渲染，其实也差不多，总之呢，前后端分离之后，前端和后端请求数据这一步是省不掉的。

既然前后端分离这么灵活，模板渲染有啥好处？第一，容易被爬虫抓到，第二，服务端可以丢到缓存里，这样性能更高。而前后端分离
则恰恰相反，爬虫不一定能抓到，很大程度上要依靠浏览器进行缓存。但是呢，前后端分离大大的解耦了前端程序员和后端程序员，
只要约定好了API，各自更改各自的程序，互不影响。

现在我基本不写模板渲染了，哈哈哈。

讲完，收工 :)
