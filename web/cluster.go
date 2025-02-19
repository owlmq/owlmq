package web

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/config"
	"github.com/owlmq/owlmq/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *OwlmqServer) NodeJoin(ctx context.Context, req *pb.NodeJoinRequest) (*pb.NodeJoinResponse, error) {
	newNodeAddr := req.Address
	newNodeID := crypto.HashKey(newNodeAddr)

	// 1. Den Successor für den neuen Knoten im Ring finden
	conn, err := grpc.NewClient(config.GetInstance().Successor, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bootstrap node: %v", err)
	}
	defer conn.Close()
	client := pb.NewOwlmqClient(conn)
	findResp, err := client.FindSuccessor(ctx, &pb.FindSuccessorRequest{Hash: newNodeID.Bytes()})
	if err != nil {
		return nil, fmt.Errorf("failed to find successor for new node: %v", err)
	}
	successorAddr := findResp.Address

	// 2. Verbindung zum Successor aufbauen und seinen aktuellen Predecessor abrufen
	connSuccessor, err := grpc.NewClient(successorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to successor: %v", err)
	}
	defer connSuccessor.Close()

	successorClient := pb.NewOwlmqClient(connSuccessor)
	predResp, err := successorClient.GetPredecessor(ctx, &pb.GetPredecessorRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get predecessor from successor: %v", err)
	}
	oldPredecessorAddr := predResp.Address

	// 3. Den neuen Knoten in den Ring einfügen
	config.SetSuccessor(newNodeAddr) // Unser neuer Nachfolger ist jetzt der neue Knoten

	// 4. Den neuen Knoten als Vorgänger des Successors setzen
	_, err = successorClient.SetPredecessor(ctx, &pb.SetPredecessorRequest{Address: newNodeAddr})
	if err != nil {
		return nil, fmt.Errorf("failed to update predecessor of successor: %v", err)
	}

	// 5. Falls der Successor schon einen Vorgänger hatte, diesen benachrichtigen
	if oldPredecessorAddr != "" {
		connOldPred, err := grpc.NewClient(oldPredecessorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			defer connOldPred.Close()
			oldPredClient := pb.NewOwlmqClient(connOldPred)
			_, _ = oldPredClient.SetSuccessor(ctx, &pb.SetSuccessorRequest{Address: newNodeAddr})
		}
	}

	return &pb.NodeJoinResponse{
		Status:   pb.NODEJOIN_STATUS_NJ_SUCCESS,
		ErrorMsg: "",
	}, nil
}

func (s *OwlmqServer) NodeList(ctx context.Context, req *pb.NodeListRequest) (*pb.NodeListResponse, error) {

	return &pb.NodeListResponse{
		Successor:   config.GetInstance().Successor,
		Predecessor: config.GetInstance().Predecessor,
		KnownNodes:  config.GetKnownNodes(),
	}, nil
}

// TODO this needs to be well tested
func (s *OwlmqServer) FindSuccessor(ctx context.Context, req *pb.FindSuccessorRequest) (*pb.FindSuccessorResponse, error) {
	//TODO do i need a mutex here

	hash := new(big.Int).SetBytes([]byte(req.Hash))
	successorID := crypto.HashKey(config.GetInstance().Successor)

	// Check if the key is between this node's ID and its successor's ID
	if config.GetInstance().Successor != config.GetInstance().Hostname {
		// If the key is within the range, return the successor
		if hash.Cmp(config.GetInstance().NodeID) > 0 && hash.Cmp(successorID) <= 0 {
			return &pb.FindSuccessorResponse{Address: config.GetInstance().Successor}, nil
		}
	}

	// Otherwise, forward the request to the closest node (which is the node closest to the key)
	if config.GetInstance().Predecessor != config.GetInstance().Hostname {
		conn, err := grpc.NewClient(config.GetInstance().Predecessor, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	return &pb.FindSuccessorResponse{Address: config.GetInstance().Hostname}, nil
}

func (s *OwlmqServer) GetPredecessor(ctx context.Context, req *pb.GetPredecessorRequest) (*pb.GetPredecessorResponse, error) {
	return &pb.GetPredecessorResponse{Address: config.GetInstance().Predecessor}, nil
}
func (s *OwlmqServer) SetPredecessor(ctx context.Context, req *pb.SetPredecessorRequest) (*pb.SetPredecessorResponse, error) {
	config.SetPredecessor(req.Address)
	return &pb.SetPredecessorResponse{}, nil
}
func (s *OwlmqServer) SetSuccessor(ctx context.Context, req *pb.SetSuccessorRequest) (*pb.SetSuccessorResponse, error) {
	config.SetSuccessor(req.Address)
	return &pb.SetSuccessorResponse{}, nil
}

func (s *OwlmqServer) Stabilize() {
	ticker := time.NewTicker(1 * time.Second) // Alle 5 Sekunden ausführen
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

func between(start, id, end *big.Int) bool {
	if start.Cmp(end) < 0 {
		return start.Cmp(id) < 0 && id.Cmp(end) < 0
	}
	return start.Cmp(id) < 0 || id.Cmp(end) < 0
}
