# 2.2 CGO Foundation

To use the CGO feature, you need to install the C/C++ build toolchain. Under macOS and Linux, you need to install GCC. Under Windows, you need to install the MinGW tool. Also ensure that the environment variable `CGO_ENABLED` is set to 1, which indicates that CGO is enabled. `CGO_ENABLED` is enabled by default when built locally, and CGO is disabled by default when cross-built. For example, to cross-build the Go program running in the ARM environment, you need to manually set the C/C++ cross-built toolchain and open the `CGO_ENABLED` environment variable. Then enable the CGO feature via the `import "C"` statement.

## 2.2.1 `import "C"` statement

If the `import "C"` statement appears in the Go code, it means that the CGO feature is used. The comment immediately before the line statement is a special syntax that contains the normal C code. When CGO is enabled, you can also include the source files corresponding to C/C++ in the current directory.

The simplest example:

```Go
Package main

/*
#include <stdio.h>

Void printint(int v) {
Printf("printint: %d\n", v);
}
*/
Import "C"

Func main() {
v := 42
C.printint(C.int(v))
}
```

This example shows the basic use of cgo. The beginning of the note writes the C function to be called and the associated header file. All C language elements in the header file after being included will be added to the virtual package "C". It should be noted that the import "C" import statement requires a separate line and cannot be imported with other packages. Passing parameters to a C function is also very simple, and it can be directly converted into a corresponding C language type. In the above example, `C.int(v)` is used to convert the int type value in a Go to an int type value in C, and then call the C language defined printint function to print.

It should be noted that Go is a strongly typed language, so the parameter type passed in cgo must be exactly the same as the declared type, and must be converted to the corresponding C type using the conversion function in "C" before passing, and cannot be directly passed into Go. A variable of type. At the same time, the C language symbols imported through the virtual C package do not need to be beginning with uppercase letters, and they are not subject to the export rules of the Go language.

Cgo puts the C language symbols referenced by the current package into the virtual C package. At the same time, other Go language packages that the current package depends on may also introduce similar virtual C packages through cgo, but different Go language packages introduce virtual ones. The types between C packages are not universal. This constraint may have a slight impact on the ability to construct some cgo helper functions yourself.

For example, we want to define a CChar type corresponding to a C character pointer in Go, and then add a GoString method to return the Go language string:

```go
Package cgo_helper

//#include <stdio.h>
Import "C"

Type CChar C.char

Func (p *CChar) GoString() string {
Return C.GoString((*C.char)(p))
}

Func PrintCString(cs *C.char) {
C.puts(cs)
}
```

Now we might want to use this helper function in other Go language packages as well:

```go
Package main

//static const char* cs = "hello";
Import "C"
Import "./cgo_helper"

Func main() {
cgo_helper.PrintCString(C.cs)
}
```

This code does not work, because the type of the `C.cs` variable introduced by the current main package is the `*char` type under the virtual C package of the cgo construct of the current `main` package (the specific point is `*). C.char`, more specific is `*main.C.char`), which is different from the `*C.char` type introduced by the cgo_helper package (specifically `*cgo_helper.C.char`). In the Go language, the methods are dependent on the type. The types of virtual C packages introduced in different Go packages are different (`main.C` not equal to `cgo_helper.C`), which leads to the extension from them. The Go type is also a different type (`*main.C.char` does not wait for `*cgo_helper.C.char`), which eventually causes the previous code to not work properly.

Users with experience with the Go language may suggest that parameters be passed in after the transformation. But this method seems to be infeasible, because the parameter of `cgo_helper.PrintCString` is the `*C.char` type introduced by its own package, which cannot be directly obtained from the outside. In other words, if a package directly uses a type of similar virtual C package such as `*C.char` in the public interface, other Go packages cannot directly use these types unless the Go package also provides `* The constructor of the C.char` type. Because of these many factors, if you want to test these cgo export types directly in the go test environment, there will be the same restrictions.

<!-- Test code; need to be sure if there is a problem -->

## 2.2.2 `#cgo` statement

In the comments before the `import "C"` statement, the relevant parameters of the compile phase and the link phase can be set by the `#cgo` statement. The parameters of the compile phase are mainly used to define related macros and the specified header file retrieval path. The parameters of the link phase are mainly to specify the library file retrieval path and the library file to be linked.

