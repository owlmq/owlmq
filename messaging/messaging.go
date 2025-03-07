package messaging

import (
	"context"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/chord"
)

type MessagingLayer interface {
	ProduceOne(ctx context.Context, msg *pb.Message) (*pb.ProduceOneResponse, error)
	Produce(stream pb.Owlmq_ProduceServer) error
	ConsumeOne(ctx context.Context, req *pb.ConsumeOneRequest) (*pb.Message, error)
	Consume(req *pb.ConsumeRequest, stream pb.Owlmq_ConsumeServer) error
	Ack(ctx context.Context, req *pb.AckRequest) (*pb.AckResponse, error)
	NewQueue(ctx context.Context, req *pb.NewQueueRequest) (*pb.NewQueueResponse, error)
}

func New(cl *chord.Chord) (MessagingLayer, error) {
	//TODO maybe add different messaging layers later
	return newSimpleMessagingLayer(cl), nil
}

type Queue struct {
	Key   string     `json:"key"` //-> "q:logs"
	Value QueueValue `json:"value"`
}

type QueueValue struct {
	Subscribers []string `json:"subscribers"`
	HeadUUID    string   `json:"headuuid"`
	//TODO add this later(storage layer need to support this)
	Persistent bool `json:"persistent"`
}

type Message struct {
	Key   string //-> "m:uuid"
	Value MessageValue
}

type MessageValue struct {
	UUID string //-> e.g.: "12345"
	//TODO maybe refactor to be the routing key
	QueueName      string //-> e.g.: "logs"
	Content        string
	PreMessageUUID string //-> e.g.: "12345"
	//genesis is the first msg in a queue
	IsGenesis bool
}
