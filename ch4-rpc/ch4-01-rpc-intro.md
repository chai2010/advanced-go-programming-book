# 4.1 RPC 入门

RPC 是远程过程调用的简称，是分布式系统中不同节点间流行的通信方式。在互联网时代，RPC 已经和 IPC 一样成为一个不可或缺的基础构件。因此 Go 语言的标准库也提供了一个简单的 RPC 实现，我们将以此为入口学习 RPC 的各种用法。

## 4.1.1 RPC 版 “Hello, World”

Go 语言的 RPC 包的路径为 net/rpc，也就是放在了 net 包目录下面。因此我们可以猜测该 RPC 包是建立在 net 包基础之上的。在第一章 “Hello, World” 革命一节最后，我们基于 http 实现了一个打印例子。下面我们尝试基于 rpc 实现一个类似的例子。

我们先构造一个 HelloService 类型，其中的 Hello 方法用于实现打印功能：

```go
type HelloService struct {}

func (p *HelloService) Hello(request string, reply *string) error {
	*reply = "hello:" + request
	return nil
}
```

其中 Hello 方法必须满足 Go 语言的 RPC 规则：方法只能有两个可序列化的参数，其中第二个参数是指针类型，并且返回一个 error 类型，同时必须是公开的方法。

然后就可以将 HelloService 类型的对象注册为一个 RPC 服务：

```go
func main() {
	rpc.RegisterName("HelloService", new(HelloService))

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}

	conn, err := listener.Accept()
	if err != nil {
		log.Fatal("Accept error:", err)
	}

	rpc.ServeConn(conn)
}
```

其中 rpc.Register 函数调用会将对象类型中所有满足 RPC 规则的对象方法注册为 RPC 函数，所有注册的方法会放在 “HelloService” 服务空间之下。然后我们建立一个唯一的 TCP 连接，并且通过 rpc.ServeConn 函数在该 TCP 连接上为对方提供 RPC 服务。

下面是客户端请求 HelloService 服务的代码：

```go
func main() {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply string
	err = client.Call("HelloService.Hello", "hello", &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(reply)
}
```

首先是通过 rpc.Dial 拨号 RPC 服务，然后通过 client.Call 调用具体的 RPC 方法。在调用 client.Call 时，第一个参数是用点号连接的 RPC 服务名字和方法名字，第二和第三个参数分别我们定义 RPC 方法的两个参数。

由这个例子可以看出 RPC 的使用其实非常简单。

## 4.1.2 更安全的 RPC 接口

在涉及 RPC 的应用中，作为开发人员一般至少有三种角色：首先是服务端实现 RPC 方法的开发人员，其次是客户端调用 RPC 方法的人员，最后也是最重要的是制定服务端和客户端 RPC 接口规范的设计人员。在前面的例子中我们为了简化将以上几种角色的工作全部放到了一起，虽然看似实现简单，但是不利于后期的维护和工作的切割。

如果要重构 HelloService 服务，第一步需要明确服务的名字和接口：

```go
const HelloServiceName = "path/to/pkg.HelloService"

type HelloServiceInterface interface {
	Hello(request string, reply *string) error
}

func RegisterHelloService(svc HelloServiceInterface) error {
	return rpc.RegisterName(HelloServiceName, svc)
}
```

我们将 RPC 服务的接口规范分为三个部分：首先是服务的名字，然后是服务要实现的详细方法列表，最后是注册该类型服务的函数。为了避免名字冲突，我们在 RPC 服务的名字中增加了包路径前缀（这个是 RPC 服务抽象的包路径，并非完全等价 Go 语言的包路径）。RegisterHelloService 注册服务时，编译器会要求传入的对象满足 HelloServiceInterface 接口。

在定义了 RPC 服务接口规范之后，客户端就可以根据规范编写 RPC 调用的代码了：

```go
func main() {
	client, err := rpc.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply string
	err = client.Call(HelloServiceName+".Hello", "hello", &reply)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}
```

