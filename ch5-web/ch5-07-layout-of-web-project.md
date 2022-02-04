# 5.7 layout 常见大型 Web 项目分层

流行的 Web 框架大多数是 MVC 框架，MVC 这个概念最早由 Trygve Reenskaug 在 1978 年提出，为了能够对 GUI 类型的应用进行方便扩展，将程序划分为：

1. 控制器（Controller）- 负责转发请求，对请求进行处理。
2. 视图（View） - 界面设计人员进行图形界面设计。
3. 模型（Model） - 程序员编写程序应有的功能（实现算法等等）、数据库专家进行数据管理和数据库设计（可以实现具体的功能）。

随着时代的发展，前端也变成了越来越复杂的工程，为了更好地工程化，现在更为流行的一般是前后分离的架构。可以认为前后分离是把 V 层从 MVC 中抽离单独成为项目。这样一个后端项目一般就只剩下 M 和 C 层了。前后端之间通过 ajax 来交互，有时候要解决跨域的问题，但也已经有了较为成熟的方案。*图 5-13* 是一个前后分离的系统的简易交互图。

![前后分离](../images/ch6-08-frontend-backend.png)

*图 5-13 前后分离交互图*

图里的 Vue 和 React 是现在前端界比较流行的两个框架，因为我们的重点不在这里，所以前端项目内的组织我们就不强调了。事实上，即使是简单的项目，业界也并没有完全遵守 MVC 框架提出者对于 M 和 C 所定义的分工。有很多公司的项目会在 Controller 层塞入大量的逻辑，在 Model 层就只管理数据的存储。这往往来源于对于 model 层字面含义的某种擅自引申理解。认为字面意思，这一层就是处理某种建模，而模型是什么？就是数据呗！

这种理解显然是有问题的，业务流程也算是一种 “模型”，是对真实世界用户行为或者既有流程的一种建模，并非只有按格式组织的数据才能叫模型。不过按照 MVC 的创始人的想法，我们如果把和数据打交道的代码还有业务流程全部塞进 MVC 里的 M 层的话，这个 M 层又会显得有些过于臃肿。对于复杂的项目，一个 C 和一个 M 层显然是不够用的，现在比较流行的纯后端 API 模块一般采用下述划分方法：

1. Controller，与上述类似，服务入口，负责处理路由，参数校验，请求转发。
2. Logic/Service，逻辑（服务）层，一般是业务逻辑的入口，可以认为从这里开始，所有的请求参数一定是合法的。业务逻辑和业务流程也都在这一层中。常见的设计中会将该层称为 Business Rules。
3. DAO/Repository，这一层主要负责和数据、存储打交道。将下层存储以更简单的函数、接口形式暴露给 Logic 层来使用。负责数据的持久化工作。

每一层都会做好自己的工作，然后用请求当前的上下文构造下一层工作所需要的结构体或其它类型参数，然后调用下一层的函数。在工作完成之后，再把处理结果一层层地传出到入口，如 *图 5-14 所示*。

![controller-logic-dao](../images/ch6-08-controller-logic-dao.png)

*图 5-14 请求处理流程*

划分为 CLD 三层之后，在 C 层之前我们可能还需要同时支持多种协议。本章前面讲到的 thrift、gRPC 和 http 并不是一定只选择其中一种，有时我们需要支持其中的两种，比如同一个接口，我们既需要效率较高的 thrift，也需要方便 debug 的 http 入口。即除了 CLD 之外，还需要一个单独的 protocol 层，负责处理各种交互协议的细节。这样请求的流程会变成 *图 5-15* 所示。

![control-flow](../images/ch6-08-control-flow.png)

*图 5-15 多协议示意图*

这样我们 Controller 中的入口函数就变成了下面这样：

```go
func CreateOrder(ctx context.Context, req *CreateOrderStruct) (
	*CreateOrderRespStruct, error,
) {
	// ...
}
```

CreateOrder 有两个参数，ctx 用来传入 trace_id 一类的需要串联请求的全局参数，req 里存储了我们创建订单所需要的所有输入信息。返回结果是一个响应结构体和错误。可以认为，我们的代码运行到 Controller 层之后，就没有任何与 “协议” 相关的代码了。在这里你找不到 `http.Request`，也找不到 `http.ResponseWriter`，也找不到任何与 thrift 或者 gRPC 相关的字眼。

在协议 (Protocol) 层，处理 http 协议的大概代码如下：

```go
// defined in protocol layer
type CreateOrderRequest struct {
	OrderID int64 `json:"order_id"`
	// ...
}

// defined in controller
type CreateOrderParams struct {
	OrderID int64
}

func HTTPCreateOrderHandler(wr http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	var params CreateOrderParams
	ctx := context.TODO()
	// bind data to req
	bind(r, &req)
	// map protocol binded to protocol-independent
	map(req, params)
	logicResp,err := controller.CreateOrder(ctx, &params)
	if err != nil {}
	// ...
}
```

