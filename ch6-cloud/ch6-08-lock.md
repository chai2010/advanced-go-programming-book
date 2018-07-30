# 6.8. 分布式锁

在单机程序并发或并行修改全局变量时，需要对修改行为加锁以创造临界区。为什么需要加锁呢？可以看看下段代码：

```go
package main

import (
    "sync"
)

// 全局变量
var counter int

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter++
        }()
    }

    wg.Wait()
    println(counter)
}
```

多次运行会得到不同的结果：

```shell
❯❯❯ go run local_lock.go                                ✭
945
❯❯❯ go run local_lock.go                                ✭
937
❯❯❯ go run local_lock.go                                ✭
959
```

## 进程内加锁

想要得到正确的结果的话，把对 counter 的操作代码部分加上锁：

```go
// ... 省略之前部分
var wg sync.WaitGroup
var l sync.Mutex
for i := 0; i < 1000; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        l.Lock()
        counter++
        l.Unlock()
    }()
}

wg.Wait()
println(counter)
// ... 省略之后部分
```

这样就可以稳定地得到计算结果了：

```shell
❯❯❯ go run local_lock.go                              ✭ ✱
1000
```

## trylock

```go
package main

import (
	"sync"
)

// Lock try lock
type Lock struct {
	c chan struct{}
}

// NewLock generate a try lock
func NewLock() Lock {
	var l Lock
	l.c = make(chan struct{}, 1)
	l.c <- struct{}{}
	return l
}

// Lock try lock, return lock result
func (l Lock) Lock() bool {
	lockResult := false
	select {
	case <-l.c:
		lockResult = true
	default:
	}
	return lockResult
}

// Unlock , Unlock the try lock
func (l Lock) Unlock() bool {
	unlockResult := false
	select {
	case l.c <- struct{}{}:
		unlockResult = true
	default:
	}
	return unlockResult
}

var counter int

func main() {
	var l = NewLock()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if !l.Lock() {
				// log error
				println("lock failed")
				return
			}
			counter++
			println("current counter", counter)
			l.Unlock()
		}()
	}
	wg.Wait()
}

```

## 基于 redis 的 setnx

```go
package main

import (
    "fmt"
    "sync"
    "time"

    "github.com/go-redis/redis"
)

func incr() {
    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    var lockKey = "counter_lock"
    var counterKey = "counter"

    // lock
    resp := client.SetNX(lockKey, 1, time.Second*5)
    lockSuccess, err := resp.Result()

    if err != nil || !lockSuccess {
        fmt.Println(err, "lock result: ", lockSuccess)
        return
    }

    // counter ++
    getResp := client.Get(counterKey)
    cntValue, err := getResp.Int64()
    if err == nil {
        cntValue++
        resp := client.Set(counterKey, cntValue, 0)
        _, err := resp.Result()
        if err != nil {
            // log err
            println("set value error!")
        }
    }
    println("current counter is ", cntValue)

    delResp := client.Del(lockKey)
    unlockSuccess, err := delResp.Result()
    if err == nil && unlockSuccess > 0 {
        println("unlock success!")
    } else {
        println("unlock failed", err)
    }
}

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            incr()
        }()
    }
    wg.Wait()
}

```

```shell
❯❯❯ go run redis_setnx.go
<nil> lock result:  false
<nil> lock result:  false
<nil> lock result:  false
<nil> lock result:  false
<nil> lock result:  false
<nil> lock result:  false
<nil> lock result:  false
<nil> lock result:  false
<nil> lock result:  false
current counter is  2028
unlock success!
```

## 基于 zk

## 基于 etcd

## redlock

```go
```

## how to choose
