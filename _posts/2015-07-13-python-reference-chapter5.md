---
layout: post
title: "Python reference note4 - The import system"
tags: [python]
---

* the `import` statement combines two operations; it searchs for the named 
module, then it bind the result of that search to a name in the local scope.
the search operation of the import statement is defined as a call to the 
`__import__()` funtion, with the appropriate arguments. the return value of 
`__import__()` is used to perform the name binding operation of the import 
statement.

* it's important to keep in mind that all packages are modules, but not all 
modules are packages.

* the first place checked during import search is `sys.module`. this mapping
serves as a cache of all modules that have been previously imported, including
the intermediate paths(e.g. `foo.bar.baz` -> `foo`, `foo.bar`, `foo.bar.baz`).

* if you keep a reference to the module object, invalidate its cache entry in 
`sys.modules`, and then re-import the named module. the two module objects
will not be the same.
