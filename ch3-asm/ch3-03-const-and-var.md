# 3.3 Constants and Global Variables

The initial values ​​of all variables in the program are generated directly or indirectly by constant or constant expressions. Many variables in the Go language are initialized with default zero values, but the variables defined in the Go assembly are preferably initialized manually by constants. With constants, you can derive global variables and initialize various other variables using expressions composed of constants. This section will briefly discuss the use of constants and global variables in Go assembly language.

## 3.3.1 Constants

Constants in Go assembly language are prefixed by the $ dollar sign. The types of constants are integer constants, floating-point constants, character constants, and string constants. The following are examples of several types of constants:

```
$1 // decimal
$0xf4f8fcff // hex
$1.5 // floating point number
$'a' // character
$"abcd" // string
```

The integer type constants are by default in decimal format, and integer constants can also be expressed in hexadecimal format. All constants must eventually match the size of the variable memory to be initialized.

For numeric constants, new constants can be constructed from constant expressions:

```
$2+2 // constant expression
$3&1<<2 // == $4
$(3&1)<<2 // == $4
```

The precedence of the operators in the constant expression is consistent with the Go language.

Constants in Go assembly language are more than just compile-time constants, they also contain runtime constants. For example, the global variables and global functions in the package are fixed at runtime, and the address of the package and the address of the function where the address does not change is also an assembly constant.

The following is the string code defined in assembly in the first section of this chapter:

```
GLOBL · NameData(SB), $8
DATA · NameData(SB)/8, $"gopher"

GLOBL · Name(SB), $16
DATA · Name+0(SB)/8, $·NameData(SB)
DATA · Name+8(SB)/8, $6
```

The `$·NameData(SB)` is also prefixed with the $ dollar sign, so it can also be thought of as a constant, which corresponds to the address of the NameData package variable. In the assembly instructions, we can also get the address of the NameData variable through the LEA instruction.


## 3.3.2 Global Variables

In Go, variables have global and local variables based on scope and life cycle. Global variables are package-level variables. Global variables generally have a relatively fixed memory address, and the declaration period spans the entire program runtime. Local variables are generally defined variables within a function. They are created on the stack only when the function is executed. When the function call is completed, it will be reclaimed (temporarily not considering the problem of closures catching local variables).

From the perspective of Go assembly language, global variables and local variables have very large differences. In Go assembly, global variables are more similar to global functions. They refer to the corresponding memory through an artificially defined symbol. The only difference is whether the memory stores data or instructions to be executed. Because the instructions in the von Neumann system are also data, and the instructions and data are stored in a uniformly addressed memory. Because there is no essential difference between instructions and data, we can even generate instructions dynamically as we manipulate data (this is the principle of all JIT techniques). Local variables need to be implicitly defined by the SP stack space after understanding the assembly function.

In Go assembly language, memory is located through the SB pseudo-register. SB is the abbreviation of Static base pointer, which means the starting address of static memory. We can think of SB as a byte array of the same size as the content capacity. All static global symbols can usually be located by SB plus an offset, and the symbol we define is actually the starting address offset relative to the SB memory. . For the SB pseudo-register, there is no difference between the global variable and the global function symbol.

To define a global variable, first declare the symbol corresponding to a variable, and the memory size corresponding to the variable. The syntax for exporting variable symbols is as follows:

```
GLOBL symbol(SB), width
```

The GLOBL assembly instruction is used to define a variable named symbol. The memory width of the variable is width, and the memory width part must be initialized with a constant. The following code defines an int32 type count variable by assembly:

```
GLOBL ·count(SB), $4
```

The symbol `·count` starts with a midpoint to indicate the current package variable, and the final symbol name is expanded to `path/to/pkg.count`. The size of the count variable is 4 bytes, and the constant must start with the $ dollar sign. The width of the memory must be an exponential multiple of 2, and the compiler will eventually ensure that the real address of the variable is aligned to the machine word multiple. It should be noted that in the Go assembly we are unable to specify a specific type for the count variable. When defining global variables in assembly, we only care about the variable name and memory size. The final type of the variable can only be declared in the Go language.

After the variable is defined, we can specify the data in the corresponding memory through the DATA assembly instruction. The syntax is as follows:

```
DATA symbol+offset(SB)/width, value
```

