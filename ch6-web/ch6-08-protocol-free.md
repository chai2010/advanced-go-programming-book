# 6.8. Protocol-free 协议无关的服务入口

我们在之前的小节中看到了如何编写一个 thrift/grpc 协议的服务。在更早的小节中，我们看到了如何编写一个 http 的服务。

如果对于我们的服务模块，提出了更进一步的要求，想要同时支持 http 和 thrift 协议。要怎么办。

在 Uncle Bob 的 《Clean Architecture》 中对插件化架构(plugin architecture) 的阐释能够给我们一些启示。我们先来看看什么是“插件化架构”。

![插件化架构](../images/ch6-08-plugin-arch.jpg)

上面这张图比较典型。 简单地来说，我们在 Business Rules 这一层定义了一些 interface，例如我们需要对数据进行存储，那么就有可能定义下面这样的接口：

```go
type RecordStore interface {
    func Save(r MyRecord) error
}
```


通过在业务逻辑层定义 interface 来对业务逻辑进行一定的保护。这样在周边的代码发生变动时，不会对业务逻辑产生任何影响。比较典型的应用，例如你们公司的服务原来是 C/S 架构，在 web 2.0 时代突然流行起了 B/S 架构，然后在移动互联网的浪潮下又开始流行 C/S 或者一些 hybrid app 架构，但是这些变化对于后端程序来说，大多数的代码应该能够做到置身事外。再比如你们公司之前因为技术实力限制，没有适合自己用的可以持久化的 kv 存储。
