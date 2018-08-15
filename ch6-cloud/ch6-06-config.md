# 6.6 分布式配置管理

在分布式系统中，常困扰我们的还有上线问题。虽然目前有一些优雅重启方案，但实际应用中可能受限于我们系统内部的运行情况而没有办法做到真正的“优雅”。比如我们为了对去下游的流量进行限制，在内存中堆积一些数据，并对堆积设定时间/总量的阈值。在任意阈值达到之后将数据统一发送给下游，以避免频繁的请求超出下游的承载能力而将下游打垮。这种情况下重启要做到优雅就比较难了。

所以我们的目标还是尽量避免采用或者绕过上线的方式，对线上程序做一些修改。比较典型的修改内容就是程序的配置项。

## 场景举例

### 报表系统

在一些偏 OLAP 或者离线的数据平台中，经过长期的叠代开发，整个系统的功能模块已经渐渐稳定。可变动的项只出现在数据层，而数据层的变动大多可以认为是 SQL 的变动，架构师们自然而然地会想着把这些变动项抽离到系统外部。比如本节所述的配置管理系统。

当业务提出了新的需求时，我们的需求是将新的 SQL 录入到系统内部，或者简单修改一下老的 SQL。不对系统进行上线，就可以直接完成这些修改。

### 业务配置

大公司的平台部门服务众多业务线，在平台内为各业务线分配唯一 id。平台本身也由多个模块构成，这些模块需要共享相同的业务线定义(要不然就乱套了)。当公司新开产品线时，需要能够在短时间内打通所有平台系统的流程。这时候每个系统都走上线流程肯定是来不及的。另外需要对这种公共配置进行统一管理，同时对其增减逻辑也做统一管理。这些信息变更时，需要自动通知到业务方的系统，而不需要人力介入(或者只需要很简单的介入，比如点击审核通过)。

除业务线管理之外，很多互联网公司会按照城市来铺展自己的业务。在某个城市未开城之前，理论上所有模块都应该认为带有该城市 id 的数据是脏数据并自动过滤掉。而如果业务开城，在系统中就应该自己把这个新的城市 id 自动加入到白名单中。这样业务流程便可以自动运转。

再举个例子，互联网公司的运营系统中会有各种类型的运营活动，有些运营活动推出后可能出现了超出预期的事件(比如公关危机)，需要紧急将系统下线。这时候会用到一些开关来快速关闭相应的功能。或者快速将想要剔除的活动 id 从白名单中剔除。在 web 章节中的 ab test 一节中，我们也提到，有时需要有这样的系统来告诉我们当前需要放多少流量到相应的功能代码上。我们可以像那一节中，使用远程 rpc 来获知这些信息，但同时，也可以结合分布式配置系统，主动地拉取到这些信息。

## 使用 etcd 实现配置更新

我们使用 etcd 实现一个简单的配置读取和动态更新流程，以此来了解线上的配置更新流程。

### 配置定义

简单的配置，可以将内容完全存储在 etcd 中。比如：

```shell
etcdctl get /configs/remote_config.json
{
    "addr" : "127.0.0.1:1080",
    "aes_key" : "01B345B7A9ABC00F0123456789ABCDAF",
    "https" : false,
    "secret" : "",
    "private_key_path" : "",
    "cert_file_path" : ""
}
```

### 新建 etcd client

```go
cfg := client.Config{
    Endpoints:               []string{"http://127.0.0.1:2379"},
    Transport:               client.DefaultTransport,
    HeaderTimeoutPerRequest: time.Second,
}
```

### 配置获取

```go
resp, err = kapi.Get(context.Background(), "/name", nil)
if err != nil {
    log.Fatal(err)
} else {
    log.Printf("Get is done. Metadata is %q\n", resp)
    log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
}
```

### 配置更新订阅