The specific meaning is that starting from the symbol+offset offset, the width-width memory is initialized with the value corresponding to the value constant. When DATA initializes memory, width must be one of several widths of 1, 2, 4, and 8, because the larger memory cannot be represented by a uint64 size at a time.

For the int32 type count variable, we can either initialize byte by byte or initialize it once:

```
DATA ·count+0(SB)/1, $1
DATA ·count+1(SB)/1, $2
DATA ·count+2(SB)/1, $3
DATA ·count+3(SB)/1, $4

// or

DATA ·count+0(SB)/4, $0x04030201
```

Since the X86 processor is a small endian, initializing all 4 bytes with hexadecimal 0x04030201 is the same effect as initializing 4 bytes one by one with 1, 2, 3, and 4.

Finally, you need to declare the corresponding variable in the Go language (similar to the C language header declaration variable), so that the garbage collector will manage the pointer-related memory data according to the type of the variable.


### 3.3.2.1 Array type

Arrays in assembly are also a very simple type. An array in the Go language is a basic type with a flat memory structure. So the `[2]byte` type and the `[1]uint16` type have the same memory structure. The situation becomes slightly more complicated only when the array and structure are combined.

Below we try to define an array variable num of type `[2]int` with assembly:

```go
Var num [2]int
```

Then define a variable corresponding to 16 bytes in the assembly and initialize it with a zero value:

```
GLOBL · num(SB), $16
DATA ·num+0(SB)/8, $0
DATA ·num+8(SB)/8, $0
```

The following figure shows the correspondence between Go statements and assembly statements when defining variables:

![](../images/ch3-4-pkg-var-decl-01.ditaa.png)

*Figure 3-4 Variable Definition*


The NOPTR flag is not needed in assembly code because the Go compiler deduces that there is no pointer data inside the variable from the `[2]int` type declared by the Go language statement.


### 3.3.2.2 bool type variable

Go assembly language definition variables cannot specify type information, so you need to declare the type of the variable first through the Go language. Here are a few bool type variables declared in the Go language:

```go
Var (
boolValue bool
trueValue bool
falseValue bool
)
```

Variables declared in the Go language cannot contain initialization statements. Then the following is the assembly definition of the amd64 environment:

```
GLOBL · boolValue(SB), $1 // uninitialized

GLOBL · trueValue(SB), $1 // var trueValue = true
DATA · trueValue(SB)/1, $1 // non 0 are true

GLOBL · falseValue(SB), $1 // var falseValue = true
DATA · falseValue(SB)/1, $0
```

The size of the bool type is 1 byte. And the variables defined in the assembly need to manually specify the initialization value, otherwise it will lead to uninitialized variables. When you need to load a 1-byte bool type variable into an 8-byte register, you need to use the MOVBQZX instruction to fill the low-order high bits with 0.

### 3.3.2.3 int variable

All integer types have similar definitions. The big difference is the integer size of the memory size and whether the integer is signed. The following are the declared int32 and uint32 type variables:

```go
Var int32Value int32

Var uint32Value uint32
```

Variables declared in the Go language cannot contain initialization statements. Then the following is the assembly definition of the amd64 environment:

```
GLOBL · int32Value(SB), $4
DATA · int32Value+0(SB)/1, $0x01 // 0th byte
DATA ·int32Value+1(SB)/1,$0x02 // 1st byte
DATA · int32Value+2(SB)/2, $0x03 // 3-4 bytes

GLOBL · uint32Value(SB), $4
DATA · uint32Value(SB)/4, $0x01020304 // 1-4 bytes
```

Initializing data when assembling variables is not a distinction between integers and symbols. Only when the CPU instruction processes the register data, will the type of the data be sorted according to the type of the instruction or whether it has a sign bit.

### 3.3.2.4 float variable

Go assembly language usually cannot distinguish whether a variable is a floating-point type, and the associated floating-point machine instructions treat the variable as a floating point number. The floating-point number of the Go language follows the IEEE754 standard and is divided into float32 single-precision floating-point numbers and float64 double-precision floating-point numbers.

In the IEEE 754 standard, the highest bit 1 bit is the sign bit, then the exponent bit (the index is represented by the frameshift format), and then the significant part (where one bit to the left of the decimal point is omitted). The following figure shows the bit layout of float32 type floating point numbers in IEEE754:

![](../images/ch3-5-ieee754.jpg)

