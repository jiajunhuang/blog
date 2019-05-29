# Python中的新式类(new style class)和老式类(old style class)

Python2.3之前，使用的是老式继承，直接看例子：

```python
>>> O = object
>>> class X(O): pass
>>> class Y(O): pass
>>> class A(X,Y): pass
>>> class B(Y,X): pass
```

这样下来，方法查找链就是这样的：

```
 -----------
|           |
|    O      |
|  /   \    |
 - X    Y  /
   |  / | /
   | /  |/
   A    B
   \   /
     ?
```

因此，不能再有一个新的类来继承 A 和 B，因为 A 的继承顺序是 X-Y，而 B 的继承顺序是 Y-X，那么到底是先在 X 里查找还是先在 Y 里查找呢？

为了解决这个问题，引入了 C3 MRO，还是以例子来说明：

```python
>>> O = object
>>> class F(O): pass
>>> class E(O): pass
>>> class D(O): pass
>>> class C(D,F): pass
>>> class B(D,E): pass
>>> class A(B,C): pass
```

那么方法查找链是这样：

```

                          6
                         ---
Level 3                 | O |                  (more general)
                      /  ---  \
                     /    |    \                      |
                    /     |     \                     |
                   /      |      \                    |
                  ---    ---    ---                   |
Level 2        3 | D | 4| E |  | F | 5                |
                  ---    ---    ---                   |
                   \  \ _ /       |                   |
                    \    / \ _    |                   |
                     \  /      \  |                   |
                      ---      ---                    |
Level 1            1 | B |    | C | 2                 |
                      ---      ---                    |
                        \      /                      |
                         \    /                      \ /
                           ---
Level 0                 0 | A |                (more specialized)
                           ---
```

计算的时候就是：

```
L[O] = O
L[D] = D O
L[E] = E O
L[F] = F O
L[B] = B + merge(DO, EO, DE)
```

规则就是，以继承时的声明为顺序，每次取方法查找链的头一个，如果这个头不在后面的方法查找链的尾部，那么就把他放到方法查找链
里，首先方法查找肯定是在 `B` 里进行，然后是 `merge(DO, EO, DE)`，`D` 是一个好的节点，因为 `D` 不在 `DO, EO, DE` 的尾部。然后是
`O，O` 在 `EO` 的尾部。然后是 `E` ，然后是 `O` 。

所以最后方法查找链就是 `B -> D -> E -> O`。

同样，拿上面的例子来看，C3 MRO的查找顺序就应该是 `A -> X -> Y -> B -> O`

---

参考资料：

- https://www.python.org/download/releases/2.3/mro/
- https://www.python.org/doc/newstyle/
