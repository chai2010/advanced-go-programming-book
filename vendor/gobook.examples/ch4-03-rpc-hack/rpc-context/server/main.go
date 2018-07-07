package main

import (
	"log"
	"net"
	"net/rpc"
)

type HelloService struct {
	conn net.Conn
}

func ServeHelloService(conn net.Conn) {
	p := rpc.NewServer()
	p.Register(&HelloService{conn: conn})
	p.ServeConn(conn)
}

func (p *HelloService) Hello(request string, reply *string) error {
	*reply = "hello:" + request + ", from" + p.conn.RemoteAddr().String()
	return nil
}

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

		go ServeHelloService(conn)
	}
}
