package main

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"github.com/docker/docker/pkg/pubsub"
	"google.golang.org/grpc"

	pb "ch4.4-3/pubsubservice"
)

type PubsubService struct {
	pub *pubsub.Publisher
}

func NewPubsubService() *PubsubService {
	return &PubsubService{
		pub: pubsub.NewPublisher(100*time.Millisecond, 10),
	}
}

func (p *PubsubService) Publish(
	ctx context.Context, arg *pb.String,
) (*pb.String, error) {
	p.pub.Publish(arg.GetValue())
	//debug
	//reply := &String{Value: "<Publish>  " + arg.GetValue()}
	//fmt.Println(reply.GetValue())
	return &pb.String{}, nil
}

func (p *PubsubService) Subscribe(
	arg *pb.String, stream pb.PubsubService_SubscribeServer,
) error {
	ch := p.pub.SubscribeTopic(func(v interface{}) bool {
		if key, ok := v.(string); ok {
			//debug
			//fmt.Printf("<debug> %t %s %s %t\n",
			//	ok,arg.GetValue(),key,strings.HasPrefix(key,arg.GetValue()))
			if strings.HasPrefix(key, arg.GetValue()) {
				return true
			}
		}
		return false
	})

	for v := range ch {
		if err := stream.Send(&pb.String{Value: v.(string)}); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	grpcServer := grpc.NewServer()
	pb.RegisterPubsubServiceServer(grpcServer, NewPubsubService())

	lis, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer.Serve(lis)
}
