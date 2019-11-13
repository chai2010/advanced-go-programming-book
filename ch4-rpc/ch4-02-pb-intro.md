# 4.2 Protobuf

Protobuf is short for Protocol Buffers, a data description language developed by Google and opened to the public in 2008. Protobuf's positioning when it was just open source is similar to XML, JSON and other data description languages. It generates code and provides serialization of structured data through the accompanying tools. But we are more concerned about Protobuf as the description language of the interface specification, which can be used as the basic tool for designing a secure cross-language PRC interface.

## 4.2.1 Getting Started with Protobuf

For readers who have not used Protobuf, it is recommended to understand the basic usage from the official website. Here we try to combine Protobuf and RPC, and finally guarantee the interface specification and security of RPC through Protobuf. The most basic unit of data in Protobuf is the message, which is similar to the structure of the Go language. Members of the message or other underlying data types can be nested in the message.

First create a hello.proto file that wraps the string type used in the HelloService service:

```protobuf
Syntax = "proto3";

Package main;

Message String {
String value = 1;
}
```

The syntax statement at the beginning indicates the syntax of proto3. The third version of Protobuf simplifies the language, and all members are initialized with zero values ​​like Go (no custom defaults are supported), so message members no longer need to support the required attribute. Then the package directive indicates that it is currently the main package (this can be consistent with the Go package name, simplifying the example code), of course, the user can also customize the corresponding package path and name for different languages. Finally, the message keyword defines a new String type, which corresponds to a String structure in the final generated Go language code. There is only one value member of the string type in the String type, and the member is encoded with a number instead of the name.

In a data description language such as XML or JSON, the corresponding data is generally bound by the name of the member. However, Protobuf encoding binds the corresponding data by the unique number of the member, so the volume of the Protobuf encoded data will be small, but it is also very inconvenient for humans to consult. We are not currently concerned with Protobuf's encoding technology. The resulting Go structure can be freely encoded in JSON or gob, so you can temporarily ignore the member encoding part of Protobuf.

The Protobuf core toolset is developed in the C++ language and does not support the Go language in the official protoc compiler. To generate the corresponding Go code based on the hello.proto file above, you need to install the appropriate plugin. The first is to install the official protoc tool, which can be downloaded from https://github.com/google/protobuf/releases. Then install the code generation plugin for Go, which can be installed via the `go get github.com/golang/protobuf/protoc-gen-go` command.

Then generate the corresponding Go code with the following command:

```
$ protoc --go_out=. hello.proto
```

The `go_out` parameter tells the protoc compiler to load the corresponding protoc-gen-go tool, then generate the code through the tool and generate the code into the current directory. Finally, there is a list of a series of protobuf files to process.

Here only a hello.pb.go file is generated, where the String structure is as follows:

```go
Type String struct {
Value string `protobuf:"bytes,1,opt,name=value" json:"value,omitempty"`
}

Func (m *String) Reset() { *m = String{} }
Func (m *String) String() string { return proto.CompactTextString(m) }
Func (*String) ProtoMessage() {}
Func (*String) Descriptor() ([]byte, []int) {
Return fileDescriptor_hello_069698f99dd8f029, []int{0}
}

Func (m *String) GetValue() string {
If m != nil {
Return m.Value
}
Return ""
}
```

The generated structure will also contain some members prefixed with the name `XXX_`, which we have hidden. At the same time, the String type also automatically generates a set of methods, in which the ProtoMessage method indicates that this is a method that implements the proto.Message interface. In addition, Protobuf generates a Get method for each member. The Get method can handle not only the null pointer type, but also the Protobuf version 2 method (the second version of the custom default feature depends on this method).

Based on the new String type, we can reimplement the HelloService service:

```go
Type HelloService struct{}

Func (p *HelloService) Hello(request *String, reply *String) error {
reply.Value = "hello:" + request.GetValue()
Return nil
}
```

The input parameters and output parameters of the Hello method are all represented by the String type defined by Protobuf. Because the new input parameter is a structure type, the pointer type is used as the input parameter, and the internal code of the function is also adjusted accordingly.

