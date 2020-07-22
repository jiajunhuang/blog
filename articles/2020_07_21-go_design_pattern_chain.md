# Go设计模式: 责任链模式

今天我们来介绍责任链模式，从名字可以看出来，它应当是一系列的操作。的确如此，看看维基百科的定义：

> 责任链模式在面向对象程式设计里是一种软件设计模式，它包含了一些命令对象和一系列的处理对象。
> 每一个处理对象决定它能处理哪些命令对象，它也知道如何将它不能处理的命令对象传递给该链中的下一个处理对象。

所以其实它就像一个工厂流水线，我们把原料丢进去，每一层处理一部分，如果处理完成，就退出，没有完成，就进入下一个环节。
我们来看看打印日志的一个例子，打印日志通常是按照优先级，当设定一个阈值之后，高于这个等级的日志才会打印，否则不打印，
通常优先级从高到低依次是：

- FATAL/PANIC
- ERROR
- WARNING
- INFO
- DEBUG

当然，最简单的实现方式是这样：

```python
DEBUG = 0
INFO = 10
WARNING = 100
ERROR = 1000
FATAL = 10000


class Logger:
    def __init__(self, default_level=INFO):
        self.__default_level = default_level

    def set_level(self, level):
        self.__default_level = level

    def __print_log(self, level, msg):
        if level >= self.__default_level:
            print(msg)

    def debug(self, msg):
        self.__print_log(DEBUG, msg)

    def info(self, msg):
        self.__print_log(INFO, msg)

    def warning(self, msg):
        self.__print_log(WARNING, msg)

    def error(self, msg):
        self.__print_log(ERROR, msg)

    def fatal(self, msg):
        self.__print_log(FATAL, msg)


if __name__ == "__main__":
    logger = Logger()
    logger.debug("[DEBUG] should not print out")
    logger.info("[INFO] should print out")
    logger.warning("[WARNING] should print out")
```

不过我们在此只是为了举个例子，说明什么叫做责任链模式，来看看如果使用责任链模式应当如何实现：

```python
DEBUG = 0
INFO = 10
WARNING = 100
ERROR = 1000
FATAL = 10000


class Logger:
    def __init__(self):
        self.next = None

    def set_next(self, next):
        self.next = next
        return next

    def print(self, level, msg):
        raise NotImplementedError("should rewrite this method in Logger")

class DebugLogger(Logger):
    def print(self, level, msg):
        if level == DEBUG:
            print("DEBUG", msg)
            return

        if self.next:
            print("not handled by DebugLogger")
            self.next.print(level, msg)

class InfoLogger(Logger):
    def print(self, level, msg):
        if level == INFO:
            print("INFO", msg)
            return

        if self.next:
            print("not handled by InfoLogger")
            self.next.print(level, msg)

class WarningLogger(Logger):
    def print(self, level, msg):
        if level == WARNING:
            print("WARNING", msg)
            return

        if self.next:
            print("not handled by WarningLogger")
            self.next.print(level, msg)

class ErrorLogger(Logger):
    def print(self, level, msg):
        if level == ERROR:
            print("ERROR", msg)
            return

        if self.next:
            print("not handled by ErrorLogger")
            self.next.print(level, msg)

class FatalLogger(Logger):
    def print(self, level, msg):
        if level == FATAL:
            print("FATAL", msg)
            return

        if self.next:
            print("not handled by FatalLogger")
            self.next.print(level, msg)


if __name__ == "__main__":
    logger = FatalLogger()
    logger.set_next(ErrorLogger()).set_next(WarningLogger()).set_next(InfoLogger()).set_next(DebugLogger())

    logger.print(INFO, "info")
    logger.print(DEBUG, "info")
```

我们可以看到责任链模式的几个核心特点，就是：

- 分层处理
- 当层没处理完成的，放下一层继续处理
- 可以把各个处理函数(handler)串成一个链

不过，这样实现未免也太麻烦了，我们来优化一下，我们不使用面向对象的模式，而是使用一个外部的函数来辅助逻辑判断：

