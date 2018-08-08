# 5.10 Service Discovery 服务发现

在微服务架构中，服务之间是存在依赖的。例如在订单系统中创建订单时，需要对用户信息做快照，这时候也就意味着这个流程要依赖: 订单、用户两个系统。当前大型网站的语境下，多服务分布式共存，单个服务也可能会跑在多台物理/虚拟机上。所以即使你知道你需要依赖的是“订单服务”这个具体的服务，实际面对的仍然是多个 ip+port 组成的集群。因此你需要: 1. 通过“订单服务”这个名字找到它对应的 ip+port 列表；2. 决定把这个请求发到哪一个 ip+port 上的订单服务。

ip+port 的组合往往被称为 endpoint。通过“订单服务”去找到这些 endpoint 的过程，叫做服务发现。选择把请求发送给哪一台机器，以最大化利用下游机器的过程，叫做负载均衡。本节主要讨论服务发现。

## 为什么不把 ip+port 写死在自己的配置文件中

在大型网站的语境下，为了高可用每一个服务都一定会有多个节点来分担压力和风险，而现在随着 docker 之类的工具的盛行，依赖服务的节点是存在迁移的可能性的。所以对于依赖服务会慢慢不再将其 ip+port 写死在自己的服务的配置文件中。否则在依赖服务迁移时，就需要跟着依赖服务修改配置，耗时费力。

如果依赖服务所在的机器挂掉了，也就意味着那个 ip+port 不可用了。这时候上游还需要有某种反馈机制能够及时知晓，并且能够在知晓之后，不经过上线就能将已经失效的 ip+port 自动从依赖中摘除。

我们先来看看怎么通过服务名字找到这些 ip+port 列表。

## 怎么通过服务名字找到 endpoints

把一个名字映射到多个 ip+port 这件事情，大多数人脑子里冒出的第一个想法应该是 "dns服务"。确实 dns 就是干这件事情的，有一些公司会提供内网 dns 服务，服务彼此之间通过 dns 来查找服务节点，这种情况下，你使用的下游服务的名字可能是类似 `api.order.service_endpoints` 的字符串，当然，如果这个 dns 服务是你来开发的话，这种字符串你可以随意定义。只要能用 namespace 按业务部门和相应的服务区分开就可以。实现内网 dns 服务的话，你还可以使用更激进的刷新策略(例如：一分钟刷新一次)，不像公网的 dns 那样需要很长时间甚至一整天才能生效。

不过使用公共的 dns 服务也存在问题，我们的 dns 服务会变成整个服务的集中的那个中心点，这样会给整个分布式系统带来一定的风险。一旦 dns 服务挂了，那么我们也就找不到自己的依赖了。我们可以使用自己本地的缓存来缓解这个问题。比如某个服务最近访问过下游服务，那么可以将下游的 ip+port 缓存在本地，如果 dns 服务挂掉了，那我至少可以用本地的缓存做个兜底，不至于什么都找不到。

```
                           ┌────────────────┐           .─────────────────.
                           │   My Service   │─────────▶(  local dns cache  )
                           └────────────────┘           `─────────────────'
                                    │
                                    │
                                    │
          ┌─────────  X   ──────────┤
          │                         │
          │                         │
          │                         │
          ▼                         │
 .─────────────────.                │
(    dns service    )               │
 `─────────────────'                │
                                    │
                                    ▼
                       ┌────────────────────────┐
                       │   dependent service    │
                       └────────────────────────┘
```

服务名和 endpoints 的对应也很直观，无非 `字符串` -> `endpoint 列表`。

我们自己来设计的话，只需要有一个 kv 存储就可以了。拿 redis 举例，我们可以用 set 来存储 endpoints 列表:

```shell
redis-cli> sadd order_service.http 100.10.1.15:1002
redis-cli> sadd order_service.http 100.10.2.11:1002
redis-cli> sadd order_service.http 100.10.5.121:1002
redis-cli> sadd order_service.http 100.10.6.1:1002
redis-cli> sadd order_service.http 100.10.10.1:1002
redis-cli> sadd order_service.http 100.10.100.11:1002
```

获取 endpoint 列表也很简单:

```shell
127.0.0.1:6379> smembers order_service.http
1) "100.10.1.15:1002"
2) "100.10.5.121:1002"
3) "100.10.10.1:1002"
4) "100.10.100.11:1002"
5) "100.10.2.11:1002"
6) "100.10.6.1:1002"
```

