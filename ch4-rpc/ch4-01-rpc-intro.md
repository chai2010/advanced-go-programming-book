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



TODO

<!--

不过上面的例子依然比较简陋：首选是RPC服务只能接受一次请求，其次客户端要通过字符串标识符来区分调用RPC服务不够友好。

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
