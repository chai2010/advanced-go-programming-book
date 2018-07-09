# 6.7. 分布式 id 生成器

有时我们需要能够生成类似 MySQL 自增 ID 这样不断增大，同时又不会重复的 id。以支持业务中的高并发场景。比较典型的，电商促销时，短时间内会有大量的订单涌入到系统，比如每秒 10w+。明星出轨时，会有大量热情的粉丝发微薄以表心意，同样产生短时间大量的消息。

在插入数据库之前，我们需要给这些消息/订单先打上一个 ID，然后再插入到我们的数据库。对这个 id 的要求是希望其中能带有一些时间信息，这样即使我们后端的系统对消息进行了分库分表，也能够以时间顺序对这些消息进行排序。

Twitter 的 snowflake 算法是这种场景下的一个典型解法。先来看看 snowflake 是怎么一回事：

```
                                                                                                     
                                                               datacenter_id          sequence_id    
    unused                                                                                           
                                                                      │                     │        
       │                                                              │                     │        
       │                                                              │                     │        
       │  │                                                      │    │                     │        
       │  │                                                      │    │                     │        
       ▼  │◀──────────────────    41 bits   ────────────────────▶│    ▼                     ▼        
    ┌─────┼──────────────────────────────────────────────────────┼────────┬────────┬────────────────┐
    │  0  │ 0000 0000 0000 0000 0000 0000 0000 0000 0000 0000 0  │ 00000  │ 00000  │ 0000 0000 0000 │
    └─────┴──────────────────────────────────────────────────────┴────────┴────────┴────────────────┘
                                      ▲                                        ▲                     
                                      │                                        │                     
                                      │                                        │                     
                                      │                                        │                     
                                      │                                        │                     
                                      │                                        │                     
                                      │                                        │                     
                                                                                                     
                            time in milliseconds                          worker_id                  

```

首先确定我们的数值是 64 位，int64 类型，被划分为四部分，不含开头的第一个 bit，因为这个 bit 是符号位。用 41 位来表示收到请求时的时间戳，单位为毫秒，然后五位来表示数据中心的 id，然后再五位来表示机器的实例 id，最后是 12 位的循环自增 id(到达 1111 1111 1111 后会归 0)。

这样的机制可以支持我们在同一台机器上，同一毫秒内产生 2 ^ 12 = 4096 条消息。一秒共 409.6w 条消息。从值域上来讲完全够用了。

数据中心 + 实例 id 共有 10 位，可以支持我们每数据中心部署 32 台机器，所有数据中心共 1024 台实例。

表示 timestamp 的 41 位，可以支持我们使用 69 年。当然，我们的时间毫秒计数不会真的从 1970 年开始记，那样我们的系统跑到 `2039/9/7 23:47:35` 就不能用了，所以这里的 timestamp 实际上只是相对于某个时间的增量，比如我们的系统上线是 2018-08-01，那么我们可以把这个 timestamp 当作是从 `2018-08-01 00:00:00.000` 的偏移量。

## worker id　分配

timestamp，datacenter_id，worker_id 和 sequence_id 这四个字段中，timestamp 和 sequence_id 是由程序在运行期生成的。但 datacenter_id 和 worker_id 需要我们在部署阶段就能够获取得到，并且一旦程序启动之后，就是不可更改的了(想想，如果可以随意更改，可能被不慎修改，造成最终生成的 id 有冲突)。

一般不同数据中心的机器，会提供对应的获取数据中心 id 的 api，所以 datacenter_id 我们可以在部署阶段轻松地获取到。而 worker_id 是我们逻辑上给机器分配的一个 id，这个要怎么办呢？比较简单的想法是由能够提供这种自增 id 功能的工具来支持，比如 MySQL:

```shell
mysql> insert into a (ip) values("10.1.2.101");
Query OK, 1 row affected (0.00 sec)

mysql> select last_insert_id();
+------------------+
| last_insert_id() |
+------------------+
|                2 |
+------------------+
1 row in set (0.00 sec)
```

从 MySQL 中获取到 worker_id 之后，就把这个 worker_id 直接持久化到本地，以避免每次上线时都需要获取新的 worker_id。让单实例的 worker_id 可以始终保持不变。

当然，使用 MySQL 相当于给我们简单的 id 生成服务增加了一个外部依赖。依赖越多，我们的服务的可运维性就越差。

考虑到集群中即使有单个 id 生成服务的实例挂了，也就是损失一段时间的一部分 id，所以我们也可以更简单暴力一些，把 worker_id 直接写在 worker 的配置中，上线时，由部署脚本完成 worker_id 字段替换。

## 开源实例

`github.com/bwmarrin/snowflake` 是一个相当轻量化的 snowflake 的 Go 实现。其文档指出：

```
+--------------------------------------------------------------------------+
| 1 Bit Unused | 41 Bit Timestamp |  10 Bit NodeID  |   12 Bit Sequence ID |
+--------------------------------------------------------------------------+
```

和标准的 snowflake 完全一致。使用上比较简单：

```go
package main

import (
    "fmt"
    "os"

    "github.com/bwmarrin/snowflake"
)

func main() {
    n, err := snowflake.NewNode(1)
    if err != nil {
        println(err)
        os.Exit(1)
    }

    for i := 0; i < 3; i++ {
        id := n.Generate()
        fmt.Println("id", id)
        fmt.Println("node: ", id.Node(), "step: ", id.Step(), "time: ", id.Time(), "\n")
    }
}

```

当然，这个库也给我们留好了定制的后路： 

```go
    // Epoch is set to the twitter snowflake epoch of Nov 04 2010 01:42:54 UTC
    // You may customize this to set a different epoch for your application.
    Epoch int64 = 1288834974657

    // Number of bits to use for Node
    // Remember, you have a total 22 bits to share between Node/Step
    NodeBits uint8 = 10

    // Number of bits to use for Step
    // Remember, you have a total 22 bits to share between Node/Step
    StepBits uint8 = 12
```

Epoch 就是本节开头讲的起始时间，NodeBits 指的是机器编号的位长，StepBits 指的是自增序列的位长。

sonyflake