```go
kapi := client.NewKeysAPI(c)
w := kapi.Watcher("/name", nil)
go func() {
    for {
        resp, err := w.Next(context.Background())
        fmt.Println(resp, err)
        fmt.Println("new values is ", resp.Node.Value)
    }
}()
```

### 整合起来

```go
package main

import (
    "fmt"
    "log"
    "time"

    "golang.org/x/net/context"

    "github.com/coreos/etcd/client"
)

func watchAndUpdate() {
}

func set() error {
    return nil
}

func get() (string, error) {
    return "", nil
}

func main() {
    cfg := client.Config{
        Endpoints:               []string{"http://127.0.0.1:2379"},
        Transport:               client.DefaultTransport,
        HeaderTimeoutPerRequest: time.Second,
    }

    c, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    kapi := client.NewKeysAPI(c)
    w := kapi.Watcher("/name", nil)
    go func() {
        for {
            resp, err := w.Next(context.Background())
            fmt.Println(resp, err)
            fmt.Println("new values is ", resp.Node.Value)
        }
    }()

    log.Print("Setting /name to alex")
    resp, err := kapi.Set(context.Background(), "/name", "alex", nil)
    if err != nil {
        log.Fatal(err)
    } else {
        log.Printf("Set is done. Metadata is %q\n", resp)
    }

    log.Print("Getting /name key value")
    resp, err = kapi.Get(context.Background(), "/name", nil)
    if err != nil {
        log.Fatal(err)
    } else {
        log.Printf("Get is done. Metadata is %q\n", resp)
        log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
    }
    time.Sleep(time.Minute)
}
```

如果业务规模不大，使用本节中的例子就可以实现功能了。TODO

## 配置膨胀

随着业务的发展，配置系统本身所承载的压力可能也会越来越大，配置文件可能成千上万。客户端同样上万，将配置内容存储在 etcd 内部便不再合适了。随着配置文件数量的膨胀，除了存储系统本身的吞吐量问题，还有配置信息的管理问题。我们需要对相应的配置进行权限管理，需要根据业务量进行配置存储的集群划分。如果客户端太多，导致了配置存储系统无法承受瞬时大量的 QPS，那可能还需要在客户端侧进行缓存优化，等等。

这也就是为什么大公司都会针对自己的业务额外开发一套复杂配置系统的原因。

## 配置版本管理

在配置管理过程中，难免出现用户误操作的情况，例如在更新配置时，输入了无法解析的配置。这种情况下我们可以通过配置校验来解决。

有时错误的配置可能不是格式上有问题，而是在逻辑上有问题。比如我们写 SQL 时少 select 了一个字段，更新配置时，不小心把丢掉了 json 字符串中的一个 field 而导致程序无法理解新的配置而进入诡异的逻辑。为了快速止损，最快且最有效的办法就是进行版本管理和，并支持按版本回滚。

在配置进行更新时，我们要为每份配置的新内容赋予一个版本号，并将修改前的内容和版本号记录下来，当发现新配置出问题时，能够及时地回滚回来。

## 客户端容错

在业务系统的配置被剥离到配置中心之后，并不意味着我们的系统可以高枕无忧了。当配置中心本身宕机时，我们也需要一定的容错能力，至少保证在其宕机期间，业务依然可以运转。这要求我们的系统能够在配置中心宕机时，也能拿到需要的配置信息。哪怕这些信息不够新。

具体来讲，在给业务提供配置读取的 sdk 时，最好能够将拿到的配置在业务机器的磁盘上也缓存一份。这样远程配置中心不可用时，可以直接用硬盘上的内容来做兜底。当重新连接上配置中心时，再把相应的内容进行更新。

加入缓存之后务必需要考虑的是数据一致性问题，当个别业务机器因为网络错误而与其它机器配置不一致时，我们也应该能够从监控系统中知晓。

我们使用一种手段解决了我们配置更新痛点，但同时可能因为使用的手段而带给我们新的问题。实际开发中，我们要对每一步决策多多思考，以使自己不在问题到来时手足无措。
