# 4.3. 玩转RPC

在不同的场景中RPC有着不同的需求，因此开源的社区就诞生了各种RPC框架。本节我们将尝试Go内置RPC框架在一些比较特殊场景的用法。

## 反向RPC

通常的RPC是基于C/S结构，RPC的服务端对应网络的服务器，RPC的客户端也对应网络客户端。但是对于一些特殊场景，比如在公司内网提供一个RPC服务，但是在外网无法链接到内网的服务器。这种时候我们可以参考类似反向代理的技术，首先从内网主动链接到外网的TCP服务器，然后基于TCP链接向外网提供RPC服务。

以下是启动反向RPC服务的代码：

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

反向RPC的内网服务将不再主导提供TCP监听服务，而是首先主动链接到对方的TCP服务器。然后基于每个建立的TCP链接向对方提供RPC服务。

而RPC客户端则需要在一个公共的地址提供一个TCP服务，用于接受RPC服务器的链接请求：

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

当每个链接建立后，基于网络链接构造RPC客户端对象并发送到clientChan管道。

客户端执行RPC调用的操作在doClientWork函数完成：

```go
func doClientWork(clientChan <-chan *rpc.Client) {
	client := <-clientChan
	defer client.Close()

	var reply string
	err = client.Call("HelloService.Hello", "hello", &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(reply)
}
```

首先从管道去取一个RPC客户端对象，并且通过defer语句指定在函数退出前关闭客户端。然后是执行正常的RPC调用。


## 上下文信息

首先是上下文信息，基于上下文我们可以针对不同客户端提供定制化的RPC服务。我们可以通过为每个信道提供独立的RPC服务来实现对上下文特性的支持。

首先改造HelloService，里面增加了对应链接的conn成员：

```go
type HelloService struct {
	conn net.Conn
}
```

然后为每个信道启动独立的RPC服务：

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

Hello方法中就可以根据conn成员识别不同信道的RPC调用：

```go
func (p *HelloService) Hello(request string, reply *string) error {
	*reply = "hello:" + request + ", from" + p.conn.RemoteAddr().String()
	return nil
}
```

基于上下文信息，我们可以方便地为RPC服务增加简单的登陆状态的验证：

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

这样可以要求在客户端链接RPC服务时，首先要执行登陆操作，登陆成功后才能正常执行其他的服务。
