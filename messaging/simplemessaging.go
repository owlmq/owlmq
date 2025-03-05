package messaging

import (
	"context"

	pb "github.com/owlmq/owlmq/api/owlmq"
)

func newSimpleMessagingLayer() MessagingLayer {
	return &SimpleMessagingLayer{}
}

type SimpleMessagingLayer struct{}

func (s *SimpleMessagingLayer) Publish(ctx context.Context, msg *pb.Message) (*pb.PublishResponse, error) {
	panic("unimplemented")
}

func (s *SimpleMessagingLayer) ConsumeOne(ctx context.Context, req *pb.ConsumeOneRequest) (*pb.Message, error) {
	panic("unimplemented")
}

func (s *SimpleMessagingLayer) Consume(req *pb.ConsumeRequest, stream pb.Owlmq_ConsumeServer) error {
	panic("unimplemented")
}

func (s *SimpleMessagingLayer) Ack(ctx context.Context, req *pb.AckRequest) (*pb.AckResponse, error) {
	panic("unimplemented")
}
