# 3.8. 汇编语言的威力

汇编语言的真正威力来自两个维度：一是突破框架限制，实现看似不可能的任务；二是突破指令限制，通过高级指令挖掘极致的性能。对于一个问题，我们将讨论如何通过Go汇编语言直接访问系统调用，和直接调用C语言函数。对于第二个问题，我们将简单讨论X64指令中AVX等高级指令的简单用法。

## 系统调用

系统调用是操作系统为外提供地公共接口。因为操作系统彻底接管了各种底层硬件设备，因此操作系统提供的系统调用成了某些操作地唯一方法。从另一个角度看，系统调用更像是一个RPC远程过程调用。在系统调用时，我们只是像操作系统发送某个调用地编号和对应地参数，然后阻塞等待系统调用地返回。因为涉及到阻塞等待，因此系统调用期间对应CPU的利用率一般是可以忽略地。同时系统调用时类似RPC地远程调用，操作系统内核处理系统调用时不会依赖用户的栈空间，一般不会导致爆栈发生。因此可以认为系统调用是最简单的一种调用了。

系统调用虽然是最简单地调用，但是它是操作系统对外的接口，因此不同的操作系统调用规范可能有很大地差异。我们先关注Linux在AMD64架构上的系统调用规范，在`syscall/asm_linux_amd64.s`文件中有注释说明：

```go
//
// System calls for AMD64, Linux
//

// func Syscall(trap int64, a1, a2, a3 uintptr) (r1, r2, err uintptr);
// Trap # in AX, args in DI SI DX R10 R8 R9, return in AX DX
// Note that this differs from "standard" ABI convention, which
// would pass 4th arg in CX, not R10.
```

这是`syscall.Syscall`函数的内部注释，简要说明了Linux系统调用的规范。系统调用一般不超过6个参数，直接由DI、SI、DX、R10、R8和R9寄存器传输，结果由AX和DX寄存器返回。

TODO

## 直接调用C函数

TODO

## AVX指令

TODO

<!--
学习汇编的原因是汇编可以挖掘功能
同时也可以挖掘性能

syscall（类似rpc，没有栈爆问题）
cgoall
avx2
-->
