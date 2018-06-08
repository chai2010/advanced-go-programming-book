# 3.6. 再论函数(Doing)

在前面的章节中我们已经简单讨论过Go的汇编函数，但是那些主要是叶子函数。叶子函数的最大特点是不会调用其他函数，也就是栈的大小是可以预期的，叶子函数也就是可以基本忽略爆栈的问题（如果已经爆了，那也是上级函数的问题）。如果没有爆栈问题，那么也就是不会有栈的分裂问题；如果没有栈的分裂也就不需要移动栈上的指针，也就不会有栈上指针管理的问题。但是是现实中Go语言的函数是可以任意深度调用的，永远不用担心爆栈的风险。那么这些近似黑科技的特殊是如何通过低级的汇编语言实现的呢？这些都是本节尝试讨论的问题。

## 递归函数: 1到n求和

递归函数是比较特殊的函数，递归函数通过调用自身并且在栈上保存状态，这可以简化很多问题的处理。Go语言中递归函数的强大之处是不用担心爆栈问题，因为栈可以根据需要进行扩容和收缩。我们现在尝试通过汇编语言实现一个递归调用的函数，为了简化目前先不考虑栈的变化。

先通过Go递归函数实现一个1到n的求和函数：

```go
// sum = 1+2+...+n
// sum(100) = 5050
func sum(n int) int {
	if n > 0 { return n+sum(n-1) } else { return 0 }
}
```

然后通过if/goto构型重新上面的递归函数，以便于转义为汇编版本：

```go
func sum(n int) (result int) {
	var AX = n
	var BX int

	if n > 0 { goto L_STEP_TO_END }
	goto L_END

L_STEP_TO_END:
	AX -= 1
	BX = sum(AX)

	AX = n // 调用函数后, AX重新恢复为n
	BX += AX

	return BX

L_END:
	return 0
}
```

在改写之后，递归调用的参数需要引入局部变量，保存中间结果也需要引入局部变量。而通过栈来保存中间的调用状态正是递归函数的核心。因为输入参数也在栈上，因为我们可以通过输入参数来保存少量的状态。同时我们模拟定义了AX和BX寄存器，寄存器在使用前需要初始化，并且在函数调用后也需要重新初始化。

下面继续改造为汇编语言版本：

```
// func sum(n int) (result int)
TEXT ·sum(SB), NOSPLIT, $16-16
	MOVQ n+0(FP), AX       // n
	MOVQ result+8(FP), BX  // result

	CMPQ AX, $0            // test n - 0
	JG   L_STEP_TO_END     // if > 0: goto L_STEP_TO_END
	JMP  L_END             // goto L_STEP_TO_END

L_STEP_TO_END:
	SUBQ $1, AX            // AX -= 1
	MOVQ AX, 0(SP)         // arg: n-1
	CALL ·sum(SB)          // call sum(n-1)
	MOVQ 8(SP), BX         // BX = sum(n-1)

	MOVQ n+0(FP), AX       // AX = n
	ADDQ AX, BX            // BX += AX
	MOVQ BX, result+8(FP)  // return BX
	RET

L_END:
	MOVQ $0, result+8(FP) // return 0
    RET
```

在汇编版本函数中并没有定义局部变量，只有用于调用自身的临时栈空间。因为函数本身的参数和返回值有16个字节，因此栈帧的大小也为16字节。L_STEP_TO_END标号部分用于处理递归调用，是函数比较复杂的部分。L_END用于处理递归终结的部分。

调用sum函数的参数在`0(SP)`位置，调用结束后的返回值在`8(SP)`位置。在函数调用之后要需要重新为需要的寄存器注入值，因为被调用的函数内部很可能会破坏了寄存器的状态。同时调用函数的参数值也可信任的，输入参数也可能在被调用函数内部被修改了值。

总得来说用汇编实现递归函数和普通函数并没有什么区别，当然是在没有考虑爆栈的前提下。我们的函数应该可以对较小的n进行求和，但是当n大到一定层度，也就是栈达到一定的深度，必然会出现爆栈的问题。爆栈是C语言的特性，不应该在哪怕是Go汇编语言中出现。

## 栈的扩容和收缩

