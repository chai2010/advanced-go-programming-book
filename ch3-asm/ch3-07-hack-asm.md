# 3.7 The power of assembly language

The true power of assembly language comes from two dimensions: one is to break through the framework constraints and achieve seemingly impossible tasks; the other is to break through the instruction limits and mine the ultimate performance through advanced instructions. For the first question, we will demonstrate how to directly access system calls through Go assembly language, and directly call C language functions. For the second question, we will demonstrate the simple use of advanced instructions such as AVX in X64 instructions.

## 3.7.1 System Call

A system call is a public interface provided by the operating system. Because the operating system completely takes over the various underlying hardware devices, the system calls provided by the operating system are the only way to do something. From another perspective, the system call is more like an RPC remote procedure call, but the channel is a register and memory. When the system is called, we send the caller's number and corresponding parameters to the operating system, and then block the wait for the system call to return. CPU utilization during system calls is generally negligible because of blocking waits. Another similarity to RPC remote calls is that the operating system kernel does not rely on the user's stack space when processing system calls, and generally does not cause a burst to occur. So system calls are the simplest and safest kind of call.

Although the system call is simple, it is the external interface of the operating system, so different operating system call specifications may vary greatly. Let's take a look at Linux's system call specification on the AMD64 architecture. There are comments in the `syscall/asm_linux_amd64.s` file:

```go
//
// System calls for AMD64, Linux
//

// func Syscall(trap int64, a1, a2, a3 uintptr) (r1, r2, err uintptr);
// Trap # in AX, args in DI SI DX R10 R8 R9, return in AX DX
// Note that this differs from "standard" ABI convention, which
// would pass 4th arg in CX, not R10.
```

This is an internal comment for the `syscall.Syscall` function, which briefly describes the specification of the Linux system call. The first six parameters of the system call are transferred directly from the DI, SI, DX, R10, R8, and R9 registers, and the result is returned by the AX and DX registers. Most of the parameter transmissions of UINX system calls such as macOS use similar rules.

The system call number of macOS is in the `/usr/include/sys/syscall.h` header file, and the Linux system call number is in the `/usr/include/asm/unistd.h` header file. Although the parameters of the system call parameters and return values ​​are similar in the UNIX family, the system calls provided by different operating systems are not identical, so the system call numbers are also very different. Take the well-known write system call in UNIX system as an example. The system call number in macOS is 4, while the system call number in Linux is 1.

We will wrap a string output function based on the write system call. The following code is the macOS version:

```
// func SyscallWrite_Darwin(fd int, msg string) int
TEXT ·SyscallWrite_Darwin(SB), NOSPLIT, $0
MOVQ $(0x2000000+4), AX // #define SYS_write 4
MOVQ fd+0(FP), DI
MOVQ msg_data+8(FP), SI
MOVQ msg_len+16(FP), DX
SYSCALL
MOVQ AX, ret+0(FP)
RET
```

The first parameter is the file descriptor number of the output file, and the second parameter is the head of the string. The string header is defined by the reflect.StringHeader structure, the first member is an 8-byte data pointer, and the second member is an 8-byte data length. In the macOS system, when executing the system call, you need to add 0x2000000 to the system call number and then pass it to AX. Then fd, data address and length are input as three parameters of the write system call, corresponding to three registers of DI, SI and DX. Finally, the system call is executed by the SYSCALL instruction, and the return value is obtained from AX after the system call returns.

This way we wrap a custom output function based on the system call. In UNIX systems, the file descriptor number for standard input stdout is 1, so we can use 1 as a parameter to implement the output of the string:

```go
Func SyscallWrite_Darwin(fd int, msg string) int

Func main() {
If runtime.GOOS == "darwin" {
SyscallWrite_Darwin(1, "hello syscall!\n")
}
}
```

If it is a Linux system, just change the number to the corresponding 1 of the write system call. Windows system calls have additional parameter transfer rules. In the X64 environment, Windows system call parameter transfer rules are very similar to the default C language rules, and will be discussed in the subsequent direct call to the C function part.


## 3.7.2 Call C function directly

In the development of computers, C language and UNIX operating system have an irreplaceable role. Therefore, the system calls, assembly language and C language function calling rules of the operating system are closely related.

In the era of 32-bit systems in X86, the C language defaults to passing parameters with the stack and returning results with the AX register, called the cdecl calling convention. The Go language function is very similar to the cdecl calling convention. They all pass arguments on the stack and the return address and BP register layout are similar. But the Go language function returns the return value through the stack, so the Go language function can support multiple return values. We can think of the Go language function as a C language function with no return value, and move the return value in the Go language function to the end of the C language function parameter, so that the stack is used not only for passing parameters but also for returning multiple result.

In the X64 era, the AMD architecture added eight general-purpose registers. In order to improve efficiency, the C language also uses registers to pass parameters by default. On X64 systems, there are two C language function call specifications, System V AMD64 ABI and Microsoft x64. The System V specification applies to many UNIX-like systems such as Linux, FreeBSD, and macOS, while Windows uses its own unique calling specification.

After understanding the calling specification of the C language function, the assembly code can bypass the CGO technique and directly call the C language function. For the sake of demonstration, let's construct a simple addition function myadd in C:

```c
#include <stdint.h>

Int64_t myadd(int64_t a, int64_t b) {
Return a+b;
}
```

Then we need to implement an asmCallCAdd function:

```go
Func asmCallCAdd(cfun uintptr, a, b int64) int64
```

Because Go assembly language and CGO features can't be used in a package at the same time (because CGO will call gcc, and gcc will treat Go assembly language as a normal assembler, causing errors), we pass the C language myadd function through a parameter. the address of. The remaining parameters of the asmCallCAdd function are consistent with the parameters of the C language myadd function.

