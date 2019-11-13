# 4.3 Fun RPC

RPC has different needs in different scenarios, so the open source community has created various RPC frameworks. In this section we will try to use the Go built-in RPC framework in some special scenarios.

## 4.3.1 Principle of Implementing Client RPC

The easiest way to use the Go language RPC library is to use the `Client.Call` method for synchronous blocking calls. The implementation of this method is as follows:

```go
Func (client *Client) Call(
serviceMethod string, args interface{},
Reply interface{},
Error {
Call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
Return call.Error
}
```

First make an asynchronous call through the `Client.Go` method, returning a `Call` structure that represents this call. Then wait for the Done pipe of the `Call` structure to return the result of the call.

We can also call the previous HelloService service asynchronously via the `Client.Go` method:

```go
Func doClientWork(client *rpc.Client) {
helloCall := client.Go("HelloService.Hello", "hello", new(string), nil)

// do some thing

helloCall = <-helloCall.Done
If err := helloCall.Error; err != nil {
log.Fatal(err)
}

Args := helloCall.Args.(string)
Reply := helloCall.Reply.(string)
fmt.Println(args, reply)
}
```

After the asynchronous call command is issued, other tasks are generally executed, so the input parameters and return values ​​of the asynchronous call can be obtained through the returned Call variable.

The `Client.Go` method that performs the asynchronous call is implemented as follows:

```go
Func (client *Client) Go(
serviceMethod string, args interface{},
Reply interface{},
Done chan *Call,
) *Call {
Call := new(Call)
call.ServiceMethod = serviceMethod
call.Args = args
call.Reply = reply
call.Done = make(chan *Call, 10) // buffered.

Client.send(call)
Return call
}
```

The first is to construct a call variable that represents the current call, and then send the complete parameters of the call to the RPC framework via `client.send`. The `client.send` method call is thread-safe, so you can send call instructions from multiple Goroutines to the same RPC link at the same time.

When the call completes or an error occurs, the call to the `call.done` method is called to complete:

```go
Func (call *Call) done() {
Select {
Case call.Done <- call:
// ok
Default:
// We don't want to block here. It is the caller's responsibility to make
// sure the channel has enough buffer space. See comment in Go().
}
}
```

From the implementation of the `Call.done` method, you can see that the `call.Done` pipeline will return the processed call.

## 4.3.2 Implementing the Watch function based on RPC

In many systems, the interface for Watch monitoring is provided. When the system meets certain conditions, the Watch method returns the result of the monitoring. Here we can try to implement a basic Watch function through the RPC framework. As described earlier, because `client.send` is thread-safe, we can also call the RPC method by concurrent blocking in different Goroutines. Monitor by calling the Watch function in a separate Goroutine.

For demonstration purposes, we plan to construct a simple memory KV database via RPC. First define the service as follows:

```go
Type KVStoreService struct {
m map[string]string
Filter map[string]func(key string)
Mu sync.Mutex
}

Func NewKVStoreService() *KVStoreService {
Return &KVStoreService{
m: make(map[string]string),
Filter: make(map[string]func(key string)),
}
}
```

The `m` member is a map type used to store KV data. The `filter` member corresponds to the list of filter functions defined at each call call. The `mu` member is a mutex that is used to protect other members when multiple Goroutines access or modify.

Then there is the Get and Set methods:

```go
Func (p *KVStoreService) Get(key string, value *string) error {
p.mu.Lock()
Defer p.mu.Unlock()

If v, ok := p.m[key]; ok {
*value = v
Return nil
}

Return fmt.Errorf("not found")
}

Func (p *KVStoreService) Set(kv [2]string, reply *struct{}) error {
p.mu.Lock()
Defer p.mu.Unlock()

Key, value := kv[0], kv[1]

If oldValue := p.m[key]; oldValue != value {
For _, fn := range p.filter {
Fn(key)
}
}

P.m[key] = value
Return nil
}
```

In the Set method, the input parameter is an array of key and value, and an anonymous empty structure is used to ignore the output parameters. Each filter function is called when the value corresponding to a key is modified.

The filter list is provided in the Watch method:

```go
Func (p *KVStoreService) Watch(timeoutSecond int, keyChanged *string) error {
Id := fmt.Sprintf("watch-%s-%03d", time.Now(), rand.Int())
Ch := make(chan string, 10) // buffered

p.mu.Lock()
P.filter[id] = func(key string) { ch <- key }
p.mu.Unlock()

Select {
Case <-time.After(time.Duration(timeoutSecond) * time.Second):
Return fmt.Errorf("timeout")
Case key := <-ch:
*keyChanged = key
Return nil
}

Return nil
}
```

