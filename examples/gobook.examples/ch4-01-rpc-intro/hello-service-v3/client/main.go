package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

var flagAddr = flag.String("addr", "localhost:1234", "server address")

func main() {
	flag.Parse()

	// nc localhost 1234
	// {"method":"HelloService.Hello","params":["hello"],"id":0}
	// echo -e '{"method":"HelloService.Hello","params":["hello2222"],"id":3}' | nc localhost 1234
	// echo -e '{"method":"HelloService.Hello","params":["hello2222"],"id":3}{"method":"HelloService.Hello","params":["hello33"],"id":4}' | nc localhost 1234

	conn, err := net.Dial("tcp", *flagAddr)
	if err != nil {
		log.Fatal("net.Dial:", err)
	}

	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))

	var reply string
	err = client.Call("HelloService.Hello", "hello", &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(reply)
}
