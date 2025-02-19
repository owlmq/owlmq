package web

import (
	"context"
	"fmt"
	"math/big"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/config"
	"github.com/owlmq/owlmq/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *OwlmqServer) NodeList(ctx context.Context, req *pb.NodeListRequest) (*pb.NodeListResponse, error) {

	return &pb.NodeListResponse{
		Successor:   config.GetInstance().Successor,
		Predecessor: config.GetInstance().Successor,
		KnownNodes:  config.GetKnownNodes(),
	}, nil
}

// TODO this needs to be well tested
func (s *OwlmqServer) FindSuccessor(ctx context.Context, req *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	//TODO do i need a mutex here
	conf := config.GetInstance()

	hash := new(big.Int).SetBytes([]byte(req.Hash))
	successorID := crypto.HashKey(conf.Successor)

	// Check if the key is between this node's ID and its successor's ID
	if conf.Successor != "" {
		// If the key is within the range, return the successor
		if hash.Cmp(conf.NodeID) > 0 && hash.Cmp(successorID) <= 0 {
			return &pb.FindSuccessorResponse{Address: conf.Successor}, nil
		}
	}

	// Otherwise, forward the request to the closest node (which is the node closest to the key)
	if conf.Predecessor != "" {
		conn, err := grpc.NewClient(conf.Predecessor, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to predecessor: %v", err)
		}
		defer conn.Close()

		client := pb.NewOwlmqClient(conn)
		resp, err := client.FindSuccessor(ctx, &pb.FindSuccessorRequest{Hash: req.Hash})
		if err != nil {
			return nil, fmt.Errorf("failed to forward FindSuccessor request: %v", err)
		}
		return resp, nil
	}

	// If there’s no predecessor or successor, this node is likely the only node in the ring
	return &pb.FindSuccessorResponse{Address: conf.Hostname}, nil
}
