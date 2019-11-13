# 3.2 Computer Structure

Assembly language is a programming language that faces the computer, so understanding the computer structure is a prerequisite for mastering assembly language. The current popular computer basically uses the von Neumann computer architecture (and Harvard architecture in some special areas). The von Neumann structure, also known as the Princeton structure, uses a storage structure that stores program instructions and data together. The instruction and data memory in the von Neumann computer actually refers to the memory in the computer, and then together with the CPU processor constitutes the simplest computer.

Assembly language is actually a very simple programming language because the computer model it is oriented to is very simple. There are several reasons why people feel that assembly language is difficult to learn: different types of CPUs have their own set of instructions; even the same CPU, 32-bit and 64-bit modes of operation will still be different; different assembly tools also have Its own unique assembly instructions; different operating systems and high-level programming languages ​​and the underlying assembly call specifications are not the same. This section will describe several interesting assembly language models, and finally streamline a streamlined instruction set for the AMD64 architecture to facilitate the learning of Go assembly language.


## 3.2.1 Turing machine and BF language

The Turing machine is an abstract computing model proposed by Turing. The machine has an infinitely long strip of paper that is divided into small squares, each with a different color, similar to the memory in a computer. At the same time, the machine has a probe that moves around the tape, similar to reading and writing data on the memory through the memory address. The machine head has a set of internal calculation states, as well as some fixed programs (more like a Harvard structure). At each moment, the machine head reads a square message from the current tape, and then outputs the information to the tape square according to its internal state and the current program instructions to be executed, and updates its internal state. And move.

Although the Turing machine is not easy to program, it is very easy to understand. There is a minimalist BrainFuck computer language that works very similarly to the Turing machine. BrainFuck was founded in 1993 by Urban Müller, referred to as BF. Müller's original design goal was to create a simple programming language that was implemented in a minimal compiler with Turing's complete idea. This language consists of eight states. The compiler (second edition) written for the Amiga machine in the early days was only 240 bytes in size!

As its name suggests, the brainfuck program is hard to read. Still, the brainfuck Turing machine can do any computing task. Although brainfuck is so different in its calculations, it does work correctly. This language is based on a simple machine model. In addition to instructions, this machine also includes: an array initialized to zero in bytes, a pointer to the array (initially pointing to the first byte of the array) And two byte streams for input and output. This is a language that is well-formed according to Turing. Its main design idea is to implement a "simple" language with minimal concepts. The BrainFuck language has only eight symbols, and all operations are done by a combination of these eight symbols.

The following is a description of these eight states, each of which is identified by a character:

| Character | C language analogy | Meaning
| --- | ----------------- | ------
| `>` | `++ptr;` | Pointer plus one
| `<` | `--ptr;` | pointer minus one
| ``` | `++*ptr;` | The value of the byte pointed to by the pointer plus one
| ``` | `--*ptr;` | The value of the byte pointed to by the pointer minus one
|`.` | `putchar(*ptr);` | The unit content pointed to by the output pointer (ASCII code)
| `,` | `*ptr = getch();` | Enter the content to the unit pointed to by the pointer (ASCII code)
|`[` | `while(*ptr) {}` | If the pointer points to a zero value, jump back to the next instruction of the corresponding `]` instruction
| `]` | | If the value of the cell pointed to by the pointer is not zero, jump forward to the next instruction of the corresponding `[` instruction

Below is a brainfuck program that prints a "hi" string to standard output:

```
++++++++++[>++++++++++<-]>++++.+.
```

In theory, we can use the BF language as the target machine language. After compiling other high-level languages ​​into the BF language, we can run it on the BF machine.

## 3.2.2 Human Resources Machine Game

"Human Resource Machine" is a well-designed assembly language programming game. In the game, the player plays a staff role to simulate the operation of the HR machine. By accomplishing every goal given by the boss to achieve the goal of promotion, the way to complete the task is to use the 11 machine instructions provided by the game to write the correct assembler, and finally get the correct output. The assembly language of the HR machine can be considered as a common assembly language across platforms and operating systems, because the gameplay is exactly the same on macOS, Windows, Linux and iOS.

