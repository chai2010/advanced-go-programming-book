# 4.6 gRPC and Protobuf Extensions

At present, the open source community has developed numerous extensions around Protobuf and gRPC, forming a huge ecosystem. In this section we will briefly introduce the validator and REST interface extensions.

## 4.6.1 Validator

So far, all we have contacted is the third edition of the Protobuf syntax. The second version of Protobuf has a default property that defines default values ​​for members of a string or numeric type.

We created the file using the second version of Protobuf syntax:

```protobuf
Syntax = "proto2";

Package main;

Message message {
Optional string name = 1 [default = "gopher"];
Optional int32 age = 2 [default = 10];
}
```

The built-in default syntax is actually implemented through the extended option feature of Protobuf. The default value feature is no longer supported in the third version of Protobuf, but we can simulate the default value feature ourselves by extending the option.

The following is a rewrite of the above proto file with the extended features of the proto3 syntax:

```protobuf
Syntax = "proto3";

Package main;

Import "google/protobuf/descriptor.proto";

Extend google.protobuf.FieldOptions {
String default_string = 50000;
Int32 default_int = 50001;
}

Message message {
String name = 1 [(default_string) = "gopher"];
Int32 age = 2[(default_int) = 10];
}
```

Inside the square brackets behind the members is the extended syntax. Regenerate the Go language code, which will contain meta information related to the extended options:

```go
Var E_DefaultString = &proto.ExtensionDesc{
ExtendedType: (*descriptor.FieldOptions)(nil),
ExtensionType: (*string)(nil),
Field: 50000,
Name: "main.default_string",
Tag: "bytes,50000,opt,name=default_string,json=defaultString",
Filename: "helloworld.proto",
}

Var E_DefaultInt = &proto.ExtensionDesc{
ExtendedType: (*descriptor.FieldOptions)(nil),
ExtensionType: (*int32)(nil),
Field: 50001,
Name: "main.default_int",
Tag: "varint, 50001, opt, name=default_int, json=defaultInt",
Filename: "helloworld.proto",
}
```

We can parse out the extension options defined by each member of the Message at runtime by a reflection-like technique, and then parse out the default values ​​we define from the associated information for each extension.

In the open source community, github.com/mwitkow/go-proto-validators have implemented more powerful validator functions based on the extended nature of Protobuf. To use this validator you first need to download the code generation plugin it provides:

```
$ go get github.com/mwitkow/go-proto-validators/protoc-gen-govalidators
```

Then add validation rules to the Message member based on the rules of the go-proto-validators validator:

```protobuf
Syntax = "proto3";

Package main;

Import "github.com/mwitkow/go-proto-validators/validator.proto";

Message message {
String important_string = 1 [
(validator.field) = {regex: "^[a-z]{2,5}$"}
];
Int32 age = 2 [
(validator.field) = {int_gt: 0, int_lt: 100}
];
}
```

In the member extension represented by square brackets, validator.field indicates that the extension is a field extension option defined in the validator package. The type of validator.field is the FieldValidator structure, defined in the imported validator.proto file.

All validation rules are defined by the FieldValidator in the validator.proto file:

```protobuf
Syntax = "proto2";
Package validator;

Import "google/protobuf/descriptor.proto";

Extend google.protobuf.FieldOptions {
Optional FieldValidator field = 65020;
}

Message FieldValidator {
// Uses a Golang RE2-syntax regex to match the field contents.
Optional string regex = 1;
// Field value of integer strictly greater than this value.
Optional int64 int_gt = 2;
// Field value of integer strictly smaller than this value.
Optional int64 int_lt = 3;

// ... more ...
}
```

From the comments defined by FieldValidator we can see some syntax for the validator extension: where regex represents a regular expression for string validation, and int_gt and int_lt represent a range of values.

Then use the following command to generate the validation function code:

```
Protoc \
--proto_path=${GOPATH}/src \
--proto_path=${GOPATH}/src/github.com/google/protobuf/src \
--proto_path=. \
--govalidators_out=. --go_out=plugins=grpc:.\
Hello.proto
```

> windows: Replace `${GOPATH}` with `%GOPATH%`.


The above command will call the protoc-gen-govalidators program to generate a separate file named hello.validator.pb.go:

```go
Var _regex_Message_ImportantString = regexp.MustCompile("^[a-z]{2,5}$")

Func (this *Message) Validate() error {
If !_regex_Message_ImportantString.MatchString(this.ImportantString) {
Return go_proto_validators.FieldError("ImportantString", fmt.Errorf(
`value '%v' must be a string conforming to regex "^[a-z]{2,5}$"`,
this.ImportantString,
))
}
If !(this.Age > 0) {
Return go_proto_validators.FieldError("Age", fmt.Errorf(
`value '%v' must be greater than '0'`, this.Age,
))
}
If !(this.Age < 100) {
Return go_proto_validators.FieldError("Age", fmt.Errorf(
`value '%v' must be less than '100'`, this.Age,
))
}
Return nil
}
```

