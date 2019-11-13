# 4.5 gRPC Advanced

As a basic RPC framework, security and extension are frequently encountered problems. This section will briefly describe how to securely authenticate gRPC. Then introduce the interceptor features through gRPC, and how to elegantly implement Token authentication, call tracking and Panic capture through the interceptor. Finally, how the gRPC service coexists with other web services is introduced.

## 4.5.1 Certificate Certification

gRPC is built on top of the HTTP/2 protocol and provides good support for TLS. The gRPC service in our previous chapter did not provide certificate support, so the client skipped the verification of the server certificate in the linked server via the `grpc.WithInsecure()` option. A gRPC service that does not have a certificate enabled is in clear text communication with the client, and the information is at risk of being monitored by any third party. In order to ensure that gRPC communication is not tampered or forged by third parties, we can enable TLS encryption on the server.

You can generate a private key and certificate for the server and client separately with the following command:

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

The above command will generate four files: server.key, server.crt, client.key, and client.crt. The private key file is suffixed with .key and needs to be kept safely. The .crt suffix is ​​a certificate file, which can also be simply understood as a public key file, and does not need to be secretly saved. The `/CN=server.grpc.io` in the subj parameter indicates that the server name is `server.grpc.io`, which is needed to verify the server's certificate.

With the certificate, we can pass in the certificate option parameters when starting the gRPC service:

```go
Func main() {
Creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
If err != nil {
log.Fatal(err)
}

Server := grpc.NewServer(grpc.Creds(creds))

...
}
```

The credentials.NewServerTLSFromFile function constructs the certificate object from the file for the server, and then wraps the certificate as an option through the grpc.Creds(creds) function and passes it as a parameter to the grpc.NewServer function.

The server can be authenticated on the client based on the server's certificate and server name:

```go
Func main() {
Creds, err := credentials.NewClientTLSFromFile(
"server.crt", "server.grpc.io",
)
If err != nil {
log.Fatal(err)
}

Conn, err := grpc.Dial("localhost:5000",
grpc.WithTransportCredentials(creds),
)
If err != nil {
log.Fatal(err)
}
Defer conn.Close()

...
}
```

Where redentials.NewClientTLSFromFile is the certificate object used to construct the client. The first parameter is the server's certificate file, and the second parameter is the name of the server that issued the certificate. Then pass the grpc.WithTransportCredentials(creds) to convert the certificate object to the parameter option to pass the grpc.Dial function.

In this way, the server's certificate needs to be notified to the client in advance, so that the client can authenticate the server certificate when the server is linked. In a complex network environment, the transmission of server certificates is itself a very dangerous issue. If the server certificate is monitored or replaced at some point in the middle, the authentication to the server will no longer be reliable.

In order to avoid tampering during the delivery of the certificate, the server and client certificates can be signed separately by a secure and reliable root certificate. In this way, the client or server can verify the validity of the certificate through the root certificate after receiving the certificate of the other party.

The root certificate is generated in a similar way to the self-signed certificate:

```
$ openssl genrsa -out ca.key 2048
$ openssl req -new -x509 -days 3650 \
-subj "/C=GB/L=China/O=gobook/CN=github.com" \
-key ca.key -out ca.crt
```

Then re-sign the server-side certificate:

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

The signature process introduces a new file with a .csr extension that represents the certificate signing request file. The .csr file can be deleted after the certificate signature is completed.

Then the client can perform certificate verification on the server based on the CA certificate:

```go
Func main() {
Certificate, err := tls.LoadX509KeyPair("client.crt", "client.key")
If err != nil {
log.Fatal(err)
}

certPool := x509.NewCertPool()
Ca, err := ioutil.ReadFile("ca.crt")
If err != nil {
log.Fatal(err)
}
If ok := certPool.AppendCertsFromPEM(ca); !ok {
log.Fatal("failed to append ca certs")
}

Creds := credentials.NewTLS(&tls.Config{
Certificates: []tls.Certificate{certificate},
ServerName: tlsServerName, // NOTE: this is required!
RootCAs: certPool,
})

Conn, err := grpc.Dial(
"localhost:5000", grpc.WithTransportCredentials(creds),
)
If err != nil {
log.Fatal(err)
}
Defer conn.Close()

...
}
```

In the new client code, we no longer rely directly on server-side certificate files. In the credentials.NewTLS function call, the client authenticates the server by introducing a CA root certificate and the name of the server. When the client links to the server, it first requests the server's certificate, and then uses the CA root certificate to verify the received server-side certificate.

