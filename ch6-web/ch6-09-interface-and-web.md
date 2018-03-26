# 6.8. interface 和 web 编程

我们在之前的小节中看到了如何编写一个 thrift/grpc 协议的服务。在更早的小节中，我们看到了如何编写一个 http 的服务。

如果对于我们的服务模块，提出了更进一步的要求，想要同时支持 http 和 thrift 协议。要怎么办。

公司内的基础架构因为技术实力原因，最早是从别人那里借来的 kv 存储方案。随着公司的发展，渐渐有大牛加入，想要甩掉这个借来的包袱自研 kv 存储，但接口与之前的 kv 存储不兼容。接入时需要业务改动接入代码，怎么写代码才能让我的核心业务逻辑不受这些外部资源变化影响呢。

## interface 与依赖反转

学习 Golang 时一般对 interface 都会建立较基本的理解。从架构的角度上来讲，interface 解决的最大的问题是依赖方向问题。例如在一个典型的 web 程序中：

TODOTODO 这里有图，标明控制流的方向

我们的控制流方向是从 controller -> logic -> dao，在不使用 interface 的前提下，我们在 controller 中需要 import logic 的 package，然后在 logic 中需要 import dao 的 package。这种 import 关系和控制流的方向是一致的，因为我们需要用到 a.x 函数，那么 import a 就显得自然而然了。而 import 意味着依赖，也就是说我们的依赖方向与控制流的方向是完全一致的。

从架构的角度讲，这个控制流会给我们带来很多问题：

```
1. dao 的变动必然会引起 logic 的变动
2. 核心的业务逻辑 logic 代码变动会给我们带来较大的风险
3.
```

interface 这时候就成为了我们的救星，如果我们在 a->b 这个控制方向上不满意，不想让 b 的变化引起 a 的不适，那么我们就在 a 与 b 之间插入一层 interface。

```go
controller -> logic (interfaces defined in package a) <- dao
```

通过插入一层 interface，代码中的依赖方向发生了变化，如图：

TODOTODO，这里是控制流和依赖流的示意图。

这样就可以让 logic 摆脱了对 dao 的依赖，从而将 logic 的代码保护了起来。就像 Uncle Bob 所描述的那样：

![插件化架构](../images/ch6-08-plugin-arch.jpg)

dao 成为了 logic 的 plugin(插件)。如果我们要把 dao 里的 kv 数据库从 rocksdb(假如) 替换为自研的 thrift 协议 kv 存储，那么新的 dao 实现也只要遵从之前定义好的 interface 就可以，logic 不需要任何变化。这就是所谓的插件化架构。

稍微具体一点：

```go
type RecordSaver interface {
    func Save(r MyRecord) error
}
```

## 