The generated code adds a Validate method to the Message structure to verify that the member satisfies the conditional constraints defined in Protobuf. Regardless of the type, all Validate methods use the same signature, so the same authentication interface can be satisfied.

Through the generated validation function, combined with the gRPC interceptor, we can easily verify the input parameters and return values ​​of each method.

## 4.6.2 REST interface

The gRPC service is generally used for intra-cluster communication. If an external exposed service is required, an equivalent REST interface is generally provided. Convenient front-end JavaScript and back-end interaction through the REST interface. The grpc-gateway project in the open source community implements the ability to turn gRPC services into REST services.

The working principle of grpc-gateway is as follows:

![](../images/ch4-2-grpc-gateway.png)

*Figure 4-2 gRPC-Gateway Workflow*

By adding routing-related meta information to the Protobuf file, routing-related processing code is generated through the custom code plug-in, and finally the REST request is forwarded to the more back-end gRPC service processing.

Routing extension meta information is also provided through Protobuf's metadata extension usage:

```protobuf
Syntax = "proto3";

Package main;

Import "google/api/annotations.proto";

Message StringMessage {
  String value = 1;
}

Service RestService {
Rpc Get(StringMessage) returns (StringMessage) {
Option (google.api.http) = {
Get: "/get/{value}"
};
}
Rpc Post(StringMessage) returns (StringMessage) {
Option (google.api.http) = {
Post: "/post"
Body: "*"
};
}
}
```

We first define the Get and Post methods for gRPC, and then add routing information after the corresponding method through the meta-extension syntax. The "/get/{value}" path corresponds to the Get method, and the `{value}` part corresponds to the value member in the parameter. The result is returned in json format. The Post method corresponds to the "/post" path, and the body contains the request information in the json format.

Then install the protoc-gen-grpc-gateway plugin with the following command:

```
Go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
```

Then generate the routing processing code necessary for grpc-gateway through the plugin:

```
$ protoc -I/usr/local/include -I. \
-I$GOPATH/src \
-I$GOPATH/src/gitHub.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
--grpc-gateway_out=. --go_out=plugins=grpc:.\
Hello.proto
```

> windows: Replace `${GOPATH}` with `%GOPATH%`.

The plugin will generate the corresponding RegisterRestServiceHandlerFromEndpoint function for the RestService service:

```go
Func RegisterRestServiceHandlerFromEndpoint(
Ctx context.Context, mux *runtime.ServeMux, endpoint string,
Opts []grpc.DialOption,
(err error) {
...
}
```

The RegisterRestServiceHandlerFromEndpoint function is used to forward requests that define the Rest interface to the real gRPC service. After registering the route handler, you can start the web service:

```go
Func main() {
Ctx := context.Background()
Ctx, cancel := context.WithCancel(ctx)
Defer cancel()

Mux := runtime.NewServeMux()

Err := RegisterRestServiceHandlerFromEndpoint(
Ctx, mux, "localhost:5000",
[]grpc.DialOption{grpc.WithInsecure()},
)
If err != nil {
log.Fatal(err)
}

http.ListenAndServe(":8080", mux)
}
```

Start grpc service, port 5000
```go
Type RestServiceImpl struct{}

Func (r *RestServiceImpl) Get(ctx context.Context, message *StringMessage) (*StringMessage, error) {
Return &StringMessage{Value: "Get hi:" + message.Value + "#"}, nil
}

Func (r *RestServiceImpl) Post(ctx context.Context, message *StringMessage) (*StringMessage, error) {
Return &StringMessage{Value: "Post hi:" + message.Value + "@"}, nil
}
Func main() {
grpcServer := grpc.NewServer()
RegisterRestServiceServer(grpcServer, new(RestServiceImpl))
Lis, _ := net.Listen("tcp", ":5000")
grpcServer.Serve(lis)
}

```

First, create a route handler through the runtime.NewServeMux() function, and then transfer the REST interface related to the RestService service to the subsequent gRPC service through the RegisterRestServiceHandlerFromEndpoint function. The runtime.ServeMux class provided by grpc-gateway also implements the http.Handler interface, so it can be used with related functions in the standard library.

After the gRPC and REST services are all started, you can request the REST service with curl:

```
$ curl localhost:8080/get/gopher
{"value":"Get: gopher"}

$ curl localhost:8080/post -X POST --data '{"value":"grpc"}'
{"value":"Post: grpc"}
```

When publishing the REST interface, we generally provide a file in Swagger format to describe this interface specification.

```
$ go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

$ protoc -I. \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --swagger_out=. \
  Hello.proto
```

Then a hello.swagger.json file will be generated. In this case, you can use the swagger-ui project to provide REST interface documentation and testing in web pages.

## 4.6.3 Nginx

The latest Nginx provides deep support for gRPC. The backend multiple gRPC services can be aggregated into one Nginx service through Nginx. At the same time, Nginx also provides the ability to register multiple backends for the same gRPC service, which makes it easy to support gRPC load balancing. Nginx's gRPC extension is a larger topic, and interested readers can refer to the relevant documentation.