The machine model of the human resource machine is very simple: the INBOX command corresponds to the input device, the OUTBOX corresponds to the output device, the player villain corresponds to a register, the floor corresponding to the temporary storage of data corresponds to the memory, and then the basic instructions of data transmission, addition, subtraction, and jump. There are a total of 11 machine instructions:

| Name | Explanation |
| -------- | ---
| INBOX | Take an integer data from the input channel and put it in the hand (register)
| OUTBOX | Put the data in the hand (register) into the output channel, and then there will be no data in hand (some instructions cannot be run at this time)
| COPYFROM | Copy the data from a numbered grid on the floor to the hand (the data before the hand is invalid), the floor grid must have data
| COPYTO | Copy the data in the hands (registers) to a numbered grid on the floor, the data in the hand is unchanged
| ADD | Add the data in the hand (register) to the data of the floor grid corresponding to a certain number, and put the new data in the hands (the data before the hand is invalid)
| SUB | Subtracts the data in the hand (register) from the data of the floor grid corresponding to a number, and puts the new data in the hands (the previous data in the hands is invalid)
| BUMP+ | Self-plus one
| BUMP- | Self-reduction
| JUMP | Jump
| JUMP =0 | Zero conditional jump
| JUMP <0 | is a negative conditional jump

In addition to machine instructions, some parts of the game also provide a register-like place for storing temporary data. Machine instructions for human resource machine games are mainly divided into the following categories:

- Input/Output (INBOX, OUTBOX): There will be only 1 new data in the hand after input, and there will be no data in the hand after the output.
- Data transfer instruction (COPYFROM/COPYTO): It is mainly used for data transmission between only one register (hand) and memory. Make sure the source data is valid during transmission.
- Arithmetic related (ADD/SUB/BUMP+/BUMP-)
- Jump instruction: If it is a conditional jump, there must be data in the register

Mainstream processors have similar instructions. In addition to the basic arithmetic and logic budget instructions, with the conditional jump instructions, you can implement common control flow structures such as branches and loops.

The following figure is a task of a certain layer: the 0 of the input data is culled, the data of non-zero is sequentially output, and the right part is the solution.

![](../images/ch3-1-arch-hsm-zero.jpg)

*Figure 3-1 Human Resources Machine*


The entire program has only one input instruction, one output instruction and two jump instructions.

```
LOOP:
INBOX
JUMP-if-zero LOOP
OUTBOX
JUMP LOOP
```

First, read a data packet through the INBOX instruction; then determine whether the data of the package is 0. If it is 0, jump to the beginning to continue reading the next data packet; otherwise, the data packet will be output, and then jump to the beginning. The data package is processed endlessly in this loop until the task is completed and promoted to a higher level, and then a similar but more complex task is handled.


## 3.2.3 X86-64 architecture

X86 is actually the abbreviation of 80X86 (the last three letters), including Intel 8086, 80286, 80386 and 80486 and other instruction sets, so its architecture is called x86 architecture. The x86-64 is a 64-bit extension of the x86 architecture designed by AMD in 1999 and is backward compatible with 16-bit and 32-bit x86 architectures. X86-64 is currently officially named AMD64, which is the AMD64 specified by the GOARCH environment variable in the Go language. The assembler in this chapter is for a 64-bit X86-64 environment, unless otherwise noted.

You must understand the corresponding CPU architecture before using assembly language. The following is an X86/AMD architecture diagram:

![](../images/ch3-2-arch-amd64-01.ditaa.png)

*Figure 3-2 AMD64 Architecture*


The memory section on the left is a common memory layout. The text generally corresponds to the code segment, and is used to store the instruction data to be executed, and the code segment is generally read-only. Then there is the data segment of rodata and data. The data segment is generally used to store global data, where rodata is a read-only data segment. The heap segment is used to manage dynamic data, and the stack segment is used to manage the data associated with each function call. In assembly language, the focus is mainly on the text code segment and the data segment. Therefore, the Go assembly language specifically provides the corresponding TEXT and DATA commands for defining code and data.