从存储的角度来讲，既然 kv 能存，那几乎所有其它的存储系统都可以存。如果我们对这些数据所在的存储系统可靠性有要求，还可以把这些服务名字和列表的对应关系存储在 MySQL 中，也没有问题。

## 故障节点摘除

上一小节讲的是存储的问题，在服务发现中，还有一个比较重要的命题，就是故障摘除。之所以开源界有很多服务发现的轮子，也正是因为这件事情并不是把 kv 映射存储下来这么简单。更重要的是我们能够在某个服务节点宕机时，让依赖该节点的其它服务感知得到这个“宕机”的变化，从而不再向其发送任何请求。

故障摘除有两种思路:

1. 客户端主动的故障摘除
2. 客户端被动故障摘除。

主动的故障摘除是指，我作为依赖其他人的上游，在下游一台机器挂掉的时候，我可以自己主动把它从依赖的节点列表里摘掉。常见的手段也有两种，一种是靠应用层心跳，还有一种靠请求投票。下面是一种根据请求时是否出错，对相应的服务节点进行投票的一个例子：

```go
// 对下游的请求正常返回时:
node := getNodeFromPool()

resp, err := remoteRPC(ctx, params)

if err != nil {
    node.Vote(status.Healthy)
} else {
    node.Vote(status.Unhealthy)
}
```

在节点管理时，会对 Unhealthy 过多的节点进行摘除，这个过程可以在 Unhealthy 的数量超过一定的阈值之后自动触发，也就是在 Vote 函数中实现即可。

如果你选择用应用层心跳，那需要下游提供 healthcheck 的接口，这个接口一般就简单返回 success 就可以了。上游要做的事情就是每隔一小段时间，去请求 healthcheck 接口，如果超时、响应失败，那么就把该节点摘除:

```go
healthcheck := func(endpoint string) {
    for {
        time.Sleep(time.Second * 10)

        resp, err := callRemoteHealthcheckAPI(endpoint)
        if err != nil {
            dropThisAPINode()
        }
    }
}()

for _, endpoint := range endpointList {
    go healthcheck(endpoint)
}
```

被动故障摘除，顾名思义。依赖出问题了要别人通知我。这个通知一般通过服务注册中心发给我。

被动故障摘除，最早的解决方案是 zookeeper 的 ephemeral node，java 技术栈的服务发现框架很多是基于此来做故障服务节点摘除。

比如我们是电商的平台部的订单系统，那么可以建立类似这样的永久节点:

```shell
/platform/order-system/create-order-service-http
```

然后把我们的 endpoints 作为临时节点，建立在上述节点之下:

```shell
ls /platform/order-system/create-order-service-http

['10.1.23.1:1023', '10.11.23.1:1023']
```

当与 zk 断开连接时，注册在该节点下的临时节点也会消失，即实现了服务节点故障时的被动摘除。

目前在企业级应用中，上述几种故障摘除方案都是存在的。读者朋友可以根据自己公司的发展阶段，灵活选用对应的方案。需要明白的一点是，并非一定要有 zk、etcd 这样的组件才能完成故障摘除。

## 基于 zk 的完整服务发现流程

用代码来实现一下上面的几个逻辑。

### 临时节点注册

```go
package main

import (
    "fmt"
    "time"

    "github.com/samuel/go-zookeeper/zk"
)

func main() {
    c, _, err := zk.Connect([]string{"127.0.0.1"}, time.Second)
    if err != nil {
        panic(err)
    }

    res, err := c.Create("/platform/order-system/create-order-service-http/10.1.13.3:1043", []byte("1"),
        zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
    if err != nil {
        panic(err)
    }
    println(res)
    time.Sleep(time.Second * 50)
}
```

在 sleep 的时候我们在 cli 中查看写入的临时节点数据：

```shell
ls /platform/order-system/create-order-service-http
['10.1.13.3:1043']
```

在程序结束之后，很快这条数据也消失了：

```shell
ls /platform/order-system/create-order-service-http
[]
```

### watch 数据变化

### 消息通知

## 总结

有了临时节点、监视功能、故障时的自动摘除功能，我们实现一套服务发现以及故障节点摘除的基本元件也就齐全了。
