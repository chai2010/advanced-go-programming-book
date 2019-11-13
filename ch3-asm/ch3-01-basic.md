# 3.1 Quick Start

The Go assembler is always a ghostly existence. We will implement a simple output program by analyzing the assembly code of a simple Go program output and then using the assembly.

## 3.1.1 Implementation and Declaration

Go assembly language is not a separate language because the Go assembler cannot be used independently. The Go assembly code must be organized in Go packages, and at least one Go language file in the package should be used to indicate basic package information such as the current package name. If the variables and functions defined in the Go assembly code are to be referenced by other Go language code, the symbols defined in the assembly need to be declared through the Go language code. Definitions of Variables and Definition of Functions The Go assembly file is similar to the .c file in C language, and the Go source file used to derive the symbols defined in the assembly is similar to the C language.h file.

## 3.1.2 Defining integer variables

For the sake of simplicity, we first define and assign an integer variable in Go, and then look at the generated assembly code.

First create a pkg.go file with the following contents:

```go
Package pkg

Var Id = 9527
```

Only one package-level variable of type int is defined in the code and initialized. Then use the following command to view the pseudo assembly code corresponding to the Go language program:

```
$ go tool compile -S pkg.go
"".Id SNOPTRDATA size=8
  0x0000 37 25 00 00 00 00 00 00 '.......
```

The `go tool compile` command is used to call the underlying command tool provided by the Go language, where the `-S` parameter indicates the output assembly format. The output assembly is relatively simple, where `"".Id` corresponds to the Id variable symbol, and the variable's memory size is 8 bytes. The initialization content of the variable is `37 25 00 00 00 00 00 00`, corresponding to 0x2537 in hexadecimal format, corresponding to 9527 decimal. SNOPTRDATA is a related flag, where NOPTR indicates that the pointer data is not included in the data.

The above content is only the assembly of the target file, and it is not exactly equivalent to the Go assembly language. The Go language official website comes with an introductory tutorial on Go assembly language at https://golang.org/doc/asm.

Go assembly language provides the DATA command to initialize the package variable. The syntax of the DATA command is as follows:

```
DATA symbol+offset(SB)/width, value
```

Where symbol is the corresponding identifier of the variable in the assembly language, offset is the offset of the symbol start address, width is the width of the memory to be initialized, and value is the value to be initialized. The symbol symbol defined by the Go language in the current package corresponds to `·symbol` in the assembly code, and the dot symbol in the "·" is a special unicode symbol.

We can use the following command to initialize the Id variable to hexadecimal 0x2537, corresponding to decimal 9527 (the constant needs to be represented by the dollar sign $):

```
DATA ·Id+0(SB)/1, $0x37
DATA ·Id+1(SB)/1, $0x25
```

Once the variables are defined, they need to be exported for reference by other code. Go assembly language provides GLOBL commands for exporting symbols:

```
GLOBL symbol(SB), width
```

Where symbol corresponds to the name of the symbol in the assembly, and width is the size of the symbol corresponding to the memory. Use the following command to export the ·Id variable in the assembly:

```
GLOBL ·Id, $8
```

The work of defining an integer variable with assembly has now been initially completed.

In order to make it easier for other packages to use the Id variable, we also need to declare the variable in the Go code and also assign a suitable type to the variable. Modify the contents of pkg.go as follows:

```go
Package pkg

Var Id int
```

The current state of the Go language code is no longer to define a variable, the semantics become a declaration of a variable (can not be initialized when a variable is declared). The definition of the Id variable has been done in assembly language.

We put the complete assembly code in the pkg_amd64.s file:

```
GLOBL ·Id(SB), $8

DATA ·Id+0(SB)/1, $0x37
DATA ·Id+1(SB)/1, $0x25
DATA ·Id+2(SB)/1, $0x00
DATA ·Id+3(SB)/1, $0x00
DATA ·Id+4(SB)/1, $0x00
DATA ·Id+5(SB)/1, $0x00
DATA ·Id+6(SB)/1, $0x00
DATA ·Id+7(SB)/1, $0x00
```

The suffix name of the file name pkg_amd64.s indicates the assembly code file in the AMD64 environment.

Although the pkg package is implemented in assembly, the usage is exactly the same as the previous Go language version:

```go
Package main

Import pkg "path to pkg package"

Func main() {
Println(pkg.Id)
}
```

For Go package users, there is no difference in using Go assembly language or Go language implementation.

## 3.1.3 Defining string variables

In the previous example, we defined an integer variable by assembly. Now let's make it a little harder and try to define a string variable by assembly. Although the definition of a string and an integer variable are basically the same from the perspective of the Go language, the underlying string has a more complex data structure than a single integer.