If the client's certificate is also signed by the CA root certificate, the server can also perform certificate authentication on the client. We use the CA root certificate to sign the client certificate:

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

Because the CA root certificate signature was introduced, the root certificate is also configured when starting the server:

```go
Func main() {
Certificate, err := tls.LoadX509KeyPair("server.crt", "server.key")
If err != nil {
log.Fatal(err)
}

certPool := x509.NewCertPool()
Ca, err := ioutil.ReadFile("ca.crt")
If err != nil {
log.Fatal(err)
}
If ok := certPool.AppendCertsFromPEM(ca); !ok {
log.Fatal("failed to append certs")
}

Creds := credentials.NewTLS(&tls.Config{
Certificates: []tls.Certificate{certificate},
ClientAuth: tls.RequireAndVerifyClientCert, // NOTE: this is optional!
ClientCAs: certPool,
})

Server := grpc.NewServer(grpc.Creds(creds))
...
}
```

The server also uses the credentials.NewTLS function to generate a certificate, selects the CA root certificate through ClientCAs, and enables the client to be authenticated through the ClientAuth option.

At this point, we have implemented a reliable gRPC system for communication between the server and the client for two-way certificate verification.

## 4.5.2 Token Certification

The certificate-based authentication described above is for each gRPC link. gRPC also provides authentication support for each gRPC method call, so that rights management is performed on different method accesses based on the user token.

To implement authentication for each gRPC method, you need to implement the grpc.PerRPCCredentials interface:

```go
Type PerRPCCredentials interface {
// GetRequestMetadata gets the current request metadata, refreshing
// tokens if required. This should be called by the transport layer on
// each request, and the data should be populated in headers or other
// context. If a status code is returned, it will be used as the status
// for the RPC. uri is the URI of the entry point for the request.
// When supported by the underlying implementation, ctx can be used for
// timeout and cancellation.
// TODO(zhaoq): Define the set of the qualified keys instead of
// it as an arbitrary string.
GetRequestMetadata(ctx context.Context, uri ...string) (
Map[string]string, error,
)
// RequireTransportSecurity indicates whether the credentials requires
// transport security.
RequireTransportSecurity() bool
}
```

Returns the necessary information for authentication in the GetRequestMetadata method. The RequireTransportSecurity method indicates whether the underlying secure link is required. In a real environment, it is recommended that the underlying security-enabled links be required, otherwise the authentication information is at risk of being compromised and tampered with.

We can create an Authentication type to authenticate the username and password:

```go
Type Authentication struct {
User string
Password string
}

Func (a *Authentication) GetRequestMetadata(context.Context, ...string) (
Map[string]string, error,
) {
Return map[string]string{"user":a.User, "password": a.Password}, ​​nil
}
Func (a *Authentication) RequireTransportSecurity() bool {
Return false
}
```

In the GetRequestMetadata method, we return the local authentication information to wrap both login and password information. To demonstrate that the code is simple, the RequireTransportSecurity method means that the underlying secure link is not required.

The token information can then be passed as a parameter option each time the gRPC service is requested:

```go
Func main() {
Auth := Authentication{
Login: "gopher",
Password: "password",
}

Conn, err := grpc.Dial("localhost"+port, grpc.WithInsecure(), grpc.WithPerRPCCredentials(&auth))
If err != nil {
log.Fatal(err)
}
Defer conn.Close()

...
}
```

The Authentication object is converted to the grpc.Dial parameter by the grpc.WithPerRPCCredentials function. Since the secure link is not enabled here, you need to pass the grpc.WithInsecure() to ignore the certificate authentication.

Then, in each method of the gRPC server, the identity is authenticated by the Authentication method of the Auth method:

```go
Type grpcServer struct { auth *Authentication }

Func (p *grpcServer) SomeMethod(
Ctx context.Context, in *HelloRequest,
) (*HelloReply, error) {
If err := p.auth.Auth(ctx); err != nil {
Return nil, err
}

Return &HelloReply{Message: "Hello " + in.Name}, nil
}

Func (a *Authentication) Auth(ctx context.Context) error {
Md, ok := metadata.FromIncomingContext(ctx)
If !ok {
Return fmt.Errorf("missing credentials")
}

Var appid string
Var appkey string

If val, ok := md["login"]; ok { appid = val[0] }
If val, ok := md["password"]; ok { appkey = val[0] }

If appid != a.Login || appkey != a.Password {
Return grpc.Errorf(codes.Unauthenticated, "invalid token")
}

Return nil
}
```

