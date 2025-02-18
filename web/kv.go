package web

import (
	"context"

	pb "github.com/owlmq/owlmq/api/owlmq"
)

func (s *OwlmqServer) Get(ctx context.Context, req *pb.KV_GetRequest) (*pb.KV_GetResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	//TODO check in the chord layer
	value, exists := s.store[req.Key]
	if !exists {
		return &pb.KV_GetResponse{Status: pb.KV_STATUS_FAILURE, Error: "Key not found"}, nil
	}
	return &pb.KV_GetResponse{Status: pb.KV_STATUS_SUCCESS, Value: value}, nil
}

func (s *OwlmqServer) Put(ctx context.Context, req *pb.KV_PutRequest) (*pb.KV_PutResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	//TODO check in the chord layer
	s.store[req.Key] = req.Value
	return &pb.KV_PutResponse{Status: pb.KV_STATUS_SUCCESS}, nil
}
