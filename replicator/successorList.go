package replicator

import (
	"context"
	"fmt"
	"time"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/chord"
	"github.com/owlmq/owlmq/config"
	"github.com/owlmq/owlmq/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newSuccessorListReplicator(s storage.StorageLayer, c chord.Chord) Replicator {
	return &SuccessorListReplicator{
		//init the replica storage once when starting the node
		replicastorage:   newReplicaStorage(),
		nodeStorageLayer: s,
		chordLayer:       c,
	}
}

type SuccessorListReplicator struct {
	replicastorage   *ReplicaStorage
	nodeStorageLayer storage.StorageLayer
	chordLayer       chord.Chord
}

func (s *SuccessorListReplicator) StartReplicationRoutine() {
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			iter := s.nodeStorageLayer.Iterator()
			//check if the iterator has a next value
			for iter.Next() {
				key, value := iter.Value()

				//loop over replicator nodes
				for _, address := range config.GetInstance().ReplicatorNodeList {
					//debugging output
					fmt.Printf("replicating the key %v with value %v to node %v\n", key, value, address)

					//move to the replicatorNode
					conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
					if err != nil {
						fmt.Printf("failed to connect to successor for replication: %v\n", err)
					}
					defer conn.Close()

					client := pb.NewOwlmqClient(conn)
					client.PutReplicaEntry(context.Background(), &pb.ReplicaPutRequest{
						Key: key,
						Value: &pb.ReplicatedEntry{
							Value:        value,
							OwnerAddress: config.GetInstance().Hostname,
							LastUpdated:  timestamppb.New(time.Now()),
						},
					})
				}
			}
		}
	}
}

func (s *SuccessorListReplicator) StartCleanupRoutine() {
	fmt.Println("CLEANUP")
	//TODO at this point we need to find out if the last replication of a key is longer than x time e.g. 4*replication cicle and if the LastUpdate is longer the replicated key will be removed
	panic("Unimplemented")
}

func (s *SuccessorListReplicator) TakeOverReplicas(address string) {
	//TODO this will get triggered if our predeccessor dies, if so we move every key if its key space into our normal node storage
	panic("Unimplemented")
}

// function so that the API call can add new replicas to this node
func (s *SuccessorListReplicator) PutEntry(key string, value ReplicatedEntry) {
	s.replicastorage.Put(key, value)
}

// function so that the API call can read a specific replica entry (maybe interesting later)
func (s *SuccessorListReplicator) GetEntry(key string) (*ReplicatedEntry, error) {
	return s.replicastorage.Get(key)
}
