# Golang defer中修改返回值

我们来看demo：

```go
package main

func foo() (i int) {
	defer func() { i++ }()
	return
}

func bar() (i int) {
	i++
	return
}

func baz() int {
	i := 0
	i++
	return i
}

func main() {
	foo()
	bar()
	baz()
}
```

可以看一下汇编码，`-N` 是指禁止编译器优化，`-S` 是指打印出汇编码。

```go
$ go tool compile -N -S main.go
"".foo STEXT size=117 args=0x8 locals=0x20
	0x0000 00000 (main.go:3)	TEXT	"".foo(SB), $32-8
	0x0000 00000 (main.go:3)	MOVQ	(TLS), CX
	0x0009 00009 (main.go:3)	CMPQ	SP, 16(CX)
	0x000d 00013 (main.go:3)	JLS	110
	0x000f 00015 (main.go:3)	SUBQ	$32, SP
	0x0013 00019 (main.go:3)	MOVQ	BP, 24(SP)
	0x0018 00024 (main.go:3)	LEAQ	24(SP), BP
	0x001d 00029 (main.go:3)	FUNCDATA	$0, gclocals·2a5305abe05176240e61b8620e19a815(SB)
	0x001d 00029 (main.go:3)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x001d 00029 (main.go:3)	MOVQ	$0, "".i+40(SP)
	0x0026 00038 (main.go:4)	LEAQ	"".i+40(SP), AX
	0x002b 00043 (main.go:4)	MOVQ	AX, 16(SP)
	0x0030 00048 (main.go:4)	MOVL	$8, (SP)
	0x0037 00055 (main.go:4)	LEAQ	"".foo.func1·f(SB), AX
	0x003e 00062 (main.go:4)	MOVQ	AX, 8(SP)
	0x0043 00067 (main.go:4)	PCDATA	$0, $0
	0x0043 00067 (main.go:4)	CALL	runtime.deferproc(SB)
	0x0048 00072 (main.go:4)	TESTL	AX, AX
	0x004a 00074 (main.go:4)	JNE	94
	0x004c 00076 (main.go:4)	JMP	78
	0x004e 00078 (main.go:5)	PCDATA	$0, $0
	0x004e 00078 (main.go:5)	XCHGL	AX, AX
	0x004f 00079 (main.go:5)	CALL	runtime.deferreturn(SB)
	0x0054 00084 (main.go:5)	MOVQ	24(SP), BP
	0x0059 00089 (main.go:5)	ADDQ	$32, SP
	0x005d 00093 (main.go:5)	RET
	0x005e 00094 (main.go:4)	PCDATA	$0, $0
	0x005e 00094 (main.go:4)	XCHGL	AX, AX
	0x005f 00095 (main.go:4)	CALL	runtime.deferreturn(SB)
	0x0064 00100 (main.go:4)	MOVQ	24(SP), BP
	0x0069 00105 (main.go:4)	ADDQ	$32, SP
	0x006d 00109 (main.go:4)	RET
	0x006e 00110 (main.go:4)	NOP
	0x006e 00110 (main.go:3)	PCDATA	$0, $-1
	0x006e 00110 (main.go:3)	CALL	runtime.morestack_noctxt(SB)
	0x0073 00115 (main.go:3)	JMP	0
	0x0000 65 48 8b 0c 25 00 00 00 00 48 3b 61 10 76 5f 48  eH..%....H;a.v_H
	0x0010 83 ec 20 48 89 6c 24 18 48 8d 6c 24 18 48 c7 44  .. H.l$.H.l$.H.D
	0x0020 24 28 00 00 00 00 48 8d 44 24 28 48 89 44 24 10  $(....H.D$(H.D$.
	0x0030 c7 04 24 08 00 00 00 48 8d 05 00 00 00 00 48 89  ..$....H......H.
	0x0040 44 24 08 e8 00 00 00 00 85 c0 75 12 eb 00 90 e8  D$........u.....
	0x0050 00 00 00 00 48 8b 6c 24 18 48 83 c4 20 c3 90 e8  ....H.l$.H.. ...
	0x0060 00 00 00 00 48 8b 6c 24 18 48 83 c4 20 c3 e8 00  ....H.l$.H.. ...
	0x0070 00 00 00 eb 8b                                   .....
	rel 5+4 t=16 TLS+0
	rel 58+4 t=15 "".foo.func1·f+0
	rel 68+4 t=8 runtime.deferproc+0
	rel 80+4 t=8 runtime.deferreturn+0
	rel 96+4 t=8 runtime.deferreturn+0
	rel 111+4 t=8 runtime.morestack_noctxt+0
"".bar STEXT nosplit size=50 args=0x8 locals=0x10
	0x0000 00000 (main.go:8)	TEXT	"".bar(SB), NOSPLIT, $16-8
	0x0000 00000 (main.go:8)	SUBQ	$16, SP
	0x0004 00004 (main.go:8)	MOVQ	BP, 8(SP)
	0x0009 00009 (main.go:8)	LEAQ	8(SP), BP
	0x000e 00014 (main.go:8)	FUNCDATA	$0, gclocals·2a5305abe05176240e61b8620e19a815(SB)
	0x000e 00014 (main.go:8)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (main.go:8)	MOVQ	$0, "".i+24(SP)
	0x0017 00023 (main.go:9)	MOVQ	$0, ""..autotmp_1(SP)
	0x001f 00031 (main.go:9)	MOVQ	$1, "".i+24(SP)
	0x0028 00040 (main.go:10)	MOVQ	8(SP), BP
	0x002d 00045 (main.go:10)	ADDQ	$16, SP
	0x0031 00049 (main.go:10)	RET
	0x0000 48 83 ec 10 48 89 6c 24 08 48 8d 6c 24 08 48 c7  H...H.l$.H.l$.H.
	0x0010 44 24 18 00 00 00 00 48 c7 04 24 00 00 00 00 48  D$.....H..$....H
	0x0020 c7 44 24 18 01 00 00 00 48 8b 6c 24 08 48 83 c4  .D$.....H.l$.H..
	0x0030 10 c3                                            ..
"".baz STEXT nosplit size=67 args=0x8 locals=0x18
	0x0000 00000 (main.go:13)	TEXT	"".baz(SB), NOSPLIT, $24-8
	0x0000 00000 (main.go:13)	SUBQ	$24, SP
	0x0004 00004 (main.go:13)	MOVQ	BP, 16(SP)
	0x0009 00009 (main.go:13)	LEAQ	16(SP), BP
	0x000e 00014 (main.go:13)	FUNCDATA	$0, gclocals·2a5305abe05176240e61b8620e19a815(SB)
	0x000e 00014 (main.go:13)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (main.go:13)	MOVQ	$0, "".~r0+32(SP)
	0x0017 00023 (main.go:14)	MOVQ	$0, "".i(SP)
	0x001f 00031 (main.go:15)	MOVQ	$0, ""..autotmp_2+8(SP)
	0x0028 00040 (main.go:15)	MOVQ	$1, "".i(SP)
	0x0030 00048 (main.go:16)	MOVQ	"".i(SP), AX
	0x0034 00052 (main.go:16)	MOVQ	AX, "".~r0+32(SP)
	0x0039 00057 (main.go:16)	MOVQ	16(SP), BP
	0x003e 00062 (main.go:16)	ADDQ	$24, SP
	0x0042 00066 (main.go:16)	RET
	0x0000 48 83 ec 18 48 89 6c 24 10 48 8d 6c 24 10 48 c7  H...H.l$.H.l$.H.
	0x0010 44 24 20 00 00 00 00 48 c7 04 24 00 00 00 00 48  D$ ....H..$....H
	0x0020 c7 44 24 08 00 00 00 00 48 c7 04 24 01 00 00 00  .D$.....H..$....
	0x0030 48 8b 04 24 48 89 44 24 20 48 8b 6c 24 10 48 83  H..$H.D$ H.l$.H.
	0x0040 c4 18 c3                                         ...
```

参考:

- https://golang.org/doc/effective_go.html#named-results
- https://golang.org/doc/asm

问我为什么不讲解一下上面的汇编代码？因为我也看不懂汇编-。-。不过可以连蒙带猜。

结合文档，named return values会在函数开始时自动声明。而匿名返回值则是相当于自动生成了一个变量
存储返回值。defer会在函数真正返回之前执行。所以defer可以修改named return values，但是匿名返回值
不能修改。