In the middle is the register provided by X86. Registers are the most important resources in the CPU. In principle, each memory data to be processed needs to be placed in a register before it can be processed by the CPU. At the same time, the processed result in the register needs to be stored in the memory. In addition to the special registers of the status register FLAGS and the instruction register IP, there are several general-purpose registers of AX, BX, CX, DX, SI, DI, BP, and SP. Eight more general-purpose registers named after R8-R15 have been added to the X86-64. For historical reasons, R0-R7 are not general-purpose registers, they are just registers specific to the MMX instructions that X87 began to introduce. In the general-purpose registers, BP and SP are two special registers: BP is used to record the start position of the current function frame, and the instruction related to the function call implicitly affects the value of BP; SP corresponds to the position of the current stack pointer. The stack-related instructions implicitly affect the SP value; some debug tools require a BP register to work properly.

On the right is the X86 instruction set. The CPU is composed of instructions and registers. The instructions are built-in algorithms for each CPU. The objects to be processed by the instructions are all registers and memory. We can think of each instruction as a function provided in the CPU built-in standard library, and then the process of constructing a more complex program based on these functions is the process of programming in assembly language.


## 3.2.4 Pseudo-registers in Go assembly

Go assembly In order to simplify the compilation of assembly code, four pseudo-registers of PC, FP, SP, and SB are introduced. The four pseudo-registers plus other general-purpose registers are the re-abstractions of the CPU by the Go assembly language. The abstract structure is also applicable to other non-X86-type architectures.

The relationship between the four pseudo registers and the memory and registers of X86/AMD64 is as follows:

![](../images/ch3-3-arch-amd64-02.ditaa.png)

*Figure 3-3 Go assembly pseudo-register*


In the AMD64 environment, the pseudo PC register is actually an alias for the IP instruction counter register. The pseudo FP register corresponds to the frame pointer of the function, which is generally used to access the parameters and return values ​​of the function. The pseudo SP stack pointer corresponds to the bottom of the current function stack frame (excluding the parameters and return value parts), and is generally used to locate local variables. The pseudo SP is a special register because there is also an SP true register with the same name. The true SP register corresponds to the top of the stack and is generally used to locate parameters and return values ​​that call other functions.

One thing to keep in mind when you need to distinguish between a pseudo-register and a true register is that a pseudo-register generally requires an identifier and an offset as a prefix, and if there is no identifier prefix, it is a true register. For example, `(SP)`, `+8(SP)` have no identifier prefix for the true SP register, and `a(SP)`, `b+8(SP)` have the identifier prefixed with the pseudo register.

## 3.2.5 X86-64 instruction set

Many assembly language tutorials emphasize that assembly language is not portable. Strictly speaking, assembly language is not portable under different CPU types, or different operating system environments, or different assembly tool chains, and the machine instructions running in the same CPU are exactly the same. The non-portability of assembly language is a great obstacle to its popularity. Although the difference in the CPU instruction set is a large factor that causes poor portability, the related toolchain of assembly language has an unshirkable responsibility for this. The Go assembly language from Plan9 has made some improvements: First, the Go assembly language is identical on the same CPU architecture, which shields the differences in the operating system; while the Go assembly language will have some basic and similar instructions. Abstraction is a pseudo-instruction of the same name, thereby reducing the difference in assembly code under different CPU architectures (the difference in register names and numbers is always present). The purpose of this section is also to find a smaller, reduced instruction set to simplify the learning of Go assembly language.

X86 is an extremely complicated system. Some people have statistics on x86-64.A thousand. Not only that, but many of the single instructions in X86 are also very powerful. For example, a paper proves that only one MOV instruction can form a Turing-complete system. These are two extreme cases. Too many instructions and too few instructions are not conducive to the preparation of the assembler, but also reflect the importance of the MOV instruction.

General-purpose basic machine instructions can be roughly classified into data transfer instructions, arithmetic operations and logic operations instructions, control flow instructions, and other instructions. So we can try to streamline an X86-64 instruction set to facilitate the learning of Go assembly language.

So let's take a look at the important MOV instructions. The MOV instruction can be used to move a literal value to a register, a literal value to a memory, a data transfer between registers, and a data transfer between a register and a memory. It should be noted that the MOV transfer instruction can only have one memory operand, and can achieve similar purposes through a temporary register. The simplest is to ignore the data transfer operation of the sign bit. Like the AMD64 instruction, the 386 has different instructions for different widths of 1, 2, 4 and 8 bytes:

