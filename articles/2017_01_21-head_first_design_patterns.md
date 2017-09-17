# 设计模式（1）- 深入浅出设计模式 阅读笔记

> 将会使用Java和Python。其中Java实现可能比较业余，但是仍然使用Java，
> 因为很多设计模式在Java中更容易看清楚。

策略模式，观察者模式，装饰器模式，工厂模式，单例模式。

## 策略模式

策略模式其实就是一句话“针对接口编程，而不是针对实现编程”。接口是在Java中
可以是对应interface也可以是abstract类。接口其实是一个抽象概念。例如，c语言虽然
语法上不支持面向对象编程，但是其头文件的声明就相当于接口。

我们先来看Python版本的策略模式：

```python
class SoundMixin:
    def sound(self):
        raise NotImplemented()


class DogSoundMixin(SoundMixin):
    def sound(self):
        print("wang wang...")


class DuckSoundMixin(SoundMixin):
    def sound(self):
        print("gua gua...")


class Animal:
    def make_noise(self):
        self.sound()


class Dog(DogSoundMixin, Animal):
    pass


class Duck(DuckSoundMixin, Animal):
    pass


if __name__ == "__main__":
    Dog().make_noise()
    Duck().make_noise()
```

运行一下：

```bash
$ python interface.py 
wang wang...
gua gua...
```

然后是Java版本：

```java
// Sound.java
public interface Sound {
    public void sound();

}

// Animal.java
public abstract class Animal {
    abstract void sound();

    public void makeNoise() {
        sound();
    }
}

// Dig.java
public class Dog extends Animal implements Sound {
    public void sound() {
        System.out.println("wang wang...");
    }
}

// Duck.java
public class Duck extends Animal implements Sound {
    public void sound() {
        System.out.println("gua gua...");
    }
}

// Main.java
public class Main {
    public static void main(String[] args) {
        Animal animal;

        animal = new Dog();
        animal.makeNoise();

        animal = new Duck();
        animal.makeNoise();
    }
}
```

运行一下：

```bash
$ javac *.java && java Main 
wang wang...
gua gua...
```

## 观察者模式

观察者模式，就是消息订阅。消息提供者能够提供接口动态注册和注销，当消息到来，
消息提供者就会通知订阅者。典型的例子就是epoll，我们注册我们想要监听的事件，
例如描述符可读，当一个描述符可读时，epoll就会通知我们（这个例子也许不是特别
明显）。

我们先用Python实现一个：

```python
class Observer:
    def __init__(self):
        self.callbacks = {}

    def register(self, name, func):
        self.callbacks[name] = func

    def cancell(self, name):
        if name in self.callbacks:
            self.callbacks.pop(name)

    def notify(self, event):
        for name, callback in self.callbacks.items():
            callback(name, event)


def callback(name, event):
    print("name(%s) with event: %s" % (name, event))


if __name__ == "__main__":
    observer = Observer()
    observer.register("marry", callback)
    observer.register("jack", callback)
    observer.register("lucy", callback)

    observer.notify("eat")

    observer.cancell("jack")
    observer.notify("sleep")
```

运行结果：

```bash
name(marry) with event: eat
name(jack) with event: eat
name(lucy) with event: eat
name(marry) with event: sleep
name(lucy) with event: sleep
```

再来Java：

```java
// Observerable.java
import java.util.ArrayList;


public class Observerable {
    private ArrayList<Callback> observerables;
    public Observerable() {
        observerables = new ArrayList<Callback>();
    }

    public void register(Callback callback) {
        observerables.add(callback);
    }

    public void cancell(Callback callback) {
        int index = observerables.indexOf(callback);
        if (index >= 0) {
            observerables.remove(index);
        }
    }

    public void runAllCallbacks() {
        Callback callback;
        for (int i = 0; i < observerables.size(); i++) {
            callback = observerables.get(i);
            callback.callback("notifying " + i);
        }
    }
}

// Callback.java
public interface Callback {
    public void callback(String s);
}

// EatCallback.java
public class EatCallback implements Callback {
    public void callback(String s) {
        System.out.println("eat callback been called");
    }
}

// Main.java
public class Main {
    public static void main(String[] args) {
        Observerable observerable = new Observerable();

        observerable.register(new EatCallback());
        observerable.register(new EatCallback());
        observerable.register(new EatCallback());

        observerable.runAllCallbacks();
    }
}
```

运行一下：

```bash
$ javac *.java && java Main
eat callback been called
eat callback been called
eat callback been called
```

## 装饰器模式

装饰器模式，就是要在原有的函数上包装一层，让它能做更多的事情，但是却不用修改
原来的函数的代码。如果用Python的同学，相信无人不知无人不晓装饰器的威力，例如：

```python
from contextlib import contextmanager


@contextmanager
def print_tag(tag):
    print("<%s>" % tag)
    yield
    print("</%s>" % tag)


with print_tag("body"):
    print("html body")
```

```bash
$ python decorator.py 
<body>
html body
</body>
```

然后再上Java实现：

