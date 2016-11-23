Python3函数参数中的星号
==========================

最近在看asyncio的代码，看到一个新的用法，查阅完文档之后，发现，在定义函数的参
数列表中，`*` 后的参数，调用者必须以 ``key=value`` 形式调用。因为平时看到的都
是 ``*args, **kwargs`` 形式的代码。

.. code:: python

    def _make_ssl_transport(self, rawsock, protocol, sslcontext, waiter=None,
                            *, server_side=False, server_hostname=None,
                            extra=None, server=None):
        """Create SSL transport."""
        raise NotImplementedError


错误示范::

    $ cat test.py
    def foo(this, *, loop):
        print("<foo> been called")

    foo(None, loop=None)
    foo(None, None)
    $ python test.py
    <foo> been called
    Traceback (most recent call last):
    File "test.py", line 6, in <module>
        foo(None, None)
    TypeError: foo() takes 1 positional argument but 2 were given


.. [#] http://stackoverflow.com/questions/14301967/python-bare-asterisk-in-function-argument

.. [#] https://docs.python.org/3/reference/compound_stmts.html#function-definitions
