---
layout: post
title: "Python reference note2 - Data model"
tags: [python]
---

### objects, values, types

* all data in a Python program is represented by objects or by relations 
between objects.

* every object has an identity, it comes since the object was born, and it will
never changes. keyword `is` compares the identity of two objects. and function
`id()` returns an integer representing of it's identity.

> in cpython, `id(x)` is the memory address where `x` is stored.

* an object's type determines the operations thath the object supports. an 
object's type will not be changed since it was created.

* some objects contain references to other objects, such as tuple, list, dict.
these are called `container`. e.g.

```python3
>>> a = [1,2,3]
>>> b = [4,5,6]
>>> c = (a,b)
>>> id(a),id(b), id(c)
(140224302651592, 140224302678088, 140224302673672)
>>> a[2] = 100
>>> id(a),id(b), id(c)
(140224302651592, 140224302678088, 140224302673672)
>>> 
```

### the standard type hierarchy

* None: there is a single object with this value. it's truth value is false.
it is accessed through the literal `None`

* NotImplemented: there is a single object with this value. it's truth value 
is true. it is accessed through the literal `NotImplemented`

* Ellipsis: there is a sigle object with this value. it is accessed through
the literal `...` or `Ellipsis`

* numbers.Number

* Sequences: 

> * the built-in function `len()` return the number of items of a sequence.
> * sequence are distinguished according to thier mutability.
> * python does not have a `char` type, instead, it is a string object with length 1.
> * Tuple: A tuple of one item can be formed by affixing a comma to an 
expression, it means that, tuple is not created by parenthese but the comma.

```python3
>>> a = "hi"
>>>type(a)
<class 'tuple'>
```

> * list: lists are formed by placing a comma-separated list of expressions in square
brackets.
> * aside from being mutable, byte arrays otherwise provide the same interface
and functionality as immutable bytes objects.

* set: set represent unordered, finite sets of unique, immutable objects. set 
cannot be indexed, but can be iterated. note that if two numric objects compared
to be equal, only one of them can be contained in a set.

* dict: the obly types of values not acceptable as keys are values containing
lists or dictionaries or other mutable types that are compared by value rather
than by object identity. also note that if two numric objects compared to be 
equal, only one of them can be contained in a set.

* instance methods
> readonly attributes
> * `__self__`: class instance object
> * `__func__`: function object
> * `__doc__`: method's documentation(same as `__func__.__doc__`)
> * `__name__`: method name(same as `__func__.__name__`)
> * `__module__`: name of module where the method was defined in, or None

* ** ??? instance method ??? need reread. **

* class instance: if no class attribute is found, and the object's class has a 
`__getattr__()` method, that is called to satisfy the loopup. note that if the
attribute is found through the normal mechanism, `__getattr__()` is not called.

* attribute assignments and deletions update the instance's dictionary, never
a class's dictionary. If the class has a `__setattr__()` or `__delattr__()`
method, this is called instead of updating the instance dictionary directly.

* if a code object represents a function, the first item in `co_consts` is the 
documentation string of the function, or `None` if undefined.

* if `__new__()` returns an instance of cls, then the new instance's `__init__()`
wil be invoked like `__init__(self[,...])`. else it will not invoke.

* `__init__()` is called after the instance has been created by `__new__()`, 
before it is returned to the caller. if a base class has an `__init__()` method, 
the derived class's `__init__()` must explictly call it to ensure proper
initialization of the base class part of the instance.

```bash
$ cat fun.py 
class A():
    def __init__(self):
        print('a init')

class B(A):
    def __init__(self):
        print('b init')

class C(A):
    def __init__(self):
        print('c init')
        super().__init__()

A()
B()
C()
$ python fun.py 
a init
b init
c init
a init
```

* rich comparison: there are no implied relationships among the comparison 
operators. the truth of `x==y` does not imply that `x!=y` is false.

* ** `object.__hash__(self)` need to reread **