We only implement the version of the System V AMD64 ABI specification. In the System V version, the register can pass up to six parameters, corresponding to the three registers DI, SI, DX, CX, R8, and R9 (if floating point numbers need to be transmitted through the XMM register), the return value is still returned by AX. By comparing the specifications of the system call, it can be found that the fourth parameter of the system call is passed in the R10 register, and the fourth parameter of the C language function is passed in CX.

The following is the implementation of the asmCallCAdd function of the System V AMD64 ABI specification:

```
// System V AMD64 ABI
// func asmCallCAdd(cfun uintptr, a, b int64) int64
TEXT · asmCallCAdd(SB), NOSPLIT, $0
MOVQ cfun+0(FP), AX // cfun
MOVQ a+8(FP), DI // a
MOVQ b+16(FP), SI // b
CALL AX
MOVQ AX, ret+24(FP)
RET
```

The first is to save the C function address indicated by the first parameter to the AX register for subsequent calls. The second and third parameters are then loaded into the DI and SI registers, respectively. The CALL instruction then calls the C function via the C language function address held in AX. Finally, the return value of the C function is obtained from the AX register and returned by the asmCallCAdd function.

The C language calling specification for Win64 environments is similar. However, in the Win64 specification, only the CX, DX, R8, and R9 four registers pass parameters (if floating point numbers need to be transferred through the XMM register), the return value is still returned by AX. Although it is possible to transfer parameters through registers, it is still necessary to prepare stack space for the first four parameters. It should be noted that Windows x64 system calls and C language functions may use the same calling rules. Because there is no Windows test environment, we do not provide the Windows version of the code implementation here, Windows users can try to achieve similar functions.

Then we can call the C function directly using the asmCallCAdd function:

```go
/*
#include <stdint.h>

Int64_t myadd(int64_t a, int64_t b) {
Return a+b;
}
*/
Import "C"

Import (
Asmpkg "path/to/asm"
)

Func main() {
If runtime.GOOS != "windows" {
Println(asmpkg.asmCallCAdd(
Uintptr(unsafe.Pointer(C.myadd)),
123, 456,
))
}
}
```

In the above code, get the address of the C function by `C.myadd`, then convert it to the appropriate type and pass the asmCallCAdd function. In this example, the assembly function assumes that the called C language function requires a small stack and can directly reuse the extra space in the Go function. If the C language function may require a larger stack, try switching to the system thread's stack like CGO.


## 3.7.3 AVX Instructions

Since Go1.11, Go assembly language has introduced support for AVX512 instructions. The AVX instruction set is part of the Intel family's SIMD instruction set. The biggest feature of the AVX512 is that the data has a width of 512 bits and can calculate 8 64-bit numbers or the same size data at a time. Therefore, the AVX instruction can be used to optimize algorithms with high degree of parallelism such as matrix or image. However, not every CPU of the X86 system supports the AVX instruction, so the first task is to determine which advanced instructions the CPU supports.

The `internal/cpu` package in the Go language standard library provides basic information on whether the CPU supports certain advanced instructions, but only the standard library can reference this package (because of the limitations of the internal path). The bottom layer of the package is the CPUID instruction provided by X86 to identify the details of the processor. The easiest way is to clone the `internal/cpu` package directly. However, in order to avoid complex dependencies, this package does not use the init function to automatically initialize, so you need to manually adjust the code to perform the doinit function initialization according to the situation.

The `internal/cpu` package provides the following feature detection for X86 processors:

```go
Package cpu

Var X86 x86

// The booleans in x86 contain the correspondingly named cpuid feature bit.
// HasAVX and HasAVX2 are only set if the OS does support XMM and YMM registers
// in addition to the cpuid feature bit being set.
// The struct is padded to avoid false sharing.
Type x86 struct {
HasAES bool
HasADX bool
HasAVX bool
HasAVX2 bool
HasBMI1 bool
HasBMI2 bool
HasERMS bool
HasFMA bool
HasOSXSAVE bool
HasPCLMULQDQ bool
HasPOPCNT bool
HasSSE2 bool
HasSSE3 bool
HasSSSE3 bool
HasSSE41 bool
HasSSE42 bool
}
```

So we can use the following code to test whether the runtime CPU supports the AVX2 instruction set:

```go
Import (
Cpu "path/to/cpu"
)

Func main() {
If cpu.X86.HasAVX2 {
// support AVX2
}
}
```

The AVX512 is a relatively new instruction set that is only supported by high-end CPUs. In order to run the code test for the mainstream CPU, we chose the AVX2 instruction to construct the example. The AVX2 instruction can process 32 bytes of data at a time and can be used to improve the efficiency of data copying.

The following example uses the AVX2 instruction to copy data, each time copying data in multiples of 32 bytes:

```
// func CopySlice_AVX2(dst, src []byte, len int)
TEXT ·CopySlice_AVX2(SB), NOSPLIT, $0
MOVQ dst_data+0(FP), DI
MOVQ src_data+24(FP), SI
MOVQ len+32(FP), BX
MOVQ $0, AX

LOOP:
VMOVDQU 0(SI)(AX*1), Y0
VMOVDQU Y0, 0(DI)(AX*1)
ADDQ $32, AX
CMPQ AX, BX
JL LOOP
RET
```

The VMOVDQU instruction first copies the 32-byte data starting at the address of `0(SI)(AX*1)` into the Y0 register, and then copies it to the target memory corresponding to `0(DI)(AX*1)`. The data address of the VMOVDQU instruction operation can be omitted.

The AVX2 has a total of 16 Y registers, each with 256 bits. If you have a lot of data to copy, you can copy multiple registers at the same time, which can optimize performance with more efficient pipeline features.