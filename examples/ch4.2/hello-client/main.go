package main

import (
	"fmt"
	"log"

	rpcpb "github.com/chai2010/advanced-go-programming-book/examples/ch4.2/hello.pb"
)


func main() {
	client, err := rpcpb.DialHelloService("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	request := &rpcpb.String{
		Value:	"hello from client",
	}
	reply := &rpcpb.String{}
	err = client.Hello(request, reply)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}