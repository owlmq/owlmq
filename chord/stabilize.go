package chord

import (
	"context"
	"fmt"
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
				log.Printf("Failed to get predecessor from successor  %s: %v updating to next successor in the list", config.GetInstance().Successor, err)
				//successor is unreachable try the next in my SuccessorList
				if len(config.GetInstance().SuccessorList) > 0 {
					config.SetSuccessor(config.GetInstance().SuccessorList[0])
				} else {
					// if its empty pick my self -> i am alone in the ring
					config.SetSuccessor(config.GetInstance().Hostname)
				}
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

			//update fingertable
			c.updateFingerTable()

			//update successor stack
			c.updateSuccessorAndReplicatorList()

			//check on Predecessor
			c.checkPredecessor()
		}
	}
}

func (c *Chord) updateSuccessorAndReplicatorList() {
	suc_list := []string{}
	rep_list := []string{}
	next := config.GetInstance().Successor

	maxIndex := 0
	if config.SuccessorCount > config.ReplicaCount {
		maxIndex = config.SuccessorCount
	} else {
		maxIndex = config.ReplicaCount
	}

	for i := 0; i < maxIndex; i++ {
		if next == config.GetInstance().Hostname || next == "" {
			break
		}
		conn, err := grpc.NewClient(next, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect to successor %s: %v", config.GetInstance().Successor, err)
		}
		client := pb.NewOwlmqClient(conn)
		resp, _ := client.GetSuccessor(context.Background(), &pb.GetSuccessorRequest{})
		conn.Close()

		//add to SuccessorList
		if i < config.SuccessorCount {
			suc_list = append(suc_list, resp.GetAddress())
		}
		if i < config.ReplicaCount {
			rep_list = append(rep_list, resp.GetAddress())
		}

		next = resp.GetAddress()
	}

	//FIXME this could lead to a memory leak (dangling pointers?)
	config.GetInstance().SuccessorList = suc_list
	config.GetInstance().ReplicatorNodeList = rep_list
}

// TODO MOVE TO CORRECT LOCATION
func (c *Chord) checkPredecessor() {
	conn, err := grpc.NewClient(config.GetInstance().Predecessor, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Vorgänger nicht erreichbar. Ersetze Vorgänger.")
		replacePredecessor()
	} else {
		client := pb.NewOwlmqClient(conn)
		_, err = client.Ping(context.Background(), &pb.PingRequest{})
		if err != nil {
			fmt.Println("Vorgänger nicht erreichbar. Ersetze Vorgänger.")
			replacePredecessor()
		} else {
			//vorgänger läuft normal
		}
	}
}

func replacePredecessor() {
	suc_list := config.GetInstance().SuccessorList

	//no successors found
	if len(suc_list) == 0 {
		config.SetPredecessor(config.GetInstance().Hostname)
		return
	}
	config.SetPredecessor(suc_list[0])
}
