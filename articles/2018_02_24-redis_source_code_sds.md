# Redis源码阅读与分析一：sds

- sds的定义： `typedef char *sds;`。所以sds是指向一个C字符串的指针。

- sds所指向的指针之前有一个头部：

```c
struct __attribute__ ((__packed__)) sdshdr5 {
    unsigned char flags; /* 3 lsb of type, and 5 msb of string length */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr8 {
    uint8_t len; /* used */
    uint8_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr16 {
    uint16_t len; /* used */
    uint16_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr32 {
    uint32_t len; /* used */
    uint32_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
struct __attribute__ ((__packed__)) sdshdr64 {
    uint64_t len; /* used */
    uint64_t alloc; /* excluding the header and null terminator */
    unsigned char flags; /* 3 lsb of type, 5 unused bits */
    char buf[];
};
```

其中 `struct __attribute__ ((__packed__))` 的意思是不要内存对齐，使用紧凑模式。
此外 `char buf[]` 是柔性数组，在结构体的最后，此前需要至少一个成员。其大小不计入结构体：

```bash
$ cat main.c 
#include <stdio.h>

struct Foo {
    int i;
    char buf[];
};

int main() {
    printf("size of int: %lu\n", sizeof(int));
    printf("size of struct Foo: %lu\n", sizeof(struct Foo));
}
$ cc main.c && ./a.out 
size of int: 4
size of struct Foo: 4
```

- sds往前挪一个byte就是flags，用来判断它到底是什么类型的。其实sds v1不是这么设计的，而是简单的统一用一个
结构体。v2这样可以节省内存。不同大小的字符串就用不同大小的头部。优点在于节省内存，缺点在于操作麻烦。
可以看到代码中很多地方访问flags是这样访问的：

```c
static inline size_t sdslen(const sds s) {
    unsigned char flags = s[-1];
    switch(flags&SDS_TYPE_MASK) {
        case SDS_TYPE_5:
            return SDS_TYPE_5_LEN(flags);
        case SDS_TYPE_8:
            return SDS_HDR(8,s)->len;
        case SDS_TYPE_16:
            return SDS_HDR(16,s)->len;
        case SDS_TYPE_32:
            return SDS_HDR(32,s)->len;
        case SDS_TYPE_64:
            return SDS_HDR(64,s)->len;
    }
    return 0;
}
```

flags的前三个bit用来表示是何种大小的头部，后5个bit暂时还没有用上。

```c
#define SDS_TYPE_5  0
#define SDS_TYPE_8  1
#define SDS_TYPE_16 2
#define SDS_TYPE_32 3
#define SDS_TYPE_64 4
#define SDS_TYPE_MASK 7
#define SDS_TYPE_BITS 3
#define SDS_HDR_VAR(T,s) struct sdshdr##T *sh = (void*)((s)-(sizeof(struct sdshdr##T)));
#define SDS_HDR(T,s) ((struct sdshdr##T *)((s)-(sizeof(struct sdshdr##T))))
#define SDS_TYPE_5_LEN(f) ((f)>>SDS_TYPE_BITS)
```

- sds相比与原生的C字符串不同之处在于：

    - 保存了长度
    - 可以动态扩展
    - 二进制安全：因为不依赖字符串内部的某位来判断是否结束，而是依赖头部里保存的长度
