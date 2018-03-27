# 6.8. layout 常见大型 web 项目分层

流行的 web 框架大多数是 MVC 框架，MVC 这个概念最早由 Trygve Reenskaug 在 1978 年提出，为了能够对 GUI 类型的应用进行方便扩展，将程序划分为：

1. 控制器（Controller）- 负责转发请求，对请求进行处理。
2. 视图（View） - 界面设计人员进行图形界面设计。
3. 模型（Model） - 程序员编写程序应有的功能（实现算法等等）、数据库专家进行数据管理和数据库设计(可以实现具体的功能)。

随着时代的发展，前端也变成了越来越复杂的工程，为了更好地工程化，现在更为流行的一般是前后分离的架构。可以认为前后分离是把 V 层从 MVC 中抽离单独成为项目。这样一个后端项目一般就只剩下 M 和 C 层了。前后端之间通过 ajax 来交互，有时候要解决跨域的问题，但也已经有了较为成熟的方案：

![前后分离](../images/ch6-08-frontend-backend.png)

事实上，即使是简单的项目，业界也并没有完全遵守 MVC 功能提出者对于 M 和 C 所定义的分工。有很多公司的项目会在 controller 层塞入大量的逻辑，在 model 层就只管理数据的存储。这往往来源于对于 model 层字面含义的某种擅自引申理解。认为字面意思，这一层就是处理某种建模，而模型是什么？就是数据呗！

这种理解显然是有问题的，业务流程也算是一种“模型”，是对真实世界用户行为或者既有流程的一种建模，并非只有按格式组织的数据才能叫模型。不过按照 MVC 的创始人的想法，我们如果把和数据打交道的代码还有业务流程全部塞进 MVC 里的 M 层的话，这个 M 层又会显得有些过于臃肿。对于复杂的项目，一个 C 和一个 M 层显然是不够用的，现在比较流行的纯后端 api 模块一般采用下述划分方法：

1. Controller，与上述类似，服务入口，负责处理路由，参数校验，请求转发
2. Logic/Service，逻辑(服务)层，一般是业务逻辑的入口，可以认为从这里开始，所有的请求参数一定是合法的。业务逻辑和业务流程也都在这一层中。常见的设计中会将该层称为 Business Rules。
3. DAO/Repository，这一层主要负责和数据、存储打交道。将下层存储以更简单的函数、接口形式暴露给 Logic 层来使用。负责数据的持久化工作。

每一层都会做好自己的工作，然后用请求当前的上下文构造下一层工作所需要的结构体或其它类型参数，然后调用下一次的函数。在工作完成之后，再把处理结果一层层地传出到入口。

TODOTODO，这里是一个请求，从 c->l->d 的流程，和返回的示意图。

划分为 CLD 三层之后，在 C 层我们可能还需要同时支持多种协议。本章前面讲到的 thrift、gRPC 和 http 并不是一定只选择其中一种，有时我们需要支持其中的两种，比如同一个接口，我们既需要效率较高的 thrift，也需要方便 debug 的 http 入口。这样请求的流程会变成下面这样：

TODOTODO，thrift protocol -> controller -> logic -> dao|| http protocol -> controller -> logic -> dao

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

我们成功地使自己的项目在入口支持了多种交互协议，但是还有一些问题没有解决。本节中所叙述的分层没有将 middleware 作为项目的分层考虑进去。如果我们考虑 middleware 的话，请求的流程是什么样的？

TODOTODO，这里是带上 middleware 之后的请求图。

之前我们学习的 middleware 是和 http 协议强相关的，在项目支持了多种协议之后，这种和协议强绑定的 middleware 成为了我们的瓶颈。如果我们坚持用之前的 middleware 方案的话，这里 thrift 的请求路线就还需要再多写一套 thrift 自己的 middleware，将业务无关的代码重复地写了两遍。请求流程变成了这样：

TODOTODO，这里是加入 middleware 之后的多协议框架请求处理流程。

这也是很多企业项目所面临的真实问题，遗憾的是开源界并没有这样方便的多协议 middleware 解决方案。

怎么解决这个问题呢，也不麻烦。把协议处理从 controller 中独立出去，新的 middleware 写在协议处理层后面。如图：

TODOTODO，这里有图，是将 middleware 后置之后的请求处理流程。

是不是感觉项目变的越来越复杂了？真实的项目就是这样一步一步，根据需求演进而来的。
