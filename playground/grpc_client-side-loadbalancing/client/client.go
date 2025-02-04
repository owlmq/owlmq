package main

import (
	"context"
	"log"
	"time"

	pb "example.com/m/protobufs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

func main() {
	// DNS Resolver aktivieren
	resolver.SetDefaultScheme("dns")

	// Verbindung mit Load Balancing herstellen
	conn, err := grpc.Dial(
		"dns:///localhost:50051",                                 // Nutzt Service Discovery
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Kein TLS
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// gRPC-Client erstellen
	client := pb.NewGreeterClient(conn)

	// Mehrere Anfragen senden, um Load Balancing zu testen
	for i := 0; i < 5; i++ {
		resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "Alice"})
		if err != nil {
			log.Fatalf("Could not greet: %v", err)
		}
		log.Printf("Response: %s", resp.Message)
		time.Sleep(1 * time.Second)
	}
}