其中唯一的变化是 client.Call 的第一个参数用 HelloServiceName+".Hello" 代替了 "HelloService.Hello"。然而通过 client.Call 函数调用 RPC 方法依然比较繁琐，同时参数的类型依然无法得到编译器提供的安全保障。

为了简化客户端用户调用 RPC 函数，我们在可以在接口规范部分增加对客户端的简单包装：

```go
type HelloServiceClient struct {
	*rpc.Client
}

var _ HelloServiceInterface = (*HelloServiceClient)(nil)

func DialHelloService(network, address string) (*HelloServiceClient, error) {
	c, err := rpc.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &HelloServiceClient{Client: c}, nil
}

func (p *HelloServiceClient) Hello(request string, reply *string) error {
	return p.Client.Call(HelloServiceName+".Hello", request, reply)
}
```

我们在接口规范中针对客户端新增加了 HelloServiceClient 类型，该类型也必须满足 HelloServiceInterface 接口，这样客户端用户就可以直接通过接口对应的方法调用 RPC 函数。同时提供了一个 DialHelloService 方法，直接拨号 HelloService 服务。

基于新的客户端接口，我们可以简化客户端用户的代码：

```go
func main() {
	client, err := DialHelloService("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	var reply string
	err = client.Hello("hello", &reply)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}
```

现在客户端用户不用再担心 RPC 方法名字或参数类型不匹配等低级错误的发生。

最后是基于 RPC 接口规范编写真实的服务端代码：

```go
type HelloService struct {}

func (p *HelloService) Hello(request string, reply *string) error {
	*reply = "hello:" + request
	return nil
}

func main() {
	RegisterHelloService(new(HelloService))

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}

		go rpc.ServeConn(conn)
	}
}
```

在新的 RPC 服务端实现中，我们用 RegisterHelloService 函数来注册函数，这样不仅可以避免命名服务名称的工作，同时也保证了传入的服务对象满足了 RPC 接口的定义。最后我们新的服务改为支持多个 TCP 连接，然后为每个 TCP 连接提供 RPC 服务。


## 4.1.3 跨语言的 RPC

标准库的 RPC 默认采用 Go 语言特有的 gob 编码，因此从其它语言调用 Go 语言实现的 RPC 服务将比较困难。在互联网的微服务时代，每个 RPC 以及服务的使用者都可能采用不同的编程语言，因此跨语言是互联网时代 RPC 的一个首要条件。得益于 RPC 的框架设计，Go 语言的 RPC 其实也是很容易实现跨语言支持的。

Go 语言的 RPC 框架有两个比较有特色的设计：一个是 RPC 数据打包时可以通过插件实现自定义的编码和解码；另一个是 RPC 建立在抽象的 io.ReadWriteCloser 接口之上的，我们可以将 RPC 架设在不同的通讯协议之上。这里我们将尝试通过官方自带的 net/rpc/jsonrpc 扩展实现一个跨语言的 RPC。

首先是基于 json 编码重新实现 RPC 服务：

```go
func main() {
	rpc.RegisterName("HelloService", new(HelloService))

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}

		go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
```

代码中最大的变化是用 rpc.ServeCodec 函数替代了 rpc.ServeConn 函数，传入的参数是针对服务端的 json 编解码器。

然后是实现 json 版本的客户端：

```go
func main() {
	conn, err := net.Dial("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("net.Dial:", err)
	}

	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))

	var reply string
	err = client.Call("HelloService.Hello", "hello", &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(reply)
}
```

先手工调用 net.Dial 函数建立 TCP 连接，然后基于该连接建立针对客户端的 json 编解码器。

在确保客户端可以正常调用 RPC 服务的方法之后，我们用一个普通的 TCP 服务代替 Go 语言版本的 RPC 服务，这样可以查看客户端调用时发送的数据格式。比如通过 nc 命令 `nc -l 1234` 在同样的端口启动一个 TCP 服务。然后再次执行一次 RPC 调用将会发现 nc 输出了以下的信息：

