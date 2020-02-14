# 4.5 gRPC进阶

作为一个基础的RPC框架，安全和扩展是经常遇到的问题。本节将简单介绍如何对gRPC进行安全认证。然后介绍通过gRPC的截取器特性，以及如何通过截取器优雅地实现Token认证、调用跟踪以及Panic捕获等特性。最后介绍了gRPC服务如何和其他Web服务共存。

## 4.5.1 证书认证

gRPC建立在HTTP/2协议之上，对TLS提供了很好的支持。我们前面章节中gRPC的服务都没有提供证书支持，因此客户端在链接服务器中通过`grpc.WithInsecure()`选项跳过了对服务器证书的验证。没有启用证书的gRPC服务在和客户端进行的是明文通讯，信息面临被任何第三方监听的风险。为了保障gRPC通信不被第三方监听篡改或伪造，我们可以对服务器启动TLS加密特性。

可以用以下命令为服务器和客户端分别生成私钥和证书：

```
$ openssl genrsa -out server.key 2048
$ openssl req -new -x509 -days 3650 \
	-subj "/C=GB/L=China/O=grpc-server/CN=server.grpc.io" \
	-key server.key -out server.crt

$ openssl genrsa -out client.key 2048
$ openssl req -new -x509 -days 3650 \
	-subj "/C=GB/L=China/O=grpc-client/CN=client.grpc.io" \
	-key client.key -out client.crt
```

以上命令将生成server.key、server.crt、client.key和client.crt四个文件。其中以.key为后缀名的是私钥文件，需要妥善保管。以.crt为后缀名是证书文件，也可以简单理解为公钥文件，并不需要秘密保存。在subj参数中的`/CN=server.grpc.io`表示服务器的名字为`server.grpc.io`，在验证服务器的证书时需要用到该信息。

有了证书之后，我们就可以在启动gRPC服务时传入证书选项参数：

```go
func main() {
	creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(grpc.Creds(creds))

	...
}
```

其中credentials.NewServerTLSFromFile函数是从文件为服务器构造证书对象，然后通过grpc.Creds(creds)函数将证书包装为选项后作为参数传入grpc.NewServer函数。

在客户端基于服务器的证书和服务器名字就可以对服务器进行验证：

```go
func main() {
	creds, err := credentials.NewClientTLSFromFile(
		"server.crt", "server.grpc.io",
	)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.Dial("localhost:5000",
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	...
}
```

其中redentials.NewClientTLSFromFile是构造客户端用的证书对象，第一个参数是服务器的证书文件，第二个参数是签发证书的服务器的名字。然后通过grpc.WithTransportCredentials(creds)将证书对象转为参数选项传人grpc.Dial函数。

以上这种方式，需要提前将服务器的证书告知客户端，这样客户端在链接服务器时才能进行对服务器证书认证。在复杂的网络环境中，服务器证书的传输本身也是一个非常危险的问题。如果在中间某个环节，服务器证书被监听或替换那么对服务器的认证也将不再可靠。

为了避免证书的传递过程中被篡改，可以通过一个安全可靠的根证书分别对服务器和客户端的证书进行签名。这样客户端或服务器在收到对方的证书后可以通过根证书进行验证证书的有效性。

根证书的生成方式和自签名证书的生成方式类似：

```
$ openssl genrsa -out ca.key 2048
$ openssl req -new -x509 -days 3650 \
	-subj "/C=GB/L=China/O=gobook/CN=github.com" \
	-key ca.key -out ca.crt
```

然后是重新对服务器端证书进行签名：

```
$ openssl req -new \
	-subj "/C=GB/L=China/O=server/CN=server.io" \
	-key server.key \
	-out server.csr
$ openssl x509 -req -sha256 \
	-CA ca.crt -CAkey ca.key -CAcreateserial -days 3650 \
	-in server.csr \
	-out server.crt
```

签名的过程中引入了一个新的以.csr为后缀名的文件，它表示证书签名请求文件。在证书签名完成之后可以删除.csr文件。

然后在客户端就可以基于CA证书对服务器进行证书验证：

```go
func main() {
	certificate, err := tls.LoadX509KeyPair("client.crt", "client.key")
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append ca certs")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates:       []tls.Certificate{certificate},
		ServerName:         tlsServerName, // NOTE: this is required!
		RootCAs:            certPool,
	})

	conn, err := grpc.Dial(
		"localhost:5000", grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	...
}
```

