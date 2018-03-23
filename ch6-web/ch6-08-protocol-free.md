# 6.8. Protocol-free 协议无关的服务入口

我们在之前的小节中看到了如何编写一个 thrift/grpc 协议的服务。在更早的小节中，我们看到了如何编写一个 http 的服务。

如果对于我们的服务模块，提出了更进一步的要求，想要同时支持 http 和 thrift 协议。要怎么办。

## clean architecture

在 Uncle Bob 的 《Clean Architecture》 中对插件化架构(plugin architecture) 的阐释能够给我们一些启示。我们先来看看什么是“插件化架构”。

![插件化架构](../images/ch6-08-plugin-arch.jpg)

上面这张图比较典型。 简单地来说，我们在 Business Rules 这一层定义了一些 interface，例如我们需要对数据进行存储，那么就有可能定义下面这样的接口：

```go
type RecordStore interface {
    func Save(r MyRecord) error
}
```

通过在业务逻辑层定义 interface 来对业务逻辑进行一定的保护。这样在周边的代码发生变动时，不会对业务逻辑产生任何影响。为什么？因为这样我们的 logic 就可以规定，我要完成我的工作，需要什么样的动作，而不用关心这个动作在实现方那里具体是怎么做的。具体实现发生变动时，logic 的代码不需要做任何修改。

比较典型的应用，例如你们公司的服务原来是 C/S 架构，在 web 2.0 时代突然流行起了 B/S 架构，然后在移动互联网的浪潮下又开始流行 C/S 或者一些 hybrid app 架构，但是这些变化对于后端程序来说，大多数的代码应该能够做到置身事外。再比如你们公司之前因为技术升级，可能多次切换底层存储，但在存储迁移过程中，哪怕是 dao 层的相关代码要做一些修改，这些修改的影响也不应该侵入到业务逻辑层。

## interface 应用

interface 的最大好处，就是帮我们完成了这样的依赖反转。针对本节的处理协议的场景具体要怎么做呢？假如我们现在有一个下订单的需求，我们可以在 logic 层(或者叫 service 层) 定义一个 interface：

```go
// project/service/dto
type CreateOrderParams struct {
    OrderID int64
    ShopID int64
    ProductID int64
    CreateTime time.Time
}

// 对订单服务入口的定义
type Entry interface {
    GetCreateOrderParams() dto.CreateOrderParams
}

func CreateOrder(e Entry) error {
    params := e.GetCreateOrderParams()

    // do some thing to create order

    return nil
}

// project/controller
type ThriftGetOrderEntry struct{
    thriftRequestParams ThriftCreateOrderRequest
}

type HTTPGetOrderEntry struct{
    r *http.Request
}

func (te ThriftGetOrderEntry) GetCreateOrderParams() dto.CreateOrderParams {
    thriftRequestParams := te.thriftRequestParams
    return logic.CreateOrderParams{
        OrderID :    thriftRequestParams.OrderID,
        ShopID :     thriftRequestParams.ShopID,
        ProductID :  thriftRequestParams.ProductID,
        CreateTime : thriftRequestParams.CreateTime,
    }
}

func (he HTTPGetOrderEntry) GetCreateOrderParams() dto.CreateOrderParams {
    // r := he.r
    // get data
    err := json.Unmarshal(data, &req) // or read body or something

    return logic.CreateOrderParams{
        OrderID : req.OrderID,
        ShopID : req.ShopID,
        ProductID : req.ProductID,
        CreateTime : req.CreateTime,
    }
}

// thrift serve on 9000
func ThriftCreateOrderHandler(req ThriftCreateOrderRequest) (resp ThriftCreateOrderResp, error){
    thriftEntryInstance  := ThriftGetOrderEntry{
       thriftRequestParams : req,
    }

    logicResp,err := logic.CreateOrder(thriftEntryInstance)
    if err != nil {}
    // ...
}

// http serve on 8000
func HTTPCreateOrderHandler(wr http.ResponseWriter, r *http.Request) {
    httpEntryInstance  := HTTPGetOrderEntry{
        r : r,
    }

    logicResp,err := logic.CreateOrder(httpEntryInstance)
    if err != nil {}
    // ...
}

```

这样在对协议层进行修改时，就可以对 logic 层没有任何影响了。像前面提到的，这样的做法我们同样可以用在 logic 和 dao 层的交接处，也没有什么问题。

## codegen 来实现同样的功能

当然，如果完全按照 clean architecture 的设计来写代码还是有一些麻烦。在可预见的范围内，我们需要处理的协议类型是有限的。现在互联网公司内部 API 常用的就只有三种 gRPC、thrift 和 http。我们甚至可以通过一些手段，使我们每写一个接口，都可以支持这三种协议。先来把 logic 的入口简化一些，这里我们使用生产环境的 logic 函数签名作为样例：

```go
func CreateOrder(ctx context.Context, req *CreateOrderStruct) (*CreateOrderRespStruct, error) {
}
```

CreateOrder 有两个参数，ctx 用来传入 trace_id 一类的需要串联请求的全局参数，req 里存储了我们创建订单所需要的所有输入信息。返回结果是一个响应结构体和错误。可以认为，我们的代码运行到 logic 层之后，就没有任何与“协议”相关的代码了。在这里你找不到 http.Request，也找不到 http.ResponseWriter，也找不到任何与 thrift 或者 gRPC 相关的字眼。

```go

// in logic
type CreateOrderRequest struct {
    OrderID `json:"order_id"`
    // ...
}

func HTTPCreateOrderHandler(wr http.ResponseWriter, r *http.Request) {
    var params CreateOrderRequest
    ctx := context.TODO()
    // bind data to req
    logicResp,err := logic.CreateOrder(ctx, &params)
    if err != nil {}
    // ...
}
```

理论上我们可以用同一个 request struct 组合上不同的 tag，来达到一个 struct 来给不同的协议复用的目的。不过遗憾的是在 thrift 中，request struct 也是通过 IDL 生成的，其内容在自动生成的 ttypes.go 文件中，我们还是需要在 thrift 的入口将这个自动生成的 struct 映射到我们 logic 入口所需要的 struct 上。gRPC 也是类似。这部分代码还是需要的。

聪明的读者可能已经可以看出来了，协议细节处理这一层实际上有大量重复劳动，每一个接口在协议这一层的处理，无非是把数据从协议特定的 struct(例如 http.Request，thrift 的被包装过了) 读出来，再绑定到我们协议相关的 struct 上，再把这个 struct 映射到 logic 入口的 struct 上，这些代码实际上长得都差不多。差不多的代码都遵循着某种模式，那么我们可以对这些模式进行简单的抽象，用 codegen 来把繁复的协议处理代码从工作内容中抽离出去。

还是举个例子：

```go
```

我们需要一个基准 request struct，来根据这个 request struct 生成我们需要的入口代码。这个基准要怎么找呢？