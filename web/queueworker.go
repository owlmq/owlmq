package web

import (
	"context"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/queueworker"
)

// TODO TEST THIS
func (s *OwlmqServer) NotifyLocalConsumers(ctx context.Context, req *pb.NotifyLocalConsumersRequest) (*pb.NotifyLocalConsumersResponse, error) {
	qw, _ := queueworker.GetInstance()
	return qw.NotifyLocalConsumers(ctx, req)
}
