package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <bootstrap address> <node address>\n", os.Args[0])
	}

	bootstrapAddress := os.Args[1] // Die Adresse des Bootstrap-Knotens (z.B. localhost:50051)
	nodeAddress := os.Args[2]      // Die Adresse des neuen Knotens (z.B. localhost:5002)

	// Verbindung zum gRPC-Server (Bootstrap-Node) herstellen
	conn, err := grpc.Dial(nodeAddress, grpc.WithInsecure()) // Verbindet sich mit dem Bootstrap-Node
	if err != nil {
		log.Fatalf("did not connect to bootstrap node: %v", err)
	}
	defer conn.Close()

	// Erstelle den gRPC-Client
	client := pb.NewOwlmqClient(conn)

	// Erstelle die Join-Anfrage
	req := &pb.NodeJoinRequest{Address: bootstrapAddress}

	// Sende die Join-Anfrage an den Bootstrap-Node
	_, err = client.NodeJoin(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to join ring: %v", err)
	}

	fmt.Printf("Node %s successfully joined the ring via %s\n", nodeAddress, bootstrapAddress)
}
