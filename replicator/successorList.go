package replicator

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/config"
	"github.com/owlmq/owlmq/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newSuccessorListReplicator(s storage.StorageLayer) Replicator {
	return &SuccessorListReplicator{
		//init the replica storage once when starting the node
		replicastorage:   newReplicaStorage(),
		nodeStorageLayer: s,
	}
}

type SuccessorListReplicator struct {
	replicastorage   *ReplicaStorage
	nodeStorageLayer storage.StorageLayer
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
					//fmt.Printf("replicating the key %v with value %v to node %v\n", key, value, address)

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
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			//TODO maybe refactor later so that the replication storage also gets a iterator
			for k, v := range s.replicastorage.store {
				n := time.Now()

				rt := n.Sub(v.LastUpdated).Seconds()
				md := (time.Duration(config.ReplicaLeaseDuration) * time.Second).Seconds()

				if rt > md {
					//debugging output
					//fmt.Printf("removing replicated key:%v with value:%v from replication storage\n", k, v)
					//replica is to old
					delete(s.replicastorage.store, k)
				}
			}

		}
	}
}

func (s *SuccessorListReplicator) TakeOverReplicas(address string) {
	//TODO this will get triggered if our predeccessor dies, if so we move every key if its key space into our normal node storage
	//TODO think about this (is there a way that we have the key 2 times in the ring if we just move it to our key space)
	fmt.Println("TAKE OVER KEY SPACE:", address)
	fmt.Println(s.replicastorage.store)

	//TODO refactor to remove deprecated
	conn, err := grpc.Dial(config.GetInstance().Hostname, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	api := pb.NewOwlmqClient(conn)

	for k, v := range s.replicastorage.store {
		fmt.Println(v.OwnerAddress)
		if v.OwnerAddress == address {
			fmt.Println("put replicated key into my key space")
			//use a PUT grpc call to transfer the key -> to only have one instance of the key
			pr := pb.KV_PutRequest{
				Key:   k,
				Value: string(v.Value),
			}
			api.Put(context.Background(), &pr)
		}
	}
}

// function so that the API call can add new replicas to this node
func (s *SuccessorListReplicator) PutEntry(key string, value ReplicatedEntry) {
	s.replicastorage.Put(key, value)
}

// function so that the API call can read a specific replica entry (maybe interesting later)
func (s *SuccessorListReplicator) GetEntry(key string) (*ReplicatedEntry, error) {
	return s.replicastorage.Get(key)
}
