# 5.6 Ratelimit Service Flow Limit

The computer program can be divided into the disk IO bottleneck type according to its bottleneck, the CPU calculates the bottleneck type, the network bandwidth bottleneck type, and sometimes the external system causes the bottleneck itself in the distributed scenario.

The most important part of the Web system is the network. Whether it is receiving, parsing user requests, accessing storage, or returning response data to users, it is necessary to go online. Before the IO multiplexing interface provided by the system such as `epoll/kqueue`, the core headache of modern computers with multiple cores is the C10k problem. The C10k problem can cause the computer to not fully utilize the CPU to handle more users. Connect, and there is no way to handle more requests by optimizing the program to increase CPU utilization.

Since Linux implemented `epoll` and FreeBSD implemented `kqueue`, this problem has been basically solved. We can easily solve the C10k problem of the year with the API provided by the kernel, which means that if your program is mainly dealing with the network, then The bottleneck must be in the user program and not in the operating system kernel.

With the development of the times, programming languages ​​have further encapsulated these system calls. Nowadays, application layer development is almost impossible to see words like `epoll` in the program. Most of the time we just focus on business logic. Just fine. Go's net library encapsulates different syscall APIs for different platforms. The `http` library is built on top of the `net` library, so in the Go language we can easily write high-performance `http with standard libraries. `Services, below is a simple `hello world` service code:

```go
Package main

Import (
"io"
"log"
"net/http"
)

Func sayhello(wr http.ResponseWriter, r *http.Request) {
wr.WriteHeader(200)
io.WriteString(wr, "hello world")
}

Func main() {
http.HandleFunc("/", sayhello)
Err := http.ListenAndServe(":9090", nil)
If err != nil {
log.Fatal("ListenAndServe:", err)
}
}
```

We need to measure the throughput of this Web service, and then specifically, the QPS of the interface. With wrk, benchmark the `hello world` service on your home Macbook Pro. The hardware of the Mac is as follows:

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

Test Results:

```shell
~ ❯❯❯ wrk -c 10 -d 10s -t10 http://localhost:9090
Running 10s test @ http://localhost:9090
  10 threads and 10 connections
  Thread Stats Avg Stdev Max +/- Stdev
Latency 339.99us 1.28ms 44.43ms 98.29%
Req/Sec 4.49k 656.81 7.47k 73.36%
  449588 requests in 10.10s, 54.88MB read
Requests/sec: 44513.22
Transfer/sec: 5.43MB

~ ❯❯❯ wrk -c 10 -d 10s -t10 http://localhost:9090
Running 10s test @ http://localhost:9090
  10 threads and 10 connections
  Thread Stats Avg Stdev Max +/- Stdev
Latency 334.76us 1.21ms 45.47ms 98.27%
Req/Sec 4.42k 633.62 6.90k 71.16%
  443582 requests in 10.10s, 54.15MB read
Requests/sec: 43911.68
Transfer/sec: 5.36MB

~ ❯❯❯ wrk -c 10 -d 10s -t10 http://localhost:9090
Running 10s test @ http://localhost:9090
  10 threads and 10 connections
  Thread Stats Avg Stdev Max +/- Stdev
Latency 379.26us 1.34ms 44.28ms 97.62%
Req/Sec 4.55k 591.64 8.20k 76.37%
  455710 requests in 10.10s, 55.63MB read
Requests/sec: 45118.57
Transfer/sec: 5.51MB
```

The result of multiple tests is about 40,000 QPS floating, and the response time is about 40ms. For a web application, this is already a very good result. We just copied the sample code of others and completed one. High-performance `hello world` server, is it a great sense of accomplishment?

This is still only a home PC, online servers are mostly 24 cores, 32G memory +, CPU is basically Intel i7. So the same program running on the server will get better results.

The `hello world` service here does not have any business logic. The procedures in the real environment are much more complicated. Some programs are biased towards network IO bottlenecks, such as some CDN services and Proxy services. Some programs are biased towards CPU/GPU bottlenecks, such as login verification services and image processing services. Some program bottlenecks are partial to disk, such as specialization. Storage system, database. Different program bottlenecks are reflected in different places, and the single-function services mentioned here are relatively easy to analyze. If you encounter a module with a large amount of business logic and a large amount of code, the bottleneck is not inferred from the three, five, and two, or you need to get more accurate conclusions from the stress test.