在新的客户端代码中，我们不再直接依赖服务器端证书文件。在credentials.NewTLS函数调用中，客户端通过引入一个CA根证书和服务器的名字来实现对服务器进行验证。客户端在链接服务器时会首先请求服务器的证书，然后使用CA根证书对收到的服务器端证书进行验证。

如果客户端的证书也采用CA根证书签名的话，服务器端也可以对客户端进行证书认证。我们用CA根证书对客户端证书签名：

```
$ openssl req -new \
	-subj "/C=GB/L=China/O=client/CN=client.io" \
	-key client.key \
	-out client.csr
$ openssl x509 -req -sha256 \
	-CA ca.crt -CAkey ca.key -CAcreateserial -days 3650 \
	-in client.csr \
	-out client.crt
```

因为引入了CA根证书签名，在启动服务器时同样要配置根证书：

```go
func main() {
	certificate, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append certs")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert, // NOTE: this is optional!
		ClientCAs:    certPool,
	})

	server := grpc.NewServer(grpc.Creds(creds))
	...
}
```

服务器端同样改用credentials.NewTLS函数生成证书，通过ClientCAs选择CA根证书，并通过ClientAuth选项启用对客户端进行验证。

到此我们就实现了一个服务器和客户端进行双向证书验证的通信可靠的gRPC系统。

## 4.5.2 Token认证

前面讲述的基于证书的认证是针对每个gRPC链接的认证。gRPC还为每个gRPC方法调用提供了认证支持，这样就基于用户Token对不同的方法访问进行权限管理。

要实现对每个gRPC方法进行认证，需要实现grpc.PerRPCCredentials接口：

```go
type PerRPCCredentials interface {
	// GetRequestMetadata gets the current request metadata, refreshing
	// tokens if required. This should be called by the transport layer on
	// each request, and the data should be populated in headers or other
	// context. If a status code is returned, it will be used as the status
	// for the RPC. uri is the URI of the entry point for the request.
	// When supported by the underlying implementation, ctx can be used for
	// timeout and cancellation.
	// TODO(zhaoq): Define the set of the qualified keys instead of leaving
	// it as an arbitrary string.
	GetRequestMetadata(ctx context.Context, uri ...string) (
		map[string]string,	error,
	)
	// RequireTransportSecurity indicates whether the credentials requires
	// transport security.
	RequireTransportSecurity() bool
}
```

在GetRequestMetadata方法中返回认证需要的必要信息。RequireTransportSecurity方法表示是否要求底层使用安全链接。在真实的环境中建议必须要求底层启用安全的链接，否则认证信息有泄露和被篡改的风险。

我们可以创建一个Authentication类型，用于实现用户名和密码的认证：

```go
type Authentication struct {
	User     string
	Password string
}

func (a *Authentication) GetRequestMetadata(context.Context, ...string) (
	map[string]string, error,
) {
	return map[string]string{"user":a.User, "password": a.Password}, nil
}
func (a *Authentication) RequireTransportSecurity() bool {
	return false
}
```

在GetRequestMetadata方法中，我们返回地认证信息包装login和password两个信息。为了演示代码简单，RequireTransportSecurity方法表示不要求底层使用安全链接。

然后在每次请求gRPC服务时就可以将Token信息作为参数选项传人：

```go
func main() {
	auth := Authentication{
		Login:    "gopher",
		Password: "password",
	}

	conn, err := grpc.Dial("localhost"+port, grpc.WithInsecure(), grpc.WithPerRPCCredentials(&auth))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	...
}
```

通过grpc.WithPerRPCCredentials函数将Authentication对象转为grpc.Dial参数。因为这里没有启用安全链接，需要传人grpc.WithInsecure()表示忽略证书认证。

然后在gRPC服务端的每个方法中通过Authentication类型的Auth方法进行身份认证：

```go
type grpcServer struct { auth *Authentication }

func (p *grpcServer) SomeMethod(
	ctx context.Context, in *HelloRequest,
) (*HelloReply, error) {
	if err := p.auth.Auth(ctx); err != nil {
		return nil, err
	}

	return &HelloReply{Message: "Hello " + in.Name}, nil
}

func (a *Authentication) Auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing credentials")
	}

	var appid string
	var appkey string

	if val, ok := md["login"]; ok { appid = val[0] }
	if val, ok := md["password"]; ok { appkey = val[0] }

	if appid != a.Login || appkey != a.Password {
		return grpc.Errorf(codes.Unauthenticated, "invalid token")
	}

	return nil
}
```

