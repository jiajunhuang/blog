程序设计实践笔记
=================

Style
-------

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

Algorithms and Data Structures
---------------------------------

Chapter2
~~~~~~~~~~

-  | if you are developing programs in a field that's new to you, you
     must find out
   | what is already known, lest you waste your time doing poorly what
     others have
   | already done well.

-  | if repeated searches are going to be made in some data set, it will
     be
   | profitable to sort once and then use binary search.

-  `big-o notation cheat sheet <http://bigocheatsheet.com/>`__

-  这一章主要介绍了常用的数据结构和主要操作,例如List, Tree, Hash
   Table.---

Chapter3
~~~~~~~~~~

-  | whoever opens an input file should do the corresponding close:
   | matching tasks should be done at the same level or place.

-  | as a principle, library routines should not just die when an error
   | occursl error status should be returned to the caller for
     appropriate
   | action.

-  do the same thing the same way everywhere. keep consistency.

Debugging
-------------

Easy bugs
~~~~~~~~~

-  look for familiar patterns. ask yourself, "have I seen this before"
   when you get a bug.

-  | examine the most recent change. source code control systems and
     other history mechanisms are
   | helpful here. e.g. git.

-  | don't make the same mistake twice. easy code can have bugs if its
     familiarity causes us to
   | let down out guard. even when code is so simple you could write it
     in your sleep, don't fall
   | asleep while writing it.

-  | debug it now, not later. don't ignore a crash when it happens;
     track it down right away,
   | since it may not happen again until it's too late.o

-  | get a stack trace. the source line number of the failure, often
     part a stack trace, is the
   | most useful single piece of debugging infomation;
     improbable(难以置信的,不会的) values of
   | arguments are also a big clue(zero pointers, integers that are huge
     when they should be
   | small, or negative when they should be positive, character strings
     that aren't alphabetic).

-  | read before typing. one effective but under-appreciated debugging
     technique is to read the
   | code very carefully and think about it for a while without making
     changes. resist the urge to
   | start typing, thinking is a worthwhile alternative.

-  explain your code to someone else.
   `小黄鸭调试法？哈哈哈哈 <https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=1&cad=rja&uact=8&ved=0CB4QFjAAahUKEwjK8PS09LTIAhWM5oAKHWwpACU&url=https%3A%2F%2Fzh.wikipedia.org%2Fzh%2F%25E5%25B0%258F%25E9%25BB%2584%25E9%25B8%25AD%25E8%25B0%2583%25E8%25AF%2595%25E6%25B3%2595&usg=AFQjCNHJAF8oTPEFyICQ_QJ9tz_gwKlcvw&sig2=REOYXrZfbO6yu1AsA7QNLQ>`__

Hard bugs
~~~~~~~~~

-  | make the bug reproducible. if the bug can't be made to happen every
     time, try to understand
   | why not. does some set of conditions make it happen more often than
     others? using a log system
   | to log some unreproducible values(such as a random number).

-  | divide and conquer. narrow down the possibilities by creating the
     smallest input where the
   | bug still shows up.

-  study the numerology of failures(研究错误出现的规律).

-  display output to localize your search. e.g. use ``grep``.

-  | write self-checking code. personally, I think ``assert`` is useful,
     and write it with a DEBUG
   | macro, just like:

.. code:: c

    #ifdef DEBUG
    ......
    #endif

and here is a trick for ``assert``:

.. code:: c

    assert(a > b), "a should bigger than b";

so the string after ``assert(a > b)`` will be displayed if assert works.

-  | write a log file. be sure to flush I/O buffers so the final log
     records appear in the log
   | file.

-  draw a picture. sometimes pictures are more effective than text for
   testing and debugging.

-  | use tools. like ``diff``, ``grep`` , etc. write tricial programs to
     test hypotheses or confirm
   | your understanding of how something works(善用工具, 弄清楚哪些坑).

-  | keep records. if the search for a bug goes on for any length of
     time, you will begin to lose
   | track of what you tried and what you learned.

Last Resorts
~~~~~~~~~~~~

| what do you do if none of this advice helps? this may be the time to
  use a good debubger to
| step through the program. it's tough to find this kind of bug, because
  your brain takes you
| right around the mistake, to follow what the program is doing, not
  what you think it is doing.

| if you can't find a bug after considerable work, take a break, clear
  your mind, do something
| else, talk to a friend and ask for help.

Other People's Bugs
~~~~~~~~~~~~~~~~~~~

| if you think that you have found a bug in someone else's program, the
  first step is to make
| absolutely sure it is a genuine bug, so you don't waste the author's
  time and lose your own
| credibility.