The input parameter of the Watch method is the number of seconds to time out. Returns the key as the return value when there is a key change. If no key is modified after the time has elapsed, a timeout error is returned. In the implementation of Watch, each Watch call is represented by a unique id, and then the corresponding filter function is registered to the `p.filter` list according to the id.

The registration and startup process of the KVStoreService service will not be repeated. Let's see how to use the Watch method from the client:

```go
Func doClientWork(client *rpc.Client) {
Go func() {
Var keyChanged string
Err := client.Call("KVStoreService.Watch", 30, &keyChanged)
If err != nil {
log.Fatal(err)
}
fmt.Println("watch:", keyChanged)
} ()

Err := client.Call(
"KVStoreService.Set", [2]string{"abc", "abc-value"},
New(struct{}),
)
If err != nil {
log.Fatal(err)
}

time.Sleep(time.Second*3)
}
```

Start by launching a separate Goroutine monitoring key change. A synchronized watch call will block until a key changes or times out. Then when the KV value is modified by the Set method, the server will return the changed key through the Watch method. This way we can monitor certain states.

## 4.3.3 Reverse RPC

The normal RPC is based on the C/S structure. The server of the RPC corresponds to the server of the network, and the client of the RPC also corresponds to the network client. However, for some special scenarios, such as providing an RPC service on the intranet, but the external network cannot be linked to the intranet server. In this case, we can refer to the technology similar to the reverse proxy. Firstly, we actively link to the TCP server of the external network from the intranet, and then provide the RPC service to the external network based on the TCP link.

Here is the code to start the reverse RPC service:

```go
Func main() {
rpc.Register(new(HelloService))

For {
Conn, _ := net.Dial("tcp", "localhost:1234")
If conn == nil {
time.Sleep(time.Second)
Continue
}

rpc.ServeConn(conn)
conn.Close()
}
}
```

The reverse RPC intranet service will no longer actively provide TCP listening services, but will first actively link to the other party's TCP server. The RPC service is then provided to each other based on each established TCP link.

The RPC client needs to provide a TCP service at a public address to accept the link request from the RPC server:

```go
Func main() {
Listener, err := net.Listen("tcp", ":1234")
If err != nil {
log.Fatal("ListenTCP error:", err)
}

clientChan := make(chan *rpc.Client)

Go func() {
For {
Conn, err := listener.Accept()
If err != nil {
log.Fatal("Accept error:", err)
}

clientChan <- rpc.NewClient(conn)
}
}()

doClientWork(clientChan)
}
```

When each link is established, the RPC client object is constructed based on the network link and sent to the clientChan pipeline.

The client performs the RPC call operation in the doClientWork function:

```go
Func doClientWork(clientChan <-chan *rpc.Client) {
Client := <-clientChan
Defer client.Close()

Var reply string
Err = client.Call("HelloService.Hello", "hello", &reply)
If err != nil {
log.Fatal(err)
}

fmt.Println(reply)
}
```

First, take an RPC client object from the pipeline, and use the defer statement to specify to close the client before the function exits. Then it is to perform a normal RPC call.


## 4.3.4 Context information

Based on the context we can provide customized RPC services for different clients. We can support contextual features by providing separate RPC services for each link.

First modify the HelloService, which adds the conn member of the corresponding link:

```go
Type HelloService struct {
Conn net.Conn
}
```

Then start a separate RPC service for each link:

```go
Func main() {
Listener, err := net.Listen("tcp", ":1234")
If err != nil {
log.Fatal("ListenTCP error:", err)
}

For {
Conn, err := listener.Accept()
If err != nil {
log.Fatal("Accept error:", err)
}

Go func() {
Defer conn.Close()

p := rpc.NewServer()
p.Register(&HelloService{conn: conn})
p.ServeConn(conn)
} ()
}
}
```

In the Hello method, you can identify RPC calls for different links based on the conn member:

```go
Func (p *HelloService) Hello(request string, reply *string) error {
*reply = "hello:" + request + ", from" + p.conn.RemoteAddr().String()
Return nil
}
```

Based on the context information, we can easily add a simple login status verification for the RPC service:

```go
Type HelloService struct {
Conn net.Conn
isLogin bool
}

Func (p *HelloService) Login(request string, reply *string) error {
If request != "user:password" {
Return fmt.Errorf("auth failed")
}
log.Println("login ok")
p.isLogin = true
Return nil
}

Func (p *HelloService) Hello(request string, reply *string) error {
If !p.isLogin {
Return fmt.Errorf("please login")
}
*reply = "hello:" + request + ", from" + p.conn.RemoteAddr().String()
Return nil
}
```

In this way, when the client links the RPC service, the login operation must be performed first, and other services can be executed normally after the login is successful.