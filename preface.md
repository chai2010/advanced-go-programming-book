# Go Language Advanced Programming (Advanced Go Programming)

This book covers high-level topics such as CGO, Go assembly language, RPC implementation, Web framework implementation, distributed system, etc., and developers who have some experience in Go language and want to know more about the various advanced usages of Go language. For those who have just learned the Go language, it is recommended to start the basics of the Go language system from [Go Language Bible] (https://github.com/golang-china/gopl-zh). If you want to know the latest trends in Go2, you can refer to [Go2 Programming Guide] (https://github.com/chai2010/go2-book).

![](cover-20190714.jpg)

- Author: fir tree wood, Github [@ chai2010] (https://github.com/chai2010), Twitter [@chaishushan] (https://twitter.com/chaishushan)
- Author: Cao Chunhui, Github [@ cch123] (https://github.com/cch123)
- Website: https://github.com/chai2010/advanced-go-programming-book

Purchase link:

- Jingdong: https://item.m.jd.com/product/12647494.html
- Asynchronous: https://www.epubit.com/book/detail/40090

If you like this book, welcome to Douban Comments:

[![](douban.png)](https://book.douban.com/subject/34442131/)

- https://book.douban.com/subject/34442131/


## read online

- https://chai2010.cn/advanced-go-programming-book/
- https://www.gitbook.com/book/chai2010/advanced-go-programming-book/


## Follow WeChat public number (guanggu-coder)

![](weixin-guanggu-coder-logo.png)


## Netease Cloud Classroom·光谷码农课

- https://study.163.com/provider/480000001914454/index.html

![](163study-go-master.jpg)

## Copyright Notice

<a rel="license" href="http://creativecommons.org/licenses/by-nc-nd/4.0/"><img alt="Creative Commons License Agreement" style="border-width:0" src ="https://i.creativecommons.org/l/by-nc-nd/4.0/88x31.png" /></a>
<br />
<span xmlns:dct="http://purl.org /dc/terms/" property="dct:title">Go Language Advanced Programming</span> by <a xmlns:cc="http://creativecommons.org/ns#" href="https://github. Com/chai2010/advanced-go-programming-book" property="cc:attributionName" rel="cc:attributionURL">Chai Shushan, Cao Chunhui</a> using <a rel="license" href="http:// Creativecommons.org/licenses/by-nc-nd/4.0/">Knowledge Sharing - Non-commercial Use - Prohibition of the interpretation of the 4.0 International License Agreement</a>.

Any commercial use or reference to all or part of this document is strictly prohibited!

Welcome everyone to provide advice!

-------

# Preface

In November 2009, Google released the Go language, which caused a sensation in the world. In 2015 and 2016, the Go language conferences in China were held in Shanghai and Beijing respectively. The developers from the Go language team made relevant reports. Throughout the past few years, Go has become the most important basic programming language in the era of cloud computing and cloud storage.

China's Go language community is the world's largest Go language community. We have not only followed the development of Go language from the beginning, but also made great contributions to the development of Go language. Wei Guangjing (vcc.163@gmail.com) from Shenzhen, China, around 2010, on the work of MinGW laid the support of the Go language for the Windows platform, and also laid the CGO support for the Windows platform. Minux (minux.ma@gmail.com), also from China, is a member of the Go core team and he has been involved in a number of Go core design and development reviews. At the same time, a large number of domestic Go language enthusiasts actively participated in the reporting and repair of BUG (the author is also one of them).

As of 2018, there are nearly 15 related Go language tutorials published in China. The content covers Go language basic programming, Web programming, concurrent programming and internal source code analysis. But as a veteran user of the Go language, the Go language topic that the author is concerned with is far more than that. The CGO feature implements Go language support for C language and C++ language, enabling Go language to seamlessly inherit the huge software assets accumulated in the C/C++ world for decades. Go assembly language provides a direct access to the underlying machine instructions, allowing us to infinitely squeeze the performance of hot code in the program. At present, the emerging projects of domestic Internet companies have gradually shifted to the Go language ecology, and the actual experience of the development of large-scale distributed systems is also of concern to everyone. These high-level or cutting-edge features are topics of interest to the author and this book.

This book is for developers who have a certain degree of Go language experience and want to know more about the various advanced usages of the Go language. For newcomers to Go, it is recommended to read D&K's [The Go Programming Language] (https://gopl.io/) before reading this book. Finally, I hope this book will help you get a deeper understanding of the Go language.

Chai2010 - August 2018 in Wuhan

# Thank you

First of all, thanks to the father of the Go language and every friend who has submitted a patch for the Go language. Thanks to fango's first online novel "Hu Wen Go.ogle" with Go language as the theme and the first Chinese Go language book "Go Language·Cloud Power". It is your sharing that brings everyone to learn Go language. enthusiasm. Thanks to Wei Guangjing for his pioneering work on the Windows platform CGO, otherwise the book may not have a dedicated CGO chapter. Thanks to the friends who submitted the issue or PR for this book (especially fuwensun, lewgun, etc.), your attention and support is the greatest motivation for the author's writing.

thank you all!