The experimental flow is the same as the previous example, but the Go function is used to implement similar functions, then the generated assembly code is observed and analyzed, and finally the Go assembly language is used for imitation. First create the pkg.go file and define the string in Go:

```go
Package pkg

Var Name = "gopher"
```

Then use the following command to view the pseudo assembly code corresponding to the Go language program:

```
$ go tool compile -S pkg.go
Go.string."gopher" SRODATA dupok size=6
  0x0000 67 6f 70 68 65 72 gopher
"".Name SDATA size=16
  0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 ................
  Rel 0+8 t=1 go.string."gopher"+0
```

A new symbol go.string."gopher" appears in the output. According to its length and content analysis, it can be guessed that it corresponds to the underlying "gopher" string data. Because the Go language string is not a value type, the Go string is actually a read-only reference type. If the same "gopher" read-only string appears in more than one code, the same symbol go.string."gopher" can be referenced after the program link. Therefore, the symbol has a SRODATA flag indicating that the data is in a read-only memory segment, and dupok indicates that only one of the same identifier data is present.

The size of the real Go string variable Name is only 16 bytes. In fact, the Name variable does not directly correspond to the "gopher" string, but corresponds to the 16-byte size of the reflect.StringHeader structure:

```go
Type reflect.StringHeader struct {
Data uintptr
Len int
}
```

From a compilation point of view, the Name variable actually corresponds to the reflect.StringHeader structure type. The first 8 bytes correspond to the pointer of the underlying real string data, which is the address corresponding to the symbol go.string."gopher". The last 8 bytes correspond to the effective length of the underlying real string data, here is 6 bytes.

Now create the pkg_amd64.s file and try to redefine and initialize the Name string with assembly code:

```
GLOBL · NameData(SB), $8
DATA · NameData(SB)/8, $"gopher"

GLOBL · Name(SB), $16
DATA · Name+0(SB)/8, $·NameData(SB)
DATA · Name+8(SB)/8, $6
```

Because in Go assembly language, go.string."gopher" is not a legal symbol, so we can't create it by hand (this is part of the privilege reserved for the compiler, because manually creating a similar symbol may break the compiler's output code. Some rules). So we have created a new NameData symbol to represent the underlying string data. Then define the Name symbol memory size is 16 bytes, of which the first 8 bytes are initialized with the address corresponding to the NameData symbol, and the last 8 bytes are constant 6 for the string length.

After defining the string variable in assembly and exporting it, you also need to declare the string variable in the Go language. Then you can test the Name variable with the Go language code:

```go
Package main

Import pkg "path/to/pkg"

Func main() {
Println(pkg.Name)
}
```

Unfortunately this run generated the following error:

```
pkgpath.NameData: missing Go type information for global symbol: size 8
```

Error message The NameData symbol defined in the assembly has no type information. In fact, the data defined in the Go assembly language does not have a so-called type, each symbol is just a corresponding piece of memory, so the NameData symbol is also untyped. But the Go language is the language that comes with the garbage collector, and the Go assembly language works within the framework of the automatic garbage collection system. When the Go garbage collector scans the NameData variable, it cannot know whether the variable contains a pointer internally, so this error occurs. The root cause of the error is not that NameData has no type, but that the NameData variable does not have an annotation that will contain pointer information.

This error can be fixed by adding a NOPTR flag to the NameData variable indicating that it will not contain pointer data:

```
#include "textflag.h"

GLOBL · NameData(SB), NOPTR, $8
```

By adding the NOPTR flag to NameData, it means that there is no pointer data. We can also modify this error by adding a pointer to the NameData variable in the Go language with no pointer and a size of 8 bytes:

```go
Package pkg

Var NameData [8]byte
Var Name string
```

We declare NameData as a byte array of length 8. The compiler can parse out through the type that the variable does not contain a pointer, so the NOPTR flag can be omitted from the assembly code. The garbage collector now stops scanning internal data when it encounters this variable.

In this implementation, the underlying Name string actually refers to the "gopher" string data corresponding to the NameData memory. Therefore, if the NameData changes, the data of the Name string will also change.

```go
Func main() {
Println(pkg.Name)

pkg.NameData[0] = '?'
Println(pkg.Name)
}
```

Of course this conflicts with the read-only definition of strings, and normal code needs to avoid this. The best way is to not export the internal NameData variable, which will prevent internal data from being inadvertently destroyed.

When defining strings with assembly, we can think differently: define the underlying string data and the string header structure together to avoid introducing the NameData symbol:

```
GLOBL · Name(SB), $24

DATA · Name+0(SB)/8, $·Name+16(SB)
DATA · Name+8(SB)/8, $6
DATA · Name+16(SB)/8, $"gopher"
```

