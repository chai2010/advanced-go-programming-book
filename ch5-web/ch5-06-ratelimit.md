# 5.6. Ratelimit 服务流量限制

计算机程序可依据其瓶颈分为 IO-bound，或 CPU-bound，我们这里先刨除掉存储类系统。web 系统打交道最多的实际上就是网络，从 linux 引入了 epoll 的 API 之后，我们可以借助其轻松解决当年的 C10k 问题，实现一个简单的 echo 服务器。随着编程语言的发展，很多编程语言对这些系统调用又进一步进行了封装，所以做应用层开发，压根儿不会在程序中看到 epoll 之类的字眼，大多数时候我们就只要聚焦中业务逻辑上就好，不用管底层是用的 epoll 还是 kqueue。时至今日，C10k 都已经很少被人所提起，我们写一个简单的 `hello world` 程序：

```go
package main

import (
	"io"
	"log"
	"net/http"
)

func sayhello(wr http.ResponseWriter, r *http.Request) {
	wr.WriteHeader(200)
	io.WriteString(wr, "hello world")
}

func main() {
	http.HandleFunc("/", sayhello)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
```

并借助 wrk，在家用电脑 Macbook Pro 上对其进行基准测试，Mac 的硬件情况如下：

```shell
CPU: Intel(R) Core(TM) i5-5257U CPU @ 2.70GHz
Core: 2
Threads: 4

Graphics/Displays:
      Chipset Model: Intel Iris Graphics 6100
          Resolution: 2560 x 1600 Retina
    Memory Slots:
          Size: 4 GB
          Speed: 1867 MHz
          Size: 4 GB
          Speed: 1867 MHz
Storage:
          Size: 250.14 GB (250,140,319,744 bytes)
          Media Name: APPLE SSD SM0256G Media
          Size: 250.14 GB (250,140,319,744 bytes)
          Medium Type: SSD
```

测试结果：

```shell
~ ❯❯❯ wrk -c 10 -d 10s -t10 http://localhost:9090
Running 10s test @ http://localhost:9090
  10 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   339.99us    1.28ms  44.43ms   98.29%
    Req/Sec     4.49k   656.81     7.47k    73.36%
  449588 requests in 10.10s, 54.88MB read
Requests/sec:  44513.22
Transfer/sec:      5.43MB

~ ❯❯❯ wrk -c 10 -d 10s -t10 http://localhost:9090
Running 10s test @ http://localhost:9090
  10 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   334.76us    1.21ms  45.47ms   98.27%
    Req/Sec     4.42k   633.62     6.90k    71.16%
  443582 requests in 10.10s, 54.15MB read
Requests/sec:  43911.68
Transfer/sec:      5.36MB

~ ❯❯❯ wrk -c 10 -d 10s -t10 http://localhost:9090
Running 10s test @ http://localhost:9090
  10 threads and 10 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   379.26us    1.34ms  44.28ms   97.62%
    Req/Sec     4.55k   591.64     8.20k    76.37%
  455710 requests in 10.10s, 55.63MB read
Requests/sec:  45118.57
Transfer/sec:      5.51MB
```

多次测试的结果在 4w 左右的 QPS浮动，响应时间最多也就是 40ms 左右，对于一个 web 程序来说，这已经是很不错的成绩了。这还只是家用 PC，线上服务器大多都是 24 核心起，32G 内存+，CPU 基本都是 Intel I7。所以同样的程序在服务器上运行会得到更好的结果。

真实环境的程序要比我们这里的 `hello world` 复杂得多，有些程序偏 IO bound，例如一些 proxy 服务、存储服务、缓存服务；有些程序偏 CPU/GPU bound，例如登陆校验服务、图像处理服务。不同的程序瓶颈会体现在不同的地方，这里提到的这些功能单一的服务相对来说还算容易分析。如果碰到业务逻辑复杂代码量巨大的模块，其瓶颈并不是三下五除二可以推测出来的，还是需要我们拿真实的环境来进行压力测试。
