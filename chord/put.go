package chord

import (
	"context"
	"fmt"
	"log"

	"github.com/owlmq/owlmq/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/owlmq/owlmq/config"

	pb "github.com/owlmq/owlmq/api/owlmq"
)

//TODO MAYBE MOVE TO OWN PACKAGE KV

func (c *Chord) Put(ctx context.Context, req *pb.KV_PutRequest) (*pb.KV_PutResponse, error) {
	//check if i am the correct node
	keyHash := crypto.HashKey(req.Key)

	if between(crypto.HashKey(config.GetInstance().Predecessor), keyHash, config.GetInstance().NodeID) {
		c.storage.Put(req.Key, req.Value)
		log.Printf("Stored key=%s at node=%s", req.Key, config.GetInstance().Hostname)
		return &pb.KV_PutResponse{Status: pb.KV_STATUS_KV_SUCCESS, ErrorMsg: "Key stored successfully"}, nil
	}

	//transfer to the successor
	//TODO fingertable
	conn, err := grpc.NewClient(config.GetInstance().Successor, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return &pb.KV_PutResponse{Status: pb.KV_STATUS_KV_FAILURE, ErrorMsg: "failed to stored key"}, fmt.Errorf("failed to connect to successor: %v", err)
	}
	defer conn.Close()

	client := pb.NewOwlmqClient(conn)
	return client.Put(ctx, req)
}
