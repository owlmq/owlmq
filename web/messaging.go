package web

import (
	"context"

	pb "github.com/owlmq/owlmq/api/owlmq"
)

// TODO TEST this
func (s *OwlmqServer) Publish(ctx context.Context, msg *pb.Message) (*pb.PublishResponse, error) {
	return s.messagingLayer.Publish(ctx, msg)
}

// TODO TEST this
func (s *OwlmqServer) ConsumeOne(ctx context.Context, req *pb.ConsumeOneRequest) (*pb.Message, error) {
	return s.messagingLayer.ConsumeOne(ctx, req)
}

// TODO TEST this
func (s *OwlmqServer) Consume(req *pb.ConsumeRequest, stream pb.Owlmq_ConsumeServer) error {
	return s.messagingLayer.Consume(req, stream)
}

// TODO TEST this
func (s *OwlmqServer) Ack(ctx context.Context, req *pb.AckRequest) (*pb.AckResponse, error) {
	return s.messagingLayer.Ack(ctx, req)
}
