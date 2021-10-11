# gevent不是黑魔法(二): gevent 实现

上一篇我们说了，gevent 是基于 greenlet，结合 event loop 实现的，这一篇
我们就来看看 gevent 是如何实现的。

首先我们来看一个简单的例子，来自 [Gevent tutorial](https://sdiehl.github.io/gevent-tutorial/)：

```python
import gevent

def foo():
    print('Running in foo')
    gevent.sleep(0)
    print('Explicit context switch to foo again')

def bar():
    print('Explicit context to bar')
    gevent.sleep(0)
    print('Implicit context switch back to bar')

gevent.joinall([
    gevent.spawn(foo),
    gevent.spawn(bar),
])
```

执行一下：

```bash
$ python main.py 
Running in foo
Explicit context to bar
Explicit context switch to foo again
Implicit context switch back to bar

```

从输出可以看到，执行的顺序是：

- 首先进入了 foo，然后执行 foo 函数第一行
- 执行 foo 函数的第二行 `gevent.sleep(0)`
- 进入 bar 函数，执行 bar 函数的第一行
- 执行 bar 函数的第二行 `gevent.sleep(0)`
- 此时再次进入 foo 函数，但是不是从头开始执行，而是接着执行 foo 函数的第三行，随即退出 foo 函数
- 最后进入 bar 函数，同样，也是从上一次没执行完的地方开始，执行了 bar 函数的第三行，随即退出

我们可以看到，使用了 gevent 库里提供的函数之后，代码就自动使用了协程，在有
`gevent.sleep` 的地方自动让出执行权。

很明显，我们的关注点在于 `gevent.sleep` 和 `gevent.joinall`，那么有什么办法可以
看到这两个函数的代码呢？有以下几种办法：

- 看 `__code__` 属性：The code object representing the compiled function body
- 看 `__module__` 属性：The name of the module the function was defined in, or None if unavailable
- 看 `__doc__` 属性：The function’s documentation string, or None if unavailable; not inherited by subclasses
- inspect.getmodulename, inspect.getmodule, inspect.getfile, inspect.getsourcelines 等，但是这个模块只能看到 Python 写的代码，看不到 C 模块。而 gevent 是 Cython + C + Python 混编的。所以这个模块只能作为辅助
- iPython 的 `?` 和 `??` 可以提供函数的实现和所在文件位置
- 还有一个就是搜索了，可以用 grep，ack 或者是 ag

先来看 `gevent.sleep`：

```python
In [1]: import gevent

In [2]: gevent.sleep.__code__
Out[2]: <code object sleep at 0x7f14ce917df0, file "/home/jiajun/.pyenv/versions/3.9.5/lib/python3.9/site-packages/gevent/hub.py", line 129>

In [3]:
```

所以我们就去 `hub.py` 里找：

```python
def sleep(seconds=0, ref=True):
    """
    Put the current greenlet to sleep for at least *seconds*.

    *seconds* may be specified as an integer, or a float if fractional
    seconds are desired.

    .. tip:: In the current implementation, a value of 0 (the default)
       means to yield execution to any other runnable greenlets, but
       this greenlet may be scheduled again before the event loop
       cycles (in an extreme case, a greenlet that repeatedly sleeps
       with 0 can prevent greenlets that are ready to do I/O from
       being scheduled for some (small) period of time); a value greater than
       0, on the other hand, will delay running this greenlet until
       the next iteration of the loop.

    If *ref* is False, the greenlet running ``sleep()`` will not prevent :func:`gevent.wait`
    from exiting.

    .. versionchanged:: 1.3a1
       Sleeping with a value of 0 will now be bounded to approximately block the
       loop for no longer than :func:`gevent.getswitchinterval`.

    .. seealso:: :func:`idle`
    """
    hub = _get_hub_noargs()
    loop = hub.loop
    if seconds <= 0:
        waiter = Waiter(hub)
        loop.run_callback(waiter.switch, None)
        waiter.get()
    else:
        with loop.timer(seconds, ref=ref) as t:
            # Sleeping is expected to be an "absolute" measure with
            # respect to time.time(), not a relative measure, so it's
            # important to update the loop's notion of now before we start
            loop.update_now()
            hub.wait(t)
```

先看注释，说这个函数，就是把当前正在执行的 greenlet 休眠至少 seconds 秒。接着我们来看代码，
首先获取 hub，然后获取 loop，咱也不知道这两个是啥，先不管。

然后是一个判断，如果 seconds 小于等于 0，那说明不需要 sleep 了，就创建一个 Waiter 对象，并且执行两行代码，
否则，就创建一个 loop.timer 然后执行 `hub.wait`。

看起来 hub 就是很重要的东西，但是咱也不知道，咱也不敢问，毕竟是看到一份陌生的代码，只能大胆假设，小心求证了。

接下来，继续看一下 `gevent.joinall`：

```python
In [1]: import gevent

In [2]: gevent.joinall?
Signature:      gevent.joinall(greenlets, timeout, raise_error, count)
Call signature: gevent.joinall(*args, **kwargs)
Type:           cython_function_or_method
String form:    <cyfunction joinall at 0x7fd5a87ba040>
Docstring:     
joinall(greenlets, timeout=None, raise_error=False, count=None)

Wait for the ``greenlets`` to finish.

:param greenlets: A sequence (supporting :func:`len`) of greenlets to wait for.
:keyword float timeout: If given, the maximum number of seconds to wait.
:return: A sequence of the greenlets that finished before the timeout (if any)
    expired.

In [3]: gevent.joinall.__code__
Out[3]: <code object joinall at 0x7fd5a87b2710, file "src/gevent/greenlet.py", line 1057>

In [4]:
```

我们就去 `greenlet.py` 里找：

```python
def joinall(greenlets, timeout=None, raise_error=False, count=None):
    """
    Wait for the ``greenlets`` to finish.

    :param greenlets: A sequence (supporting :func:`len`) of greenlets to wait for.
    :keyword float timeout: If given, the maximum number of seconds to wait.
    :return: A sequence of the greenlets that finished before the timeout (if any)
        expired.
    """
    if not raise_error:
        return wait(greenlets, timeout=timeout, count=count)

    done = []
    for obj in iwait(greenlets, timeout=timeout, count=count):
        if getattr(obj, 'exception', None) is not None:
            if hasattr(obj, '_raise_exception'):
                obj._raise_exception()
            else:
                raise obj.exception
        done.append(obj)
    return done
```

先看注释，说这个函数，就是等待传入的所有 greenlets 完成，然后退出，默认参数
`raise_error=False`，所以我们上面的例子，会执行函数的第一个分支，我们就去看看
`wait(greenlets, timeout=timeout, count=count)` 里都做了什么，由于 `wait` 是
`src/gevent/_hub_primitives.py` 里 `wait_on_objects` 的别名，所以我们直接看后者：

```python
def wait_on_objects(objects=None, timeout=None, count=None):
    """
    Wait for ``objects`` to become ready or for event loop to finish.

    If ``objects`` is provided, it must be a list containing objects
    implementing the wait protocol (rawlink() and unlink() methods):

    - :class:`gevent.Greenlet` instance
    - :class:`gevent.event.Event` instance
    - :class:`gevent.lock.Semaphore` instance
    - :class:`gevent.subprocess.Popen` instance

    If ``objects`` is ``None`` (the default), ``wait()`` blocks until
    the current event loop has nothing to do (or until ``timeout`` passes):

    - all greenlets have finished
    - all servers were stopped
    - all event loop watchers were stopped.

    If ``count`` is ``None`` (the default), wait for all ``objects``
    to become ready.

    If ``count`` is a number, wait for (up to) ``count`` objects to become
    ready. (For example, if count is ``1`` then the function exits
    when any object in the list is ready).

    If ``timeout`` is provided, it specifies the maximum number of
    seconds ``wait()`` will block.

    Returns the list of ready objects, in the order in which they were
    ready.

    .. seealso:: :func:`iwait`
    """
    if objects is None:
        hub = get_hub()
        return hub.join(timeout=timeout) # pylint:disable=
    return list(iwait_on_objects(objects, timeout, count))
```

先看注释，注释说明了这个函数做了什么，要做什么。然后看代码，在我们的例子里，
`objects` 明显不是 `None`，所以执行的是 `list(iwait_on_objects(objects, timeout, count))`，
但是值得一提的是，这里再次看到了 `hub` 的身影。

```python
def iwait_on_objects(objects, timeout=None, count=None):
    """
    Iteratively yield *objects* as they are ready, until all (or *count*) are ready
    or *timeout* expired.

    If you will only be consuming a portion of the *objects*, you should
    do so inside a ``with`` block on this object to avoid leaking resources::

        with gevent.iwait((a, b, c)) as it:
            for i in it:
                if i is a:
                    break

    :param objects: A sequence (supporting :func:`len`) containing objects
        implementing the wait protocol (rawlink() and unlink()).
    :keyword int count: If not `None`, then a number specifying the maximum number
        of objects to wait for. If ``None`` (the default), all objects
        are waited for.
    :keyword float timeout: If given, specifies a maximum number of seconds
        to wait. If the timeout expires before the desired waited-for objects
        are available, then this method returns immediately.

    .. seealso:: :func:`wait`

    .. versionchanged:: 1.1a1
       Add the *count* parameter.
    .. versionchanged:: 1.1a2
       No longer raise :exc:`LoopExit` if our caller switches greenlets
       in between items yielded by this function.
    .. versionchanged:: 1.4
       Add support to use the returned object as a context manager.
    """
    # QQQ would be nice to support iterable here that can be generated slowly (why?)
    hub = get_hub()
    if objects is None:
        return [hub.join(timeout=timeout)]
    return _WaitIterator(objects, hub, timeout, count)
```

这里我们看到了，之所以叫做 `iwait xxx`，就是因为他们是使用了 iterator 的版本，
所以我们重点关注 `_WaitIterator` 的 `__init__`, `__iter__`, `__next__` 方法都做了什么：

```python
class _WaitIterator(object):
    def __init__(self, objects, hub, timeout, count):
        self._hub = hub
        self._waiter = MultipleWaiter(hub) # pylint:disable=undefined-variable
        self._switch = self._waiter.switch
        self._timeout = timeout
        self._objects = objects

        self._timer = None
        self._begun = False

        # Even if we're only going to return 1 object,
        # we must still rawlink() *all* of them, so that no
        # matter which one finishes first we find it.
        self._count = len(objects) if count is None else min(count, len(objects))

    def _begin(self):
        if self._begun:
            return

        self._begun = True

        # XXX: If iteration doesn't actually happen, we
        # could leave these links around!
        for obj in self._objects:
            obj.rawlink(self._switch)

        if self._timeout is not None:
            self._timer = self._hub.loop.timer(self._timeout, priority=-1)
            self._timer.start(self._switch, self)

    def __iter__(self):
        return self

    def __next__(self):
        self._begin()

        if self._count == 0:
            # Exhausted
            self._cleanup()
            raise StopIteration()

        self._count -= 1
        try:
            item = self._waiter.get()
            self._waiter.clear()
            if item is self:
                # Timer expired, no more
                self._cleanup()
                raise StopIteration()
            return item
        except:
            self._cleanup()
            raise

    # 其余略
```

`__init__` 里，主要就是初始化工作，我们稍微记住一下都初始化了什么。然后是 `__next__`
的一开始，就调用了 `self._begin`，这个的最后一句，是说如果 `timeout` 不为空，那么时间到了
就执行 `self._switch`，而 `self._switch` 在 `__init__` 已经写明了，是
`self._waiter.switch`，`self._waiter` 是 `MultipleWaiter(hub)` 的实例。追踪一下，发现
`MultipleWaiter` 就是 `src/gevent/_waiter.py` 中的 `MultipleWaiter`。

到这里，我们先打住，继续看也很难看出什么来。到目前为止，我们发现一个非常重要的东西，那就是 `hub`，
基本上到处都和这个 `hub` 有关。这个时候我们再回过头来，看看 `hub` 到底是什么，首先我们要看，怎么拿到的 `hub`，翻看上面的代码：

```python
hub = _get_hub_noargs()
hub = get_hub()
```

我们来看看这两个函数的实现：

```python
def get_hub(*args, **kwargs): # pylint:disable=unused-argument
    """
    Return the hub for the current thread.

    If a hub does not exist in the current thread, a new one is
    created of the type returned by :func:`get_hub_class`.

    .. deprecated:: 1.3b1
       The ``*args`` and ``**kwargs`` arguments are deprecated. They were
       only used when the hub was created, and so were non-deterministic---to be
       sure they were used, *all* callers had to pass them, or they were order-dependent.
       Use ``set_hub`` instead.

    .. versionchanged:: 1.5a3
       The *args* and *kwargs* arguments are now completely ignored.
    """

    return get_hub_noargs()

def get_hub_noargs():
    # Just like get_hub, but cheaper to call because it
    # takes no arguments or kwargs. See also a copy in
    # gevent/greenlet.py
    hub = _threadlocal.hub
    if hub is None:
        hubtype = get_hub_class()
        hub = _threadlocal.hub = hubtype()
    return hub

def get_hub_class():
    """Return the type of hub to use for the current thread.

    If there's no type of hub for the current thread yet, 'gevent.hub.Hub' is used.
    """
    hubtype = _threadlocal.Hub
    if hubtype is None:
        hubtype = _threadlocal.Hub = Hub
    return hubtype
```

绕了半天也看不出 Hub 是什么，在哪里，但是别忘了，我们研究的可是 Python，既然直接看代码
看不出来，那直接运行一下呀：

```python
In [1]: import gevent

In [2]: hub = gevent.get_hub()

In [3]: hub?
Type:        Hub
String form: <Hub '' at 0x7fa538159a00 epoll default pending=0 ref=0 fileno=12 thread_ident=0x7fa53aade740>
File:        ~/.pyenv/versions/3.9.5/lib/python3.9/site-packages/gevent/hub.py
Docstring:  
A greenlet that runs the event loop.

It is created automatically by :func:`get_hub`.

.. rubric:: Switching

Every time this greenlet (i.e., the event loop) is switched *to*,
if the current greenlet has a ``switch_out`` method, it will be
called. This allows a greenlet to take some cleanup actions before
yielding control. This method should not call any gevent blocking
functions.

In [4]: hub.__class__?
Init signature: hub.__class__(loop=None, default=None)
Docstring:     
A greenlet that runs the event loop.

It is created automatically by :func:`get_hub`.

.. rubric:: Switching

Every time this greenlet (i.e., the event loop) is switched *to*,
if the current greenlet has a ``switch_out`` method, it will be
called. This allows a greenlet to take some cleanup actions before
yielding control. This method should not call any gevent blocking
functions.
File:           ~/.pyenv/versions/3.9.5/lib/python3.9/site-packages/gevent/hub.py
Type:           type
Subclasses:     

In [5]:
```

看来一切奥秘，都在 `src/gevent/hub.py` 里：

```python
class Hub(WaitOperationsGreenlet):
    """
    A greenlet that runs the event loop.

    It is created automatically by :func:`get_hub`.

    .. rubric:: Switching

    Every time this greenlet (i.e., the event loop) is switched *to*,
    if the current greenlet has a ``switch_out`` method, it will be
    called. This allows a greenlet to take some cleanup actions before
    yielding control. This method should not call any gevent blocking
    functions.
    """
    # 略


set_default_hub_class(Hub)  # 导入 `gevent.hub` 的时候就会设置好 Hub 的类型
```

注释已经说的很清楚了：

- Hub 是一个 greenlet
- Hub 是带有 event loop 的 greenlet

从代码中可以看到，`hub.loop` 就是事件循环，我们来看下是什么：

```python
In [1]: import gevent

In [2]: hub = gevent.get_hub()

In [3]: hub.loop?
Type:        loop
String form: <loop at 0x7f48460519e0 epoll default pending=0 ref=0 fileno=12>
File:        ~/.pyenv/versions/3.9.5/lib/python3.9/site-packages/gevent/libev/corecext.cpython-39-x86_64-linux-gnu.so
Docstring:   <no docstring>

In [4]: type(hub.loop)
Out[4]: gevent.libev.corecext.loop

In [5]:
```

看来 `gevent.loop` 是 libev。对 libev，我读了一下文档，了解了一下大概，文档链接在 [这里](https://manpages.ubuntu.com/manpages/cosmic/man3/EV::libev.3pm.html)，
读者需要了解的就是，libev 是一个对事件循环的封装库，不仅封装了 I/O，而且还封装了
timer, signal 等，建议读者去读一下文档，作者很风趣，吐槽了一大堆的技术，如国内吹上天的 epoll，MacOS。

但是在这里，我们只需要知道上面这些即可。我们了解了 Hub 是什么之后，就可以继续我们最开始的分析了。刚才看到了 `MultipleWaiter`：

```python
class MultipleWaiter(Waiter):
    """
    An internal extension of Waiter that can be used if multiple objects
    must be waited on, and there is a chance that in between waits greenlets
    might be switched out. All greenlets that switch to this waiter
    will have their value returned.

    This does not handle exceptions or throw methods.
    """
    __slots__ = ['_values']

    def __init__(self, hub=None):
        Waiter.__init__(self, hub)
        # we typically expect a relatively small number of these to be outstanding.
        # since we pop from the left, a deque might be slightly
        # more efficient, but since we're in the hub we avoid imports if
        # we can help it to better support monkey-patching, and delaying the import
        # here can be impractical (see https://github.com/gevent/gevent/issues/652)
        self._values = []

    def switch(self, value):
        self._values.append(value)
        Waiter.switch(self, True)

    def get(self):
        if not self._values:
            Waiter.get(self)
            Waiter.clear(self)

        return self._values.pop(0)
```

`MultipleWaiter` 本质上还是对 `Waiter` 的继承和扩展，支持等待多个 greenlet：

```python
class Waiter(object):
    """
    A low level communication utility for greenlets.

    Waiter is a wrapper around greenlet's ``switch()`` and ``throw()`` calls that makes them somewhat safer:

    * switching will occur only if the waiting greenlet is executing :meth:`get` method currently;
    * any error raised in the greenlet is handled inside :meth:`switch` and :meth:`throw`
    * if :meth:`switch`/:meth:`throw` is called before the receiver calls :meth:`get`, then :class:`Waiter`
      will store the value/exception. The following :meth:`get` will return the value/raise the exception.

    The :meth:`switch` and :meth:`throw` methods must only be called from the :class:`Hub` greenlet.
    The :meth:`get` method must be called from a greenlet other than :class:`Hub`.

        >>> from gevent.hub import Waiter
        >>> from gevent import get_hub
        >>> result = Waiter()
        >>> timer = get_hub().loop.timer(0.1)
        >>> timer.start(result.switch, 'hello from Waiter')
        >>> result.get() # blocks for 0.1 seconds
        'hello from Waiter'
        >>> timer.close()

    If switch is called before the greenlet gets a chance to call :meth:`get` then
    :class:`Waiter` stores the value.

        >>> from gevent.time import sleep
        >>> result = Waiter()
        >>> timer = get_hub().loop.timer(0.1)
        >>> timer.start(result.switch, 'hi from Waiter')
        >>> sleep(0.2)
        >>> result.get() # returns immediately without blocking
        'hi from Waiter'
        >>> timer.close()

    .. warning::

        This is a limited and dangerous way to communicate between
        greenlets. It can easily leave a greenlet unscheduled forever
        if used incorrectly. Consider using safer classes such as
        :class:`gevent.event.Event`, :class:`gevent.event.AsyncResult`,
        or :class:`gevent.queue.Queue`.
    """

    def __init__(self, hub=None):
        self.hub = get_hub() if hub is None else hub
        self.greenlet = None
        self.value = None
        self._exception = _NONE

    def switch(self, value):
        """
        Switch to the greenlet if one's available. Otherwise store the
        *value*.

        .. versionchanged:: 1.3b1
           The *value* is no longer optional.
        """
        greenlet = self.greenlet
        if greenlet is None:
            self.value = value
            self._exception = None
        else:
            if getcurrent() is not self.hub: # pylint:disable=undefined-variable
                raise AssertionError("Can only use Waiter.switch method from the Hub greenlet")
            switch = greenlet.switch
            try:
                switch(value)
            except: # pylint:disable=bare-except
                self.hub.handle_error(switch, *sys.exc_info())

    def get(self):
        """If a value/an exception is stored, return/raise it. Otherwise until switch() or throw() is called."""
        if self._exception is not _NONE:
            if self._exception is None:
                return self.value
            getcurrent().throw(*self._exception) # pylint:disable=undefined-variable
        else:
            if self.greenlet is not None:
                raise ConcurrentObjectUseError('This Waiter is already used by %r' % (self.greenlet, ))
            self.greenlet = getcurrent() # pylint:disable=undefined-variable
            try:
                return self.hub.switch()
            finally:
                self.greenlet = None
```

很重要的一个方法，就是 `Waiter.switch`，我们来看看它做了啥：

```python
    def switch(self, value):
        """
        Switch to the greenlet if one's available. Otherwise store the
        *value*. 首先看注释，说如果当前 greenlet 可用，切换执行它，否则把 value 存储下来。

        .. versionchanged:: 1.3b1
           The *value* is no longer optional.
        """
        greenlet = self.greenlet  # 尝试找到 self.greenlet，但是 __init__ 的时候，其实是设置成了 None
        if greenlet is None:
            self.value = value  # 如果是None，就保存，如注释所说
            self._exception = None
        else:
            if getcurrent() is not self.hub: # pylint:disable=undefined-variable
                raise AssertionError("Can only use Waiter.switch method from the Hub greenlet")
            switch = greenlet.switch
            try:  # 否则，就切换到当前的 greenlet
                switch(value)
            except: # pylint:disable=bare-except
                self.hub.handle_error(switch, *sys.exc_info())
```

怎么理解呢？我建议还是需要结合注释中的例子来理解：

```bash
$ python
>>> from gevent.hub import Waiter
>>> from gevent import get_hub
>>> result = Waiter()  # 初始化
>>> timer = get_hub().loop.timer(0.1)
>>> timer.start(result.switch, 'hello from Waiter') # 设置时间到期了就执行 `result.switch`
>>> result.get() # blocks for 0.1 seconds  # 说此处会阻塞 0.1 秒
'hello from Waiter'
>>> timer.close()
```

我们来看下 `Waiter.get`：

```python
    def get(self):
        """If a value/an exception is stored, return/raise it. Otherwise until switch() or throw() is called."""
        if self._exception is not _NONE:  # 说明已经有值存储了
            if self._exception is None:  # 无异常，则返回
                return self.value
            getcurrent().throw(*self._exception) # pylint:disable=undefined-variable  # 有异常，则抛出
        else:
            if self.greenlet is not None:  # 一个检查
                raise ConcurrentObjectUseError('This Waiter is already used by %r' % (self.greenlet, ))
            self.greenlet = getcurrent() # pylint:disable=undefined-variable  # 获取当前 greenlet
            try:
                return self.hub.switch()  # 切换到 `self.hub` 执行
            finally:
                self.greenlet = None
```

上述代码，由于设置 timer 之后，立即执行了 `result.get`，所以实际上执行的是 `else` 分支，也就是说切换到 hub
里去了，而 0.1 秒之后，hub 里 loop的timer 时间到了，唤醒并且执行 `Waiter.switch` 函数，最终执行了 `greenlet.switch`
函数，输出了返回值。

到这里，我们基本上就可以理解 gevent 的实现原理了，首先，有一个东西叫做 Hub，Hub也不是什么很特殊的东西，它就是一个
greenlet，也就是协程。但是它也有一个特点，就是它带了一个事件循环，默认情况下，使用的就是 libev。当我们遇到了一些
会阻塞当前协程的函数时，由于我们调用的是 gevent 提供的实现版本，它在底层实际上是封装成了 libev 所支持的 watcher，
然后切换到hub来执行，而上一篇我们看到了，greenlet 执行之后，是会不断的去找可以执行的其它 greenlet 来执行的。

就这样，gevent 结合 greenlet 和 event loop，实现了一套写起来同步，看起来是阻塞，单实际执行起来却是异步的非阻塞的
协程库。但是使用的时候有一个限制条件，那就是必须使用 gevent 提供的实现，例如 `gevent.sleep`，`gevent.joinall` 等，
一旦使用标准库自带的 sleep 等，就会出问题。也正是因此，gevent提供了 monkey patch，当然，这也是 gevent 被称作黑魔法
的主要原因。

## 总结

通过这一篇文章，我们了解了 gevent 是如何基于 greenlet 和 event loop 实现的一套协程网络库，这对我们使用 gevent 起到了
充实信心的作用，毕竟了解了底层原理，也就知道 gevent 的实现，是不是真的如传说中的黑魔法那般碰不得。我们通过两篇文章，
第一篇了解了 gevent 的基础，也就是 greenlet，第二篇结合 event loop 一起看 gevent 是如何实现异步非阻塞的高并发网络库的。
相信这两篇文章会对读者带来帮助。


---

ref:

- https://sdiehl.github.io/gevent-tutorial/
- https://manpages.ubuntu.com/manpages/cosmic/man3/EV::libev.3pm.html
