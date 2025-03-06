package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/chord"
)

func newSimpleMessagingLayer(cl *chord.Chord) MessagingLayer {
	return &SimpleMessagingLayer{
		chordLayer: cl,
	}
}

type SimpleMessagingLayer struct {
	chordLayer *chord.Chord
}

func (s *SimpleMessagingLayer) ProduceOne(ctx context.Context, msg *pb.Message) (*pb.ProduceOneResponse, error) {
	panic("unimplemented")
}

func (s *SimpleMessagingLayer) Produce(stream pb.Owlmq_ProduceServer) error {
	panic("unimplemented")
}

func (s *SimpleMessagingLayer) ConsumeOne(ctx context.Context, req *pb.ConsumeOneRequest) (*pb.Message, error) {
	panic("unimplemented")
}

func (s *SimpleMessagingLayer) Consume(req *pb.ConsumeRequest, stream pb.Owlmq_ConsumeServer) error {
	panic("unimplemented")
}

func (s *SimpleMessagingLayer) Ack(ctx context.Context, req *pb.AckRequest) (*pb.AckResponse, error) {
	panic("unimplemented")
}

func (s *SimpleMessagingLayer) NewQueue(ctx context.Context, req *pb.NewQueueRequest) (*pb.NewQueueResponse, error) {
	q := Queue{
		Key: fmt.Sprintf("q:%s", req.QueueName),
		Value: QueueValue{
			Subscribers: []string{},
		},
	}

	//check if the queue already exists
	pg := pb.KV_GetRequest{
		Key: q.Key,
	}
	val, _ := s.chordLayer.Get(context.Background(), &pg)
	if val.GetValue() != "" {
		return &pb.NewQueueResponse{
			Status: pb.MESSAGE_STATUS_MSG_FAILURE,
		}, errors.New("queue already exists")
	}

	//generate a new Queue
	qvJSON, err := json.Marshal(q.Value)
	if err != nil {
		return &pb.NewQueueResponse{
			Status: pb.MESSAGE_STATUS_MSG_FAILURE,
		}, errors.New("error while serializing")
	}
	pr := pb.KV_PutRequest{
		Key:   q.Key,
		Value: string(qvJSON),
	}
	// store the new Queue on the chord ring
	_, err = s.chordLayer.Put(context.Background(), &pr)
	if err != nil {
		return &pb.NewQueueResponse{
			Status: pb.MESSAGE_STATUS_MSG_FAILURE,
		}, errors.New("error while putting new queue")
	}

	// send an initial genisis msg to the queue
	genesis := Message{
		Key: fmt.Sprintf("m:%s", uuid.New().String()),
		Value: MessageValue{
			QueueName:      q.Key,
			Content:        "GENESIS MESSAGE",
			PreMessageUUID: "",
			IsGenesis:      true,
		},
	}
	gvJSON, err := json.Marshal(genesis.Value)
	if err != nil {
		return &pb.NewQueueResponse{
			Status: pb.MESSAGE_STATUS_MSG_FAILURE,
		}, errors.New("error while serializing genesis")
	}
	gr := pb.KV_PutRequest{
		Key:   genesis.Key,
		Value: string(gvJSON),
	}
	_, err = s.chordLayer.Put(context.Background(), &gr)
	if err != nil {
		return &pb.NewQueueResponse{
			Status: pb.MESSAGE_STATUS_MSG_FAILURE,
		}, errors.New("error while putting genesis")
	}

	return &pb.NewQueueResponse{
		Status: pb.MESSAGE_STATUS_MSG_SUCCESS,
	}, nil
}
