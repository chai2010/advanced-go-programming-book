# 4.8 grpcurl tool

Protobuf itself has a reflection function that gets the Proto file of the object at runtime. gRPC also provides a reflection package called reflection to provide queries for gRPC services. gRPC officially provides a C++ implementation of the grpc_cli tool, which can be used to query the gRPC list or call the gRPC method. But the C++ version of grpc_cli installation is more complicated, we recommend the grpcurl tool implemented in pure Go language. This section will briefly introduce the use of the grpcurl tool.

## 4.8.1 Starting the reflection service

There is only one Register function in the reflection package, which is used to register grpc.Server to the reflection service. The reflection package documentation gives a simple usage:

```go
Import (
"google.golang.org/grpc/reflection"
)

Func main() {
s := grpc.NewServer()
pb.RegisterYourOwnServer(s, &server{})

// Register reflection service on gRPC server.
reflection.Register(s)

s.Serve(lis)
}
```

If the gprc reflection service is enabled, then the gRPC service can be queried or invoked by the reflection service provided by the reflection package.

## 4.8.2 Viewing the list of services

Grpcurl is a tool developed by the Go language open source community and needs to be installed manually:

```
$ go get github.com/fullstorydev/grpcurl
$ go install github.com/fullstorydev/grpcurl/cmd/grpcurl
```

The most commonly used in grpcurl is the list command, which is used to get a list of services or service methods. For example, the `grpcurl localhost:1234 list` command will get a list of grpc services on the local 1234 port. When using grpcurl, you need to set the public and private key files with the `-cert` and `-key` parameters, and link the services that enable the tls protocol. For the grpc service without the tls protocol, the verification process of the tls certificate is ignored by the `-plaintext` parameter. If it is a Unix Socket protocol, you need to specify the `-unix` parameter.

If the public and private key files are not configured and the certificate verification process is not ignored, you will get an error similar to the following:

```shell
$ grpcurl localhost:1234 list
Failed to dial target host "localhost:1234": tls: first record does not \
Look like a TLS handshake
```

If the grpc service is normal, but the service does not start the reflection reflection service, you will encounter the following error:

```shell
$ grpcurl -plaintext localhost:1234 list
Failed to list services: server does not support the reflection API
```

Assuming that the grpc service has started the reflection reflection service, the Protobuf file of the service is as follows:

```protobuf
Syntax = "proto3";

Package HelloService;

Message String {
String value = 1;
}

Service HelloService {
Rpc Hello (String) returns (String);
Rpc Channel (stream String) returns (stream String);
}
```

Grpcurl will see the following output when viewing the list of services with the list command:

```shell
$ grpcurl -plaintext localhost:1234 list
HelloService.HelloService
grpc.reflection.v1alpha.ServerReflection
```

Where HelloService.HelloService is the service defined in the protobuf file. The ServerReflection service is a reflection service registered by the reflection package. Through the ServerReflection service, you can query all gRPC service information including itself.

## 4.8.3 List of methods of service

Continue to use the list subcommand to also view the list of methods for the HelloService service:

```shell
$ grpcurl -plaintext localhost:1234 list HelloService.HelloService
Channel
Hello
```

From the output, you can see that the HelloService service provides two methods, Channel and Hello, which are consistent with the definition of the Protobuf file.

If you want to know the details of the method, you can use the describe subcommand provided by grpcurl to view more detailed descriptions:

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

The output lists each method of the service, each method input parameter and the type corresponding to the return value.


## 4.8.4 Get Type Information

After you get the parameters of the method and the type of the return value, you can continue to view the type information. The following is to use the describe command to view the information of the parameter HelloService.String:

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

The json information corresponding to the HelloService.String type is defined in Protobuf as follows:

```protobuf
Message String {
String value = 1;
}
```

The output of json data is just another representation of a Protobuf file.

## 4.8.5 Calling method

After getting the details of the gRPC service, you can json call the gRPC method.

The following command passes a json string as an input parameter via the `-d` parameter, calling the Hello method of the HelloService service:

```shell
$ grpcurl -plaintext -d '{"value": "gopher"}' \
Localhost:1234 HelloService.HelloService/Hello
{
  "value": "hello:gopher"
}
```

If the `-d` parameter is `@`, it means reading the json input parameter from the standard input. This is generally used to compare the input json data, or it can be used to test the stream method.

The following command is a link to the Channel stream method by reading the input stream parameters from standard input:

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

With the grpcurl tool, we can test the gRPC service in an environment without client code.