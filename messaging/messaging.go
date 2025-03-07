package messaging

import (
	"context"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/chord"
)

type MessagingLayer interface {
	ProduceOne(ctx context.Context, msg *pb.Message) (*pb.ProduceOneResponse, error)
	Produce(stream pb.Owlmq_ProduceServer) error
	ConsumeOne(ctx context.Context, req *pb.ConsumeOneRequest) (*pb.Message, error)
	Consume(req *pb.ConsumeRequest, stream pb.Owlmq_ConsumeServer) error
	Ack(ctx context.Context, req *pb.AckRequest) (*pb.AckResponse, error)
	NewQueue(ctx context.Context, req *pb.NewQueueRequest) (*pb.NewQueueResponse, error)
}

func New(cl *chord.Chord) (MessagingLayer, error) {
	//TODO maybe add different messaging layers later
	return newSimpleMessagingLayer(cl), nil
}