```json
{"method":"HelloService.Hello","params":["hello"],"id":0}
```

这是一个 json 编码的数据，其中 method 部分对应要调用的 rpc 服务和方法组合成的名字，params 部分的第一个元素为参数，id 是由调用端维护的一个唯一的调用编号。

请求的 json 数据对象在内部对应两个结构体：客户端是 clientRequest，服务端是 serverRequest。clientRequest 和 serverRequest 结构体的内容基本是一致的：

```go
type clientRequest struct {
	Method string         `json:"method"`
	Params [1]interface{} `json:"params"`
	Id     uint64         `json:"id"`
}

type serverRequest struct {
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
	Id     *json.RawMessage `json:"id"`
}
```

在获取到 RPC 调用对应的 json 数据后，我们可以通过直接向架设了 RPC 服务的 TCP 服务器发送 json 数据模拟 RPC 方法调用：

```
$ echo -e '{"method":"HelloService.Hello","params":["hello"],"id":1}' | nc localhost 1234
```

返回的结果也是一个 json 格式的数据：

```json
{"id":1,"result":"hello:hello","error":null}
```

其中 id 对应输入的 id 参数，result 为返回的结果，error 部分在出问题时表示错误信息。对于顺序调用来说，id 不是必须的。但是 Go 语言的 RPC 框架支持异步调用，当返回结果的顺序和调用的顺序不一致时，可以通过 id 来识别对应的调用。

返回的 json 数据也是对应内部的两个结构体：客户端是 clientResponse，服务端是 serverResponse。两个结构体的内容同样也是类似的：

```go
type clientResponse struct {
	Id     uint64           `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  interface{}      `json:"error"`
}

type serverResponse struct {
	Id     *json.RawMessage `json:"id"`
	Result interface{}      `json:"result"`
	Error  interface{}      `json:"error"`
}
```

因此无论采用何种语言，只要遵循同样的 json 结构，以同样的流程就可以和 Go 语言编写的 RPC 服务进行通信。这样我们就实现了跨语言的 RPC。

## 4.1.4 Http 上的 RPC

Go 语言内在的 RPC 框架已经支持在 Http 协议上提供 RPC 服务。但是框架的 http 服务同样采用了内置的 gob 协议，并且没有提供采用其它协议的接口，因此从其它语言依然无法访问的。在前面的例子中，我们已经实现了在 TCP 协议之上运行 jsonrpc 服务，并且通过 nc 命令行工具成功实现了 RPC 方法调用。现在我们尝试在 http 协议上提供 jsonrpc 服务。

新的 RPC 服务其实是一个类似 REST 规范的接口，接收请求并采用相应处理流程：

```go
func main() {
	rpc.RegisterName("HelloService", new(HelloService))

	http.HandleFunc("/jsonrpc", func(w http.ResponseWriter, r *http.Request) {
		var conn io.ReadWriteCloser = struct {
			io.Writer
			io.ReadCloser
		}{
			ReadCloser: r.Body,
			Writer:     w,
		}

		rpc.ServeRequest(jsonrpc.NewServerCodec(conn))
	})

	http.ListenAndServe(":1234", nil)
}
```

RPC 的服务架设在 “/jsonrpc” 路径，在处理函数中基于 http.ResponseWriter 和 http.Request 类型的参数构造一个 io.ReadWriteCloser 类型的 conn 通道。然后基于 conn 构建针对服务端的 json 编码解码器。最后通过 rpc.ServeRequest 函数为每次请求处理一次 RPC 方法调用。

模拟一次 RPC 调用的过程就是向该连接发送一个 json 字符串：

```
$ curl localhost:1234/jsonrpc -X POST \
	--data '{"method":"HelloService.Hello","params":["hello"],"id":0}'
```

返回的结果依然是 json 字符串：

```json
{"id":0,"result":"hello:hello","error":null}
```

这样就可以很方便地从不同语言中访问 RPC 服务了。

