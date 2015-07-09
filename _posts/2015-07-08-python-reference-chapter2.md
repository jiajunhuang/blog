---
layout: post
title: "Python reference note1"
tags: [python]
---

* `\` only valid in string literals or in the end of the line.

* tabs are replaced by 1~8 spaces. so it is recommented to replace tab with 
4 or more spaces in your editor. After replacement, the total number of spaces
determines the indentation of the line.

> Exception:TabError

* the indentation levels of the consecutive are used to generate `INDENT` and 
`DEDENT` tokens, using a stack:
> * before the first line is read, a 0 is pushed on the stack
> * at the befinning of each logical line, the line's indentation level is 
compared to the top of the stack. 
`=`: nothing happens
`>`: pushed on the stack, generate one `INDENT` token
`<`: if it is elements on the stack, all the elements larger than it will be
poped off, else raise an exception.

* at the end of the file, a `DEDENT` token is generate for each number 
remaining on the stack that is larger than 0.

* valid characters for identifiers: `A-Z`, `a-z`, `_`, `0-9`(digits can't
	be used as the head of a identifier)

* keywords: we have 33 keywords in [here](https://docs.python.org/3/reference/lexical_analysis.html#keywords)

> * `_*`: stand for the result of last evaluation in interpreter. identifiers 
start with `_` will not be imported by `from module import *`

> * `__*`: class-private names.

> * `__*__`: system-defined names. one can redefine such methods to change 
behavor of some operations.

### String literals

* white spaces is not allowed between stringprefix and byteprefix and the
rest of the literals.

* (in python3) UTF-8 is the default encoding declaration.

* bytes with a numeric value of 128 or greater must be expressed with escapes.

* escaped sequences are similar to those used by C.

* unlike standard c, all unrecognized escape sequences are left in the string
, keep it unchanged.

* a raw literal cannot end in a single backslash, because the backslash would 
escape the following quote character. raw string still processing backslash.

### Numric literals

* numeric literals fo not include a sign, a phrase like `-1` is actually
an expression composed of the unary operator `-` and the literal `1`.
