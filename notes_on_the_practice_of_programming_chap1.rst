:Date: 10/05/2015

程序设计实践笔记1 -- Style
==========================

Names
~~~~~

-  use descriptive names for globals, short names for locals.

.. code:: c

    int i = 0;
    for (; i < 10; i++)
        ;

-  | be consistent. give related things related names that show their
     relationship
   | and highlight their diffrence.

-  | use active names for functions. function names should be based on
     active
   | verbs, perhaps followed by nouns.

-  be accurate(保证准确).

Expressions and Statements
~~~~~~~~~~~~~~~~~~~~~~~~~~

-  indent to show structure (python have a native support for it,
   hahah).

-  use the nature form for expressions.

.. code:: c

    // do not use expressions like 
    // if (!(block_id < actblks) || !(block_id >= unblocks))
    // but in this way:
    if ((block_id >= actblks) || (block_id < unblocks))

-  parenthesize to resolve ambiguity. parentheses specify grouping and
   can be
   used to make the indent clear
   even when they are not required.

.. code:: c

    // instead of writing expressions like 
    // leap_year = y % 4 == 0 && y % 100 != 0 || y % 400 == 0;
    // write:
    leap_year = ((y % 4 == 0) && (y % 100 != 0) || (y % 400 == 0));

-  break up complex expressions.

-  be careful with side effects. like:

.. code:: c

    array[i++] = array[i++] = ' ';

Consistency and Idioms
~~~~~~~~~~~~~~~~~~~~~~

-  | use a consistent indentation and brace style. the program;s
     consistency is
   | more import than your own,
   | because it makes life easire for those who follow.

-  use idioms for consistency(采用惯用写法).

.. code:: c

    // instead of wrting :
    // int i = 0;
    // for (i = 0; i < n; )
    //     array[i++] = 1;
    int i = 0;
    for (i = 0; i < n; i++)
        array[i] = 1;

.. code:: c

    // write 
    while (1):
        //
    // instead of writing for (;;)

-  sprawling layouts also force code onto multiple screens or pages, and
   thus
   detract from readability.
   so, set cc to 79 in your vim! :)

.. code:: vim

    set cc=79

-  | the return value from ``malloc``, ``realloc``, ``strdup`` or any
     other allocation
   | routine should always
   | be checked!

-  | use else-ifs for multi-way decisions. put the most possible choice
     in the
   | first statement can improve
   | performance.

-  cases should always end with a ``break``, though longer.

.. code:: c

    switch (c) {
        case 'a': blablabla; break;
        case 'b': blablabla; break;
        ...
    }

| but, an acceptable use of fall-through occurs when serveral cases have
| identical(相同的) code, the
| conventional layout is like this:

.. code:: c

    switch (c) {
        case '0':
        case '1':
        case '2':
            blablabla
            break;
    }

Function Macros
~~~~~~~~~~~~~~~

-  avoid function macros.

    | in c++, inline functions render function macros unnecessary;
    | in java, there are no macros;
    | in c, they cause more problems than they solve.

-  parenthesize the macro body and arguments.

.. code:: c

    1/square(x) // works well if square is a function, but not macro:
    // #define square(x) (x)*(x), it will be evaluated to:
    1/(x) * (x)
    // this version works well:
    // #define square(x) ((x) * (x))

Magic Numbers
~~~~~~~~~~~~~

-  | ``magic numbers`` are the constants, array sizes, character
     posiitions,
   | conversion factors, and other literal numeric values that appear in
     programs.

-  | give name toi magic numbers. by given names to the principal
     numbers in the
   | calculation, we can make the code easier to follow.

-  | define numbers as constants, not macros. macros are dangerous ways
     to program
   | because they change the lexical structure of the program underfoot.

-  use character constants, not integers.

.. code:: c

    // instead of using:
    if (c >= 65 && c <= 90) 
    // using: 
    if (c >= 'A' && c <= 'Z')
    // this way is the best(use the standard library):
    if (isupper(c))

-  use the language to calculate the size of an object.

.. code:: c

    #define NELEMS(array) (sizeof(array) / sizeof(array[0]))

Comments
~~~~~~~~

-  | the best comments aid the understanding of a program by briefly
     pointing out
   | salient details or by providing a larger-scale view of the
     proceedings.

-  | don't belabor the obvious. comments should't report self0evident
     information,
   | such as the usage of ``i++``.

-  | comment functions and global data. we comment functions, global
     variables,
   | constant definitions, fields in structures and classes, and
     anything else
   | where a brief summary can aid understanding.

-  don't comment bad code, rewrite it.

-  | don't contradict the code(代码与注释要保持同步修改,以免冲突).
     comments
   | should not only agree with code, they should support it.

-  | clarify, don't confuse. comments are supposed to help readers over
     the hard
   | parts, not create more obstacles. when it takes more than a few
     words to
   | explain what's happening, it's often an indication that the code
     should be
   | rewritten.

吐槽
~~~~

| 《程序设计实践》这本书是作者经验总结,原书很棒.但是我买的是一本评注版,
| 原以为评注版会更好,没想到反而是影响阅读.他评注的我都会,我不会的他也没评注=
  =!
| 价格还比原书贵.英文还可以的同学直接原书走起吧.所以说以后都不要买评注版本
| 的书了,更贵而且没什么用.
