# 6.2 Distributed lock

When a stand-alone program modifies global variables concurrently or in parallel, the modification behavior needs to be locked to create a critical section. Why do you need to lock? Let's see what happens when there is a concurrent count without locking:

```go
Package main

Import (
"sync"
)

// global variable
Var counter int

Func main() {
Var wg sync.WaitGroup
For i := 0; i < 1000; i++ {
wg.Add(1)
Go func() {
Defer wg.Done()
Counter++
}()
}

wg.Wait()
Println(counter)
}
```

Running multiple times will result in different results:

```shell
❯❯❯ go run local_lock.go
945
❯❯❯ go run local_lock.go
937
❯❯❯ go run local_lock.go
959
```

## 6.2.1 In-process locking

To get the correct result, lock the operation code portion of the counter:

```go
// ... omit the previous part
Var wg sync.WaitGroup
Var l sync.Mutex
For i := 0; i < 1000; i++ {
wg.Add(1)
Go func() {
Defer wg.Done()
l.Lock()
Counter++
l.Unlock()
}()
}

wg.Wait()
Println(counter)
// ... after omitting the part
```

This will result in a stable calculation:

```shell
❯❯❯ go run local_lock.go
1000
```

## 6.2.2 trylock

In some scenarios, we just want a single executor for a task. Unlike the counter scene, all goroutines execute successfully. Later, after the goroutine failed to steal, it needs to abandon its process. This time you need a trylock.

Trylock, as the name implies, attempts to lock, and the lock succeeds in executing the subsequent process. If the lock fails, it will not block, but will directly return the result of the lock. In the Go language we can simulate a trylock with a channel of size 1:

```go
Package main

Import (
"sync"
)

// Lock try lock
Type lock struct {
c chan struct{}
}

// NewLock generate a try lock
Func NewLock() Lock {
Var l Lock
L.c = make(chan struct{}, 1)
L.c <- struct{}{}
Return l
}

// Lock try lock, return lock result
Func (l Lock) Lock() bool {
lockResult := false
Select {
Case <-l.c:
lockResult = true
Default:
}
Return lockResult
}

// Unlock , Unlock the try lock
Func (l Lock) Unlock() {
L.c <- struct{}{}
}

Var counter int

Func main() {
Var l = NewLock()
Var wg sync.WaitGroup
For i := 0; i < 10; i++ {
wg.Add(1)
Go func() {
Defer wg.Done()
If !l.Lock() {
// log error
Println("lock failed")
Return
}
Counter++
Println("current counter", counter)
l.Unlock()
}()
}
wg.Wait()
}
```

Because our logic limits each goroutine to execute the following logic only if it successfully executes `Lock`, so in `Unlock` it can be guaranteed that the channel in the Lock structure must be empty, so it will not block or fail. The above code uses a channel of size 1 to simulate a trylock. In theory, you can use the CAS in the standard library to achieve the same functionality and at a lower cost. Readers can try it themselves.

In a stand-alone system, trylock is not a good choice. Because a large number of goroutine locks can cause a waste of meaningless resources in the CPU. There is a proper noun used to describe this lock-up scenario: livelock.

The live lock means that the program seems to be executing normally, but the CPU cycle is wasted on grabbing the lock instead of executing the task, so that the overall execution of the program is inefficient. The problem of livelocks is a lot more troublesome to locate. Therefore, in a stand-alone scenario, it is not recommended to use this type of lock.

## 6.2.3 Redis based setnx

In a distributed scenario, we also need this kind of "preemption" logic. What should we do at this time? We can use the `setnx` command provided by Redis:

```go
Package main

Import (
"fmt"
"sync"
"time"

"github.com/go-redis/redis"
)

Func incr() {
Client := redis.NewClient(&redis.Options{
Addr: "localhost:6379",
Password: "", // no password set
DB: 0, // use default DB
})

Var lockKey = "counter_lock"
Var counterKey = "counter"

// lock
Resp := client.SetNX(lockKey, 1, time.Second*5)
lockSuccess, err := resp.Result()

If err != nil || !lockSuccess {
fmt.Println(err, "lock result: ", lockSuccess)
Return
}

// counter ++
getResp := client.Get(counterKey)
cntValue, err := getResp.Int64()
If err == nil || err == redis.Nil {
cntValue++
Resp := client.Set(counterKey, cntValue, 0)
_, err := resp.Result()
If err != nil {
// log err
Println("set value error!")
}
}
Println("current counter is ", cntValue)

delResp := client.Del(lockKey)
unlockSuccess, err := delResp.Result()
If err == nil && unlockSuccess > 0 {
Println("unlock success!")
} else {
Println("unlock failed", err)
}
}

Func main() {
Var wg sync.WaitGroup
For i := 0; i < 10; i++ {
wg.Add(1)
Go func() {
Defer wg.Done()
Incr()
}()
}
wg.Wait()
}
```

Look at the results of the run:

```shell
❯❯❯ go run redis_setnx.go
<nil> lock result: false
<nil> lock result: false
<nil> lock result: false
<nil> lock result: false
<nil> lock result: false
<nil> lock result: false
<nil> lock result: false
<nil> lock result: false
<nil> lock result: false
Current counter is 2028
Unlock success!
```