```java
// Origin.java
public class Origin {
    private String description = "Origin";

    public String getDescription() {
        return description;
    }
}

// Decorator.java
public class Decorator {
    Origin origin;

    public Decorator(Origin origin) {
        this.origin = origin;
    }

    public String getDescription() {
        return origin.getDescription() + " -> Decorator";
    }
}

// Main.java
public class Main {
    public static void main(String[] args) {
        Origin origin = new Origin();
        Decorator decorator = new Decorator(origin);
        System.out.println(decorator.getDescription());
    }
}
```

运行一下：

```bash
$ javac *.java && java Main
Origin -> Decorator
```

## 工厂模式，抽象工厂

工厂模式和抽象工厂的差别在于，一个使用继承来实现，是一种金字塔关系，超类在上。
而抽象工厂虽然具体实现是金字塔关系，但是实际使用的时候只需要指定接口，所以
总体看来是一个倒金字塔关系。Python上不是很能看得出来，但是Java版本的很明显。
我们举个例子，现在有NFS和ISCSI的库供我们使用，我们希望能够做到指定 `type`就
返回给我们对应的实例。

先来Python版本：

```python
TYPE_NFS = "nfs"
TYPE_ISCSI = "iscsi"


class Storage:
    pass


class NFS(Storage):
    pass


class ISCSI(Storage):
    pass


class StorageFactory:
    @staticmethod
    def instance(type):
        return NFS() if type == TYPE_NFS else ISCSI()


if __name__ == "__main__":
    print(StorageFactory.instance(TYPE_NFS).__class__)
    print(StorageFactory.instance(TYPE_ISCSI).__class__)
```

运行一下：

```bash
$ python factory.py 
<class '__main__.NFS'>
<class '__main__.ISCSI'>
```

再来Java：

```java
// Storage.java
public interface Storage {
    public String getName();
}

// NFS.java
public class NFS implements Storage {
    public String getName() {
        return "it's NFS";
    }
}

// ISCSI.java
public class ISCSI implements Storage {
    public String getName() {
        return "it's ISCSI";
    }
}

// StorageFactory.java
public class StorageFactory {
    public Storage instance(String type) {
        if (type.equals("nfs")) {
            return new NFS();
        } else {
            return new ISCSI();
        }
    }
}

// Main.java
public class Main {
    public static void main(String[] args) {
        StorageFactory storageFactory = new StorageFactory();
        System.out.println(storageFactory.instance("nfs").getName());
        System.out.println(storageFactory.instance("iscsi").getName());
    }
}
```

运行一下：

```bash
$ javac *.java && java Main
it's NFS
it's ISCSI
```

上面的Java代码是抽象工厂，其优点是简洁，调用类无需修改代码即可运行，但是每次
增加新的存储类型的时候，都需要修改一下instance方法。而工厂方法则无需如此：

```java
// Storage.java
public abstract class Storage {
    public abstract String getName();
    public static Storage instance(String type) {
        return null;
    }
}

// NFS.java
public class NFS extends Storage {
    public String getName() {
        return "it's NFS";
    }

    public static Storage instance(String type) {
        return new NFS();
    }
}

// ISCSI.java
public class ISCSI extends Storage {
    public String getName() {
        return "it's ISCSI";
    }

    public static Storage instance(String type) {
        return new ISCSI();
    }
}

// Main.java
public class Main {
    public static void main(String[] args) {
        System.out.println(NFS.instance("nfs").getName());
        System.out.println(ISCSI.instance("iscsi").getName());
    }
}
```

所以工厂模式和抽象工厂的区别在于，工厂模式由子类决定具体实例化哪个类，而
抽象工厂自己决定初始化哪个类。

## 单例模式

单例模式是在开发中经常能用到的。单例，名字比较官方，如果改成“全局唯一”，
那就瞬间更好理解了。例如我们需要初始化一个全局唯一的配置类。

Python版本：

```python
class Singleton(type):
    _instances = {}

    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            cls._instances[cls] = super().__call__(*args, **kwargs)
        return cls._instances[cls]


class SingletonConfig(metaclass=Singleton):
    def __init__(self):
        print("SingletonConfig initializing...")


if __name__ == "__main__":
    for i in range(10):
        SingletonConfig()
```

运行结果：

```bash
$ python singleton.py 
SingletonConfig initializing...
```

更多实现方式见：http://stackoverflow.com/questions/6760685/creating-a-singleton-in-python

Java版本：

```java
// SingletonConfig.java
public class SingletonConfig {
    private static SingletonConfig singletonConfig;

    private SingletonConfig() {};

    public static synchronized SingletonConfig getConfig() {
        if (singletonConfig == null) {
            System.out.println("initializing a new Singleton Config");
            singletonConfig = new SingletonConfig();
        }
        return singletonConfig;
    }
}

// SingletonDemo.java
public class SingletonDemo {
    public static void main(String[] args) {
        for (int i = 0; i < 10; i++) {
            SingletonConfig config = SingletonConfig.getConfig();
        }
    }
}
```

运行结果：

```bash
$ javac *.java && java SingletonDemo 
initializing a new Singleton Config
```
