# Thinking in JAVA学习笔记

- static声明的东西与实例无关。类中，static声明的东西是类内共享。static声明的变量会在类第一次初始化时初始化。默认为零值。

```java
class Cups {
    static int i;
    int j;

    Cups(int num) {
        System.out.println("cup " + num + " constructed!, i is: " + i + " j is: " + j);
        i++;
        j++;
    }

    public void f() {
        System.out.println("i is: " + i + " j is: " + j);
    }
}

class Ideas {
    static int i;

    Ideas(int num) {
        System.out.println("idea " + num + " constructed! i is: " + i);
    }
}

public class Main {
    public static void main(String []args) {
        new Cups(1);
        new Cups(1).f();
    }
}
```

输出：

```bash
$ javac Main.java && java Main 
cup 1 constructed!, i is: 0 j is: 0
cup 1 constructed!, i is: 1 j is: 0
i is: 2 j is: 1
```

- 每个编译单元(.java文件)只能有一个public的类

- 访问权限控制有四个等级，从强到弱

    - public 包外可访问
    - 不写，就是包内可访问，包外不可访问
    - protected 就是该类和子类可访问，其他不可以
    - private 仅该类内可访问

- 可以向上转型，不过向下转型编译器会报错：

```java
class Person {
    protected String name;

    Person(String name) {
        this.name = name;
    }
}

class Zengnima extends Person {
    Zengnima(String name) {
        super(name);
    }

    void printName() {
        System.out.println("name: " + this.name);
    }
}

public class Main{
    public static void main(String []args) {
        //Person p = new Zengnima("shadiao");
        //Zengnima z = (Zengnima)p;
        //z.printName();
        Person p = new Person("shadiao");
        Zengnima z = (Zengnima)p;
        z.printName();
    }
}
```

执行结果：

```bash
$ javac Main.java && java Main 
Exception in thread "main" java.lang.ClassCastException: Person cannot be cast to Zengnima
	at Main.main(Main.java:25)
```

- java中除了static和final是前期绑定，其他方法都是后期绑定(幸好是这样)。
