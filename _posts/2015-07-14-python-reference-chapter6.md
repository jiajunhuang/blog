---
layout: post
title: "Python reference note5 - Expressions"
tags: [python]
---

* Arithmetic conversions, below is the rules:

> * if either argument is a complex number, the other is converted to complex.
> * otherwise, if either argument is a floating point number, the other is 
converted to floating point.
> * otherwise, both must be integers and no conversion is necessaey.

* all literals correspond to immutable data tyoes.

* a parenthesized expression list yield whatever that expression list yields:
if the list contains at least one comma, it yields a tuple; otherwise, it 
yields the single expression that makes up the expression list. an empty pair
of parentheses yields an empty tuple object. **Note that tuples are not formed
by the parentheses, but rather by use of the comma operator.the exception is 
the empty tuple, for which parentheses are required.**

* if you need to pass a dict as argument, it should **always** appear in the 
end of arguments.

* if keyword arguments are present, they are first converted to positional 
arguments.

* default values are calculated, once, when the function is defined; thus, a 
mutable object such as a list or dictionary used as default value will be 
shared by all calls that don't specify an argument value for the corresponding
slot.

```python3
>>> a = [1,2,3]
>>> def foo(lst = a):
...   lst.append("foo")
...
>>> def bar(lst = a):
...   lst.append("bar")
...
>>> foo()
>>> bar()
>>> a
[1, 2, 3, 'foo', 'bar']
```

* comparisons can be chained arbitrarily, e.g. `x < y < z` is equivalent to
`x < y and y < z`, except that y is evaluated only once, (but in both cases
z is not evaluated at all when x < y is found to be false).

* tuples and lists are compared lexicographically using comparison of  corresponding 
elements. this means that to compare equal, each element must compare equal and
the two sequence must be of the same type and have the same length. if the 
coreesponding element does not exist, the shorter sequence is ordered first.

* Most other objects of built-in types compare unequal unless they are the 
same object; the choice whether one object is considered smaller or larger 
than another one is made arbitrarily but consistently within one execution 
of a program. (if you have not implement `__gt__()` ... methods).

* in the context of boolean operations, and also when expressions are used 
by control flow statements, the following values are interpreted as false:
False, None, numeric zero of all types, and empty strings and containers
(including strigns, tuples, lists, dictionaries, sets, and frozensets).
all other values are interpreted as true. user-defined objects can customize 
their value by providing a `__bool__()` method.

* `and` and `or` expressions obey the short-circuit evaluation.

* note that neither and nor or restrict the value and type they return to
False and True, but rather return the last evaluated argument. but because
`not` has to create a new value, so it returns a boolean value regardless
of the type of its argument.

* The trailing comma is required only to create a single tuple 
(a.k.a. a singleton); it is optional in all other cases. A single expression 
without a trailing comma doesn’t create a tuple, but rather yields the value 
of that expression. (To create an empty tuple, use an empty pair of 
parentheses: ().)

* python evaluates expressions from left to right. notice that while evaluating
an assignment, the right-hand side is evaluated before the left hand side.

* python operator precedence table is [here](https://docs.python.org/3/reference/expressions.html#operator-precedence)
