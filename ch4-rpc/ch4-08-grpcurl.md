
# 4.8 grpcurl工具

grpc子包中还提供了一个名为reflection的反射包，用于为grpc服务提供查询。reflection包中只有一个Register函数，用于将grpc.Server注册到反射服务中。

reflection包文档给出了简单的使用方法：

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

如果启动了gprc反射服务，那么就可以通过reflection包提供的反射服务查询GRPC服务或调用GRPC方法。

GRPC官方提供了一个C++实现的grpc_cli工具，可以用于查询GRPC列表或调用GRPC方法。不过我们推荐用纯Go语言实现的grpcurl工具，因为grpcurl工具的安装更加简单。

比如程序4.4-2在本地启动了反射服务，就可以用grpcurl来查看GRPC服务信息：

```shell
$ grpcurl -plaintext localhost:1234 list
HelloService.HelloService
grpc.reflection.v1alpha.ServerReflection
```

其中`-plaintext`参数表示跳过TLS证书验证流程，list子命令表示列出所有的服务。从输出可以发现出了我们实现的HelloService服务外，还有一个ServerReflection服务。ServerReflection服务就是reflection包注册的反射服务。

继续使用list子命令还可以查看HelloService服务的方法列表：

```shell
$ grpcurl -plaintext localhost:1234 list HelloService.HelloService
Channel
Hello
```

grpcurl还提供了describe子命令用于描述服务信息：

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

可以查看每个方法输入参数和返回值对应的类型。

describe子命令也可以查看参数HelloService.String类型的信息：

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

在获取GRPC服务的详细信息之后就可以json调用GRPC方法了：

```shell
$ grpcurl -plaintext -d '{"value": "gopher"}' \
	localhost:1234 HelloService.HelloService/Hello
{
  "value": "hello:gopher"
}
```

通过grpcurl工具，我们可以在没有服务端代码的环境下测试GRPC服务。

