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

	//// ======== RECV
	re := pb.ConsumeRequest{
		QueueName: qName,
		//TODO ConsumerGroup: ,
	}
	stream, err := api.Consume(context.Background(), &re)
	if err != nil {
		log.Fatalf("Fehler beim Abrufen des Streams: %v", err)
	}
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("Stream-Ende erreicht oder Fehler: %v", err)
			break // Beende den Stream, wenn ein Fehler auftritt
		}
		fmt.Printf("Empfangen: %s - Inhalt: %s\n", msg.QueueName, msg.Content)
		log.Printf("Received a message: %s", msg)
	}
}