The detailed authentication work is mainly done in the Authentication.Auth method. First, the meta information is obtained from the ctx context through the metadata.FromIncomingContext, and then the corresponding authentication information is taken out for authentication. If the authentication fails, it returns a code.Unauthenticated type error.

## 4.5.3 Interceptor

The grpc.UnaryInterceptor and grpc.StreamInterceptor in gRPC provide interceptor support for common methods and stream methods, respectively. Here we briefly introduce the use of interceptors for common methods.

To implement an interceptor for a normal method, you need to implement a function for the argument to grpc.UnaryInterceptor:

```go
Func filter(ctx context.Context,
Req interface{}, info *grpc.UnaryServerInfo,
Handler grpc.UnaryHandler,
) (resp interface{}, err error) {
log.Println("fileter:", info)
Return handler(ctx, req)
}
```

The ctx and req parameters of the function are the first two parameters of each normal RPC method. The third info parameter indicates that the corresponding gRPC method is currently in use, and the fourth handler parameter corresponds to the current gRPC method function. The first function in the above is the log output info parameter, and then call the corresponding gRPC method function of the handler.

To use the filter interceptor function, you only need to enter it as a parameter when starting the gRPC service:

```go
Server := grpc.NewServer(grpc.UnaryInterceptor(filter))
```

Then the server will first output a log before receiving each gRPC method call, and then call the other party's method.

If the interceptor function returns an error, then the gRPC method call will be treated as a failure. Therefore, we can do some simple verification work on the input parameters in the interceptor. Similarly, you can do some verification work on the results returned by the handler. The interceptor is also very suitable for the previous work on Token certification.

The following is an interceptor that adds an exception to the gRPC method exception:

```go
Func filter(
Ctx context.Context, req interface{},
Info *grpc.UnaryServerInfo,
Handler grpc.UnaryHandler,
) (resp interface{}, err error) {
log.Println("fileter:", info)

Defer func() {
If r := recover(); r != nil {
Err = fmt.Errorf("panic: %v", r)
}
}()

Return handler(ctx, req)
}
```

However, only one interceptor can be set for each service in the gRPC framework, so all interception can only be done in one function. The go-grpc-middleware package in the open source grpc-ecosystem project has implemented chain interceptor support for interceptors based on gRPC.

The following is a simple usage of the chain interceptor in the go-grpc-middleware package.

```go
Import "github.com/grpc-ecosystem/go-grpc-middleware"

myServer := grpc.NewServer(
grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
Filter1, filter2, ...
)),
grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
Filter1, filter2, ...
)),
)
```

Interested students can refer to the code of the go-grpc-middleware package.

## 4.5.4 Coexistence with Web Services

gRPC is built on top of the HTTP/2 protocol, so we can put the gRPC service on the same port as the normal web service.

For services that do not have the TLS protocol enabled, you need to make appropriate adjustments to the HTTP2/2 features:

```go
Func main() {
Mux := http.NewServeMux()

h2Handler := h2c.NewHandler(mux, &http2.Server{})
Server = &http.Server{Addr: ":3999", Handler: h2Handler}
server.ListenAndServe()
}
```

Enabling a normal https server is very simple:

```go
Func main() {
Mux := http.NewServeMux()
mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
fmt.Fprintln(w, "hello")
})

http.ListenAndServeTLS(port, "server.crt", "server.key",
http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
mux.ServeHTTP(w, r)
Return
}),
)
}
```

It is equally simple to enable the gRPC service with certificates separately:

```go
Func main() {
Creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
If err != nil {
log.Fatal(err)
}

grpcServer := grpc.NewServer(grpc.Creds(creds))

...
}
```

Because the gRPC service has implemented the ServeHTTP method, it can be directly used as a Web routing processing object. If you put gRPC and Web services together, it will lead to conflicts between gRPC and Web path. We need to distinguish between two types of services when processing.

We can generate routing handlers that support both the Web and gRPC protocols in the following ways:

```go
Func main() {
...

http.ListenAndServeTLS(port, "server.crt", "server.key",
http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
If r.ProtoMajor != 2 {
mux.ServeHTTP(w, r)
Return
}
If strings.Contains(
r.Header.Get("Content-Type"), "application/grpc",
) {
grpcServer.ServeHTTP(w, r) // gRPC Server
Return
}

mux.ServeHTTP(w, r)
Return
}),
)
}
```

First gRPC is built on top of the HTTP/2 version. If HTTP is not an HTTP/2 protocol, gRPC support will not be available. At the same time, the Content-Type type of each gRPC call request will be marked as "application/grpc" type.

This way we can provide web services on the gRPC port at the same time.