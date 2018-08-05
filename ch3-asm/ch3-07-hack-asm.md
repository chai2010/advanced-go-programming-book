# 3.7 汇编语言的威力

汇编语言的真正威力来自两个维度：一是突破框架限制，实现看似不可能的任务；二是突破指令限制，通过高级指令挖掘极致的性能。对于第一个问题，我们将演示如何通过Go汇编语言直接访问系统调用，和直接调用C语言函数。对于第二个问题，我们将演示X64指令中AVX等高级指令的简单用法。

## 3.7.1 系统调用

系统调用是操作系统为外提供的公共接口。因为操作系统彻底接管了各种底层硬件设备，因此操作系统提供的系统调用成了实现某些操作的唯一方法。从另一个角度看，系统调用更像是一个RPC远程过程调用，不过信道是寄存器和内存。在系统调用时，我们向操作系统发送调用的编号和对应的参数，然后阻塞等待系统调用地返回。因为涉及到阻塞等待，因此系统调用期间的CPU利用率一般是可以忽略的。另一个和RPC地远程调用类似的地方是，操作系统内核处理系统调用时不会依赖用户的栈空间，一般不会导致爆栈发生。因此系统调用是最简单安全的一种调用了。

系统调用虽然简单，但是它是操作系统对外的接口，因此不同的操作系统调用规范可能有很大地差异。我们先看看Linux在AMD64架构上的系统调用规范，在`syscall/asm_linux_amd64.s`文件中有注释说明：

```go
//
// System calls for AMD64, Linux
//

// func Syscall(trap int64, a1, a2, a3 uintptr) (r1, r2, err uintptr);
// Trap # in AX, args in DI SI DX R10 R8 R9, return in AX DX
// Note that this differs from "standard" ABI convention, which
// would pass 4th arg in CX, not R10.
```

这是`syscall.Syscall`函数的内部注释，简要说明了Linux系统调用的规范。系统调用的前6个参数直接由DI、SI、DX、R10、R8和R9寄存器传输，结果由AX和DX寄存器返回。macOS等类UINX系统调用的参数传输大多数都采用类似的规则。

macOS的系统调用编号在`/usr/include/sys/syscall.h`头文件，Linux的系统调用号在`/usr/include/asm/unistd.h`头文件。虽然在UNIX家族中是系统调用的参数和返回值的传输规则类似，但是不同操作系统提供的系统调用却不是完全相同的，因此系统调用编号也有很大的差异。以UNIX系统中著名的write系统调用为例，在macOS的系统调用编号为4，而在Linux的系统调用编号却是1。

我们将基于write系统调用包装一个字符串输出函数。下面的代码是macOS版本：

```
// func SyscallWrite_Darwin(fd int, msg string) int
TEXT ·SyscallWrite_Darwin(SB), NOSPLIT, $0
	MOVQ $(0x2000000+4), AX // #define SYS_write 4
	MOVQ fd+0(FP),       DI
	MOVQ msg_data+8(FP), SI
	MOVQ msg_len+16(FP), DX
	SYSCALL
	MOVQ AX, ret+0(FP)
	RET
```

其中第一个参数是输出文件的文件描述符编号，第二个参数是字符串的头部。字符串头部是由reflect.StringHeader结构定义，第一成员是8字节的数据指针，第二个成员是8字节的数据长度。在macOS系统中，执行系统调用时还需要将系统调用的编号加上0x2000000后再行传入AX。然后再将fd、数据地址和长度作为write系统调用的三个参数输入，分别对应DI、SI和DX三个寄存器。最后通过SYSCALL指令执行系统调用，系统调用返回后从AX获取返回值。

这样我们就基于系统调用包装了一个定制的输出函数。在UNIX系统中，标准输入stdout的文件描述符编号是1，因此我们可以用1作为参数实现字符串的输出：

```go
func SyscallWrite_Darwin(fd int, msg string) int

func main() {
	if runtime.GOOS == "darwin" {
		SyscallWrite_Darwin(1, "hello syscall!\n")
	}
}
```

如果是Linux系统，只需要将编号改为write系统调用对应的1即可。而Windows的系统调用则有另外的参数传输规则。在X64环境Windows的系统调用参数传输规则和默认的C语言规则非常相似，在后续的直接调用C函数部分再行讨论。


## 3.7.2 直接调用C函数

在计算机的发展的过程中，C语言和UNIX操作系统有着不可替代的作用。因此操作系统的系统调用、汇编语言和C语言函数调用规则几个技术是密切相关的。

在X86的32位系统时代，C语言一般默认的是用栈传递参数并用AX寄存器返回结果，称为cdecl调用约定。Go语言函数和cdecl调用约定非常相似，它们都是以栈来传递参数并且返回地址和BP寄存器的布局都是类似的。但是Go语言函数将返回值也通过栈返回，因此Go语言函数可以支持多个返回值。我们可以将Go语言函数看作是没有返回值的C语言函数，同时将Go语言函数中的返回值挪到C语言函数参数的尾部，这样栈不仅仅用于传入参数也用于返回多个结果。

在X64时代，AMD架构增加了8个通用寄存器，为了提高效率C语言也默认改用寄存器来传递参数。在X64系统，默认有System V AMD64 ABI和Microsoft x64两种C语言函数调用规范。其中System V的规范适用于Linux、FreeBSD、macOS等诸多类UNIX系统，而Windows则是用自己特有的调用规范。

在理解了C语言函数的调用规范之后，汇编代码就可以绕过CGO技术直接调用C语言函数。为了便于演示，我们先用C语言构造一个简单的加法函数myadd：

```c
#include <stdint.h>

int64_t myadd(int64_t a, int64_t b) {
	return a+b;
}
```

然后我们需要实现一个asmCallCAdd函数：

