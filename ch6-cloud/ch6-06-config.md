# 6.3 分布式配置管理

在分布式系统中，常困扰我们的还有上线问题。虽然目前有一些优雅重启方案，但实际应用中可能受限于我们系统内部的运行情况而没有办法做到真正的“优雅”。比如我们为了对去下游的流量进行限制，在内存中堆积一些数据，并对堆积设定时间/总量的阈值。在任意阈值达到之后将数据统一发送给下游，以避免频繁的请求超出下游的承载能力而将下游打垮。这种情况下重启要做到优雅就比较难了。

所以我们的目标还是能尽量避免或者绕过上线的情况下，对线上程序做一些修改。比较典型的修改内容就是程序的配置项。

## 使用 etcd 实现配置更新

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

### 配置膨胀之后

随着业务的发展，配置系统本身所承载的压力可能也会越来越大，配置文件可能成千上万。客户端同样上万，将配置内容存储在 etcd 内部便不再合适了。随着配置文件数量的膨胀，除了存储系统本身的吞吐量问题，还有配置信息的管理问题。我们需要对相应的配置进行权限管理，需要根据业务量进行配置存储的集群划分。如果客户端太多，导致了配置存储系统无法承受瞬时大量的 QPS，那可能还需要在客户端侧进行缓存优化，等等。

这也就是为什么大公司都会针对自己的业务额外开发一套复杂配置系统的原因。
