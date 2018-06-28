package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type HelloService struct{}

func (p *HelloService) Hello(request string, reply *string) error {
	*reply = "hello:" + request
	return nil
}

var flagPort = flag.Int("port", 1234, "listen port")

func main() {
	flag.Parse()

	rpc.RegisterName("HelloService", new(HelloService))

	// nc -l 2399
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *flagPort))
	if err != nil {
		log.Fatal("ListenTCP error:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Accept error:", err)
		}

		go rpc.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
