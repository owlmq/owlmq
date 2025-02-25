package web

import (
	"context"
	"fmt"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/config"
	"github.com/owlmq/owlmq/crypto"
	"github.com/owlmq/owlmq/replicator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TODO maybe move to chordLayer package
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
	return s.chordLayer.FindSuccessor(ctx, req)
}

func (s *OwlmqServer) GetSuccessor(ctx context.Context, req *pb.GetSuccessorRequest) (*pb.GetSuccessorResponse, error) {
	return &pb.GetSuccessorResponse{Address: config.GetInstance().Successor}, nil
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

func (s *OwlmqServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{}, nil
}

func (s *OwlmqServer) PutReplicaEntry(ctx context.Context, req *pb.ReplicaPutRequest) (*pb.ReplicaPutResponse, error) {
	replicator.GetInstance().PutEntry(req.Key, replicator.ReplicatedEntry{
		Value:        req.GetValue().GetValue(),
		OwnerAddress: req.GetValue().GetOwnerAddress(),
		LastUpdated:  req.GetValue().GetLastUpdated().AsTime(),
	})
	return &pb.ReplicaPutResponse{
		Status:   pb.REPLICA_STATUS_REPLICA_SUCCESS,
		ErrorMsg: "",
	}, nil
}

func (s *OwlmqServer) GetReplicaEntry(ctx context.Context, req *pb.ReplicaGetRequest) (*pb.ReplicaGetResponse, error) {
	entry, err := replicator.GetInstance().GetEntry(req.Key)
	if err != nil {
		return &pb.ReplicaGetResponse{
			Status:   pb.REPLICA_STATUS_REPLICA_FAILURE,
			ErrorMsg: err.Error(),
		}, err
	}

	return &pb.ReplicaGetResponse{
		Value: &pb.ReplicatedEntry{
			Value:        entry.Value,
			OwnerAddress: entry.OwnerAddress,
			LastUpdated:  timestamppb.New(entry.LastUpdated),
		},
		Status:   pb.REPLICA_STATUS_REPLICA_SUCCESS,
		ErrorMsg: "",
	}, nil

}
