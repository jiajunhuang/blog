## Typescript

Typescript是JS的超集，最主观的，增加了类型系统，这样可以减少很多错误，TS编译成JS，例如：

```bash
$ cat > greeter.ts
class Student {
    fullName: string;

    constructor(public firstName: string, public middleInitial: string, public lastName: string) {
        this.fullName = firstName + " " + middleInitial + " " + lastName;
    }
}

interface Person {
    firstName: string;
    lastName: string;
}

function greeter(person: Person) {
    return "Hello, " + person.firstName + " " + person.lastName;
}

let user = new Student("Jane", "M.", "User");

document.body.textContent = greeter(user);
$ tsc greeter.ts 
$ cat greeter.js 
var Student = /** @class */ (function () {
    function Student(firstName, middleInitial, lastName) {
        this.firstName = firstName;
        this.middleInitial = middleInitial;
        this.lastName = lastName;
        this.fullName = firstName + " " + middleInitial + " " + lastName;
    }
    return Student;
}());
function greeter(person) {
    return "Hello, " + person.firstName + " " + person.lastName;
}
var user = new Student("Jane", "M.", "User");
document.body.textContent = greeter(user);
```
TS主要有这么几种类型：

- Boolean: `let isDone: boolean = false;`
- Number: `let aint: number = 1;`
- String: `let name: string = "hello";`
- Array: `let list: number[] = [1, 2, 3];` 或者使用泛型：`let list: Array<number> = [1, 2, 3];`
- Tuple: `let x: [string, number] = ["hello", 10];`
- Enum: `enum Color {Red, Green, Blue}`
- Any: 类似Go语言中的 `interface{}` 或者C语言中的 `void *`：`let notSure: any = 4;`
- Void: `function warnUser(): void { console.log("This is my warning message"); }`
- `null` 和 `undefined`: `let u: undefined = undefined; let n: null = null;`
- Never: 永不执行，例如如果是函数返回Never，那么说明函数只能通过异常终止，而不会返回
- Object: 不是上面所说的类型的，就是 `object` 类型

---

参考资料：

- https://www.typescriptlang.org/docs/handbook/basic-types.html
