// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

#include "textflag.h"

// https://en.wikipedia.org/wiki/X86_calling_conventions#x86-64_calling_conventions

/*
System V AMD64 ABI

The calling convention of the System V AMD64 ABI is followed on
Solaris, Linux, FreeBSD, macOS,[18] and is the de facto standard
among Unix and Unix-like operating systems. The first six integer
or pointer arguments are passed in registers RDI, RSI, RDX, RCX, R8, R9
(R10 is used as a static chain pointer in case of nested functions[19]:21),
while XMM0, XMM1, XMM2, XMM3, XMM4, XMM5, XMM6 and XMM7 are used for
certain floating point arguments.[19]:22 As in the Microsoft x64
calling convention, additional arguments are passed on the stack[19]:22.
Integral return values up to 64 bits in size are stored in RAX
while values up to 128 bit are stored in RAX and RDX. Floating-point
return values are similarly stored in XMM0 and XMM1.[19]:25.

If the callee wishes to use registers RBX, RBP, and R12–R15,
it must restore their original values before returning control to the caller.
All other registers must be saved by the caller if it wishes to preserve
their values.[19]:16

For leaf-node functions (functions which do not call any other function(s)),
a 128-bytes space is stored just beneath the stack pointer of the function.
The space is called red-zone. This zone will not be clobbered by any signal
or interrupt handlers. Compilers can thus utilize this zone to save
local variables. Compilers may omit some instructions at the starting of
the function (adjustment of RSP, RBP) by utilizing this zone.
However, other function may clobber this zone. Therefore, this zone
should only be used for leaf-node functions. gcc and clang takes
-mno-red-zone flag to disable red-zone optimizations.

If the callee is a variadic function, then the number of floating point
arguments passed to the function in vector registers must be provided by
the caller in the AL register.[19]:55

Unlike the Microsoft calling convention, a shadow space is not provided;
on function entry, the return address is adjacent to the seventh
integer argument on the stack.
*/

// func CallCAdd(cfun uintptr, a, b int64) int64
TEXT ·CallCAdd_SystemV_ABI(SB), NOSPLIT, $0
	MOVQ cfun+0(FP), AX // cfun
	MOVQ a+8(FP),    DI // a
	MOVQ b+16(FP),   SI // b
	CALL AX
	MOVQ AX, ret+24(FP)
	RET


/*
Microsoft x64 calling convention

The Microsoft x64 calling convention[14][15] is followed on Windows
and pre-boot UEFI (for long mode on x86-64). It uses registers RCX, RDX, R8, R9
for the first four integer or pointer arguments (in that order),
and XMM0, XMM1, XMM2, XMM3 are used for floating point arguments.
Additional arguments are pushed onto the stack (right to left).
Integer return values (similar to x86) are returned in RAX if 64 bits or less.
Floating point return values are returned in XMM0.
Parameters less than 64 bits long are not zero extended;
the high bits are not zeroed.

When compiling for the x64 architecture in a Windows context
(whether using Microsoft or non-Microsoft tools),
there is only one calling convention – the one described here,
so that stdcall, thiscall, cdecl, fastcall, etc.,
are now all one and the same.

In the Microsoft x64 calling convention, it is the caller's responsibility to
allocate 32 bytes of "shadow space" on the stack right before calling the function
(regardless of the actual number of parameters used),
and to pop the stack after the call. The shadow space is used to spill
RCX, RDX, R8, and R9,[16] but must be made available to all functions,
even those with fewer than four parameters.

The registers RAX, RCX, RDX, R8, R9, R10, R11 are
considered volatile (caller-saved).[17]

The registers RBX, RBP, RDI, RSI, RSP, R12, R13, R14, and R15 are
considered nonvolatile (callee-saved).[17]

For example, a function taking 5 integer arguments will take the
first to fourth in registers, and the fifth will be pushed
on the top of the shadow space. So when the called function is entered,
the stack will be composed of (in ascending order) the return address,
followed by the shadow space (32 bytes) followed by the fifth parameter.

In x86-64, Visual Studio 2008 stores floating point numbers in XMM6 and XMM7
(as well as XMM8 through XMM15); consequently, for x86-64,
user-written assembly language routines must preserve XMM6 and XMM7
(as compared to x86 wherein user-written assembly language routines
did not need to preserve XMM6 and XMM7). In other words, user-written
assembly language routines must be updated to save/restore
XMM6 and XMM7 before/after the function when being ported from x86 to x86-64.

Starting with Visual Studio 2013, Microsoft introduced the__vectorcall
calling convention which extends the x64 convention.
*/

// func CallCAdd(cfun uintptr, a, b int64) int64
TEXT ·CallCAdd_Win64_ABI(SB), NOSPLIT, $0
	MOVQ cfun+0(FP), AX // cfun
	MOVQ a+8(FP),    CX // a
	MOVQ b+16(FP),   DX // b
	CALL AX
	MOVQ AX, ret+24(FP)
	RET

// func SyscallWrite_Darwin(fd int, msg string) int
TEXT ·SyscallWrite_Darwin(SB), NOSPLIT, $0
	MOVQ $(0x2000000+4), AX // #define SYS_write 4
	MOVQ fd+0(FP),       DI
	MOVQ msg_data+8(FP), SI
	MOVQ msg_len+16(FP), DX
	SYSCALL
	MOVQ AX, ret+0(FP)
	RET

// func SyscallWrite_Linux(fd int, msg string) int
TEXT ·SyscallWrite_Linux(SB), NOSPLIT, $0
	MOVQ $1,             AX // #define SYS_write 1
	MOVQ fd+0(FP),       DI
	MOVQ msg_data+8(FP), SI
	MOVQ msg_len+16(FP), DX
	SYSCALL
	MOVQ AX, ret+0(FP)
	RET

// func SyscallWrite_Windows(fd int, msg string) int
TEXT ·SyscallWrite_Windows(SB), NOSPLIT, $0
	RET // TODO

// func CopySlice_AVX2(dst, src []byte, len int)
TEXT ·CopySlice_AVX2(SB), NOSPLIT, $0
	MOVQ dst_data+0(FP),  DI
	MOVQ src_data+24(FP), SI
	MOVQ len+32(FP),      BX
	MOVQ $0,              AX

LOOP:
	VMOVDQU 0(SI)(AX*1), Y0
	VMOVDQU Y0, 0(DI)(AX*1)
	ADDQ $32, AX
	CMPQ AX, BX
	JL   LOOP
	RET
