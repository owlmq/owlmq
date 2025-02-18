package web

import (
	"context"

	pb "github.com/owlmq/owlmq/api/owlmq"
)

func (s *OwlmqServer) Get(ctx context.Context, req *pb.KV_GetRequest) (*pb.KV_GetResponse, error) {
	value, err := s.chordLayer.Get(req.Key)
	if err != nil {
		return &pb.KV_GetResponse{Status: pb.KV_STATUS_FAILURE, Error: "Key not found"}, nil
	}
	return &pb.KV_GetResponse{Status: pb.KV_STATUS_SUCCESS, Value: value}, nil
}

func (s *OwlmqServer) Put(ctx context.Context, req *pb.KV_PutRequest) (*pb.KV_PutResponse, error) {
	//TODO error handling
	s.chordLayer.Put(req.Key, req.Value)
	return &pb.KV_PutResponse{Status: pb.KV_STATUS_SUCCESS}, nil
}