Go语言的编译器在生成函数的机器代码时，会在开头插入以小段代码。插入的代码可以做很多事情，包括触发runtime.Gosched进行协作式调度，还包括栈的动态增长等。其实栈等扩容工作主要在runtime包的runtime·morestack_noctxt函数实现，这是一个底层函数，只有汇编层面才可以调用。

在新版本的sum汇编函数中，我们在开头和末尾都引入了部分代码：

```
// func sum(n int) int
TEXT ·sum(SB), $16-16
	NO_LOCAL_POINTERS

L_START:
	MOVQ TLS, CX
	MOVQ 0(CX)(TLS*1), AX
	CMPQ SP, 16(AX)
	JLS  L_MORE_STK

	// 原来的代码

L_MORE_STK:
	CALL runtime·morestack_noctxt(SB)
	JMP  L_START
```

其中NO_LOCAL_POINTERS表示没有局部指针。因为新引入的代码可能导致调用runtime·morestack_noctxt函数，而栈的扩容必然要涉及函数参数和局部编指针的调整，如果缺少局部指针信息将导致扩容工作无法进行。不仅仅是栈的扩容需要函数的参数和局部指针标记表格，在GC进行垃圾回收时也将需要。函数的参数和返回值的指针状态可以通过在Go语言中的函数声明中获取，函数的局部变量则需要手工指定。因为手工指定指针表格是一个非常繁琐的工作，因此一般要避免在手写汇编中出现局部指针。

喜欢深究的读者可能会有一个问题：如果进行垃圾回收或栈调整时，寄存器中的指针时如何维护的？前文说过，Go语言的函数调用时通过栈进行传递参数的，并没有使用寄存器传递参数。同时函数调用之后所有的寄存器视为失效。因此在调整和维护指针时，只需要扫描内存中的指针数据，寄存器中的数据在垃圾回收器函数返回后都需要重新加载，因此寄存器是不需要扫描的。

在Go语言的Goroutine实现中，每个TlS线程局部变量会保存当前Goroutine的信息结构体的指针。通过`MOVQ TLS, CX`和`MOVQ 0(CX)(TLS*1), AX`两条指令将表示当前Goroutine信息的g结构体加载到CX寄存器。g结构体在`$GOROOT/src/runtime/runtime2.go`文件定义，开头的结构成员如下：

```go
type g struct {
	// Stack parameters.
	// stack describes the actual stack memory: [stack.lo, stack.hi).
	// stackguard0 is the stack pointer compared in the Go stack growth prologue.
	// It is stack.lo+StackGuard normally, but can be StackPreempt to trigger a preemption.
	// stackguard1 is the stack pointer compared in the C stack growth prologue.
	// It is stack.lo+StackGuard on g0 and gsignal stacks.
	// It is ~0 on other goroutine stacks, to trigger a call to morestackc (and crash).
	stack       stack   // offset known to runtime/cgo
	stackguard0 uintptr // offset known to liblink
	stackguard1 uintptr // offset known to liblink

	...
}
```

第一个成员是stack类型，表示当前栈的开始和结束地址。stack的定义如下：

```go
// Stack describes a Go execution stack.
// The bounds of the stack are exactly [lo, hi),
// with no implicit data structures on either side.
type stack struct {
	lo uintptr
	hi uintptr
}
```

在g结构体中的stackguard0成员是出现爆栈前的警戒线。stackguard0的偏移量是16个字节，因此上述代码中的`CMPQ SP, 16(AX)`表示将当前的真实SP和爆栈警戒线比较，如果超出警戒线则表示需要进行栈扩容，也就是跳转到L_MORE_STK。在L_MORE_STK标号处，线调用runtime·morestack_noctxt进行栈扩容，然后又跳回到函数到开始位置，此时此刻函数到栈已经调整了。然后再进行一次栈大小到检测，如果依然不足则继续扩容，直到栈足够大为止。

以上是栈的扩容，但是栈到收缩是在何时处理到呢？我们知道Go运行时会定期进行垃圾回收操作，这其中栈的回收工作。如果栈使用到比例小于一定到阈值，则分配一个较小到栈空间，然后将栈上面到数据移动到新的栈中，栈移动的过程和栈扩容的过程类似。

## PCDATA和PCDATA

TODO

## 方法函数

TODO
