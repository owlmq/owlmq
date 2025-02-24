package chord

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

func (c *Chord) FindSuccessor(ctx context.Context, req *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	id := new(big.Int).SetBytes(req.Hash)
	successorID := crypto.HashKey(config.GetInstance().Successor)

	// 1. Prüfen, ob der aktuelle Knoten direkt zuständig ist
	if between(config.GetInstance().NodeID, id, successorID) {
		return &pb.FindSuccessorResponse{Address: config.GetInstance().Successor}, nil
	}

	// 2. Mit Finger Table den nächsten Sprung ermitteln
	nextaddr := c.closestPrecedingNode(id)

	// 3. Anfrage an den nächsten Finger-Eintrag weiterleiten
	conn, err := grpc.NewClient(nextaddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node %s: %v", nextaddr, err)
	}
	defer conn.Close()

	client := pb.NewOwlmqClient(conn)
	return client.FindSuccessor(ctx, req)
}

func (c *Chord) closestPrecedingNode(id *big.Int) string {
	// search finger table from the last to the first
	for i := len(config.GetInstance().FingerTable) - 1; i >= 0; i-- {
		finger := config.GetInstance().FingerTable[i]
		if between(config.GetInstance().NodeID, crypto.HashKey(finger.Address), id) {
			return finger.Address
		}
	}
	// if not found Successor is fallback
	return config.GetInstance().Successor
}

func (c *Chord) updateFingerTable() {
	for i := 0; i < int(config.M); i++ {
		start := new(big.Int).Add(config.GetInstance().NodeID, new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(i)), nil))
		start.Mod(start, new(big.Int).Exp(big.NewInt(2), big.NewInt(config.M), nil))

		// call FindSuccessor for every startID
		resp, err := c.FindSuccessor(context.Background(), &pb.FindSuccessorRequest{Hash: start.Bytes()})
		if err == nil {
			config.GetInstance().FingerTable[i] = &config.FingerEntry{Start: start, Address: resp.Address}
		} else {
			// use Successor as fallback
			config.GetInstance().FingerTable[i] = &config.FingerEntry{Start: start, Address: config.GetInstance().Successor}
		}
	}
}