So far, we have initially realized the combination of Protobuf and RPC. When starting the RPC service, we can still choose the default gob or manually specify the json code, and even re-implement a plugin based on the protobuf code. Although I have done so much work, it seems that I have not seen any gains!

Looking back at the more secure RPC interface part of Chapter 1, we spent a great deal of effort to add security to RPC services. The resulting code for the more secure RPC interface itself is very cumbersome to use manual maintenance, while all security-related code is only available for the Go language environment! Since the input and output parameters defined by Protobuf are used, can the RPC service interface be defined by Protobuf? Its practical Protobuf defines the language-independent RPC service interface is its real value!

Update the hello.proto file below to define the HelloService service via Protobuf:

```protobuf
Service HelloService {
Rpc Hello (String) returns (String);
}
```

But the regenerated Go code hasn't changed. This is because there are millions of RPC implementations in the world, and the protoc compiler does not know how to generate code for the HelloService service.

However, a plugin named `grpc` has been integrated inside protoc-gen-go to generate code for gRPC:

```
$ protoc --go_out=plugins=grpc:. hello.proto
```

In the generated code, there are some new types like HelloServiceServer and HelloServiceClient. These types are for gRPC and do not meet our RPC requirements.

However, the gRPC plugin provides us with an improved idea. Below we will explore how to generate secure code for our RPC.


## 4.2.2 Custom Code Generation Plugin

Protobuf's protoc compiler implements support for different languages ​​through a plugin mechanism. For example, if the protoc command has the parameter of `--xxx_out` format, then protoc will first query whether there is a built-in xxx plugin. If there is no built-in xxx plugin, it will continue to query whether there is a protoc-gen-xxx named executable in the current system. Finally, the code is generated by the plugin that is queried. For the Protoc-gen-go plugin for Go, there is a layer of static plugin system. For example, protoc-gen-go has a built-in gRPC plugin. Users can generate gRPC related code through the `--go_out=plugins=grpc` parameter. Otherwise, only relevant code will be generated for the message.

Referring to the code of the gRPC plugin, you can find that the generator.RegisterPlugin function can be used to register the plugin. The plugin is a generator.Plugin interface:

```go
// A Plugin provides functionality to add to the output during
// Go code generation, such as to produce RPC stubs.
Type Plugin interface {
// Name identifies the plugin.
Name() string
// Init is called once after data structures are built but before
// code generation begins.
Init(g *Generator)
// Generate produces the code generated by the plugin for this file,
//except for the imports, by calling the generator's methods P, In,
// and Out.
Generate(file *FileDescriptor)
// GenerateImports produces the import declarations for this file.
// It is called after Generate.
GenerateImports(file *FileDescriptor)
}
```

The Name method returns the name of the plugin. This is the plugin system for the Protobuf implementation of the Go language. It has nothing to do with the name of the protoc plugin. Then the Init function initializes the plugin with the g parameter, which contains all the information about the Proto file. The final Generate and GenerateImports methods are used to generate the body code and the corresponding import package code.

So we can design a netrpcPlugin plugin to generate code for the standard library's RPC framework:

```go
Import (
"github.com/golang/protobuf/protoc-gen-go/generator"
)

Type netrpcPlugin struct{ *generator.Generator }

Func (p *netrpcPlugin) Name() string { return "netrpc" }
Func (p *netrpcPlugin) Init(g *generator.Generator) { p.Generator = g }

Func (p *netrpcPlugin) GenerateImports(file *generator.FileDescriptor) {
If len(file.Service) > 0 {
p.genImportCode(file)
}
}

Func (p *netrpcPlugin) Generate(file *generator.FileDescriptor) {
For _, svc := range file.Service {
p.genServiceCode(svc)
}
}
```

