
# 4.8 grpcurl工具

Protobuf本身具有反射功能，可以在运行时获取对象的Proto文件。gRPC同样也提供了一个名为reflection的反射包，用于为gRPC服务提供查询。gRPC官方提供了一个C++实现的grpc_cli工具，可以用于查询gRPC列表或调用gRPC方法。但是C++版本的grpc_cli安装比较复杂，我们推荐用纯Go语言实现的grpcurl工具。本节将简要介绍grpcurl工具的用法。

## 4.8.1 启动反射服务

reflection包中只有一个Register函数，用于将grpc.Server注册到反射服务中。reflection包文档给出了简单的使用方法：

```go
import (
	"google.golang.org/grpc/reflection"
)

func main() {
	s := grpc.NewServer()
	pb.RegisterYourOwnServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	s.Serve(lis)
}
```

如果启动了gprc反射服务，那么就可以通过reflection包提供的反射服务查询gRPC服务或调用gRPC方法。

## 4.8.2 查看服务列表

grpcurl是Go语言开源社区开发的工具，需要手工安装：

```
$ go get github.com/fullstorydev/grpcurl
$ go install github.com/fullstorydev/grpcurl/cmd/grpcurl
```

grpcurl中最常使用的是list命令，用于获取服务或服务方法的列表。比如`grpcurl localhost:1234 list`命令将获取本地1234端口上的grpc服务的列表。在使用grpcurl时，需要通过`-cert`和`-key`参数设置公钥和私钥文件，链接启用了tls协议的服务。对于没有没用tls协议的grpc服务，通过`-plaintext`参数忽略tls证书的验证过程。如果是Unix Socket协议，则需要指定`-unix`参数。

如果没有配置好公钥和私钥文件，也没有忽略证书的验证过程，那么将会遇到类似以下的错误：

```shell
$ grpcurl localhost:1234 list
Failed to dial target host "localhost:1234": tls: first record does not \
look like a TLS handshake
```

如果grpc服务正常，但是服务没有启动reflection反射服务，将会遇到以下错误：

```shell
$ grpcurl -plaintext localhost:1234 list
Failed to list services: server does not support the reflection API
```

假设grpc服务已经启动了reflection反射服务，服务的Protobuf文件如下：

```protobuf
syntax = "proto3";

package HelloService;

message String {
	string value = 1;
}

service HelloService {
	rpc Hello (String) returns (String);
	rpc Channel (stream String) returns (stream String);
}
```

grpcurl用list命令查看服务列表时将看到以下输出：

```shell
$ grpcurl -plaintext localhost:1234 list
HelloService.HelloService
grpc.reflection.v1alpha.ServerReflection
```

其中HelloService.HelloService是在protobuf文件定义的服务。而ServerReflection服务则是reflection包注册的反射服务。通过ServerReflection服务可以查询包括本身在内的全部gRPC服务信息。

## 4.8.3 服务的方法列表

继续使用list子命令还可以查看HelloService服务的方法列表：

```shell
$ grpcurl -plaintext localhost:1234 list HelloService.HelloService
Channel
Hello
```

从输出可以看到HelloService服务提供了Channel和Hello两个方法，和Protobuf文件的定义是一致的。

如果还想了解方法的细节，可以使用grpcurl提供的describe子命令查看更详细的描述信息：

```
$ grpcurl -plaintext localhost:1234 describe HelloService.HelloService
HelloService.HelloService is a service:
{
  "name": "HelloService",
  "method": [
    {
      "name": "Hello",
      "inputType": ".HelloService.String",
      "outputType": ".HelloService.String",
      "options": {

      }
    },
    {
      "name": "Channel",
      "inputType": ".HelloService.String",
      "outputType": ".HelloService.String",
      "options": {

      },
      "clientStreaming": true,
      "serverStreaming": true
    }
  ],
  "options": {

  }
}
```

输出列出了服务的每个方法，每个方法输入参数和返回值对应的类型。


## 4.8.4 获取类型信息

在获取到方法的参数和返回值类型之后，还可以继续查看类型的信息。下面是用describe命令查看参数HelloService.String类型的信息：

```shell
$ grpcurl -plaintext localhost:1234 describe HelloService.String
HelloService.String is a message:
{
  "name": "String",
  "field": [
    {
      "name": "value",
      "number": 1,
      "label": "LABEL_OPTIONAL",
      "type": "TYPE_STRING",
      "options": {

      },
      "jsonName": "value"
    }
  ],
  "options": {

  }
}
```

json信息对应HelloService.String类型在Protobuf中的定义如下：

```protobuf
message String {
	string value = 1;
}
```

输出的json数据只不过是Protobuf文件的另一种表示形式。

## 4.8.5 调用方法

在获取gRPC服务的详细信息之后就可以json调用gRPC方法了。

下面命令通过`-d`参数传入一个json字符串作为输入参数，调用的是HelloService服务的Hello方法：

```shell
$ grpcurl -plaintext -d '{"value": "gopher"}' \
	localhost:1234 HelloService.HelloService/Hello
{
  "value": "hello:gopher"
}
```

如果`-d`参数是`@`则表示从标准输入读取json输入参数，这一般用于比较输入复杂的json数据，也可以用于测试流方法。

下面命令是链接Channel流方法，通过从标准输入读取输入流参数：

```shell
$ grpcurl -plaintext -d @ localhost:1234 HelloService.HelloService/Channel
{"value": "gopher"}
{
  "value": "hello:gopher"
}

{"value": "wasm"}
{
  "value": "hello:wasm"
}
```

通过grpcurl工具，我们可以在没有客户端代码的环境下测试gRPC服务。
