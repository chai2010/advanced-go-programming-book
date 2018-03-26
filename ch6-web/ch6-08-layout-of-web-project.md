# 6.8. layout 常见大型 web 项目分层

讲解多协议支持，和没有 v 的 api 分层设计。

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