```go
func asmCallCAdd(cfun uintptr, a, b int64) int64
```

因为Go汇编语言和CGO特性不能同时在一个包中使用（因为CGO会调用gcc，而gcc会将Go汇编语言当做普通的汇编程序处理，从而导致错误），我们通过一个参数传入C语言myadd函数的地址。asmCallCAdd函数的其余参数和C语言myadd函数的参数保持一致。

我们只实现System V AMD64 ABI规范的版本。在System V版本中，寄存器可以最多传递六个参数，分别对应DI、SI、DX、CX、R8和R9六个寄存器（如果是浮点数则需要通过XMM寄存器传送），返回值依然通过AX返回。通过对比系统调用的规范可以发现，系统调用的第四个参数是用R10寄存器传递，而C语言函数的第四个参数是用CX传递。

下面是System V AMD64 ABI规范的asmCallCAdd函数的实现：

```
// System V AMD64 ABI
// func asmCallCAdd(cfun uintptr, a, b int64) int64
TEXT ·asmCallCAdd(SB), NOSPLIT, $0
	MOVQ cfun+0(FP), AX // cfun
	MOVQ a+8(FP),    DI // a
	MOVQ b+16(FP),   SI // b
	CALL AX
	MOVQ AX, ret+24(FP)
	RET
```

首先是将第一个参数表示的C函数地址保存到AX寄存器便于后续调用。然后分别将第二和第三个参数加载到DI和SI寄存器。然后CALL指令通过AX中保持的C语言函数地址调用C函数。最后从AX寄存器获取C函数的返回值，并通过asmCallCAdd函数返回。

Win64环境的C语言调用规范类似。不过Win64规范中只有CX、DX、R8和R9四个寄存器传递参数（如果是浮点数则需要通过XMM寄存器传送），返回值依然通过AX返回。虽然是可以通过寄存器传输参数，但是调用这依然要为前四个参数准备栈空间。需要注意的是，Windows x64的系统调用和C语言函数可能是采用相同的调用规则。因为没有Windows测试环境，我们这里就不提供了Windows版本的代码实现了，Windows用户可以自己尝试实现类似功能。

然后我们就可以使用asmCallCAdd函数直接调用C函数了：

```go
/*
#include <stdint.h>

int64_t myadd(int64_t a, int64_t b) {
	return a+b;
}
*/
import "C"

import (
	asmpkg "path/to/asm"
)

func main() {
	if runtime.GOOS != "windows" {
		println(asmpkg.asmCallCAdd(
			uintptr(unsafe.Pointer(C.myadd)),
			123, 456,
		))
	}
}
```

在上面的代码中，通过`C.myadd`获取C函数的地址，然后转换为合适的类型再传人asmCallCAdd函数。在这个例子中，汇编函数假设调用的C语言函数需要的栈很小，可以直接复用Go函数中多余的空间。如果C语言函数可能需要较大的栈，可以尝试像CGO那样切换到系统线程的栈上运行。


## 3.7.3 AVX指令

从Go1.11开始，Go汇编语言引入了AVX512指令的支持。AVX指令集是属于Intel家的SIMD指令集中的一部分。AVX512的最大特点是数据有512位宽度，可以一次计算8个64位数或者是等大小的数据。因此AVX指令可以用于优化矩阵或图像等并行度很高的算法。不过并不是每个X86体系的CPU都支持了AVX指令，因此首要的任务是如何判断CPU支持了哪些高级指令。

在Go语言标准库的`internal/cpu`包提供了CPU是否支持某些高级指令的基本信息，但是只有标准库才能引用这个包（因为internal路径的限制）。该包底层是通过X86提供的CPUID指令来识别处理器的详细信息。最简便的方法是直接将`internal/cpu`包克隆一份。不过这个包为了避免复杂的依赖没有使用init函数自动初始化，因此需要根据情况手工调整代码执行doinit函数初始化。

`internal/cpu`包针对X86处理器提供了以下特性检测：

```go
package cpu

var X86 x86

// The booleans in x86 contain the correspondingly named cpuid feature bit.
// HasAVX and HasAVX2 are only set if the OS does support XMM and YMM registers
// in addition to the cpuid feature bit being set.
// The struct is padded to avoid false sharing.
type x86 struct {
	HasAES       bool
	HasADX       bool
	HasAVX       bool
	HasAVX2      bool
	HasBMI1      bool
	HasBMI2      bool
	HasERMS      bool
	HasFMA       bool
	HasOSXSAVE   bool
	HasPCLMULQDQ bool
	HasPOPCNT    bool
	HasSSE2      bool
	HasSSE3      bool
	HasSSSE3     bool
	HasSSE41     bool
	HasSSE42     bool
}
```

因此我们可以用以下的代码测试运行时的CPU是否支持AVX2指令集：

```go
import (
	cpu "path/to/cpu"
)

func main() {
	if cpu.X86.HasAVX2 {
		// support AVX2
	}
}
```

AVX512是比较新的指令集，只有高端的CPU才会提供支持。为了主流的CPU也能运行代码测试，我们选择AVX2指令来构造例子。AVX2指令每次可以处理32字节的数据，可以用来提升数据复制的工作的效率。

下面的例子是用AVX2指令复制数据，每次复制数据32字节倍数大小的数据：

```
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
```

其中VMOVDQU指令先将`0(SI)(AX*1)`地址开始的32字节数据复制到Y0寄存器中，然后再复制到`0(DI)(AX*1)`对应的目标内存中。VMOVDQU指令操作的数据地址可以不用对齐。

AVX2共有16个Y寄存器，每个寄存器有256bit位。如果要复制的数据很多，可以多个寄存器同时复制，这样可以利用更高效的流水特性优化性能。