For the IO/Network bottleneck class, the performance is that the NIC/disk IO will be full before the CPU. In this case, even if the CPU is optimized, the throughput of the entire system cannot be improved, and the read and write speed of the disk can be increased. The memory size increases the bandwidth of the NIC to improve overall performance. The CPU bottleneck program is that the CPU usage first reaches 100% before the storage and network cards are not full. The CPU is busy with various computing tasks, and the IO devices are relatively idle.

Regardless of the type of service, when the resource is used to the limit, the request will accumulate, timeout, system hang, and eventually harm to the end user. For distributed Web services, the bottleneck is not always inside the system, and it may be external. Non-computation-intensive systems tend to fall into the relational database, and at this time the Web module itself is far from reaching the bottleneck.

No matter where our service bottlenecks are, the final thing to do is the same, that is, traffic restrictions.

## 5.6.1 Common Traffic Limiting Means

There are many ways to limit traffic. The most common ones are: leaky buckets and token buckets:

1. A leaky bucket means that we have a bucket that has been filled with water and leaks a drop of water every time it has been fixed for a fixed period of time. If you receive this drip, then you can continue the service request, if not received, then you need to wait for the next drop.
2. The token bucket refers to adding a token to the bucket at a constant speed. When the service request is received, the token needs to be obtained from the bucket. The number of tokens can be adjusted according to the resources that need to be consumed. If you don't have a token, you can choose to wait or give up.

These two methods look a lot like, but there are still differences. The rate at which the leaking bucket exits is fixed, and the token bucket can be taken as long as there is a token in the bucket. That is to say, the token bucket allows a certain degree of concurrency. For example, at the same time, there are 100 user requests. As long as there are 100 tokens in the token bucket, all 100 requests will be put forward. The token bucket also degenerates into a leaky bucket model if there is no token in the bucket.

![token bucket](../images/ch5-token-bucket.png)

*Figure 5-12 Token Bucket*

In actual applications, token buckets are widely used, and most of the popular current limiters in the open source world are based on token buckets. And on this basis, a certain degree of expansion, such as `github.com/juju/ratelimit` provides several different ways to fill the token bucket:

```go
Func NewBucket(fillInterval time.Duration, capacity int64) *Bucket
```

The default token bucket, `fillInterval`, means that each time a token is placed in the bucket, `capacity` is the capacity of the bucket, and the portion that exceeds the bucket capacity is discarded directly. The bucket is initially full.

```go
Func NewBucketWithQuantum(fillInterval time.Duration, capacity, quantum int64) *Bucket
```

The difference from the normal `NewBucket()` is that each time a token is placed in the bucket, a `quantum` token is placed instead of a token.

```go
Func NewBucketWithRate(rate float64, capacity int64) *Bucket
```

This is a bit special, and the number of tokens per second is filled according to the ratio provided. For example, `capacity` is 100, and `rate` is 0.1, then 10 tokens are filled every second.

Getting a token from the bucket also provides several APIs:

```go
Func (tb *Bucket) Take(count int64) time.Duration {}
Func (tb *Bucket) TakeAvailable(count int64) int64 {}
Func (tb *Bucket) TakeMaxDuration(count int64, maxWait time.Duration) (
time.Duration, bool,
) {}
Func (tb *Bucket) Wait(count int64) {}
Func (tb *Bucket) WaitMaxDuration(count int64, maxWait time.Duration) bool {}
```

The names and functions are relatively straightforward, so I won't go into details here. Compared to the ratelimiter provided by Google's Java tool library Guava, which is more famous in the open source world, this library does not support token bucket warm-up and cannot modify the initial token capacity, so the requirements in individual extreme cases may not be met. However, after understanding the basic principles of the token bucket, if there is no way to meet the demand, I believe that you can also modify it and support your own business scenario.

## 5.6.2 Principle

From the functional point of view, the token bucket model is the addition and subtraction operation process for the global count, but the use of the count requires us to add the read-write lock, which has a small burden of thought. If we are already familiar with the Go language, it is easy to think of a buffered channel to complete a simple token-token operation:

```go
Var tokenBucket = make(chan struct{}, capacity)
```

Add `token` to `tokenBucket` every once in a while, and if `bucket` is full, give up:

```go
fillToken := func() {
Ticker := time.NewTicker(fillInterval)
	for {
		select {
		case <-ticker.C:
			select {
			case tokenBucket <- struct{}{}:
			default:
			}
			fmt.Println("current token cnt:", len(tokenBucket), time.Now())
		}
	}
}
```

Combine the code:

