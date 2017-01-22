# 设计模式（2）

## 命令模式

命令模式就是不改变原有代码，而在原有代码的基础之上封装一层，为每个命令创建一个
类，而这个类又实现了统一的一个接口。例如我们有电灯，电风扇，他们的开关方法
分别叫 `light`, `lights_out`, `on`, `off`。

用Python描述：

```python
class Light:
    def light(self):
        print("light on...")

    def lights_off(self):
        print("lights off...")


class Fan:
    def on(self):
        print("fan on...")

    def off(self):
        print("fan off...")


class Command:
    def execute(self):
        raise NotImplemented()


class LightOnCommand(Command, Light):
    def execute(self):
        self.light()


class LightOffCommand(Command, Light):
    def execute(self):
        self.lights_off()


class FanOnCommand(Command, Fan):
    def execute(self):
        self.on()


class FanOffCommand(Command, Fan):
    def execute(self):
        self.off()


if __name__ == "__main__":
    LightOnCommand().execute()
    LightOffCommand().execute()
    FanOnCommand().execute()
    FanOffCommand().execute()
```

执行结果：

```bash
$ python command.py 
light on...
lights off...
fan on...
fan off...
```

用Java描述：

```java
// Fan.java
public class Fan {
    public void on() {
        System.out.println("fan on...");
    }

    public void off() {
        System.out.println("fan off...");
    }
}

// Light.java
public class Light {
    public void light() {
        System.out.println("light on...");
    }

    public void lights_off() {
        System.out.println("light off...");
    }
}


// Command.java
public interface Command {
    public void execute();
}


// LightOnCommand.java
public class LightOnCommand implements Command {
    Light light;

    public LightOnCommand(Light light) {
        this.light = light;
    }

    public void execute() {
        light.light();
    }
}


// LightOffCommand.java
public class LightOffCommand implements Command {
    Light light;

    public LightOffCommand (Light light) {
        this.light = light;
    }

    public void execute() {
        light.lights_off();
    }
}


// FanOnCommand.java
public class FanOnCommand implements Command {
    Fan fan;

    public FanOnCommand(Fan fan) {
        this.fan = fan;
    }

    public void execute() {
        fan.on();
    }
}


// FanOffCommand.java
public class FanOffCommand implements Command {
    Fan fan;

    public FanOffCommand(Fan fan) {
        this.fan = fan;
    }

    public void execute() {
        fan.off();
    }
}


// Main.java
public class Main {
    public static void main(String[] args) {
        Light light = new Light();
        Fan fan = new Fan();

        LightOnCommand lightOnCommand = new LightOnCommand(light);
        LightOffCommand lightOffCommand = new LightOffCommand(light);

        lightOnCommand.execute();
        lightOffCommand.execute();

        FanOnCommand fanOnCommand = new FanOnCommand(fan);
        FanOffCommand fanOffCommand = new FanOffCommand(fan);

        fanOnCommand.execute();
        fanOffCommand.execute();
    }
}
```

执行结果：

```bash
$ javac *.java && java Main 
light on...
light off...
fan on...
fan off...
```

## 适配器模式

适配器模式要理解起来就比较简单，就是通过适配器，把两个原本不兼容的类通过
适配器转换，达成可以单向或双向调用。典型的生活中的例子就是，例如假设我买了
一个港版的iPhone，随机附赠的充电头也是港版的，使用电压和大陆不一样，插头
也不一样，所以就要购买一个转换接口，把国标插头转换成港版插头能用的插头。

## 模板方法模式

模板方法就是规定了大概的步骤，但是留有细节步骤让子类去实现。例如某调度系统
可能要求支持多种算法，但是大体步骤都是相似的，于是就可以通过模板方法模式把
相似的步骤都写下来，然后把具体细节实现留给子类完成。当最后调用的时候就会根据
实例化不同而产生不同细节上的效果。

## 迭代器模式

用Python的同学想必对迭代器非常的熟悉，我们的类只要实现了 `__iter__` 和 
`__next__` 协议，就可以用于 for 循环中。


## 状态模式

状态模式和策略模式非常的想象，区别在于，策略模式是用户来设置具体使用哪一种
实现，而状态模式是由类内部的状态机来完成的，例如为每一个类都实现可能的操作
会带来的影响和下一个状态，于是当启动状态机以后，就开始运转了。

## 代理模式

代理模式由于涉及Java远程调用，就暂时不上代码了，但是要理解起来，很简单，请看
自己的爬土啬方法。

## 最后

看完了《Head First 设计模式》。一直很喜欢《Head First》系列书籍，这本书也不
例外。设计模式是软件开发过程中常遇到的解决复杂度的一些方法，经过总结，而得出
来的“模式”。虽然看完了这本书，但是真正掌握他们还得在
“写代码-思考-领悟设计模式”的循环中才能掌握。
