# Chapter 3 Go Assembly Language

* Can run on the line, no machine. ——rfyiamcool & Sun boss who loves to learn*

* Follow the right person and do the right thing. - Rhichy*

Many of the design ideas and tools in the Go language are inherited from the Plan9 operating system, and the Go assembly language is also based on the evolution of the Plan9 assembly. According to Rob Pike, the assembly pseudo code output by the C-compiler written by Ken Thompson for the Plan9 system in 1986 is the predecessor of Plan9 assembly. The so-called Plan9 assembly language is just a handy way to write the assembly pseudocode output by the C compiler.

No matter how the high-level language develops, the status as the assembly language closest to the CPU is still not completely replaced. Only the assembly language can fully exploit the full functionality of the CPU chip, so the boot process of the operating system must rely on the help of assembly language. Only the assembly language can completely squeeze the performance of the CPU chip, so many performance-sensitive algorithms such as the underlying encryption and decryption will consider performance optimization through assembly language.

For every serious Gopher, Go assembly language is a technology that cannot be ignored. Because even if you only understand a little bit of assembly, it is easier to understand the computer principle better, and it is easier to understand the implementation principle of advanced features such as dynamic stack/interface in Go language. And after mastering the Go assembly language, you will re-stand at the top of the programming language contempt chain, without worrying about being despised by any other so-called high-level programming language users.

In this chapter, we will use AMD64 as the main development environment to briefly explore the basic usage of Go assembly language.