```go
// #cgo CFLAGS: -DPNG_DEBUG=1 -I./include
// #cgo LDFLAGS: -L/usr/local/lib -lpng
// #include <png.h>
Import "C"
```

In the above code, the CFLAGS part, the `-D` part defines the macro PNG_DEBUG, the value is 1; `-I` defines the search directory contained in the header file. In the LDFLAGS section, `-L` specifies the library file retrieval directory when linking, and `-l` specifies the link to the png library.


Because of the problems left by C/C++, the C header file retrieval directory can be a relative directory, but the library file retrieval directory requires an absolute path. The absolute path of the current package directory can be represented by the `${SRCDIR}` variable in the search directory of the library file:

```
// #cgo LDFLAGS: -L${SRCDIR}/libs -lfoo
```

The above code will be expanded when linked:

```
// #cgo LDFLAGS: -L/go/src/foo/libs -lfoo
```

The `#cgo` statement mainly affects several compiler environment variables such as CFLAGS, CPPFLAGS, CXXFLAGS, FFLAGS, and LDFLAGS. LDFLAGS is used to set the parameters of the link, in addition to several variables used to change the build parameters of the compilation phase (CFLAGS is used to set the compilation parameters for C language code).

For users who use C and C++ in a cgo environment, there may be three different compile options: CFLAGS for C language-specific compile options, CXXFLAGS for C++-specific compile options, and CPPFLAGS for C and C++ compiles. Option. However, in the link phase, the C and C++ link options are generic, so there is no longer a difference between C and C++ at this time, and their target files are of the same type.

The `#cgo` directive also supports conditional selection, and subsequent compile or link options take effect when an operating system or a CPU architecture type is met. For example, the following are the compile and link options for windows and non-windows platforms:

```
// #cgo windows CFLAGS: -DX86=1
// #cgo !windows LDFLAGS: -lm
```

In the windows platform, the X86 macro is predefined to 1 before compiling; under the non-widnows platform, the math library is required to be linked in the link phase. This usage is useful for scenarios where there are only a few differences in compilation options on different platforms.

If cgo corresponds to different c code under different systems, we can use the `#cgo` directive to define different C language macros, and then use macros to distinguish different codes:

```go
Package main

/*
#cgo windows CFLAGS: -DCGO_OS_WINDOWS=1
#cgo darwin CFLAGS: -DCGO_OS_DARWIN=1
#cgo linux CFLAGS: -DCGO_OS_LINUX=1

#if defined(CGO_OS_WINDOWS)
Const char* os = "windows";
#elifdefined(CGO_OS_DARWIN)
Const char* os = "darwin";
#elifdefined(CGO_OS_LINUX)
Const char* os = "linux";
#else
# error(unknown os)
#endif
*/
Import "C"

Func main() {
Print(C.GoString(C.os))
}
```

This way we can use the techniques commonly used in C to handle the difference code between different platforms.

## 2.2.3 build tag conditional compilation

The build tag is a special comment at the beginning of a C/C++ file in a Go or cgo environment. Conditional compilation is similar to macros defined for different platforms by the `#cgo` directive. The corresponding code is built only after the macro of the corresponding platform is defined. However, there is a limit to defining macros via the `#cgo` directive. It can only be based on operating systems supported by Go, such as windows, darwin, and linux. If we want to define a macro for the DEBUG flag, the `#cgo` command will be powerless. The build tag conditional compilation feature provided by the Go language can be easily implemented.

For example, the following source files will only be built when the debug build flag is set:

```go
// +build debug

Package main

Var buildMode = "debug"
```

Can be built with the following command:

```
Go build -tags="debug"
Go build -tags="windows debug"
```

We can specify multiple build flags at the same time via the `-tags` command line argument, separated by spaces.

When there are multiple build tags, we combine multiple flags through the rules of logical operation. For example, the following build flags indicate that the build is only done under "linux/386" or "non-cgo environment" under the darwin platform.

```go
// +build linux,386 darwin,!cgo
```

Among them, `linux, 386`, linux and 386 use comma link to mean AND; and `linux, 386` and `darwin, !cgo` mean the meaning of OR by blank segmentation.