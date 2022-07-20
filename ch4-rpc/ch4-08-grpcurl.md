
# 4.8 grpcurl 工具

Protobuf 本身具有反射功能，可以在运行时获取对象的 Proto 文件。gRPC 同样也提供了一个名为 reflection 的反射包，用于为 gRPC 服务提供查询。gRPC 官方提供了一个 C++ 实现的 grpc_cli 工具，可以用于查询 gRPC 列表或调用 gRPC 方法。但是 C++ 版本的 grpc_cli 安装比较复杂，我们推荐用纯 Go 语言实现的 grpcurl 工具。本节将简要介绍 grpcurl 工具的用法。

## 4.8.1 启动反射服务

reflection 包中只有一个 Register 函数，用于将 grpc.Server 注册到反射服务中。reflection 包文档给出了简单的使用方法：

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

如果启动了 gprc 反射服务，那么就可以通过 reflection 包提供的反射服务查询 gRPC 服务或调用 gRPC 方法。

## 4.8.2 查看服务列表

grpcurl 是 Go 语言开源社区开发的工具，需要手工安装：

```
$ go get github.com/fullstorydev/grpcurl
$ go install github.com/fullstorydev/grpcurl/cmd/grpcurl
```

grpcurl 中最常使用的是 list 命令，用于获取服务或服务方法的列表。比如 `grpcurl localhost:1234 list` 命令将获取本地 1234 端口上的 grpc 服务的列表。在使用 grpcurl 时，需要通过 `-cert` 和 `-key` 参数设置公钥和私钥文件，连接启用了 tls 协议的服务。对于没有没用 tls 协议的 grpc 服务，通过 `-plaintext` 参数忽略 tls 证书的验证过程。如果是 Unix Socket 协议，则需要指定 `-unix` 参数。

如果没有配置好公钥和私钥文件，也没有忽略证书的验证过程，那么将会遇到类似以下的错误：

```shell
$ grpcurl localhost:1234 list
Failed to dial target host "localhost:1234": tls: first record does not \
look like a TLS handshake
```

如果 grpc 服务正常，但是服务没有启动 reflection 反射服务，将会遇到以下错误：

```shell
$ grpcurl -plaintext localhost:1234 list
Failed to list services: server does not support the reflection API
```

假设 grpc 服务已经启动了 reflection 反射服务，服务的 Protobuf 文件如下：

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

grpcurl 用 list 命令查看服务列表时将看到以下输出：

```shell
$ grpcurl -plaintext localhost:1234 list
HelloService.HelloService
grpc.reflection.v1alpha.ServerReflection
```

其中 HelloService.HelloService 是在 protobuf 文件定义的服务。而 ServerReflection 服务则是 reflection 包注册的反射服务。通过 ServerReflection 服务可以查询包括本身在内的全部 gRPC 服务信息。

## 4.8.3 服务的方法列表

继续使用 list 子命令还可以查看 HelloService 服务的方法列表：

```shell
$ grpcurl -plaintext localhost:1234 list HelloService.HelloService
Channel
Hello
```

从输出可以看到 HelloService 服务提供了 Channel 和 Hello 两个方法，和 Protobuf 文件的定义是一致的。

如果还想了解方法的细节，可以使用 grpcurl 提供的 describe 子命令查看更详细的描述信息：

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

在获取到方法的参数和返回值类型之后，还可以继续查看类型的信息。下面是用 describe 命令查看参数 HelloService.String 类型的信息：

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

json 信息对应 HelloService.String 类型在 Protobuf 中的定义如下：

```protobuf
message String {
	string value = 1;
}
```

输出的 json 数据只不过是 Protobuf 文件的另一种表示形式。

## 4.8.5 调用方法

在获取 gRPC 服务的详细信息之后就可以 json 调用 gRPC 方法了。

下面命令通过 `-d` 参数传入一个 json 字符串作为输入参数，调用的是 HelloService 服务的 Hello 方法：

```shell
$ grpcurl -plaintext -d '{"value":"gopher"}' \
	localhost:1234 HelloService.HelloService/Hello
{
  "value": "hello:gopher"
}
```

如果 `-d` 参数是 `@` 则表示从标准输入读取 json 输入参数，这一般用于比较输入复杂的 json 数据，也可以用于测试流方法。

下面命令是连接 Channel 流方法，通过从标准输入读取输入流参数：

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

通过 grpcurl 工具，我们可以在没有客户端代码的环境下测试 gRPC 服务。
