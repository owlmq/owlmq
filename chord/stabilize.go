package chord

import (
	"context"
	"log"
	"time"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/config"
	"github.com/owlmq/owlmq/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (c *Chord) Stabilize() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 1. Verbindung zum aktuellen Successor aufbauen
			conn, err := grpc.NewClient(config.GetInstance().Successor, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Printf("Failed to connect to successor %s: %v", config.GetInstance().Successor, err)
				continue
			}
			client := pb.NewOwlmqClient(conn)

			// 2. Successor nach seinem Predecessor fragen
			resp, err := client.GetPredecessor(context.Background(), &pb.GetPredecessorRequest{})
			conn.Close()

			if err != nil {
				log.Printf("Failed to get predecessor from successor %s: %v", config.GetInstance().Successor, err)
				continue
			}

			// 3. Falls der Predecessor des Successors zwischen diesem Knoten und dem aktuellen Successor liegt, setzen wir ihn als neuen Successor
			if resp.Address != "" {
				newSuccessorID := crypto.HashKey(resp.Address)
				if between(crypto.HashKey(config.GetInstance().Hostname), newSuccessorID, crypto.HashKey(config.GetInstance().Successor)) {
					config.SetSuccessor(resp.Address)
					log.Printf("Updated successor to %s", resp.Address)
				}
			}

			// 4. Dem neuen Successor mitteilen, dass dieser Knoten sein Predecessor ist
			conn, err = grpc.NewClient(config.GetInstance().Successor, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err == nil {
				defer conn.Close()
				client = pb.NewOwlmqClient(conn)
				_, err = client.SetPredecessor(context.Background(), &pb.SetPredecessorRequest{Address: config.GetInstance().Hostname})
				if err != nil {
					log.Printf("Failed to update predecessor of successor %s: %v", config.GetInstance().Successor, err)

				}
			}
		}
	}
}