In the new structure, the memory corresponding to the Name symbol is changed from 16 bytes to 24 bytes, and the extra 8 bytes store the underlying "gopher" string. The first 16 bytes of the Name symbol still correspond to the reflect.StringHeader structure: the Data part corresponds to `$·Name+16(SB)`, indicating that the address of the data is the position where the Name symbol is shifted by 16 bytes backward; the Len part Still corresponds to the length of 6 bytes. This is a technique often used by C programmers.


## 3.1.4 Defining the main function

The previous examples have shown how to define integer and string type variables by assembly. We will now try to implement the function in assembly and then output a string.

First create the main.go file, create and initialize the string variable, and declare maIn function:

```go
Package main

Var helloworld = "Hello, the world"

Func main()
```

Then create the main_amd64.s file, which corresponds to the implementation of the main function:

```
TEXT · main(SB), $16-0
MOVQ · helloworld+0(SB), AX; MOVQ AX, 0(SP)
MOVQ · helloworld+8(SB), BX; MOVQ BX, 8(SP)
CALL runtime·printstring(SB)
CALL runtime·printnl(SB)
RET
```

`TEXT · main(SB), $16-0` is used to define the `main` function, where `$16-0` indicates that the frame size of the `main` function is 16 bytes (corresponding to the size of the string header structure, To pass arguments to the `runtime·printstring` function, `0` indicates that the `main` function has no arguments and return values. The `main` function internally prints a string by calling the runtime's internal `runtime·printstring(SB)` function. Then call `runtime·printnl` to print the newline symbol.

The Go language function passes the call parameters and return values ​​completely through the stack when the function is called. First, the 16 bytes of the string header structure corresponding to helloworld are copied to the 16-byte space corresponding to the stack pointer SP by the MOVQ instruction, and then the corresponding function is called by the CALL instruction. Finally, the RET instruction is used to indicate that the current function returns.


## 3.1.5 Special characters

After the Go language function or method symbol is compiled into the target file, each symbol in the object file contains the absolute import path of the corresponding package. Therefore, the symbol of the target file can be very complicated, such as "path/to/pkg.(*SomeType).SomeMethod" or "go.string."abc"". The symbolic name of the target file contains not only ordinary letters, but also many special characters such as dot, asterisk, braces, and double quotation marks. The Go language assembler is a two-knife transplanted from Plan9, and can't handle these special characters, which leads to various limitations when manually implementing Go features in Go assembly language.

Go assembly language also follows the philosophy that Go is less and more, it only retains the most basic features: defining variables and global functions. Among them, special delimiters are introduced in names such as variables and global functions to support the package system such as Go language. In order to simplify the implementation of the lexical scanner of the Go assembler, the midpoint ``` in Unicode and the division `/` in uppercase are specifically introduced, and the corresponding Unicode code points are `U+00B7` and `U+2215`. After the assembler is compiled, the midpoint ``` will be replaced with the dot "." in ASCII, and the uppercase division will be replaced by the division "/" in the ASCII code. For example, `math/rand·Int` will be replaced with `math/rand.Int`. This simplifies the implementation of the lexical analysis part of the assembler by separating the decimal point in the midpoint and floating point numbers, the division in uppercase, and the division symbol in the expression.

Even if the problem of Go assembly language design trade-off is temporarily put aside, how to input the midpoint `·` and division `/` characters in different input methods of different operating systems is a challenge. These two characters are described in the https://golang.org/doc/asm documentation, so copying directly from this page is the easiest and most reliable way.

If it is a macOS system, there are several ways to input the midpoint `·`: when you do not open the input method, you can directly use option+shift+9; if it is the built-in simplified pinyin input method, enter the upper left corner `~ `Key corresponds to `·`, if it is the built-in Unicode input method, you can enter the corresponding Unicode code point. Among them, the Unicode input method may be the safest and most reliable input method.


## 3.1.6 No semicolon

A semicolon in Go assembly language can be used to separate multiple statements within the same line. Here is the assembly code for chaotic typography with semicolons:

```
TEXT · main(SB), $16-0; MOVQ · helloworld+0(SB), AX; MOVQ · helloworld+8(SB), BX;
MOVQ AX, 0(SP); MOVQ BX, 8(SP); CALL runtime·printstring(SB);
CALL runtime·printnl(SB);
RET;
```

As with the Go language, you can also omit the semicolon at the end of the line. When the end is encountered, the assembler automatically inserts a semicolon. Here is the code after omitting the semicolon:

```
TEXT · main(SB), $16-0
MOVQ · helloworld+0(SB), AX; MOVQ AX, 0(SP)
MOVQ · helloworld+8(SB), BX; MOVQ BX, 8(SP)
CALL runtime·printstring(SB)
CALL runtime·printnl(SB)
RET
```

As with the Go language, multiple consecutive whitespace characters and one space between statements are equivalent.