详细地认证工作主要在Authentication.Auth方法中完成。首先通过metadata.FromIncomingContext从ctx上下文中获取元信息，然后取出相应的认证信息进行认证。如果认证失败，则返回一个codes.Unauthenticated类型地错误。

## 4.5.3 截取器

gRPC中的grpc.UnaryInterceptor和grpc.StreamInterceptor分别对普通方法和流方法提供了截取器的支持。我们这里简单介绍普通方法的截取器用法。

要实现普通方法的截取器，需要为grpc.UnaryInterceptor的参数实现一个函数：

```go
func filter(ctx context.Context,
	req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	log.Println("filter:", info)
	return handler(ctx, req)
}
```

函数的ctx和req参数就是每个普通的RPC方法的前两个参数。第三个info参数表示当前是对应的那个gRPC方法，第四个handler参数对应当前的gRPC方法函数。上面的函数中首先是日志输出info参数，然后调用handler对应的gRPC方法函数。

要使用filter截取器函数，只需要在启动gRPC服务时作为参数输入即可：

```go
server := grpc.NewServer(grpc.UnaryInterceptor(filter))
```

然后服务器在收到每个gRPC方法调用之前，会首先输出一行日志，然后再调用对方的方法。

如果截取器函数返回了错误，那么该次gRPC方法调用将被视作失败处理。因此，我们可以在截取器中对输入的参数做一些简单的验证工作。同样，也可以对handler返回的结果做一些验证工作。截取器也非常适合前面对Token认证工作。

下面是截取器增加了对gRPC方法异常的捕获：

```go
func filter(
	ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	log.Println("filter:", info)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	return handler(ctx, req)
}
```

不过gRPC框架中只能为每个服务设置一个截取器，因此所有的截取工作只能在一个函数中完成。开源的grpc-ecosystem项目中的go-grpc-middleware包已经基于gRPC对截取器实现了链式截取器的支持。

以下是go-grpc-middleware包中链式截取器的简单用法

```go
import "github.com/grpc-ecosystem/go-grpc-middleware"

myServer := grpc.NewServer(
	grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		filter1, filter2, ...
	)),
	grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
		filter1, filter2, ...
	)),
)
```

感兴趣的同学可以参考go-grpc-middleware包的代码。

## 4.5.4 和Web服务共存

gRPC构建在HTTP/2协议之上，因此我们可以将gRPC服务和普通的Web服务架设在同一个端口之上。

对于没有启动TLS协议的服务则需要对HTTP2/2特性做适当的调整：

```go
func main() {
	mux := http.NewServeMux()

	h2Handler := h2c.NewHandler(mux, &http2.Server{})
	server = &http.Server{Addr: ":3999", Handler: h2Handler}
	server.ListenAndServe()
}
```

启用普通的https服务器则非常简单：

```go
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "hello")
	})

	http.ListenAndServeTLS(port, "server.crt", "server.key",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mux.ServeHTTP(w, r)
			return
		}),
	)
}
```

而单独启用带证书的gRPC服务也是同样的简单：

```go
func main() {
	creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(creds))

	...
}
```

因为gRPC服务已经实现了ServeHTTP方法，可以直接作为Web路由处理对象。如果将gRPC和Web服务放在一起，会导致gRPC和Web路径的冲突，在处理时我们需要区分两类服务。

我们可以通过以下方式生成同时支持Web和gRPC协议的路由处理函数：

```go
func main() {
	...

	http.ListenAndServeTLS(port, "server.crt", "server.key",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor != 2 {
				mux.ServeHTTP(w, r)
				return
			}
			if strings.Contains(
				r.Header.Get("Content-Type"), "application/grpc",
			) {
				grpcServer.ServeHTTP(w, r) // gRPC Server
				return
			}

			mux.ServeHTTP(w, r)
			return
		}),
	)
}
```

首先gRPC是建立在HTTP/2版本之上，如果HTTP不是HTTP/2协议则必然无法提供gRPC支持。同时，每个gRPC调用请求的Content-Type类型会被标注为"application/grpc"类型。

这样我们就可以在gRPC端口上同时提供Web服务了。