```go
package main

import (
	"fmt"
	"time"
)

func main() {
	var fillInterval = time.Millisecond * 10
	var capacity = 100
	var tokenBucket = make(chan struct{}, capacity)

	fillToken := func() {
		ticker := time.NewTicker(fillInterval)
		for {
			select {
			case <-ticker.C:
				select {
				case tokenBucket <- struct{}{}:
				default:
				}
				fmt.Println("current token cnt:", len(tokenBucket), time.Now())
			}
		}
	}

	go fillToken()
	time.Sleep(time.Hour)
}

```

Look at the results of the run:

```shell
current token cnt: 98 2018-06-16 18:17:50.234556981 +0800 CST m=+0.981524018
current token cnt: 99 2018-06-16 18:17:50.243575354 +0800 CST m=+0.990542391
current token cnt: 100 2018-06-16 18:17:50.254628067 +0800 CST m=+1.001595104
current token cnt: 100 2018-06-16 18:17:50.264537143 +0800 CST m=+1.011504180
current token cnt: 100 2018-06-16 18:17:50.273613018 +0800 CST m=+1.020580055
current token cnt: 100 2018-06-16 18:17:50.2844406 +0800 CST m=+1.031407637
current token cnt: 100 2018-06-16 18:17:50.294528695 +0800 CST m=+1.041495732
current token cnt: 100 2018-06-16 18:17:50.304550145 +0800 CST m=+1.051517182
current token cnt: 100 2018-06-16 18:17:50.313970334 +0800 CST m=+1.060937371
```

At the time of 1s, it just filled 100, without much deviation. However, it can be seen here that Go's timer has an error of about 0.001 s, so if the token bucket size is more than 1000, there may be some error. For general services, this error does not matter.

The token token operation of the token bucket above is also simpler to implement, simplifying the problem, we only take one token here:

```go
func TakeAvailable(block bool) bool{
	var takenResult bool
	if block {
		select {
		case <-tokenBucket:
			takenResult = true
		}
	} else {
		select {
		case <-tokenBucket:
			takenResult = true
		default:
			takenResult = false
		}
	}

	return takenResult
}
```
Some companies have built their own current-limiting wheels in this way, but if the open source version of the ratelimit is the case, then we have nothing to say. The reality is not like this.

Let's think about it, the token bucket puts the token into the bucket every fixed time. If we record the last time the token was put, it is t1, and the token number k1 at that time, the token interval is Ti, each time you put x tokens into the token bucket, the token bucket capacity is cap. Now if someone calls `TakeAvailable` to get n tokens, we will record this moment as t2. At t2, how many tokens should there be in the token bucket? The pseudo code is as follows:

```go
cur = k1 + ((t2 - t1)/ti) * x
cur = cur > cap ? cap : cur
```
We use the time difference between two time points, combined with other parameters, theoretically, we can know how many tokens are in the bucket before taking the token. It is theoretically unnecessary to use the operation of filling the token into the channel in front of this section. As long as you make a simple calculation of the number of tokens in the token bucket each time you call `Take`, you can get the correct number of tokens. Is it a lot like the feeling of "inert evaluation"?

After getting the correct number of tokens, then the actual `Take` operation is fine. This `Take` operation only needs to perform a simple subtraction on the number of tokens. Remember to lock to ensure concurrent security. `github.com/juju/ratelimit` This library does just that.

## 5.6.3 Service bottlenecks and QoS

Earlier we talked about a lot of CPU bottlenecks, IO bottlenecks and other concepts, this performance bottleneck can be located relatively quickly from most companies have monitoring systems, if a system encounters performance problems, then the monitoring chart response It is the fastest.

While performance metrics are important, the overall QoS of the service should also be considered when providing services to users. The full name of QoS is Quality of Service, as the name suggests is the quality of service. QoS includes metrics such as availability, throughput, latency, delay variation, and loss. In general, we can improve the CPU utilization of Web services by optimizing the system, thereby increasing the throughput of the entire system. But as throughput improves, the user experience is likely to get worse. User-sensitive is more sensitive than latency. Although your system throughput is high, but you can't open the page for a long time, it will cause a lot of user loss. Therefore, in the large company's Web service performance indicators, in addition to the average response delay, the response time of 95 cents, 99 points is also taken out as a performance standard. The average response does not have much impact on improving CPU utilization. The response time of 95-bit and 99-digit may increase dramatically. Then it is worth considering whether the cost of improving these CPU utilization is worthwhile.

Machines in online systems generally keep the CPU at a certain margin.