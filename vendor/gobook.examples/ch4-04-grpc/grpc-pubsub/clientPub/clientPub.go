package main

import (
	"google.golang.org/grpc"
	"log"
	."gobook.examples/ch4-04-grpc/grpc-pubsub/helloservice"
	"context"
)

func main() {
	conn, err := grpc.Dial("localhost:1234", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := NewPubsubServiceClient(conn)

	_, err = client.Publish(context.Background(), &String{Value: "golang: hello Go"})
	if err != nil {
		log.Fatal(err)
	}
	_, err = client.Publish(context.Background(), &String{Value: "docker: hello Docker"})
	if err != nil {
		log.Fatal(err)
	}
}