| Data Type | 386/AMD64 | Comment |
| --------- | ----------- | ------------- |
| [1]byte | MOVB | B => Byte |
| [2]byte | MOVW | W => Word |
| [4]byte | MOVL | L => Long |
| [8]byte | MOVQ | Q => Quadword |

The MOV instruction is not only used to transfer data between registers and memory, but it can also be used to handle data expansion and truncation operations. When the data width and the width of the register are different and the sign bit needs to be processed, 386 and AMD64 have different instructions:

| Data Type | 386 | AMD64 | Comment |
| --------- | ------- | ------- | ------------- |
| int8 | MOVBLSX | MOVBQSX | sign extend |
| uint8 | MOVBLZX | MOVBQZX | zero extend |
| int16 | MOVWLSX | MOVWQSX | sign extend |
| uint16 | MOVWLZX | MOVWQZX | zero extend |

For example, when you need to convert an int64 type data to a bool type, you need to use the MOVBQZX instruction.

Basic arithmetic instructions include ADD, SUB, MUL, DIV, etc. Among them, ADD, SUB, MUL, and DIV are used for addition, subtraction, multiplication, and division, and the final result is stored in the target register. The basic logical operation instructions have several instructions such as AND, OR, and NOT, corresponding to several instructions such as logical AND, OR and inversion.

| Name | Explanation |
| ------ | ---
| ADD | Addition
| SUB | Subtraction
| MUL | Multiplication
| DIV | Division
| AND | Logic and
| OR | Logical OR
| NOT | Logic inversion

Among them, arithmetic and logic instructions are the basis of sequential programming. A more complex branch or loop structure can be implemented by a logical comparison affecting the status register, combined with a conditional jump instruction. It should be noted that multiply and divide instructions such as MUL and DIV may implicitly use some registers. Please refer to the relevant manual for details of the instructions.

Control flow instructions include CMP, JMP-if-x, JMP, CALL, and RET. The CMP instruction is used for subtraction of two operands. The sign bit and zero bit of the status register are set according to the comparison result, which can be used for the jump condition of conditional jump. JMP-if-x is a set of conditional jump instructions. Commonly used are JL, JLZ, JE, JNE, JG, JGE, etc., which are skipped when the conditions are less than, less than or equal to, equal to, not equal to, greater than, greater than or equal to turn. The JMP instruction corresponds to an unconditional jump, and the jump is set by setting the address to be jumped to the IP instruction register. The CALL and RET instructions return instructions for calling functions and functions, respectively.

| Name | Explanation |
| -------- | ---
| JMP | Unconditional Jump
| JMP-if-x | Conditional Jump, JL, JLZ, JE, JNE, JG, JGE
| CALL | Call function
| RET | function returns

Unconditional and conditional adjustment instructions are the basic instructions for implementing branch and loop control flows. In theory, we can also implement the function call and return functions through jump instructions. However, because the current function is already the most basic abstraction in modern computers, most CPUs provide proprietary instructions and registers for function calls and returns.

Other important instructions are LEA, PUSH, POP, etc. The LEA instruction loads the memory address in the standard parameter format into the register (instead of loading the contents of the memory location). PUSH and POP are push and pop instructions, respectively. SP in the general-purpose register is the stack pointer, and the stack grows toward the lower address.

| Name | Explanation |
| ------ | ---
| LEA | Address
| PUSH | Push stack
| POP | Popping

When you need to access the memory corresponding to some members such as an array or a structure by indirect indexing, you can use the LEA instruction to access the current internal access address first, and then operate the corresponding memory data. The stack instruction can be used to adjust the size of its stack space.

Finally, it should be noted that Go assembly language may not support all CPU instructions. If you encounter an unsupported CPU instruction, you can fill the corresponding machine code of the real CPU instruction to the corresponding location by the BYTE command provided by the Go assembly language. The full X86 instructions are defined in the https://github.com/golang/arch/blob/master/x86/x86.csv file. At the same time, Go assembly is also defining aliases for some instructions. For details, please refer to https://golang.org/src/cmd/internal/obj/x86/anames.go.