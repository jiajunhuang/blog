# 网络乞讨之合并支付宝和微信的收款二维码

在博客的右边加了一个网络乞讨的二维码，但是支付宝和微信分别有两个二维码，要想个法子把它们合并一下。于是把支付宝和微信的收款
二维码解码了一下(可以找二维码解码工具，如果是iOS用户，直接用自带的相机扫描，就会弹出来是否打开)，发现支付宝的是一个链接，
但是微信的是一个自有链接。

扫码试了一下，发现打开链接的时候，可以根据 `UserAgent` 来判断，支付宝会带一个 `AlipayClient`，而微信会带一个 `MicroMessenger`，
因此，根据 `UserAgent` 然后跳转就好了。但是有一个问题是，微信解码出来的数据，如果直接跳转的话，会是一个空白页面，所以，
解决办法就是，把微信的付款二维码的图片留下，如果是微信，那么就把微信付款二维码展示给用户看，否则就跳到支付宝的收款页面：

```python
@app.route("/reward")
def reward():
    user_agent = request.user_agent.string
    if "MicroMessenger" in user_agent:
        return redirect(config.WECHAT_PAY_URL)
    else:
        return redirect(config.ALIPAY_URL)
```

然后我们再生成一个二维码，内容是 `https://jiajunhuang.com/reward`，这样子别人扫码就会跳转到这个URL，然后我们根据 `UserAgent`
进行跳转，大功告成！
