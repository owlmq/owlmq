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
	//if api.Produce(
	//"",            // exchange
	//qName,         // routing key
	//"hello world", // content
	//) != nil {
	//fmt.Println("Error")
	//}

	//// ======== RECV
	//msg, err := api.Consume(
	//qName, // queue name
	//)
	//if err != nil {
	//fmt.Println("Error")
	//}

	////print msg
	//fmt.Println(msg)
}
