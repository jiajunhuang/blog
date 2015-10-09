---
layout:     post
title:      "程序设计实践笔记3 --Debugging"
tags: [programming, notes]
---

### Easy bugs

* look for familiar patterns. ask yourself, "have I seen this before" when you get a bug.

* examine the most recent change. source code control systems and other history mechanisms are 
helpful here. e.g. git.

* don't make the same mistake twice. easy code can have bugs if its familiarity  causes us to 
let down out guard. even when code is so simple you could write it in your sleep, don't fall 
asleep while writing it.

* debug it now, not later. don't ignore a crash when it happens; track it down right away,
since it may not happen again until it's too late.o

* get a stack trace. the source line number of the failure, often part a stack trace, is the 
most useful single piece of debugging infomation; improbable(难以置信的，不会的) values of 
arguments are also a big clue(zero pointers, integers that are huge when they should be
small, or negative when they should be positive, character strings that aren't alphabetic).

* read before typing. one effective but under-appreciated debugging technique is to read the
code very carefully and think about it for a while without making changes. resist the urge to 
start typing, thinking is a worthwhile alternative.

* explain your code to someone else. [小黄鸭调试法？哈哈哈哈](https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=1&cad=rja&uact=8&ved=0CB4QFjAAahUKEwjK8PS09LTIAhWM5oAKHWwpACU&url=https%3A%2F%2Fzh.wikipedia.org%2Fzh%2F%25E5%25B0%258F%25E9%25BB%2584%25E9%25B8%25AD%25E8%25B0%2583%25E8%25AF%2595%25E6%25B3%2595&usg=AFQjCNHJAF8oTPEFyICQ_QJ9tz_gwKlcvw&sig2=REOYXrZfbO6yu1AsA7QNLQ)

### Hard bugs

* make the bug reproducible. if the bug can't be made to happen every time, try to understand
why not. does some set of conditions make it happen more often than others? using a log system
to log some unreproducible values(such as a random number).

* divide and conquer. narrow down the possibilities by creating the smallest input where the 
bug still shows up.

* study the numerology of failures(研究错误出现的规律).

* display output to localize your search. e.g. use `grep`.

* write self-checking code. personally, I think `assert` is useful, and write it with a DEBUG
macro, just like:

```c
#ifdef DEBUG
......
#endif
```

and here is a trick for `assert`:

```c
assert(a > b), "a should bigger than b";
```

so the string after `assert(a > b)` will be displayed if assert works.

* write a log file. be sure to flush I/O buffers so the final log records appear in the log 
file.

* draw a picture. sometimes pictures are more effective than text for testing and debugging.

* use tools. like `diff`, `grep` , etc. write tricial programs to test hypotheses or confirm
your understanding of how something works(善用工具， 弄清楚哪些坑).

* keep records. if the search for a bug goes on for any length of time, you will begin to lose 
track of what you tried and what you learned.

### Last Resorts

what do you do if none of this advice helps? this may be the time to use a good debubger to
step through the program. it's tough to find this kind of bug, because your brain takes you 
right around the mistake, to follow what the program is doing, not what you think it is doing.

if you can't find a bug after considerable work, take a break, clear your mind, do something 
else, talk to a friend and ask for help.

### Other People's Bugs

if you think that you have found a bug in someone else's program, the first step is to make
absolutely sure it is a genuine bug, so you don't waste the author's time and lose your own 
credibility.
