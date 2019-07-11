# 6.2 分布式锁

在单机程序并发或并行修改全局变量时，需要对修改行为加锁以创造临界区。为什么需要加锁呢？我们看看在不加锁的情况下并发计数会发生什么情况：

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
❯❯❯ go run local_lock.go
945
❯❯❯ go run local_lock.go
937
❯❯❯ go run local_lock.go
959
```

## 6.2.1 进程内加锁

想要得到正确的结果的话，要把对计数器（counter）的操作代码部分加上锁：

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
❯❯❯ go run local_lock.go
1000
```

## 6.2.2 trylock

在某些场景，我们只是希望一个任务有单一的执行者。而不像计数器场景一样，所有goroutine都执行成功。后来的goroutine在抢锁失败后，需要放弃其流程。这时候就需要trylock了。

trylock顾名思义，尝试加锁，加锁成功执行后续流程，如果加锁失败的话也不会阻塞，而会直接返回加锁的结果。在Go语言中我们可以用大小为1的Channel来模拟trylock：

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
func (l Lock) Unlock() {
	l.c <- struct{}{}
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

因为我们的逻辑限定每个goroutine只有成功执行了`Lock`才会继续执行后续逻辑，因此在`Unlock`时可以保证Lock结构体中的channel一定是空，从而不会阻塞，也不会失败。上面的代码使用了大小为1的channel来模拟trylock，理论上还可以使用标准库中的CAS来实现相同的功能且成本更低，读者可以自行尝试。

在单机系统中，trylock并不是一个好选择。因为大量的goroutine抢锁可能会导致CPU无意义的资源浪费。有一个专有名词用来描述这种抢锁的场景：活锁。

活锁指的是程序看起来在正常执行，但CPU周期被浪费在抢锁，而非执行任务上，从而程序整体的执行效率低下。活锁的问题定位起来要麻烦很多。所以在单机场景下，不建议使用这种锁。

## 6.2.3 基于Redis的setnx

在分布式场景下，我们也需要这种“抢占”的逻辑，这时候怎么办呢？我们可以使用Redis提供的`setnx`命令：

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
	if err == nil || err == redis.Nil {
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

看看运行结果：

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

通过代码和执行结果可以看到，我们远程调用`setnx`运行流程上和单机的trylock非常相似，如果获取锁失败，那么相关的任务逻辑就不应该继续向前执行。

`setnx`很适合在高并发场景下，用来争抢一些“唯一”的资源。比如交易撮合系统中卖家发起订单，而多个买家会对其进行并发争抢。这种场景我们没有办法依赖具体的时间来判断先后，因为不管是用户设备的时间，还是分布式场景下的各台机器的时间，都是没有办法在合并后保证正确的时序的。哪怕是我们同一个机房的集群，不同的机器的系统时间可能也会有细微的差别。

所以，我们需要依赖于这些请求到达Redis节点的顺序来做正确的抢锁操作。如果用户的网络环境比较差，那也只能自求多福了。

## 6.2.4 基于ZooKeeper

```go
package main

import (
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

func main() {
	c, _, err := zk.Connect([]string{"127.0.0.1"}, time.Second) //*10)
	if err != nil {
		panic(err)
	}
	l := zk.NewLock(c, "/lock", zk.WorldACL(zk.PermAll))
	err = l.Lock()
	if err != nil {
		panic(err)
	}
	println("lock succ, do your business logic")

	time.Sleep(time.Second * 10)

	// do some thing
	l.Unlock()
	println("unlock succ, finish business logic")
}
```

基于ZooKeeper的锁与基于Redis的锁的不同之处在于Lock成功之前会一直阻塞，这与我们单机场景中的`mutex.Lock`很相似。

其原理也是基于临时Sequence节点和watch API，例如我们这里使用的是`/lock`节点。Lock会在该节点下的节点列表中插入自己的值，只要节点下的子节点发生变化，就会通知所有watch该节点的程序。这时候程序会检查当前节点下最小的子节点的id是否与自己的一致。如果一致，说明加锁成功了。

这种分布式的阻塞锁比较适合分布式任务调度场景，但不适合高频次持锁时间短的抢锁场景。按照Google的Chubby论文里的阐述，基于强一致协议的锁适用于`粗粒度`的加锁操作。这里的粗粒度指锁占用时间较长。我们在使用时也应思考在自己的业务场景中使用是否合适。

## 6.2.5 基于etcd

etcd是分布式系统中，功能上与ZooKeeper类似的组件，这两年越来越火了。上面基于ZooKeeper我们实现了分布式阻塞锁，基于etcd，也可以实现类似的功能：

```go
package main

import (
	"log"

	"github.com/zieckey/etcdsync"
)

func main() {
	m, err := etcdsync.New("/lock", 10, []string{"http://127.0.0.1:2379"})
	if m == nil || err != nil {
		log.Printf("etcdsync.New failed")
		return
	}
	err = m.Lock()
	if err != nil {
		log.Printf("etcdsync.Lock failed")
		return
	}

	log.Printf("etcdsync.Lock OK")
	log.Printf("Get the lock. Do something here.")

	err = m.Unlock()
	if err != nil {
		log.Printf("etcdsync.Unlock failed")
	} else {
		log.Printf("etcdsync.Unlock OK")
	}
}
```

etcd中没有像ZooKeeper那样的Sequence节点。所以其锁实现和基于ZooKeeper实现的有所不同。在上述示例代码中使用的etcdsync的Lock流程是：

1. 先检查`/lock`路径下是否有值，如果有值，说明锁已经被别人抢了
2. 如果没有值，那么写入自己的值。写入成功返回，说明加锁成功。写入时如果节点被其它节点写入过了，那么会导致加锁失败，这时候到 3
3. watch `/lock`下的事件，此时陷入阻塞
4. 当`/lock`路径下发生事件时，当前进程被唤醒。检查发生的事件是否是删除事件（说明锁被持有者主动unlock），或者过期事件（说明锁过期失效）。如果是的话，那么回到 1，走抢锁流程。

值得一提的是，在etcdv3的API中官方已经提供了可以直接使用的锁API，读者可以查阅etcd的文档做进一步的学习。

## 6.2.7 如何选择合适的锁

业务还在单机就可以搞定的量级时，那么按照需求使用任意的单机锁方案就可以。

如果发展到了分布式服务阶段，但业务规模不大，qps很小的情况下，使用哪种锁方案都差不多。如果公司内已有可以使用的ZooKeeper、etcd或者Redis集群，那么就尽量在不引入新的技术栈的情况下满足业务需求。

业务发展到一定量级的话，就需要从多方面来考虑了。首先是你的锁是否在任何恶劣的条件下都不允许数据丢失，如果不允许，那么就不要使用Redis的`setnx`的简单锁。

对锁数据的可靠性要求极高的话，那只能使用etcd或者ZooKeeper这种通过一致性协议保证数据可靠性的锁方案。但可靠的背面往往都是较低的吞吐量和较高的延迟。需要根据业务的量级对其进行压力测试，以确保分布式锁所使用的etcd或ZooKeeper集群可以承受得住实际的业务请求压力。需要注意的是，etcd和Zookeeper集群是没有办法通过增加节点来提高其性能的。要对其进行横向扩展，只能增加搭建多个集群来支持更多的请求。这会进一步提高对运维和监控的要求。多个集群可能需要引入proxy，没有proxy那就需要业务去根据某个业务id来做分片。如果业务已经上线的情况下做扩展，还要考虑数据的动态迁移。这些都不是容易的事情。

在选择具体的方案时，还是需要多加思考，对风险早做预估。
