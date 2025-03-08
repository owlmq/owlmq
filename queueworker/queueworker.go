package queueworker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/config"
	"github.com/owlmq/owlmq/types"
	"google.golang.org/grpc"
)

type QueueWorker interface {
	NotifyConsumers(queuename string)
	NotifyLocalConsumers(ctx context.Context, req *pb.NotifyLocalConsumersRequest) (*pb.NotifyLocalConsumersResponse, error)
	AddConsumer(consumerUUID string, stream pb.Owlmq_ConsumeServer)
	CheckConsumer(consumerUUID string) error
}

// This struct holds the logic for a QueueWorker which is responsible for distributing the messages inside a queue to the subscribers
var queueworkerInstance QueueWorker
var once sync.Once

// maybe add other queueworker strategies later
func GetInstance() (QueueWorker, error) {
	once.Do(func() {
		queueworkerInstance = &SimpleQueueWorker{
			Consumers: make(map[string]pb.Owlmq_ConsumeServer),
		}
	})
	return queueworkerInstance, nil
}

type SimpleQueueWorker struct {
	//!! map is not thread save
	Consumers map[string]pb.Owlmq_ConsumeServer
	mu        sync.Mutex
}

func (s *SimpleQueueWorker) NotifyConsumers(queuename string) {
	//get the q:queuename value from the kv.store
	//TODO refactor to remove deprecated
	conn, err := grpc.Dial(config.GetInstance().Hostname, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	api := pb.NewOwlmqClient(conn)

	pg := pb.KV_GetRequest{
		Key: fmt.Sprintf("q:%s", queuename),
	}
	qvJSON, err := api.Get(context.Background(), &pg)
	if err != nil {
		log.Printf("Error accessing the queue: %v", err)
	}
	var qval types.QueueValue
	err = json.Unmarshal([]byte(qvJSON.GetValue()), &qval)
	if err != nil {
		log.Printf("Error Unmarshal: %v", err)
	}

	fmt.Println("HERE we are NotifyConsumers")
	fmt.Println(qvJSON.GetValue())
	fmt.Println(qval.Consumers)
	// iterate the nodes where consumers are connected
	for _, v := range qval.Consumers {
		conn, err := grpc.Dial(v.QueueWorkerAdress, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		remote := pb.NewOwlmqClient(conn)
		nr := pb.NotifyLocalConsumersRequest{
			QueueName: queuename,
		}
		// inform inform the nodes queueworkers that a new message is available
		remote.NotifyLocalConsumers(context.Background(), &nr)
	}

}

func (s *SimpleQueueWorker) NotifyLocalConsumers(ctx context.Context, req *pb.NotifyLocalConsumersRequest) (*pb.NotifyLocalConsumersResponse, error) {
	fmt.Println("HERE we are NotifyLOCALConsumers")
	//get queuename value
	//TODO refactor to remove deprecated
	conn, err := grpc.Dial(config.GetInstance().Hostname, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	api := pb.NewOwlmqClient(conn)

	pg := pb.KV_GetRequest{
		Key: fmt.Sprintf("q:%s", req.QueueName),
	}
	qvJSON, err := api.Get(context.Background(), &pg)
	if err != nil {
		log.Printf("Error accessing the queue: %v", err)
	}
	var qval types.QueueValue
	_ = json.Unmarshal([]byte(qvJSON.Value), &qval)

	//read the newest message from the queue -> TODO change this to get different behaviours
	pg = pb.KV_GetRequest{
		Key: fmt.Sprintf("m:%s", qval.HeadUUID),
	}
	qvJSON, err = api.Get(context.Background(), &pg)
	if err != nil {
		log.Printf("Error accessing the queue: %v", err)
	}
	var vmsg types.Message
	_ = json.Unmarshal([]byte(qvJSON.Value), &vmsg)

	//iterate the value to inform local consumers
	for _, v := range qval.Consumers {
		//this means that the client is connected localy
		if v.QueueWorkerAdress == config.GetInstance().Hostname {
			//TODO inform local clients that here we have a new message
			s, err := s.GetStream(v.ConsumerUUID)
			if err != nil {
				fmt.Printf("Unable to connect to client stream %s\n", err.Error())
			}
			if err := s.Send(&pb.Message{
				QueueName: vmsg.Value.QueueName,
				Content:   vmsg.Value.Content,
			}); err != nil {
				fmt.Printf("Unable to send client %s\n", err.Error())
			}
		}
	}

	return &pb.NotifyLocalConsumersResponse{}, nil
}
