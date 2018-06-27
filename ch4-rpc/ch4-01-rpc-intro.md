# 4.1. RPC入门(Doing)

RPC是远程过程调用的简称，是分布式系统中不同节点间流行的交互方式。在互联网时代，RPC已经是一个不可或缺的基础构件。因此Go语言的标准库提供了一个简单的RPC实现，我们将以此为入口学习RPC的各种用法。

## RPC版"Hello, World"

Go语言的RPC包的路径为net/rpc，也就是放在了net包目录下面。因此我们可以猜测GoRPC包是建立在net包基础之上的。在第一章“Hello, World”革命一节最后，我们基于http实现了一个打印例子。下面我们尝试基于rpc实现一个类似的例子。

我们先构造一个HelloService类型，其中的Hello方法用于实现打印功能：

```go
type HelloService struct {}

func (p *HelloService) Hello(request string, reply *string) error {
	*reply = "hello:" + request
	return nil
}
```

其中Hello方法必须满足Go语言的RPC规则：方法只能有两个可序列化的参数，其中第二个参数是指针类型，并且返回一个error类型。

然后就可以将HelloService类型的对象注册为一个RPC服务：

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

其中rpc.Register会将对象类型中所有满足RPC规则的对象方法注册为RPC函数。然后我们建立一个唯一的TCP链接，并且通过rpc.ServeConn方法在该TCP链接上建立RPC服务。

下面是客户端请求HelloService服务的代码：

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

首选是通过rpc.Dial拨号RPC服务，然后通过client.Call调用具体的RPC方法。在调用client.Call时，第一个参数是用点号链接的RPC服务名字和方法名字，第二和第三个参数分别我们定义RPC方法的两个参数。

由这个例子可以看出RPC的使用其实非常简单。

## 更安全的PRC接口

在涉及RPC的应用中，作为开发人员一般至少有三种角色：首选是服务端实现RPC方法的开发人员，其次是客户端调用RPC方法的人员，最后也是最重要的是制定服务端和客户端RPC接口规范的设计人员。在前面的例子中我们为了简化将以上几种角色的工作全部放到了一起，虽然看似实现简单，但是不利于后期的维护和工作的切割。

如果要重构HelloService服务，第一步需要明确服务的名字和接口：

```go
const HelloServiceName = "path/to/pkg.HelloService"

type HelloServiceInterface = interface {
	Hello(request string, reply *string) error
}

func RegisterHelloService(svc HelloServiceInterface) error {
	rpc.RegisterName(HelloServiceName, svc)
}
```

我们将RPC服务的接口规范分为三个部分：首选是服务的名字，然后是服务要实现的详细方法列表，最后是注册该类型服务的函数。为了避免名字冲突，我们在RPC服务的名字中增加了包路径前缀（这个是RPC服务抽象的包路径，并非完全等价Go语言的包路径）。RegisterHelloService注册服务时，编译器会要求传入的对象满足HelloServiceInterface接口。

在定义了RPC服务接口规范之后，客户端就可以根据规范编写RPC调用的代码了：

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
}
```

其中唯一的变化是client.Call的第一个参数用`HelloServiceName+".Hello"`代理了"HelloService.Hello"。然后通过client.Call函数调用RPC方法依然比较繁琐，同时参数的类型依然无法得到编译器提供的安全保障。

为了简化客户端用户调用RPC函数，我们在可以在接口规范部分增加对客户端的简单包装：

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
	return &HelloServiceClient{Client: c}
}

func (p *HelloServiceClient) Hello(request string, reply *string) error {
	return client.Call(HelloServiceName+".Hello", request, reply)
}
```

我们在接口规范中针对客户端新增加了HelloServiceClient类型，改类型也必须满足HelloServiceInterface接口，这样客户端用户就可以直接通过接口对应的方法调用RPC函数。同时提供了一个DialHelloService方法，直接拨号HelloService服务。

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
}
```

现在客户端用户不用再担心RPC方法名字或参数类型不匹配等低级错误的发生。

最后是基于RPC接口规范编写真实的代码：

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

在新的RPC服务端实现中，我们用RegisterHelloService函数来注册函数，这样不仅可以避免服务名称的工作，同时也保证了传入的服务对象满足了RPC接口定义的定义。最后我们支持多个TCP链接，然后为每个TCP链接建立RPC服务。


## RPC不应该绑定到某个语言

TODO

<!--

不过上面的例子依然比较简陋：首选是RPC服务只能接受一次请求，其次客户端要通过字符串标识符来区分调用RPC服务不够友好。

同时改进，支持多个链接

netrpc简单例子
通过接口给服务端和客户端增加类型约束，缺点是繁琐
http模式，但是只能gob，无法跨语言

简单的例子

http 例子
jsonrpc 例子

jsonrpc on http？标准库不支持

手工 json on http
nodejs 调用（跨语言）

缺点，函数名时字符串，容易出错（可编译）。
手工摸索一个基于接口的规范，手工遵循

同步/异步

名字空间

-->
