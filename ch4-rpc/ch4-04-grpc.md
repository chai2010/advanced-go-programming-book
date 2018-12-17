# 4.4 gRPC入门

gRPC是Google公司基于Protobuf开发的跨语言的开源RPC框架。gRPC基于HTTP/2协议设计，可以基于一个HTTP/2链接提供多个服务，对于移动设备更加友好。本节将讲述gRPC的简单用法。

## 4.4.1 gRPC技术栈

Go语言的gRPC技术栈如图4-1所示：

![](../images/ch4-1-grpc-go-stack.png)

*图4-1 gRPC技术栈*

最底层为TCP或Unix Socket协议，在此之上是HTTP/2协议的实现，然后在HTTP/2协议之上又构建了针对Go语言的gRPC核心库。应用程序通过gRPC插件生产的Stub代码和gRPC核心库通信，也可以直接和gRPC核心库通信。

## 4.4.2 gRPC入门

如果从Protobuf的角度看，gRPC只不过是一个针对service接口生成代码的生成器。我们在本章的第二节中手工实现了一个简单的Protobuf代码生成器插件，只不过当时生成的代码是适配标准库的RPC框架的。现在我们将学习gRPC的用法。

创建hello.proto文件，定义HelloService接口：

```proto
syntax = "proto3";

package main;

message String {
	string value = 1;
}

service HelloService {
	rpc Hello (String) returns (String);
}
```

使用protoc-gen-go内置的gRPC插件生成gRPC代码：

```
$ protoc --go_out=plugins=grpc:. hello.proto
```

gRPC插件会为服务端和客户端生成不同的接口：

```go
type HelloServiceServer interface {
	Hello(context.Context, *String) (*String, error)
}

type HelloServiceClient interface {
	Hello(context.Context, *String, ...grpc.CallOption) (*String, error)
}
```

gRPC通过context.Context参数，为每个方法调用提供了上下文支持。客户端在调用方法的时候，可以通过可选的grpc.CallOption类型的参数提供额外的上下文信息。

基于服务端的HelloServiceServer接口可以重新实现HelloService服务：

```go
type HelloServiceImpl struct{}

func (p *HelloServiceImpl) Hello(
	ctx context.Context, args *String,
) (*String, error) {
	reply := &String{Value: "hello:" + args.GetValue()}
	return reply, nil
}
```

gRPC服务的启动流程和标准库的RPC服务启动流程类似：

```go
func main() {
	grpcServer := grpc.NewServer()
	RegisterHelloServiceServer(grpcServer, new(HelloServiceImpl))

	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer.Serve(lis)
}
```

首先是通过`grpc.NewServer()`构造一个gRPC服务对象，然后通过gRPC插件生成的RegisterHelloServiceServer函数注册我们实现的HelloServiceImpl服务。然后通过`grpcServer.Serve(lis)`在一个监听端口上提供gRPC服务。

然后就可以通过客户端链接gRPC服务了：

```go
func main() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := NewHelloServiceClient(conn)
	reply, err := client.Hello(context.Background(), &String{Value: "hello"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply.GetValue())
}
```

其中grpc.Dial负责和gRPC服务建立链接，然后NewHelloServiceClient函数基于已经建立的链接构造HelloServiceClient对象。返回的client其实是一个HelloServiceClient接口对象，通过接口定义的方法就可以调用服务端对应的gRPC服务提供的方法。

gRPC和标准库的RPC框架有一个区别，gRPC生成的接口并不支持异步调用。不过我们可以在多个Goroutine之间安全地共享gRPC底层的HTTP/2链接，因此可以通过在另一个Goroutine阻塞调用的方式模拟异步调用。

## 4.4.3 gRPC流

RPC是远程函数调用，因此每次调用的函数参数和返回值不能太大，否则将严重影响每次调用的响应时间。因此传统的RPC方法调用对于上传和下载较大数据量场景并不适合。同时传统RPC模式也不适用于对时间不确定的订阅和发布模式。为此，gRPC框架针对服务器端和客户端分别提供了流特性。

服务端或客户端的单向流是双向流的特例，我们在HelloService增加一个支持双向流的Channel方法：

```proto
service HelloService {
	rpc Hello (String) returns (String);

	rpc Channel (stream String) returns (stream String);
}
```

关键字stream指定启用流特性，参数部分是接收客户端参数的流，返回值是返回给客户端的流。

重新生成代码可以看到接口中新增加的Channel方法的定义：

```go
type HelloServiceServer interface {
	Hello(context.Context, *String) (*String, error)
	Channel(HelloService_ChannelServer) error
}
type HelloServiceClient interface {
	Hello(ctx context.Context, in *String, opts ...grpc.CallOption) (
		*String, error,
	)
	Channel(ctx context.Context, opts ...grpc.CallOption) (
		HelloService_ChannelClient, error,
	)
}
```

在服务端的Channel方法参数是一个新的HelloService_ChannelServer类型的参数，可以用于和客户端双向通信。客户端的Channel方法返回一个HelloService_ChannelClient类型的返回值，可以用于和服务端进行双向通信。

HelloService_ChannelServer和HelloService_ChannelClient均为接口类型：

```go
type HelloService_ChannelServer interface {
	Send(*String) error
	Recv() (*String, error)
	grpc.ServerStream
}

type HelloService_ChannelClient interface {
	Send(*String) error
	Recv() (*String, error)
	grpc.ClientStream
}
```

可以发现服务端和客户端的流辅助接口均定义了Send和Recv方法用于流数据的双向通信。

