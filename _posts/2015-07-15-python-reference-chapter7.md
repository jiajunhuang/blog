---
layout: post
title: "Python reference note6 - simple statements"
tags: [python]
---

* assignment statements are used to (rebind) names to values and to modify 
attributes or items of mutable objects.

* assignment statements **need reread**, [here](https://docs.python.org/3/reference/simple_stmts.html#assignment-statements)

* unlike normal assignments, argumented assignments evaluate the left-hand side
before evaluating the right-hand side.

* if a name binding operation occurs anywhere within a code block, all uses of 
the name within the name within the block are treated as references to the 
current block. also note that, del is also a name binding operation, althrough
it is unbinding.

* return may only occur syntactically nested in a function definition, not 
within a nested class definition.

* when return passes control out of a try statement with a finally clause, that
finally clause is executed before really leaving the function.

* **diffrence between expressions and statement** 
[copy from stackoverflow](http://stackoverflow.com/questions/4728073/what-is-the-difference-between-an-expression-and-a-statement-in-python)
expressions only contain 
identifiers, literals and operators, where operators include arithmetic and 
boolean operators, the function call operator () the subscription operator
[] and similar, and can be reduced to some kind of "value", which can be any 
python object. Examples:

```python3
3+5
map(lambda x: x*x, range(10))
[a.x for a in some_iterable]
yield 7
```

statements, on the other hand, are everything that can make uo a line(or 
several lines) of python code, note that expressions are statements as well,
examples:

```python3
print(42)
if x: do_y()
a = 7
```

* break may only occur syntactically nested in a for or while loop, but not
nested in a function or class definition with that loop. it terminates the 
nearest enclosing loop, skipping the optional else clause if the loop has one.
if a for loop is terminated by break, the loop control target keeps its current
value. when break passes control out of a try statement with a finally cause,
that finally clause is executed before really leacving the loop.

* when continue passes control out of a try statement with a finally clause, 
that finally clause is executed before really starting the next loop cycle.

* the public names defined by a module are determined by checking the module's
namespace for a variable named `__all__`, if defined, it must be a sequence of 
strings which are names defined or imported by that module. the names given in 
`__all__` are all considered public and are required to exist. if `__all__` is 
not defined, the set of public names includes all names found in the module's
namespace which do not begin with an underscore character(`_`). `__all__` should
contain the entire public API.

* a future statement must appear near the top of the module. the only lines 
that can appear before a future statement are:
> * the module docstring(if any)
> * comments
> * blank lines
> * other future statements

* the `nonlocal` statement causes the listed identifiers to refer to previously
bound variables in the nearest enclosing scope excluding globals. names listed
in a nonlocal statement, unlike those listed in a global statement, must refer
to pre-existing binding in an enclosing scope.
