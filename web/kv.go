package web

import (
	"context"

	pb "github.com/owlmq/owlmq/api/owlmq"
)

func (s *OwlmqServer) Get(ctx context.Context, req *pb.KV_GetRequest) (*pb.KV_GetResponse, error) {
	return s.chordLayer.Get(ctx, req)
}

func (s *OwlmqServer) Put(ctx context.Context, req *pb.KV_PutRequest) (*pb.KV_PutResponse, error) {
	//TODO error handling
	return s.chordLayer.Put(ctx, req)
}