现在我们可以实现流服务：

```go
func (p *HelloServiceImpl) Channel(stream HelloService_ChannelServer) error {
	for {
		args, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		reply := &String{Value: "hello:" + args.GetValue()}

		err = stream.Send(reply)
		if err != nil {
			return err
		}
	}
}
```

服务端在循环中接收客户端发来的数据，如果遇到io.EOF表示客户端流被关闭，如果函数退出表示服务端流关闭。生成返回的数据通过流发送给客户端，双向流数据的发送和接收都是完全独立的行为。需要注意的是，发送和接收的操作并不需要一一对应，用户可以根据真实场景进行组织代码。

客户端需要先调用Channel方法获取返回的流对象：

```go
stream, err := client.Channel(context.Background())
if err != nil {
	log.Fatal(err)
}
```

在客户端我们将发送和接收操作放到两个独立的Goroutine。首先是向服务端发送数据：

```go
go func() {
	for {
		if err := stream.Send(&String{Value: "hi"}); err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Second)
	}
}()
```

然后在循环中接收服务端返回的数据：

```go
for {
	reply, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			break
		}
		log.Fatal(err)
	}
	fmt.Println(reply.GetValue())
}
```

这样就完成了完整的流接收和发送支持。

## 4.4.4 发布和订阅模式

在前一节中，我们基于Go内置的RPC库实现了一个简化版的Watch方法。基于Watch的思路虽然也可以构造发布和订阅系统，但是因为RPC缺乏流机制导致每次只能返回一个结果。在发布和订阅模式中，由调用者主动发起的发布行为类似一个普通函数调用，而被动的订阅者则类似gRPC客户端单向流中的接收者。现在我们可以尝试基于gRPC的流特性构造一个发布和订阅系统。

发布订阅是一个常见的设计模式，开源社区中已经存在很多该模式的实现。其中docker项目中提供了一个pubsub的极简实现，下面是基于pubsub包实现的本地发布订阅代码：

```go
import (
	"github.com/docker/docker/pkg/pubsub"
)

func main() {
	p := pubsub.NewPublisher(100*time.Millisecond, 10)

	golang := p.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(string); ok {
			if strings.HasPrefix(key, "golang:") {
				return true
			}
		}
		return false
	})
	docker := p.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(string); ok {
			if strings.HasPrefix(key, "docker:") {
				return true
			}
		}
		return false
	})

	go p.Publish("hi")
	go p.Publish("golang: https://golang.org")
	go p.Publish("docker: https://www.docker.com/")
	time.Sleep(1)

	go func() {
		fmt.Println("golang topic:", <-golang)
	}()
	go func() {
		fmt.Println("docker topic:", <-docker)
	}()

	<-make(chan bool)
}
```

其中`pubsub.NewPublisher`构造一个发布对象，`p.SubscribeTopic()`可以通过函数筛选感兴趣的主题进行订阅。

现在尝试基于gRPC和pubsub包，提供一个跨网络的发布和订阅系统。首先通过Protobuf定义一个发布订阅服务接口：

```protobuf
service PubsubService {
	rpc Publish (String) returns (String);
	rpc Subscribe (String) returns (stream String);
}
```

其中Publish是普通的RPC方法，Subscribe则是一个单向的流服务。然后gRPC插件会为服务端和客户端生成对应的接口：

```go
type PubsubServiceServer interface {
	Publish(context.Context, *String) (*String, error)
	Subscribe(*String, PubsubService_SubscribeServer) error
}
type PubsubServiceClient interface {
	Publish(context.Context, *String, ...grpc.CallOption) (*String, error)
	Subscribe(context.Context, *String, ...grpc.CallOption) (
		PubsubService_SubscribeClient, error,
	)
}

type HelloService_SubscribeServer interface {
	Send(*String) error
	grpc.ServerStream
}
```

因为Subscribe是服务端的单向流，因此生成的HelloService_SubscribeServer接口中只有Send方法。

然后就可以实现发布和订阅服务了：

```go
type PubsubService struct {
	pub *pubsub.Publisher
}

func NewPubsubService() *PubsubService {
	return &PubsubService{
		pub: pubsub.NewPublisher(100*time.Millisecond, 10),
	}
}
```

然后是实现发布方法和订阅方法：

```go
func (p *PubsubService) Publish(
	ctx context.Context, arg *String,
) (*String, error) {
	p.pub.Publish(arg.GetValue())
	return &String{}, nil
}

func (p *PubsubService) Subscribe(
	arg *String, stream PubsubService_SubscribeServer,
) error {
	ch := p.pub.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(string); ok {
			if strings.HasPrefix(key,arg.GetValue()) {
				return true
			}
		}
		return false
	})

	for v := range ch {
		if err := stream.Send(&String{Value: v.(string)}); err != nil {
			return err
		}
	}

	return nil
}
```

这样就可以从客户端向服务器发布信息了：

```go
func main() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := NewPubsubServiceClient(conn)

	_, err = client.Publish(
		context.Background(), &String{Value: "golang: hello Go"},
	)
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Publish(
		context.Background(), &String{Value: "docker: hello Docker"},
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

然后就可以在另一个客户端进行订阅信息了：

```go
func main() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := NewPubsubServiceClient(conn)
	stream, err := client.SubscribeTopic(
		context.Background(), &String{Value: "golang:"},
	)
	if err != nil {
		log.Fatal(err)
	}

	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		fmt.Println(reply.GetValue())
	}
}
```

到此我们就基于gRPC简单实现了一个跨网络的发布和订阅服务。

