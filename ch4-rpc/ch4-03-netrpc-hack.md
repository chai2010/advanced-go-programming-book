# 4.3 玩转 RPC

在不同的场景中 RPC 有着不同的需求，因此开源的社区就诞生了各种 RPC 框架。本节我们将尝试 Go 内置 RPC 框架在一些比较特殊场景的用法。

## 4.3.1 客户端 RPC 的实现原理

Go 语言的 RPC 库最简单的使用方式是通过 `Client.Call` 方法进行同步阻塞调用，该方法的实现如下：

```go
func (client *Client) Call(
	serviceMethod string, args interface{},
	reply interface{},
) error {
	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}
```

首先通过 `Client.Go` 方法进行一次异步调用，返回一个表示这次调用的 `Call` 结构体。然后等待 `Call` 结构体的 Done 管道返回调用结果。

我们也可以通过 `Client.Go` 方法异步调用前面的 HelloService 服务：

```go
func doClientWork(client *rpc.Client) {
	helloCall := client.Go("HelloService.Hello", "hello", new(string), nil)

	// do some thing

	helloCall = <-helloCall.Done
	if err := helloCall.Error; err != nil {
		log.Fatal(err)
	}

	args := helloCall.Args.(string)
	reply := helloCall.Reply.(*string)
	fmt.Println(args, *reply)
}
```

在异步调用命令发出后，一般会执行其他的任务，因此异步调用的输入参数和返回值可以通过返回的 Call 变量进行获取。

执行异步调用的 `Client.Go` 方法实现如下：

```go
func (client *Client) Go(
	serviceMethod string, args interface{},
	reply interface{},
	done chan *Call,
) *Call {
	call := new(Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply
	call.Done = make(chan *Call, 10) // buffered.

	client.send(call)
	return call
}
```

首先是构造一个表示当前调用的 call 变量，然后通过 `client.send` 将 call 的完整参数发送到 RPC 框架。`client.send` 方法调用是线程安全的，因此可以从多个 Goroutine 同时向同一个 RPC 连接发送调用指令。

当调用完成或者发生错误时，将调用 `call.done` 方法通知完成：

```go
func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here. It is the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
	}
}
```

从 `Call.done` 方法的实现可以得知 `call.Done` 管道会将处理后的 call 返回。

## 4.3.2 基于 RPC 实现 Watch 功能

在很多系统中都提供了 Watch 监视功能的接口，当系统满足某种条件时 Watch 方法返回监控的结果。在这里我们可以尝试通过 RPC 框架实现一个基本的 Watch 功能。如前文所描述，因为 `client.send` 是线程安全的，我们也可以通过在不同的 Goroutine 中同时并发阻塞调用 RPC 方法。通过在一个独立的 Goroutine 中调用 Watch 函数进行监控。

为了便于演示，我们计划通过 RPC 构造一个简单的内存 KV 数据库。首先定义服务如下：

```go
type KVStoreService struct {
	m      map[string]string
	filter map[string]func(key string)
	mu     sync.Mutex
}

func NewKVStoreService() *KVStoreService {
	return &KVStoreService{
		m:      make(map[string]string),
		filter: make(map[string]func(key string)),
	}
}
```

其中 `m` 成员是一个 map 类型，用于存储 KV 数据。`filter` 成员对应每个 Watch 调用时定义的过滤器函数列表。而 `mu` 成员为互斥锁，用于在多个 Goroutine 访问或修改时对其它成员提供保护。

然后就是 Get 和 Set 方法：

```go
func (p *KVStoreService) Get(key string, value *string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if v, ok := p.m[key]; ok {
		*value = v
		return nil
	}

	return fmt.Errorf("not found")
}

func (p *KVStoreService) Set(kv [2]string, reply *struct{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	key, value := kv[0], kv[1]

	if oldValue := p.m[key]; oldValue != value {
		for _, fn := range p.filter {
			fn(key)
		}
	}

	p.m[key] = value
	return nil
}
```

在 Set 方法中，输入参数是 key 和 value 组成的数组，用一个匿名的空结构体表示忽略了输出参数。当修改某个 key 对应的值时会调用每一个过滤器函数。

而过滤器列表在 Watch 方法中提供：

```go
func (p *KVStoreService) Watch(timeoutSecond int, keyChanged *string) error {
	id := fmt.Sprintf("watch-%s-%03d", time.Now(), rand.Int())
	ch := make(chan string, 10) // buffered

	p.mu.Lock()
	p.filter[id] = func(key string) { ch <- key }
	p.mu.Unlock()

	select {
	case <-time.After(time.Duration(timeoutSecond) * time.Second):
		return fmt.Errorf("timeout")
	case key := <-ch:
		*keyChanged = key
		return nil
	}

	return nil
}
```

