package main

import (
	"context"
	"fmt"
	"log"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"google.golang.org/grpc"
)

// This example demonstrates sending a message and later recving it
func main() {
	//Connect
	conn, err := grpc.Dial("localhost:9000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	api := pb.NewOwlmqClient(conn)

	// ======== Create Queue
	qName := "abc-queue"
	_, err = api.NewQueue(context.Background(), &pb.NewQueueRequest{
		QueueName: qName,
	})
	if err != nil {
		fmt.Println("Error", err)
	}

	// ======== Send
	//Produce msg to owlmq
	msg := pb.Message{
		QueueName: qName,         // queue name -> later routing key
		Content:   "hello world", // content
	}
	_, err = api.ProduceOne(context.Background(), &msg)
	if err != nil {
		fmt.Println("Error", err)
	}

	// ======== RECV
	for {
		re := pb.ConsumeOneRequest{
			QueueName: qName,
			//TODO ConsumerGroup: ,
		}
		msg, err := api.ConsumeOne(context.Background(), &re)
		if err != nil {
			fmt.Println("Error")
		}
		//print msg
		log.Printf("Received a message: %s", msg)
	}
}
