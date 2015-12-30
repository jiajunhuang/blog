:Date: 2015-12-30

python import机制
=================

module
-------

import语句结合了两个操作，拿``import xxx``来说：

- `搜索xxx模块，搜索顺序是build-in module -> sys.path <https://docs.python.org/3/tutorial/modules.html#the-module-search-path>`__

      sys.path: https://docs.python.org/3/library/sys.html#sys.path

- 然后把搜索到的结果binding到本地作用域中(local scope)
  其中搜索动作由``__import__()``完成，如果搜索到了，就会初始化一个`module object`_.

.. _`module object`: https://docs.python.org/3/library/types.html#types.ModuleType

package
-------

- python只有一种类型的module object，无论模块是用C写的，用Python写的还是其他。
  为了把模块组织起来，并且提供命名空间，Python有一个 packages_ 的概念。

.. _packages: https://docs.python.org/3/glossary.html#term-package

- You can think of packages as the directories on a file system and modules
  as files within directories, but don’t take this analogy too literally
  since packages and modules need not originate from the file system.

- It’s important to keep in mind that all packages are modules, but not
  all modules are packages. Or put another way, packages are just a special
  kind of module. Specifically, any module that contains a ``__path__``
  attribute is considered a package.

- 访问子模块是把package和subpackage之间用`.`分隔。

    reguular packages:

    这是你看到的最多的package, Python3.2之前的package都是这样，例如::

        parent/
        __init__.py
        one/
            __init__.py
        two/
            __init__.py
        three/
            __init__.py

    如果执行``import parent.one``, 那么会执行``parent/__init__.py``和``parent/one/__init__,py``


    `namespace packages <https://www.python.org/dev/peps/pep-0420/>`__ TODO

searching
----------

- to begin the search, python needs the `fully qualified name <https://docs.python.org/3/glossary.html#term-qualified-name>`__.
  e.g. foo.bar.baz. In this case, Python first tries to import foo,
  then foo.bar, and finally foo.bar.baz. If any of the intermediate
  imports fail, an ImportError is raised.

- The first place checked during import search is `sys.modules <https://docs.python.org/3/library/sys.html#sys.modules>`__.

Special considerations for __main__
------------------------------------

The __main__ module is a special case relative to Python’s import system.
As noted elsewhere, the __main__ module is directly initialized at
interpreter startup, much like sys and builtins.

the import statement
--------------------

The basic import statement (no from clause) is executed in two steps:

1. find a module, loading and initializing it if necessary

#. define a name or names in the local namespace for the scope where the
   import statement occurs.

If the requested module is retrieved successfully, it will be made available
in the local namespace in one of three ways:

1. If the module name is followed by as, then the name following as is bound
directly to the imported module.

#. If no other name is specified, and the module being imported is a top
level module, the module’s name is bound in the local namespace as a
reference to the imported module

#. If the module being imported is not a top level module, then the
name of the top level package that contains the module is bound in
the local namespace as a reference to the top level package.
The imported module must be accessed using its full qualified name
rather than directly

the from statement(this paragrah is stolen from ``Python/Doc/reference/simple_stmts.rst``)
--------------------------------------------------------------------------------------

The `from` form uses a slightly more complex process:

#. find the module specified in the `from` clause, loading and
   initializing it if necessary;
#. for each of the identifiers specified in the `import` clauses:

   #. check if the imported module has an attribute by that name
   #. if not, attempt to import a submodule with that name and then
      check the imported module again for that attribute
   #. if the attribute is not found, `ImportError` is raised.
   #. otherwise, a reference to that value is stored in the local namespace,
      using the name in the `as` clause if it is present,
      otherwise using the attribute name

Examples::

   import foo                 # foo imported and bound locally
   import foo.bar.baz         # foo.bar.baz imported, foo bound locally
   import foo.bar.baz as fbb  # foo.bar.baz imported and bound as fbb
   from foo.bar import baz    # foo.bar.baz imported and bound as baz
   from foo import attr       # foo imported and foo.attr bound as attr

If the list of identifiers is replaced by a star (``'*'``), all public
names defined in the module are bound in the local namespace for the scope
where the `import` statement occurs.

.. index:: single: __all__ (optional module attribute)

The *public names* defined by a module are determined by checking the module's
namespace for a variable named ``__all__``; if defined, it must be a sequence
of strings which are names defined or imported by that module.  The names
given in ``__all__`` are all considered public and are required to exist.  If
``__all__`` is not defined, the set of public names includes all names found
in the module's namespace which do not begin with an underscore character
(``'_'``).  ``__all__`` should contain the entire public API. It is intended
to avoid accidentally exporting items that are not part of the API (such as
library modules which were imported and used within the module).

The wild card form of import --- ``from module import *`` --- is only allowed at
the module level.  Attempting to use it in class or function definitions will
raise a `SyntaxError`.
