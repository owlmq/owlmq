package replicator

import (
	"sync"
	"time"

	"github.com/owlmq/owlmq/storage"
)

type Replicator interface {
	StartReplicationRoutine()
	StartCleanupRoutine()
	TakeOverReplicas(address string)

	PutEntry(key string, value ReplicatedEntry)
	GetEntry(key string) (*ReplicatedEntry, error)
}

type ReplicatedEntry struct {
	Value        string
	OwnerAddress string
	LastUpdated  time.Time
}

var replicatorInstance Replicator
var once sync.Once

func New(s storage.StorageLayer) (Replicator, error) {
	once.Do(func() {
		replicatorInstance = newSuccessorListReplicator(s)
	})
	return replicatorInstance, nil
	//maybe add other replicators later
}

func GetInstance() Replicator {
	return replicatorInstance
}
