# GCC默认的头文件搜索路径

```bash
$ cpp -v
Using built-in specs.
COLLECT_GCC=cpp
Target: x86_64-pc-linux-gnu
Configured with: /build/gcc/src/gcc/configure --prefix=/usr --libdir=/usr/lib --libexecdir=/usr/lib --mandir=/usr/share/man --infodir=/usr/share/info --with-bugurl=https://bugs.archlinux.org/ --enable-languages=c,c++,ada,fortran,go,lto,objc,obj-c++ --enable-shared --enable-threads=posix --with-system-zlib --with-isl --enable-__cxa_atexit --disable-libunwind-exceptions --enable-clocale=gnu --disable-libstdcxx-pch --disable-libssp --enable-gnu-unique-object --enable-linker-build-id --enable-lto --enable-plugin --enable-install-libiberty --with-linker-hash-style=gnu --enable-gnu-indirect-function --enable-multilib --disable-werror --enable-checking=release --enable-default-pie --enable-default-ssp --enable-cet=auto
Thread model: posix
gcc version 9.1.0 (GCC) 
COLLECT_GCC_OPTIONS='-E' '-v' '-mtune=generic' '-march=x86-64'
 /usr/lib/gcc/x86_64-pc-linux-gnu/9.1.0/cc1 -E -quiet -v - -mtune=generic -march=x86-64
ignoring nonexistent directory "/usr/lib/gcc/x86_64-pc-linux-gnu/9.1.0/../../../../x86_64-pc-linux-gnu/include"
#include "..." search starts here:
#include <...> search starts here:
 /usr/lib/gcc/x86_64-pc-linux-gnu/9.1.0/include
 /usr/local/include
 /usr/lib/gcc/x86_64-pc-linux-gnu/9.1.0/include-fixed
 /usr/include
End of search list.
^C
$
```

中显示出了gcc的搜索路径：

```
#include "..." search starts here:
#include <...> search starts here:
 /usr/lib/gcc/x86_64-pc-linux-gnu/9.1.0/include
 /usr/local/include
 /usr/lib/gcc/x86_64-pc-linux-gnu/9.1.0/include-fixed
 /usr/include
```

如果是 `#include "header.h"` 那么就是相对于当前文件进行搜索，如果是 `#include <header.h>` 那么就是从上述路径搜索。

---

参考资料：

- https://commandlinefanatic.com/cgi-bin/showarticle.cgi?article=art026
- https://gcc.gnu.org/onlinedocs/cpp/Search-Path.html
