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

// Get sucht den Wert für einen gegebenen Key
func (c *Chord) Get(ctx context.Context, req *pb.KV_GetRequest) (*pb.KV_GetResponse, error) {
	keyHash := crypto.HashKey(req.Key)

	if between(crypto.HashKey(config.GetInstance().Predecessor), keyHash, config.GetInstance().NodeID) {
		value, err := c.storage.Get(req.Key)
		if err != nil {
			return nil, fmt.Errorf("Key not found")
		}
		log.Printf("Retrieved key=%s from node=%s", req.Key, config.GetInstance().Hostname)
		return &pb.KV_GetResponse{Status: pb.KV_STATUS_KV_SUCCESS, Value: value, ErrorMsg: "Key stored successfully"}, nil
	}

	// 2. Weiterleiten an den Successor
	conn, err := grpc.NewClient(config.GetInstance().Successor, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to successor: %v", err)
	}
	defer conn.Close()

	client := pb.NewOwlmqClient(conn)
	return client.Get(ctx, req)
}