First the Name method returns the name of the plugin. The netrpc Plugin has an anonymous `*generator.Generator` member built in, and then initialized with the parameter g when Init is initialized, so the plugin inherits all public methods from the g parameter object. The GenerateImports method calls the custom genImportCode function to generateImport the code. The Generate method calls the custom genServiceCode method to generate the code for each service.

Currently, the custom genImportCode and genServiceCode methods simply output a simple comment:

```go
Func (p *netrpcPlugin) genImportCode(file *generator.FileDescriptor) {
p.P("// TODO: import code")
}

Func (p *netrpcPlugin) genServiceCode(svc *descriptor.ServiceDescriptorProto) {
p.P("// TODO: service code, Name = " + svc.GetName())
}
```

To use the plugin, you need to register the plugin with the generator.RegisterPlugin function, which can be done in the init function:

```go
Func init() {
generator.RegisterPlugin(new(netrpcPlugin))
}
```

Because the Go language package can only be imported statically, we can't add our newly written plugin to the installed protoc-gen-go. We will re-clone the main function corresponding to protoc-gen-go:

```go
Package main

Import (
"io/ioutil"
"os"

"github.com/golang/protobuf/proto"
"github.com/golang/protobuf/protoc-gen-go/generator"
)

Func main() {
g := generator.New()

Data, err := ioutil.ReadAll(os.Stdin)
If err != nil {
g.Error(err, "reading input")
}

If err := proto.Unmarshal(data, g.Request); err != nil {
g.Error(err, "parsing input proto")
}

If len(g.Request.FileToGenerate) == 0 {
g.Fail("no files to generate")
}

g.CommandLineParameters(g.Request.GetParameter())

// Create a wrapped version of the Descriptors and EnumDescriptors that
// point to the file that defines them.
g.WrapTypes()

g.SetPackageNames()
g.BuildTypeNameMap()

g.GenerateAllFiles()

// Send back the results.
Data, err = proto.Marshal(g.Response)
If err != nil {
g.Error(err, "failed to marshal output proto")
}
_, err = os.Stdout.Write(data)
If err != nil {
g.Error(err, "failed to write output proto")
}
}
```

In order to avoid interference with the protoc-gen-go plugin, we named our executable program protoc-gen-go-netrpc, which means that the netrpc plugin is included. Then recompile the hello.proto file with the following command:

```
$ protoc --go-netrpc_out=plugins=netrpc:. hello.proto
```

The `--go-netrpc_out` parameter tells the protoc compiler to load a plugin named protoc-gen-go-netrpc. The `plugins=netrpc` in the plugin indicates that the internal unique netrpcPlugin plugin named netrpc is enabled. The added comment code will be included in the newly generated hello.pb.go file.

At this point, the hand-customized Protobuf code generation plugin is finally working.

## 4.2.3 Automatically generate complete RPC code

In the previous example we have built the minimal netrpcPlugin plugin and created a new plugin for protoc-gen-go-netrpc by cloning the main program of protoc-gen-go. Now continue to improve the netrpcPlugin plugin, the ultimate goal is to generate an RPC security interface.

The first is to generate the code for the import package in the custom genImportCode method:

```go
Func (p *netrpcPlugin) genImportCode(file *generator.FileDescriptor) {
p.P(`import "net/rpc"`)
}
```

Then generate the relevant code for each service in the custom genServiceCode method. Analysis can find that the most important thing for each service is the name of the service, and then each service has a set of methods. For the service definition method, the most important is the name of the method, as well as the names of the input parameters and output parameter types.

For this we define a ServiceSpec type that describes the meta-information of the service:

```go
Type ServiceSpec struct {
ServiceName string
MethodList []ServiceMethodSpec
}

Type ServiceMethodSpec struct {
MethodName string
InputTypeName string
OutputTypeName string
}
```

Then we create a new buildServiceSpec method to parse the ServiceSpec meta information for each service:

```go
Func (p *netrpcPlugin) buildServiceSpec(
Svc *descriptor.ServiceDescriptorProto,
) *ServiceSpec {
Spec := &ServiceSpec{
ServiceName: generator.CamelCase(svc.GetName()),
}

For _, m := range svc.Method {
spec.MethodList = append(spec.MethodList, ServiceMethodSpec{
MethodName: generator.CamelCase(m.GetName()),
InputTypeName: p.TypeName(p.ObjectNamed(m.GetInputType())),
OutputTypeName: p.TypeName(p.ObjectNamed(m.GetOutputType())),
})
}

Return spec
}
```

The input parameter is the `*descriptor.ServiceDescriptorProto` type, which fully describes all the information of a service. Then you can get the name of the service defined in the Protobuf file by `svc.GetName()`. After the name in the Protobuf file is changed to the name of the Go language, a conversion is required through the `generator.CamelCase` function. Similarly, in the for loop we get the name of the method via `m.GetName()` and then the corresponding name in Go. More complicated is the resolution of the input and output parameter names: first you need to get the type of the input parameter through `m.GetInputType()`, then get the class object information corresponding to the type through `p.ObjectNamed`, and finally get the name of the class object. .

Then we can generate the code of the service based on the meta information of the service constructed by the buildServiceSpec method:

```go
Func (p *netrpcPlugin) genServiceCode(svc *descriptor.ServiceDescriptorProto) {
Spec := p.buildServiceSpec(svc)

Var buf bytes.Buffer
t := template.Must(template.New("").Parse(tmplService))
Err := t.Execute(&buf, spec)
If err != nil {
log.Fatal(err)
}

p.P(buf.String())
}
```

For ease of maintenance, we generate service code based on a Go language template, where tmplService is the template for the service.

Before writing a template, let's look at what the final code we expect to generate looks like:

```go
Type HelloServiceInterface interface {
Hello(in String, out *String) error
}

Func RegisterHelloService(srv *rpc.Server, x HelloService) error {
If err := srv.RegisterName("HelloService", x); err != nil {
Return err
}
Return nil
}

Type HelloServiceClient struct {
*rpc.Client
}

Var _ HelloServiceInterface = (*HelloServiceClient)(nil)

Func DialHelloService(network, address string) (*HelloServiceClient, error) {
c, err := rpc.Dial(network, address)
If err != nil {
Return nil, err
}
Return &HelloServiceClient{Client: c}, nil
}

Func (p *HelloServiceClient) Hello(in String, out *String) error {
Return p.Client.Call("HelloService.Hello", in, out)
}
```

The HelloService is the name of the service, along with a series of method-related names.

The following template can be built with reference to the final code to be generated:

```go
Const tmplService = `
{{$root := .}}

Type {{.ServiceName}}Interface interface {
{{- range $_, $m := .MethodList}}
{{$m.MethodName}}(*{{$m.InputTypeName}}, *{{$m.OutputTypeName}}) error
{{- end}}
}

Func Register{{.ServiceName}}(
Srv *rpc.Server, x {{.ServiceName}}Interface,
Error {
If err := srv.RegisterName("{{.ServiceName}}", x); err != nil {
Return err
}
Return nil
}

Type {{.ServiceName}}Client struct {
*rpc.Client
}

Var _ {{.ServiceName}}Interface = (*{{.ServiceName}}Client)(nIl)

Func Dial{{.ServiceName}}(network, address string) (
*{{.ServiceName}}Client, error,
) {
c, err := rpc.Dial(network, address)
If err != nil {
Return nil, err
}
Return &{{.ServiceName}}Client{Client: c}, nil
}

{{range $_, $m := .MethodList}}
Func (p *{{$root.ServiceName}}Client) {{$m.MethodName}}(
In *{{$m.InputTypeName}}, out *{{$m.OutputTypeName}},
Error {
Return p.Client.Call("{{$root.ServiceName}}.{{$m.MethodName}}", in, out)
}
{{end}}
`
```

When Protobuf's plugin customization work is completed, the code can be automatically generated each time the RPC service changes in the hello.proto file. You can also adjust or increase the content of the generated code by updating the template of the plugin. After mastering the custom Protobuf plugin technology, you will have this technology completely.