*Figure 3-5 IEEE754 floating point structure*


IEEE754 floating-point numbers also have some wonderful features: for example, there are two positive and negative zeros; in addition to infinity and infinitesimal Inf and non-numbered NaN; and if two floating-point numbers are ordered, the corresponding signed integers are also ordered (or vice versa) Not necessarily true, because the non-numbers in the floating point number are not sortable). Floating point numbers are the most difficult corners in the program, because many handwritten floating-point numeric denomination constants in the program cannot be accurately expressed at all. The error rounding method involved in floating-point calculations may also be random.

The following is to declare two floating-point numbers in the Go language (if no variables are defined in the assembly, the declaration will also define the variables).

```go
Var float32Value float32

Var float64Value float64
```

Then define and initialize the two floating point numbers declared above in the assembly:

```
GLOBL ·float32Value(SB), $4
DATA ·float32Value+0(SB)/4,$1.5 // var float32Value = 1.5

GLOBL ·float64Value(SB), $8
DATA ·float64Value(SB)/8,$0x01020304 // bit mode initialization
```

In the previous section of the simplified arithmetic instructions, we are all targeting integers. If you want to process floating-point numbers by integer instructions, the addition and subtraction of floating-point numbers must be based on the floating-point arithmetic rules: first align the decimal point, then integer addition and subtraction, and finally The results are normalized and the precision rounding problem is handled. However, in the current mainstream CPU, it provides a proprietary calculation instruction for floating point numbers.

### 3.3.2.5 string type variable

From the perspective of Go assembly language, a string is just a structure. The header structure of string is defined as follows:

```go
Type reflect.StringHeader struct {
Data uintptr
Len int
}
```

StringHeader has a size of 16 bytes in the amd64 environment, so we first declare the string variable in Go code and then define a 16-byte variable in the assembly:

```go
Var helloworld string
```

```
GLOBL · helloworld (SB), $16
```

