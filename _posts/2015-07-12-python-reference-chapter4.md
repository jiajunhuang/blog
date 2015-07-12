---
layout: post
title: "Python reference note3 - Execution model"
tags: [python]
---

* a block is a piece of Python program text that is executed as a unit. The 
following are blocks: a module, a function body, and a class definition.

* a scope defines the visibility of a name within a block.

* the scope of names defined in a class block is limited to the class block;
it does not extend to the code blocks of methods - this includes comprehensions
and generator expressions sice they are implemented using a function scope.

```python3
>>> class A:
...   a = 42
...   b = list(a+i for i in range(10))
... 
Traceback (most recent call last):
  File "<stdin>", line 1, in <module>
  File "<stdin>", line 3, in A
  File "<stdin>", line 3, in <genexpr>
NameError: name 'a' is not defined
```

* if a name is bound in a block, it is a local variable of that block, unless
declared as nonlock. if a name is bound at the module level, it is a global 
variable.

* if a variable is used in a (inner, usually) code block, but not defined 
there(instead, define in outter, usually), it is a free variable.

* the following constructs bind names: formal parameters to functions, import
statements, class and function  definitions(these bind the class or function
name in the defining block), and targets that are identifiers if occurring in 
an assignment, for loop header, or after as in a with statement or except clause.
