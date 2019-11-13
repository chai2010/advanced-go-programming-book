# 4.1 Getting Started with RPC

RPC is short for remote procedure call and is a popular communication method between different nodes in a distributed system. In the Internet age, RPC has become an indispensable infrastructure as well as IPC. Therefore, the standard library of the Go language also provides a simple RPC implementation, and we will use this as an entry to learn the various uses of RPC.

## 4.1.1 RPC version "Hello, World"

The path of the Go language RPC package is net/rpc, which is placed under the net package directory. So we can guess that the RPC package is based on the net package. At the end of the first chapter of the "Hello, World" revolution, we implemented a print example based on http. Below we try to implement a similar example based on rpc.

We first construct a HelloService type, in which the Hello method is used to implement the printing function:

```go
Type HelloService struct {}

Func (p *HelloService) Hello(request string, reply *string) error {
*reply = "hello:" + request
Return nil
}
```

The Hello method must satisfy the RPC rules of the Go language: the method can only have two serializable parameters, the second one is a pointer type, and returns an error type, and must be a public method.

Then you can register the object of type HelloService as an RPC service:

```go
Func main() {
rpc.RegisterName("HelloService", new(HelloService))

Listener, err := net.Listen("tcp", ":1234")
If err != nil {
log.Fatal("ListenTCP error:", err)
}

Conn, err := listener.Accept()
If err != nil {
log.Fatal("Accept error:", err)
}

rpc.ServeConn(conn)
}
```

The rpc.Register function call registers all object methods in the object type that satisfy the RPC rules as RPC functions, and all registered methods are placed under the "HelloService" service space. Then we create a unique TCP link and provide RPC services to the other party over the TCP link via the rpc.ServeConn function.

Here is the code for the client to request the HelloService service:

```go
Func main() {
Client, err := rpc.Dial("tcp", "localhost:1234")
If err != nil {
log.Fatal("dialing:", err)
}

Var reply string
Err = client.Call("HelloService.Hello", "hello", &reply)
If err != nil {
log.Fatal(err)
}

fmt.Println(reply)
}
```

The first is to dial the RPC service through rpc.Dial, and then call the specific RPC method through client.Call. When calling client.Call, the first parameter is the RPC service name and method name linked by a dot, and the second and third parameters respectively define two parameters of the RPC method.

From this example, it can be seen that the use of RPC is actually very simple.

## 4.1.2 Safer RPC Interface

In applications involving RPC, there are generally at least three roles as developers: first, the developer who implements the RPC method on the server side, followed by the person who invokes the RPC method on the client side, and finally, the most important thing is to develop the server and client RPC. The designer of the interface specification. In the previous example, we put all the above roles together in order to simplify, although it seems to be simple to implement, but it is not conducive to the later maintenance and work cutting.

If you want to refactor the HelloService service, the first step is to clarify the name and interface of the service:

```go
Const HelloServiceName = "path/to/pkg.HelloService"

Type HelloServiceInterface = interface {
Hello(request string, reply *string) error
}

Func RegisterHelloService(svc HelloServiceInterface) error {
Return rpc.RegisterName(HelloServiceName, svc)
}
```

We divide the interface specification of the RPC service into three parts: first, the name of the service, then the list of detailed methods to be implemented by the service, and finally the function to register the type of service. In order to avoid name conflicts, we added the package path prefix to the name of the RPC service (this is the package path of the RPC service abstraction, not the package path of the Go language). When RegisterHelloService registers the service, the compiler will ask the incoming object to satisfy the HelloServiceInterface interface.

After defining the RPC service interface specification, the client can write the code for the RPC call according to the specification:

```go
Func main() {
Client, err := rpc.Dial("tcp", "localhost:1234")
If err != nil {
log.Fatal("dialing:", err)
}

Var reply string
Err = client.Call(HelloServiceName+".Hello", "hello", &reply)
If err != nil {
log.Fatal(err)
}
}
```

The only change is that the first parameter of client.Call replaces "HelloService.Hello" with HelloServiceName+".Hello". However, calling the RPC method through the client.Call function is still cumbersome, and the type of the parameter still cannot obtain the security provided by the compiler.

In order to simplify the client user's call to the RPC function, we can add a simple wrapper to the client in the interface specification section:

```go
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

Func (p *HelloServiceClient) Hello(request string, reply *string) error {
Return p.Client.Call(HelloServiceName+".Hello", request, reply)
}
```

We added a new HelloServiceClient type to the client in the interface specification. The type must also satisfy the HelloServiceInterface interface, so that the client user can directly call the RPC function through the corresponding method of the interface. At the same time, a DialHelloService method is provided to directly dial the HelloService service.

Based on the new client interface, we can simplify the code of the client user:

```go
Func main() {
Client, err := DialHelloService("tcp", "localhost:1234")
If err != nil {
log.Fatal("dialing:", err)
}

Var reply string
Err = client.Hello("hello", &reply)
If err != nil {
log.Fatal(err)
}
}
```

Now client users no longer have to worry about low-level errors such as RPC method names or parameter type mismatches.

Finally, the actual server code is written based on the RPC interface specification:

```go
Type HelloService struct {}

Func (p *HelloService) Hello(request string, reply *string) error {
*reply = "hello:" + request
Return nil
}

Func main() {
RegisterHelloService(new(HelloService))

Listener, err := net.Listen("tcp", ":1234")
If err != nil {
log.Fatal("ListenTCP error:", err)
}

For {
Conn, err := listener.Accept()
If err != nil {
log.Fatal("Accept error:", err)
}

Go rpc.ServeConn(conn)
}
}
```

