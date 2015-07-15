---
layout: post
title: "Python reference note7 - compound statements"
tags: [python]
---

* for statement: `for i in expression_list`, the expression list is evaluated 
once, it should yield an iterable object.

* in for statements, a `break` statement executed in the first suite terminates 
the loop without execting the `else` clauses's suite. a `continue` statement
executed in the first suite skips the rest of the suite and continues with the
next item, **or with the `else` clause if there is no next item** .

* if two nested handlers exist for the same exception, and the exception occurs
in the try clause of the inner handler, the outer handler will not handle the 
exception.

* in `try...else...finally...` statements, the optional else clause is executed
if and when control flows off the end of the try clause. exceptions in the else
clause are not handled by the preceding except clauses.

* if finally is present, it specifies a 'cleanup' handler. the try clause is 
executed, including any except and clauses. if an exception occurs in any of 
the clauses and is not handled, the excetion is temporarily saved. the finally 
clause is executed. if there is a saved exception, it is re-raised at the end 
of the finally clause. if the finally clause raises another exception, the 
saved exception is set as the context of the new exception. if the finally 
clause executes a return or break statement, the saved exception is discarded.
the execption information is not availabel to the program during exection of 
the finally clause.

* a function definition is an executeable statement. its execution binds the 
function name in the current local namespace to a function object. this 
function object contains a reference to the current global namespace as the 
global namespace to be used when the function is called. the function definition
does not execute the function body, this gets executed only when the function 
is caled.

* a function definition may be wrapped by one or more decorator expressions. 
decorator expressions are evaluated when the function is defined, in the scope
that contains the function definition. the result must be a callable, which is
invoked with the function object as the only argument. the returnd value is 
bound to the function name instead of the function object.

* default parameter values are evaluated from left to right when the function
definition is executed. this means that the expression is evaluated once, when 
the function is defined, and that the same 'pre-computed' value is used for 
each call.

```python3
class A():
    def foo(self, lst = []):
        lst.append('foo')
        print(lst)

a = A()
a.foo()
b = A()
b.foo()
```

and the execution:

```bash
$ python fun.py 
['foo']
['foo', 'foo']
```
* variables defined in the definition are class attributes, they are shared
by instances. instance attributes can be set in a method with `self.name = value`
