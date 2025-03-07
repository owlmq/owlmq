package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

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
	//new message
	mid := uuid.New().String()
	m := Message{
		Key: fmt.Sprintf("m:%s", mid),
		Value: MessageValue{
			UUID:      mid,
			QueueName: msg.QueueName,
			Content:   msg.Content,
			IsGenesis: false,
		},
	}

	//TODO  this is a synchronization problem, i am getting the value of q:... here
	//		then i am putting the updated value with my message as new head of queue
	//		solution 1: implement a atomic PUTGET() operation (prefered solution)
	//		solution 2: implement shared locking
	//request predecessor-message
	pg := pb.KV_GetRequest{
		Key: fmt.Sprintf("q:%s", msg.QueueName),
	}
	qvJSON, err := s.chordLayer.Get(context.Background(), &pg)
	if err != nil {
		log.Printf("Error accessing the queue: %v", err)
	}
	var qval QueueValue
	_ = json.Unmarshal([]byte(qvJSON.Value), &qval)

	//set the Unmarshaled headUUID as predecessor message to the new message
	m.Value.PreMessageUUID = qval.HeadUUID

	//store the new message in the chord ring
	mvJSON, err := json.Marshal(m.Value)
	pr := pb.KV_PutRequest{
		Key:   m.Key,
		Value: string(mvJSON),
	}
	_, err = s.chordLayer.Put(context.Background(), &pr)

	//put the new q:.. value back with the new message as new head
	q := Queue{
		Key: pg.Key,
		Value: QueueValue{
			Subscribers: qval.Subscribers,
			HeadUUID:    mid,
			//TODO Persistent: ,
		},
	}
	qJSON, err := json.Marshal(qvJSON.Value)
	pr = pb.KV_PutRequest{
		Key:   q.Key,
		Value: string(qJSON),
	}
	_, err = s.chordLayer.Put(context.Background(), &pr)

	return &pb.ProduceOneResponse{
		Status: pb.MESSAGE_STATUS_MSG_SUCCESS,
	}, nil
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
	//this need to be here because i need the uuid for the head of the queue
	gid := uuid.New().String()

	q := Queue{
		Key: fmt.Sprintf("q:%s", req.QueueName),
		Value: QueueValue{
			Subscribers: []string{},
			HeadUUID:    gid,
			//TODO Persistent: ,
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
		Key: fmt.Sprintf("m:%s", gid),
		Value: MessageValue{
			UUID:           gid,
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
