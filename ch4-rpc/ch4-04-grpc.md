# 4.4. GRPC入门

GRPC是Google公司基于Protobuf开发的跨语言的开源RPC框架。GRPC基于HTTP/2协议设计，可以基于一个HTTP/2链接提供多个服务，对于移动设备更加友好。本节将讲述GRPC的简单用法。

## GRPC入门

如果从Protobuf的角度看，GRPC只不过是针对service接口一个生成代码生成器。我们在本章的第二节中一节手工实现了一个简单的Protobuf代码生成器插件，只不过当时生成的代码是适配标准库的RPC框架的。

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

TODO

## 认证

TODO

<!--

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
