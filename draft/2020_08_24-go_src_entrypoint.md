```bash
$ gdb ./demo
GNU gdb (Debian 8.2.1-2+b3) 8.2.1
Copyright (C) 2018 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
Type "show copying" and "show warranty" for details.
This GDB was configured as "x86_64-linux-gnu".
Type "show configuration" for configuration details.
For bug reporting instructions, please see:
<http://www.gnu.org/software/gdb/bugs/>.
Find the GDB manual and other documentation resources online at:
    <http://www.gnu.org/software/gdb/documentation/>.

For help, type "help".
Type "apropos word" to search for commands related to "word"...
Reading symbols from ./demo...done.
Loading Go Runtime support.
(gdb) info files
Symbols from "/home/jiajun/Code/go/demo".
Local exec file:
        `/home/jiajun/Code/go/demo', file type elf64-x86-64.
        Entry point: 0x463860
        0x0000000000401000 - 0x0000000000497718 is .text
        0x0000000000498000 - 0x00000000004db284 is .rodata
        0x00000000004db420 - 0x00000000004dbb44 is .typelink
        0x00000000004dbb60 - 0x00000000004dbbb0 is .itablink
        0x00000000004dbbb0 - 0x00000000004dbbb0 is .gosymtab
        0x00000000004dbbc0 - 0x000000000053678e is .gopclntab
        0x0000000000537000 - 0x0000000000537020 is .go.buildinfo
        0x0000000000537020 - 0x00000000005451e4 is .noptrdata
        0x0000000000545200 - 0x000000000054c770 is .data
        0x000000000054c780 - 0x000000000057c648 is .bss
        0x000000000057c660 - 0x000000000057eea8 is .noptrbss
        0x0000000000400f9c - 0x0000000000401000 is .note.go.buildid
(gdb) b *0x463860
Breakpoint 1 at 0x463860: file /home/jiajun/Code/go/src/runtime/rt0_linux_amd64.s, line 8.
(gdb) r
Starting program: /home/jiajun/Code/go/demo

Breakpoint 1, _rt0_amd64_linux () at /home/jiajun/Code/go/src/runtime/rt0_linux_amd64.s:8
8               JMP     _rt0_amd64(SB)
(gdb) n
_rt0_amd64 () at /home/jiajun/Code/go/src/runtime/asm_amd64.s:15
15              MOVQ    0(SP), DI       // argc
(gdb)
16              LEAQ    8(SP), SI       // argv
(gdb)
17              JMP     runtime·rt0_go(SB)
(gdb)
runtime.rt0_go () at /home/jiajun/Code/go/src/runtime/asm_amd64.s:89
89              MOVQ    DI, AX          // argc
(gdb)
90              MOVQ    SI, BX          // argv
(gdb)
91              SUBQ    $(4*8+7), SP            // 2args 2auto
(gdb)
runtime.rt0_go () at /home/jiajun/Code/go/src/runtime/asm_amd64.s:92
92              ANDQ    $~15, SP
(gdb)
runtime.rt0_go () at /home/jiajun/Code/go/src/runtime/asm_amd64.s:93
93              MOVQ    AX, 16(SP)
(gdb)
94              MOVQ    BX, 24(SP)
(gdb)
98              MOVQ    $runtime·g0(SB), DI
(gdb)
99              LEAQ    (-64*1024+104)(SP), BX
(gdb)
100             MOVQ    BX, g_stackguard0(DI)
(gdb)
101             MOVQ    BX, g_stackguard1(DI)
(gdb)
102             MOVQ    BX, (g_stack+stack_lo)(DI)
(gdb)
103             MOVQ    SP, (g_stack+stack_hi)(DI)
(gdb)
106             MOVL    $0, AX
(gdb)
107             CPUID
(gdb)
108             MOVL    AX, SI
(gdb)
109             CMPL    AX, $0
(gdb)
110             JE      nocpuinfo
(gdb)
115             CMPL    BX, $0x756E6547  // "Genu"
(gdb)
116             JNE     notintel
(gdb)
117             CMPL    DX, $0x49656E69  // "ineI"
(gdb)
118             JNE     notintel
(gdb)
119             CMPL    CX, $0x6C65746E  // "ntel"
(gdb)
120             JNE     notintel
(gdb)
121             MOVB    $1, runtime·isIntel(SB)
(gdb)
122             MOVB    $1, runtime·lfenceBeforeRdtsc(SB)
(gdb)
126             MOVL    $1, AX
(gdb)
127             CPUID
(gdb)
128             MOVL    AX, runtime·processorVersionInfo(SB)
(gdb)
132             MOVQ    _cgo_init(SB), AX
(gdb)
133             TESTQ   AX, AX
(gdb)
134             JZ      needtls
(gdb)
183             LEAQ    runtime·m0+m_tls(SB), DI
(gdb)
184             CALL    runtime·settls(SB)
(gdb)
188             MOVQ    $0x123, g(BX)
(gdb)
189             MOVQ    runtime·m0+m_tls(SB), AX
(gdb)
190             CMPQ    AX, $0x123
(gdb)
191             JEQ 2(PC)
(gdb)
196             LEAQ    runtime·g0(SB), CX
(gdb)
197             MOVQ    CX, g(BX)
(gdb)
198             LEAQ    runtime·m0(SB), AX
(gdb)
201             MOVQ    CX, m_g0(AX)
(gdb)
203             MOVQ    AX, g_m(CX)
(gdb)
205             CLD                             // convention is D is always left cleared
(gdb)
206             CALL    runtime·check(SB)
(gdb)
208             MOVL    16(SP), AX              // copy argc
(gdb)
209             MOVL    AX, 0(SP)
(gdb)
210             MOVQ    24(SP), AX              // copy argv
(gdb)
211             MOVQ    AX, 8(SP)
(gdb)
212             CALL    runtime·args(SB)
(gdb)
213             CALL    runtime·osinit(SB)
(gdb)
214             CALL    runtime·schedinit(SB)
(gdb)
217             MOVQ    $runtime·mainPC(SB), AX         // entry
(gdb)
218             PUSHQ   AX
(gdb)
219             PUSHQ   $0                      // arg size
(gdb)
220             CALL    runtime·newproc(SB)
(gdb)
221             POPQ    AX
(gdb)
222             POPQ    AX
(gdb)
225             CALL    runtime·mstart(SB)
(gdb)
[New LWP 6192]
[New LWP 6193]
[New LWP 6194]
[New LWP 6195]
[New LWP 6196]
hello world
[LWP 6196 exited]
[LWP 6195 exited]
[LWP 6194 exited]
[LWP 6193 exited]
[LWP 6192 exited]
[Inferior 1 (process 6185) exited normally]
(gdb)
The program is not being run.
```
