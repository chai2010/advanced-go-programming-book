# 4.4. GRPC入门

GRPC是Google公司基于Protobuf开发的跨语言的开源RPC框架。GRPC基于HTTP/2协议设计，可以基于一个HTTP/2链接提供多个服务，对于移动设备更加友好。本节将讲述GRPC的简单用法。

## GRPC入门

如果从Protobuf的角度看，GRPC只不过是一个针对service接口生成代码的生成器。我们在本章的第二节中手工实现了一个简单的Protobuf代码生成器插件，只不过当时生成的代码是适配标准库的RPC框架的。

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

使用protoc-gen-go内置的grpc插件生成GRPC代码：

```
$ protoc --go_out=plugins=grpc:. hello.proto
```

GRPC插件会为服务端和客户端生成不同的接口：

```go
type HelloServiceServer interface {
	Hello(context.Context, *String) (*String, error)
}

type HelloServiceClient interface {
	Hello(ctx context.Context, in *String, opts ...grpc.CallOption) (*String, error)
}
```

GRPC通过context.Context参数，为每个方法调用提供了上下文支持。客户端在调用方法的时候，可以通过可选的grpc.CallOption类型的参数提供额外的上下文信息。

基于服务端的HelloServiceServer接口可以重新实现HelloService服务：

```go
type HelloServiceImpl struct{}

func (p *HelloServiceImpl) Hello(ctx context.Context, args *String) (*String, error) {
	reply := &String{Value: "hello:" + args.GetValue()}
	return reply, nil
}
```

GRPC服务的启动流程和标准库的RPC服务启动流程类似：

```go
func main() {
	grpcServer := grpc.NewServer()
	RegisterHelloServiceServer(grpcServer, &HelloServiceImpl{})

	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}
	grpcServer.Serve(lis)
}
```

首先是通过`grpc.NewServer()`构造一个GRPC服务对象，然后通过GRPC插件生成的RegisterHelloServiceServer函数注册我们实现的HelloServiceImpl服务。然后通过`grpcServer.Serve(lis)`在一个监听端口上提供GRPC服务。

然后就可以通过客户端链接GRPC服务了：

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

其中grpc.Dial负责和GRPC服务建立链接，然后NewHelloServiceClient函数基于已经建立的链接构造HelloServiceClient对象。返回的client其实是一个HelloServiceClient接口对象，通过接口定义的方法就可以调用服务端对应的GRPC服务提供的方法。

GRPC和标准库的RPC框架还有一个区别，GRPC生成的接口并不支持异步调用。

## GRPC流

RPC是远程函数调用，因此每次调用的函数参数和返回值不能太大，负责将严重影响每次调用的性能。因此传统的RPC方法调用对于上传和下载较大数据量场景并不适合。同时传统RPC模式也不适用于对于时间不确定的订阅和发布模式。为此，GRPC框架分别提供了服务器端和客户端的流特性。

服务端或客户端的单向流是双向流的特例，我们在HelloService增加一个支持双向流的Channel方法：

```proto
service HelloService {
	rpc Hello (String) returns (String);

	rpc Channel (stream String) returns (stream String);
}
```

关键字stream指定启用流特性，参数部分是接收客户端参数的流，返回值是返回给客户端的流。

重新生成代码可以可以看到接口中新增加的Channel方法的定义：

```go
type HelloServiceServer interface {
	Hello(context.Context, *String) (*String, error)
	Channel(HelloService_ChannelServer) error
}
type HelloServiceClient interface {
	Hello(ctx context.Context, in *String, opts ...grpc.CallOption) (*String, error)
	Channel(ctx context.Context, opts ...grpc.CallOption) (HelloService_ChannelClient, error)
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

服务端在循环中接收客户端发来的数据，如果遇到io.EOF表示客户端流被关闭，如果函数退出表示服务端流关闭。然后生成返回的数据通过流发送给客户端。需要主要的是，发送和接收的操作并不需要一一对应，用户可以根据真实场景进行组织代码。

客户端需要先调用Channel方法获取返回的流对象：

```go
stream, err := client.Channel(context.Background())
if err != nil {
	log.Fatal(err)
}
```

在客户端我们将发送和接收操作放到两个独立的Goroutine。首先是向服务端发生数据：

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

这样就完成了完整的流接收和发生支持。


<!--
Publish
Watch

TODO

## 认证

TODO



入门/流/认证

--

简单介绍

同步/异步
流

验证/密码

日志截取器，panic 捕获

gtpc到rest扩展

参数的自动验证，在截取器进行
-->