In the new RPC server implementation, we use the RegisterHelloService function to register the function, which not only avoids the work of naming the service name, but also ensures that the incoming service object satisfies the definition of the RPC interface. Finally, our new service instead supports multiple TCP links and then provides RPC services for each TCP link.


## 4.1.3 Cross-language RPC

The standard library RPC defaults to Go-specific gob encoding, so it is more difficult to call Go-based RPC services from other languages. In the era of micro-services in the Internet, each RPC and users of services may use different programming languages, so cross-language is a primary condition for RPC in the Internet age. Thanks to the RPC framework design, Go language RPC is also very easy to achieve cross-language support.

The Go language RPC framework has two more distinctive designs: one is to enable custom encoding and decoding through plug-ins when RPC data is packaged; the other is that RPC is built on the abstract io.ReadWriteCloser interface, we can RPC is built on top of different communication protocols. Here we will try to implement a cross-language RPC through the official native net/rpc/jsonrpc extension.

The first is to re-implement the RPC service based on json encoding:

```go
Func main() {
rpc.RegisterName("HelloService", new(HelloService))

Listener, err := net.Listen("tcp", ":1234")
If err != nil {
log.Fatal("ListenTCP error:", err)
}

For {
Conn, err := listener.Accept()
If err != nil {
log.Fatal("Accept error:", err)
}

Go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
}
}
```

The biggest change in the code is to replace the rpc.ServeConn function with the rpc.ServeCodec function. The argument passed in is the json codec for the server.

Then the client that implements the json version:

```go
Func main() {
Conn, err := net.Dial("tcp", "localhost:1234")
If err != nil {
log.Fatal("net.Dial:", err)
}

Client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))

Var reply string
Err = client.Call("HelloService.Hello", "hello", &reply)
If err != nil {
log.Fatal(err)
}

fmt.Println(reply)
}
```

First call the net.Dial function to establish a TCP link, and then build a json codec for the client based on the link.

After ensuring that the client can call the RPC service normally, we replace the Go language version of the RPC service with a normal TCP service, so that we can view the data format sent by the client. For example, start a TCP service on the same port with the nc command `nc -l 1234`. Then executing an RPC call again will find that nc outputs the following information:

```json
{"method":"HelloService.Hello","params":["hello"],"id":0}
```

This is a json-encoded data, where the method part corresponds to the name of the rpc service and method to be called, the first element of the params part is the parameter, and the id is a unique call number maintained by the caller.

The requested json data object internally corresponds to two structures: the client is clientRequest and the server is serverRequest. The contents of the clientRequest and serverRequest structures are basically the same:

```go
Type clientRequest struct {
Method string `json:"method"`
Params [1]interface{} `json:"params"`
Id uint64 `json:"id"`
}

Type serverRequest struct {
Method string `json:"method"`
Params *json.RawMessage `json:"params"`
Id *json.RawMessage `json:"id"`
}
```

After obtaining the json data corresponding to the RPC call, we can send the json data simulation RPC method call directly to the TCP server that has set up the RPC service:

```
$ echo -e '{"method":"HelloService.Hello","params":["hello"],"id":1}' | nc localhost 1234
```

The returned result is also a json formatted data:

```json
{"id":1,"result":"hello:hello","error":null}
```

Where id corresponds to the input id parameter, result is the returned result, and error part indicates the error message when the problem occurs. For sequential calls, id is not required. However, the RPC framework of the Go language supports asynchronous calls. When the order of the returned results is inconsistent with the order of the calls, the corresponding call can be identified by the id.

The returned json data is also the corresponding two internal structures: the client is clientResponse, and the server is serverResponse. The contents of the two structures are also similar:

```go
Type clientResponse struct {
Id uint64 `json:"id"`
Result *json.RawMessage `json:"result"`
Error interface{} `json:"error"`
}

Type serverResponse struct {
Id *json.RawMessage `json:"id"`
Result interface{} `json:"result"`
Error interface{} `json:"error"`
}
```

So no matter what language you use, just follow the same json structure, you can communicate with the RPC service written by Go in the same process. This way we have implemented cross-language RPC.

## 4.1.4 RPC on Http

The RPC framework inherent in the Go language already supports the provision of RPC services on the Http protocol. However, the framework's http service also uses the built-in gob protocol, and does not provide an interface that uses other protocols, so it is still inaccessible from other languages. In the previous example, we have implemented the jsonrpc service on top of the TCP protocol, and successfully implemented the RPC method call through the nc command line tool. Now we try to provide the jsonrpc service on the http protocol.

The new RPC service is actually a REST-like interface that receives requests and uses the appropriate processing flow:

```go
Func main() {
rpc.RegisterName("HelloService", new(HelloService))

http.HandleFunc("/jsonrpc", func(w http.ResponseWriter, r *http.Request) {
Var conn io.ReadWriteCloser = struct {
io.Writer
io.ReadCloser
}{
ReadCloser: r.Body,
Writer: w,
}

rpc.ServeRequest(jsonrpc.NewServerCodec(conn))
})

http.ListenAndServe(":1234", nil)
}
```

The RPC service is set up in the "/jsonrpc" path, and a conn channel of type io.ReadWriteCloser is constructed in the handler based on parameters of type http.ResponseWriter and http.Request. A json codec for the server is then built based on conn. Finally, the RPC method call is processed once for each request through the rpc.ServeRequest function.

The process of simulating an RPC call is to send a json string to the link:

```
$ curl localhost:1234/jsonrpc -X POST \
--data '{"method":"HelloService.Hello","params":["hello"],"id":0}'
```

The result returned is still a json string:

```json
{"id":0,"result":"hello:hello","error":null}
```

This makes it easy to access RPC services from different languages.