Watch 方法的输入参数是超时的秒数。当有 key 变化时将 key 作为返回值返回。如果超过时间后依然没有 key 被修改，则返回超时的错误。Watch 的实现中，用唯一的 id 表示每个 Watch 调用，然后根据 id 将自身对应的过滤器函数注册到 `p.filter` 列表。

KVStoreService 服务的注册和启动过程我们不再赘述。下面我们看看如何从客户端使用 Watch 方法：

```go
func doClientWork(client *rpc.Client) {
	go func() {
		var keyChanged string
		err := client.Call("KVStoreService.Watch", 30, &keyChanged)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("watch:", keyChanged)
	} ()

	err := client.Call(
		"KVStoreService.Set", [2]string{"abc", "abc-value"},
		new(struct{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second*3)
}
```

首先启动一个独立的 Goroutine 监控 key 的变化。同步的 watch 调用会阻塞，直到有 key 发生变化或者超时。然后在通过 Set 方法修改 KV 值时，服务器会将变化的 key 通过 Watch 方法返回。这样我们就可以实现对某些状态的监控。

## 4.3.3 反向 RPC

通常的 RPC 是基于 C/S 结构，RPC 的服务端对应网络的服务器，RPC 的客户端也对应网络客户端。但是对于一些特殊场景，比如在公司内网提供一个 RPC 服务，但是在外网无法连接到内网的服务器。这种时候我们可以参考类似反向代理的技术，首先从内网主动连接到外网的 TCP 服务器，然后基于 TCP 连接向外网提供 RPC 服务。

以下是启动反向 RPC 服务的代码：

```go
func main() {
	rpc.Register(new(HelloService))

	for {
		conn, _ := net.Dial("tcp", "localhost:1234")
		if conn == nil {
			time.Sleep(time.Second)
			continue
		}

		rpc.ServeConn(conn)
		conn.Close()
	}
}
```

反向 RPC 的内网服务将不再主动提供 TCP 监听服务，而是首先主动连接到对方的 TCP 服务器。然后基于每个建立的 TCP 连接向对方提供 RPC 服务。

而 RPC 客户端则需要在一个公共的地址提供一个 TCP 服务，用于接受 RPC 服务器的连接请求：

```go
func main() {
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}

	clientChan := make(chan *rpc.Client)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatal("Accept error:", err)
			}

			clientChan <- rpc.NewClient(conn)
		}
	}()

	doClientWork(clientChan)
}
```

当每个连接建立后，基于网络连接构造 RPC 客户端对象并发送到 clientChan 管道。

客户端执行 RPC 调用的操作在 doClientWork 函数完成：

```go
func doClientWork(clientChan <-chan *rpc.Client) {
	client := <-clientChan
	defer client.Close()

	var reply string
	err := client.Call("HelloService.Hello", "hello", &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(reply)
}
```

首先从管道去取一个 RPC 客户端对象，并且通过 defer 语句指定在函数退出前关闭客户端。然后是执行正常的 RPC 调用。


## 4.3.4 上下文信息

基于上下文我们可以针对不同客户端提供定制化的 RPC 服务。我们可以通过为每个连接提供独立的 RPC 服务来实现对上下文特性的支持。

首先改造 HelloService，里面增加了对应连接的 conn 成员：

```go
type HelloService struct {
	conn net.Conn
}
```

然后为每个连接启动独立的 RPC 服务：

```go
func main() {
	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}

		go func() {
			defer conn.Close()

			p := rpc.NewServer()
			p.Register(&HelloService{conn: conn})
			p.ServeConn(conn)
		} ()
	}
}
```

Hello 方法中就可以根据 conn 成员识别不同连接的 RPC 调用：

```go
func (p *HelloService) Hello(request string, reply *string) error {
	*reply = "hello:" + request + ", from" + p.conn.RemoteAddr().String()
	return nil
}
```

基于上下文信息，我们可以方便地为 RPC 服务增加简单的登陆状态的验证：

```go
type HelloService struct {
	conn    net.Conn
	isLogin bool
}

func (p *HelloService) Login(request string, reply *string) error {
	if request != "user:password" {
		return fmt.Errorf("auth failed")
	}
	log.Println("login ok")
	p.isLogin = true
	return nil
}

func (p *HelloService) Hello(request string, reply *string) error {
	if !p.isLogin {
		return fmt.Errorf("please login")
	}
	*reply = "hello:" + request + ", from" + p.conn.RemoteAddr().String()
	return nil
}
```

这样可以要求在客户端连接 RPC 服务时，首先要执行登陆操作，登陆成功后才能正常执行其他的服务。