Through the code and execution results, we can see that the remote call to `setnx` is very similar to the single trylock. If the lock fails, the relevant task logic should not continue to execute.

`setnx` is great for high-concurrency scenarios and is used to compete for some "unique" resources. For example, in the transaction matching system, the seller initiates an order, and multiple buyers will compete for it. In this scenario, we have no way to rely on specific time to judge the sequence, because no matter whether it is the time of the user equipment or the time of each machine in the distributed scene, there is no way to ensure the correct timing after the merger. Even if we are clustered in the same room, the system time of different machines may have subtle differences.

Therefore, we need to rely on the order of these requests to reach the Redis node to do the correct lock-up operation. If the user's network environment is relatively poor, then they can only ask for more.

## 6.2.4 Based on ZooKeeper

```go
Package main

Import (
"time"

"github.com/samuel/go-zookeeper/zk"
)

Func main() {
c, _, err := zk.Connect([]string{"127.0.0.1"}, time.Second) //*10)
If err != nil {
Panic(err)
}
l := zk.NewLock(c, "/lock", zk.WorldACL(zk.PermAll))
Err = l.Lock()
If err != nil {
Panic(err)
}
Println("lock succ, do your business logic")

time.Sleep(time.Second * 10)

// do some thing
l.Unlock()
Println("unlock succ, finish business logic")
}
```

The ZooKeeper-based lock differs from the Redis-based lock in that it locks until the lock succeeds, which is similar to `mutex.Lock` in our stand-alone scenario.

The principle is also based on the temporary Sequence node and the watch API, for example, we are using the `/lock` node. Lock will insert its own value in the node list under this node. As long as the child nodes under the node change, it will notify all programs that watch the node. At this time, the program will check whether the id of the smallest child node under the current node is consistent with its own. If they are consistent, the lock is successful.

This kind of distributed blocking lock is more suitable for distributed task scheduling scenarios, but it is not suitable for stealing scenarios with high frequency locking time. According to Google's Chubby paper, locks based on strong consistent protocols apply to the "coarse-grained" locking operation. The coarse grain size here means that the lock takes a long time. We should also consider whether it is appropriate to use it in our own business scenarios.

## 6.2.5 Based on etcd

Etcd is a component of a distributed system that is functionally similar to ZooKeeper and has been getting hotter in the past two years. Based on ZooKeeper, we implemented a distributed blocking lock. Based on etcd, we can also implement similar functions:

```go
Package main

Import (
"log"

"github.com/zieckey/etcdsync"
)

Func main() {
m, err := etcdsync.New("/lock",10, []string{"http://127.0.0.1:2379"})
If m == nil || err != nil {
log.Printf("etcdsync.New failed")
Return
}
Err = m.Lock()
If err != nil {
log.Printf("etcdsync.Lock failed")
Return
}

log.Printf("etcdsync.Lock OK")
log.Printf("Get the lock. Do something here.")

Err = m.Unlock()
If err != nil {
log.Printf("etcdsync.Unlock failed")
} else {
log.Printf("etcdsync.Unlock OK")
}
}
```

There is no Sequence node like ZooKeeper in etcd. So its lock implementation is different from the ZooKeeper implementation. The lock process for etcdsync used in the sample code above is:

1. Check if there is a value in the `/lock` path. If there is a value, the lock has been robbed by others.
2. If there is no value, write your own value. The write returns successfully, indicating that the lock is successful. If the node was written by another node during writing, it will cause the lock to fail. This time to 3
3. Watch the event under `/lock`, which is stuck at this time.
4. When an event occurs in the `/lock` path, the current process is awakened. Check if the event that occurred is a delete event (indicating that the lock is actively unlocked by the holder), or an expired event (indicating that the lock expires). If so, then go back and take the lock-up process.

It is worth mentioning that in the etcdv3 API, the official has provided a lock API that can be used directly. Readers can consult the etcd documentation for further study.

## 6.2.7 How to choose the right lock

When the business is still in the order of magnitude that can be achieved in a single machine, then any single lock solution can be used as required.

If you develop into a distributed service phase, but the business scale is not large, qps is small, the use of which lock scheme is similar. If you have a ZooKeeper, etcd, or Redis cluster available in your company, try to meet your business needs without introducing a new technology stack.

If the business develops to a certain level, it needs to be considered in many aspects. The first is whether your lock does not allow data loss under any harsh conditions. If not, then don't use Redis's simple lock of `setnx`.

If the reliability of the lock data is extremely high, then only the etcd or ZooKeeper lock scheme that guarantees data reliability through a coherent protocol can be used. But reliable backs tend to be lower throughput and higher latency. It needs to be stress tested according to the level of business to ensure that the etcd or ZooKeeper cluster used by distributed locks can withstand the pressure of actual business requests. It should be noted that there is no way to improve the performance of etcd and Zookeeper clusters by adding nodes. To scale it horizontally, you can only increase the number of clusters to support more requests. This will further increase the requirements for operation and maintenance and monitoring. Multiple clusters may need to be introduced with a proxy. Without a proxy, the service needs to be fragmented according to a certain service id. If the service is already expanded, you should also consider the dynamic migration of data. These are not easy things.

When choosing a specific plan, you still need to think more and make an early prediction of the risk.