理论上我们可以用同一个请求结构体组合上不同的 tag，来达到一个结构体来给不同的协议复用的目的。不过遗憾的是在 thrift 中，请求结构体也是通过 IDL 生成的，其内容在自动生成的 ttypes.go 文件中，我们还是需要在 thrift 的入口将这个自动生成的结构体映射到我们 logic 入口所需要的结构体上。gRPC 也是类似。这部分代码还是需要的。

聪明的读者可能已经可以看出来了，协议细节处理这一层有大量重复劳动，每一个接口在协议这一层的处理，无非是把数据从协议特定的结构体 (例如 `http.Request`，thrift 的被包装过了) 读出来，再绑定到我们协议无关的结构体上，再把这个结构体映射到 Controller 入口的结构体上，这些代码长得都差不多。差不多的代码都遵循着某种模式，那么我们可以对这些模式进行简单的抽象，用代码生成的方式，把繁复的协议处理代码从工作内容中抽离出去。

先来看看 HTTP 对应的结构体、thrift 对应的结构体和我们协议无关的结构体分别长什么样子：

```go
// http 请求结构体
type CreateOrder struct {
	OrderID   int64  `json:"order_id" validate:"required"`
	UserID    int64  `json:"user_id" validate:"required"`
	ProductID int    `json:"prod_id" validate:"required"`
	Addr      string `json:"addr" validate:"required"`
}

// thrift 请求结构体
type FeatureSetParams struct {
	DriverID  int64  `thrift:"driverID,1,required"`
	OrderID   int64  `thrift:"OrderID,2,required"`
	UserID    int64  `thrift:"UserID,3,required"`
	ProductID int    `thrift:"ProductID,4,required"`
	Addr      string `thrift:"Addr,5,required"`
}

// controller input struct
type CreateOrderParams struct {
	OrderID int64
	UserID int64
	ProductID int
	Addr string
}

```

我们需要通过一个源结构体来生成我们需要的 HTTP 和 thrift 入口代码。再观察一下上面定义的三种结构体，我们只要能用一个结构体生成 thrift 的 IDL，以及 HTTP 服务的 “IDL（只要能包含 json 或 form 相关 tag 的结构体定义信息）” 就可以了。这个初始的结构体我们可以把结构体上的 HTTP 的 tag 和 thrift 的 tag 揉在一起：

```go
type FeatureSetParams struct {
	DriverID  int64  `thrift:"driverID,1,required" json:"driver_id"`
	OrderID   int64  `thrift:"OrderID,2,required" json:"order_id"`
	UserID    int64  `thrift:"UserID,3,required" json:"user_id"`
	ProductID int    `thrift:"ProductID,4,required" json:"prod_id"`
	Addr      string `thrift:"Addr,5,required" json:"addr"`
}
```

然后通过代码生成把 thrift 的 IDL 和 HTTP 的请求结构体都生成出来，如 *图 5-16 所示*

![code gen](../images/ch6-08-code-gen.png)

*图 5-16 通过 Go 代码定义结构体生成项目入口*

至于用什么手段来生成，你可以通过 Go 语言内置的 Parser 读取文本文件中的 Go 源代码，然后根据 AST 来生成目标代码，也可以简单地把这个源结构体和 Generator 的代码放在一起编译，让结构体作为 Generator 的输入参数（这样会更简单一些），都是可以的。

当然这种思路并不是唯一选择，我们还可以通过解析 thrift 的 IDL，生成一套 HTTP 接口的结构体。如果你选择这么做，那整个流程就变成了 *图 5-17* 所示。

![code gen](../images/ch6-08-code-gen-2.png)

*图 5-17 也可以从 thrift 生成其它部分*

看起来比之前的图顺畅一点，不过如果你选择了这么做，你需要自行对 thrift 的 IDL 进行解析，也就是相当于可能要手写一个 thrift 的 IDL 的 Parser，虽然现在有 Antlr 或者 peg 能帮你简化这些 Parser 的书写工作，但在 “解析” 的这一步我们不希望引入太多的工作量，所以量力而行即可。

既然工作流已经成型，我们可以琢磨一下怎么让整个流程对用户更加友好。

比如在前面的生成环境引入 Web 页面，只要让用户点点鼠标就能生成 SDK，这些就靠读者自己去探索了。

虽然我们成功地使自己的项目在入口支持了多种交互协议，但是还有一些问题没有解决。本节中所叙述的分层没有将中间件作为项目的分层考虑进去。如果我们考虑中间件的话，请求的流程是什么样的？见 *图 5-18* 所示。

![control flow 2](../images/ch6-08-control-flow-2.png)

*图 5-18 加入中间件后的控制流*

之前我们学习的中间件是和 HTTP 协议强相关的，遗憾的是在 thrift 中看起来没有和 HTTP 中对等的解决这些非功能性逻辑代码重复问题的中间件。所以我们在图上写 `thrift stuff`。这些 `stuff` 可能需要你手写去实现，然后每次增加一个新的 thrift 接口，就需要去写一遍这些非功能性代码。

这也是很多企业项目所面临的真实问题，遗憾的是开源界并没有这样方便的多协议中间件解决方案。当然了，前面我们也说过，很多时候我们给自己保留的 HTTP 接口只是用来做调试，并不会暴露给外人用。这种情况下，这些非功能性的代码只要在 thrift 的代码中完成即可。
