# Go设计模式：简单工厂模式

原始的工厂模式太过于繁冗，我几乎不用。但是简单工厂模式是经常用到的，比如：

```python
def get_duck(color):
    if color == "yello":
        return YelloDuck()
    elif color == "blue":
        return BlueDuck()
    else:
        return DefaultDuck()
```

这就是传说中的简单工厂模式，不过，在Go里我们一般不会返回多个struct，而是返回一个interface，而具体实现，都满足这个interface，
比如，如果我们做一个短信服务，肯定要对接多个短信渠道，比如阿里云、腾讯云，那么就可以这样：

```go
type SMSSender interface {
	Send(content string, receivers []string) error
}

type AliyunSMS struct{}

func (a *AliyunSMS) Send(content string, receivers []string) error {
	// pass
}

type TencentSMS struct{}

func (t *TencentSMS) Send(content string, receivers []string) error {
	// pass
}

// 简单工厂在这里
func getSMSSender(channel string) SMSSender {
	if channel == "aliyun" {
		return &AliyunSMS{}
	} else if channel == "tencent" {
		return &TencentSMS{}
	} else {
		// 略
	}
}
```

但是这只是一种用法，还有一种写法上的变种：

```go
type SMSSender interface {
	Send(content string, receivers []string) error
}

type AliyunSMS struct{}

func (a *AliyunSMS) Send(content string, receivers []string) error {
	return nil
}

type TencentSMS struct{}

func (t *TencentSMS) Send(content string, receivers []string) error {
	return nil
}

var senderMapper = map[string]SMSSender{
	"aliyun":  &AliyunSMS{},
	"tencent": &TencentSMS{},
}

// 简单工厂在这里
func getSMSSender(channel string) SMSSender {
	sender, exist := senderMapper[channel]
	if !exist {
		// 略
	}

	return sender
}
```

当然，这里的区别在于，使用一个mapper之后，就节省了一堆的 `if...else...`，不过缺点就是并非每次都
实例化了对应的sender，当然也是可以通过反射做到的，不过不推荐，所以实际上用哪种
写法，还是要结合实际情况来看。

工厂模式还有一种，是抽象工厂模式，这个似乎不太常用，至少我没有在代码里遇到过，也许这个模式就是适用于Java这种擅长
把小项目做成 "大项目" 的语言的吧。
