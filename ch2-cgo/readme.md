# Chapter 2 CGO Programming

* Past experience is often the shackles of the future, because the sunk costs invested in gas technology can hinder people from embracing new technologies. ——chai2010*

* It was once painful because I couldn't learn the dazzling new standards of C++; Go's "less is more" avenue to Jane's philosophy made me regain my confidence and find the long-lost programming fun. ——Ending*

After decades of development, C/C++ has accumulated huge software assets, many of which have been tested and their performance has been optimized. The Go language must be able to stand on the shoulders of the giant C/C++. With a huge amount of C/C++ software assets, we can confidently program in Go. C language as a general language, many libraries will choose to provide a C-compatible API, and then implemented in a different programming language. The Go language supports C language function calls through a tool called CGO, and we can use the Go language to export the C dynamic library interface to other languages. This chapter focuses on some of the issues involved in CGO programming.