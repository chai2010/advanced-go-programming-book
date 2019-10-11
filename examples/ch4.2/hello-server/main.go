package main

import (
	"log"
	"net"
	"net/rpc"

	rpcpb "github.com/chai2010/advanced-go-programming-book/examples/ch4.2/hello.pb"
)

type HelloService struct{}

func (p *HelloService) Hello(request *rpcpb.String, reply *rpcpb.String) error {
	reply.Value = "hello:" + request.GetValue()
	return nil
}

func main() {
	rpcpb.RegisterHelloService(rpc.DefaultServer, new(HelloService))

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}

		go rpc.ServeConn(conn)
	}
}