At the same time we can prepare real data for the string. In the following compilationIn the code, we define a private variable in the current file of text (with the suffix ``>`), the content is "Hello World!":

```
GLOBL text<>(SB),NOPTR,$16
DATA text<>+0(SB)/8,$"Hello Wo"
DATA text<>+8(SB)/8,$"rld!"
```

Although the string represented by the `text<>` private variable is only 12 characters long, we still need to extend the length of the variable to an exponential multiple of 2, which is the length of 16 bytes. Where `NOPTR` indicates that `text<>` does not contain pointer data.

Then use the constant corresponding to the memory address corresponding to the text private variable to initialize the Data part of the string header structure, and manually specify the Len part as the length of the string:

```
DATA ·helloworld+0(SB)/8,$text<>(SB) // StringHeader.Data
DATA ·helloworld+8(SB)/8,$12 // StringHeader.Len
```

It should be noted that the string is a read-only type, to avoid directly modifying the contents of the underlying data of the string in the assembly.

### 3.3.2.6 slice type variable

The slice variable is similar to the string variable, except that it corresponds to the slice header structure. The structure of the slice header is as follows:

```go
Type reflect.SliceHeader struct {
Data uintptr
Len int
Cap int
}
```

The comparison shows that the first two member strings of the sliced ​​header are the same. So we can extend the Cap member to the slice type based on the previous string variable:

```go
Var helloworld []byte
```

```
GLOBL · helloworld (SB), $24 // var helloworld []byte("Hello World!")
DATA ·helloworld+0(SB)/8,$text<>(SB) // StringHeader.Data
DATA ·helloworld+8(SB)/8,$12 // StringHeader.Len
DATA ·helloworld+16(SB)/8,$16 // StringHeader.Cap

GLOBL text<>(SB), $16
DATA text<>+0(SB)/8,$"Hello Wo" // ...string data...
DATA text<>+8(SB)/8,$"rld!" // ...string data...
```

Because of the compatibility of slices and strings, we can temporarily use the first 16 bytes of the slice header as a string, which eliminates unnecessary conversions.

### 3.3.2.7 map/channel type variable

Types such as map/channel do not have an exposed internal structure. They are just pointers of unknown types and cannot be initialized directly. In assembly code we can only define and initialize a value for a similar variable:

```go
Var m map[string]int

Var ch chan int
```

```
GLOBL ·m(SB), $8 // var m map[string]int
DATA ·m+0(SB)/8, $0

GLOBL ·ch(SB), $8 // var ch chan int
DATA ·ch+0(SB)/8, $0
```

In fact, some helper functions are provided for assembly in the runtime package. For example, in the assembly you can create map and chan variables through the runtime.makemap and runtime.makechan internal functions. The signature of the helper function is as follows:

```go
Func makemap(mapType *byte, hint int, mapbuf *any) (hmap map[any]any)
Func makechan(chanType *byte, size int) (hchan chan any)
```

It should be noted that makemap is a generic function that can create different types of maps. The specific type of map is specified by the mapType parameter.


## 3.3.3 Memory layout of variables

We have repeatedly stressed that variables are not typed in Go assembly language. So there are different types of variables in the Go language, and the underlying may correspond to the same memory structure. A deep understanding of the memory layout of each variable is a prerequisite for assembly programming.

First look at the memory layout of the `[2]int` type array we have seen before:

![](../images/ch3-6-pkg-var-decl-02.ditaa.png)

*Figure 3-6 Variable Definition*


The variable allocates space in the data segment, and the element addresses of the array are sequentially arranged from low to high.

Then look at the memory layout of the `image.Point` structure type variable in the standard library image package:

![](../images/ch3-7-pkg-var-decl-03.ditaa.png)

*Figure 3-7 Structure Variable Definition*


The variable also allocates space in the data segment, and the address of the variable structure member is also sequentially arranged from low to high.

So the bottom of the `[2]int` and `image.Point` types have approximately the same memory layout.

## 3.3.4 Identifier rules and special signs

The identifier of the Go language can be located by the absolute packet path plus the identifier itself, so the identifiers in different packages will have no problem even if they have the same name. Go assembly uses special symbols to represent slashes and dot notation, because it simplifies the assembly of the assembler lexical part of the code, as long as the string is replaced.

The following are the common uses of several identifiers in assembly (usually also for function identifiers):

```
GLOBL ·pkg_name1(SB), $1
GLOBL main·pkg_name2(SB), $1
GLOBL my/pkg·pkg_name(SB), $1
```

In addition, Go assembly can define a private identifier that can be accessed only by the current file (similar to the statically modified variable in the C language), with the suffix of `<>`:

```
GLOBL file_private<>(SB), $1
```

This can reduce the interference of private identifiers on the naming of identifiers in other files.

In addition, Go assembly language also defines some flags in the "textflag.h" file. The flags used for variables are DUPOK, RODATA, and NOPTR. DUPOK means that there may be more than one identifier corresponding to the variable. Only one of them can be selected when linking (usually used to merge the same constant string to reduce the space occupied by duplicate data). The RODATA flag indicates that the variable is defined in a read-only memory segment, so any subsequent modification of this variable will result in an exception (recover also cannot be captured). NOPTR indicates that this variable contains no pointer data internally, and the garbage collector ignores the scan of the variable. If the variable has already been declared in the Go code, the Go compiler will automatically analyze whether the variable contains a pointer. In this case, you can skip the handwritten NOPTR flag.

For example, the following example is to define a read-only variable of type int by assembly:

```go
Var const_id int // readonly
```

```
#include "textflag.h"

GLOBL · const_id(SB), NOPTR|RODATA, $8
DATA ·const_id+0(SB)/8, $9527
```

We use the #include statement to include the "textflag.h" header file that defines the flag (same as the preprocessing in C). Then the GLOBL assembly command adds two flags NOPTR and RODATA to the variable (multiple flags are separated by a vertical bar), indicating that there is no pointer data in the variable and is defined in the read-only data segment.

Variables are also generally called the value of the address, but const_id can take the address, but it can't be modified. The limit that cannot be modified is not provided by the compiler, but because changes to the variable cause writes to the read-only memory segment, causing an exception.

## 3.3.5 Summary

Above we have initially demonstrated the use of global variables defined by assembly. But in a real environment we don't recommend defining variables by assembly - because it's easier and safer to define variables in Go. Defining variables in the Go language, the compiler can help us calculate the size of the variable, generate the initial value of the variable, and also contain enough type information. The advantage of assembly language is the nature and performance of the excavator. It is not possible to use these variables to define these variables. Therefore, after understanding the usage of the assembly definition variables, it is recommended that you use it with caution.