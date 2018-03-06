# 6.8. Protocol-free 协议无关的服务入口

我们在之前的小节中看到了如何编写一个 thrift/grpc 协议的服务。在更早的小节中，我们看到了如何编写一个 http 的服务。

如果对于我们的服务模块，提出了更进一步的要求，想要同时支持 http 和 thrift 协议。要怎么办。

在 Uncle Bob 的 《Clean Architecture》 中对插件化架构(plugin architecture) 能够给我们一些启示。
