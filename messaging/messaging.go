package messaging

import (
	"context"

	pb "github.com/owlmq/owlmq/api/owlmq"
)

type MessagingLayer interface {
	Publish(ctx context.Context, msg *pb.Message) (*pb.PublishResponse, error)
	Consume(req *pb.ConsumeRequest, stream pb.Owlmq_ConsumeServer) error
	ConsumeOne(ctx context.Context, req *pb.ConsumeOneRequest) (*pb.Message, error)
	Ack(ctx context.Context, req *pb.AckRequest) (*pb.AckResponse, error)
}

func New() (MessagingLayer, error) {
	//TODO maybe add different messaging layers later
	return newSimpleMessagingLayer(), nil
}