```python
DEBUG = 0
INFO = 10
WARNING = 100
ERROR = 1000
FATAL = 10000


class BaseLogger:
    def print_log(self, msg):
        raise NotImplementedError("should rewrite this method in Logger")


class DebugLogger(BaseLogger):
    def print_log(self, msg):
        print("DEBUG", msg)


class InfoLogger(BaseLogger):
    def print_log(self, msg):
        print("INFO", msg)


class ErrorLogger(BaseLogger):
    def print_log(self, msg):
        print("ERROR", msg)


class WarningLogger(BaseLogger):
    def print_log(self, msg):
        print("WARNING", msg)


class FatalLogger(BaseLogger):
    def print_log(self, msg):
        print("FATAL", msg)


class Logger:
    def __init__(self):
        self.__loggers = [FatalLogger(), ErrorLogger(), WarningLogger(), InfoLogger(), DebugLogger()]
        self.__levels = [FATAL, ERROR, WARNING, INFO, DEBUG]

    def print_log(self, level, msg):
        for i, lvl in enumerate(self.__levels):
            print(f"trying lvl {lvl}")
            if level == lvl:
                logger = self.__loggers[i]
                logger.print_log(msg)
                return



if __name__ == "__main__":
    Logger().print_log(ERROR, "hello world")
    Logger().print_log(INFO, "hello world")
```

这样子我们就把优先级，或者说分层的逻辑集中在 `Logger.print_log` 里了。

我们来看看Go语言里该要如何实现：

```go
package main

import (
	"fmt"
)

const (
	Debug = iota
	Info
	Warning
	Error
	Fatal
)

type Logger interface {
	PrintLog(level int, msg string)
}

var (
	_ Logger = &DebugLogger{}
	_ Logger = &InfoLogger{}
	_ Logger = &ErrorLogger{}
)

type BaseLogger struct {
	next Logger
}

func (b *BaseLogger) SetNext(logger Logger) {
	b.next = logger
}

type DebugLogger struct {
	BaseLogger
}

func (d *DebugLogger) PrintLog(level int, msg string) {
	if level == Debug {
		fmt.Printf("[DEBUG] %s\n", msg)
	} else {
		fmt.Printf("ignore [DEBUG]\n")
		d.next.PrintLog(level, msg)
	}
}

type InfoLogger struct {
	BaseLogger
}

func (d *InfoLogger) PrintLog(level int, msg string) {
	if level == Info {
		fmt.Printf("[INFO] %s\n", msg)
	} else {
		fmt.Printf("ignore [INFO]\n")
		d.next.PrintLog(level, msg)
	}
}

type ErrorLogger struct {
	BaseLogger
}

func (e *ErrorLogger) PrintLog(level int, msg string) {
	if level == Error {
		fmt.Printf("[ERROR] %s\n", msg)
	} else {
		fmt.Printf("ignore [ERROR]\n")
		e.next.PrintLog(level, msg)
	}
}

func main() {
	errorLogger := &ErrorLogger{}
	infoLogger := &InfoLogger{}
	debugLogger := &DebugLogger{}

	infoLogger.SetNext(debugLogger)
	errorLogger.SetNext(infoLogger)

	errorLogger.PrintLog(Info, "info")
	errorLogger.PrintLog(Debug, "debug")
}
```

搞定！我们来继续回忆一下常见的责任链模式的使用之处，比如Web开发中，对于一个请求，我们需要加一些中间件对不对，例如接到
请求需要判断是否是正常请求，是否登录了，参数是否合法等，这个时候我们就会用上中间件，而中间件就可以用责任链模式来实现，
把handler串成一个串，依次处理，如果已处理完成，那么提前终止，否则进入下一环。责任链模式还有很多处理流程类应用的例子，
比如：请假流程，审批流程，异常处理流程等等。

责任链模式的好处是解耦，如果不用责任链模式把处理流程分层，那么就需要一个硕大的函数来处理各种判断，那么这个函数很容易就
臃肿不堪，里面包含各种 `if...else...`。

以上就是对责任链模式的介绍。

---

参考资料：

- https://zh.wikipedia.org/wiki/%E8%B4%A3%E4%BB%BB%E9%93%BE%E6%A8%A1%E5%BC%8F
