# 求值策略：Applicative Order vs Normal Order

- https://cs.stackexchange.com/questions/40758/difference-between-normal-order-and-applicative-order-evaluation
- https://mitpress.mit.edu/sites/default/files/sicp/full-text/book/book-Z-H-10.html#%_sec_1.1.5

举个例子：

`(test 0 (p))`，如果`test`的定义是

```lisp
(define test (x y)
 (if (= x 0) 0 y)
)
```

`p` 的定义是 `(define p (p))`

- Applicative order执行到test的参数时，会立即对 0和p进行求值。0求值得到0，`p`求值得到`p`，`p`继续求值得到`p`，所以会
陷入无线循环
- Normal order执行到test的参数时，不会立即对参数进行求值，而是把函数进行展开，上面的表达式会被展开成

```lisp
(if (= 0 0) 0 (p))
```

然后开始执行，`(= 0 0)` 为true，直接取值为0。因此不会陷入循环。Normal order会将表达式进行展开，递归的将函数体替换
原表达式中的引用。当到达无法展开时才会开始求值。这种玩法也叫 lazy evaluation，Haskell就是这么玩的。所以如果你看Haskell
相关的书，他们一定会鼓吹说Haskell是惰性求值的，可以避免多余的计算。
