package main

import (
	"google.golang.org/grpc"
	"log"
	"fmt"
	."github.com/advanced-go-programming-book-code/ch4/s04/e01grpc/helloservice"
	"context"
)

func main() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := NewHelloServiceClient(conn)
	reply, err := client.Hello(context.Background(), &String{Value: "hello"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply.GetValue())
}