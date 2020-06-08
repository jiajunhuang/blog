# 适配器模式

我买过一个港版的手机，香港的插座和大陆不一样，电压也不一样，因此为了充电，我还买了一个转换头，这个转换头，就是起的
适配器的使用。

原本，大陆插座是两个孔的；而香港的是三个孔，也就是说我手机是没有办法充电的。但是通过转换头之后，转换头插入大陆插座，
而它自身提供一个三孔插座可以让香港的充电头插入，完成了适配的工作。

编程中也有类似的操作，应用的最多的莫过于代码/接口升级，而又需要保证老接口的兼容性，这个时候，为了让老接口继续工作，
我们需要提供一个中间层，让老接口对外的接口不变，但实际上代码却调用了新代码。

举个例子，假设我们有一个老项目，他提供如下接口：

```go
package main

import (
	"fmt"
)

func printName(firstName, secondName string) {
	fmt.Printf("firstName is %s, secondName is %s\n", firstName, secondName)
}

func main() {
	printName("Jiajun", "Huang")
}
```

看其中的 `printName` 函数的签名 `(firstName, secondName string)`，打印结果就是：

```bash
$ go build && ./test 
firstName is Jiajun, secondName is Huang
```

但是不巧的是，这个项目现在被山西煤老板1一个亿收购了，煤老板是中国人，我们和外国人的名字不同之处，在于他们的名字在前，
姓氏在后，我们是姓氏在前，名字在后，所以老板使用这套系统的时候很不开心：

```bash
$ go build && ./test 
firstName is 三, secondName is 张
```

怎么叫三张呢？应该叫张三才对！因此我们需要把接口换一下，伟大的架构师决定重构！毕竟老板特别生气，但是老系统还有很多
老外的资料在跑，虽然所有中国人都用中国人的输出方式，但是还是希望不要把老外的也统一过来，要不然客户又要生气了。

所以架构师决定，用一个新的函数来替代以前的这个 `printName`，并且能够实现，中国人用中国人的方式打印，外国人用外国人的，
并且要提供扩展性：

```go
package main

import (
	"fmt"
)

func isChinese(string) bool {
	// 略
	return true
}

func isEnglish(string) bool {
	// 略
	return true
}

func printEnglishName(firstName, secondName string) {
	fmt.Printf("firstName is %s, secondName is %s\n", firstName, secondName)
}

func printChineseName(familyName, name string) {
	fmt.Printf("姓: %s，名: %s", familyName, name)
}

func printName(familyName, name string) {
	if isChinese(familyName) {
		printEnglishName(name, familyName)
	} else if isEnglish(familyName) {
		printChineseName(familyName, name)
	} else {
		fmt.Println("暂不支持/Not support yet")
	}
}

func main() {
	// 略
}
```

如此，便成功的实现老板的需求，并且可以遇见，下次项目再被法国人一亿欧元买回去的时候，他们仍然可以愉快的使用这个系统。

这就是适配器的作用，对接多个端，如果你想不起什么是适配器的话，想象一下我的港版手机就知道了 :)
