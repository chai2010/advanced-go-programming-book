package main

import (
	"context"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"

	hs "ch4.4-2/HelloService"
)

type HelloServiceImpl struct{}

func (p *HelloServiceImpl) Hello(
	ctx context.Context, args *hs.String,
) (*hs.String, error) {
	reply := &hs.String{Value: "hello:" + args.GetValue()}
	return reply, nil
}

//grpc stream
func (p *HelloServiceImpl) Channel(stream hs.HelloService_ChannelServer) error {
	for {
		args, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		reply := &hs.String{Value: "hello:" + args.GetValue()}

		err = stream.Send(reply)
		if err != nil {
			return err
		}
	}
}

func main() {
	grpcServer := grpc.NewServer()
	hs.RegisterHelloServiceServer(grpcServer, new(HelloServiceImpl))

	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer.Serve(lis)
}
