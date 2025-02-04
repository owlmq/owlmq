package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "example.com/m/protobufs"

	"google.golang.org/grpc"
)

type server struct {
	port int
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	message := fmt.Sprintf("Hello %s from server on port %d", req.Name, s.port)
	return &pb.HelloReply{Message: message}, nil
}

func startServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGreeterServer(grpcServer, &server{port: port})

	log.Printf("Server listening on port %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func main() {
	go startServer(50051) // Server 1 auf Port 50051
	go startServer(50052) // Server 2 auf Port 50052

	// Halte das Programm am Leben
	select {}
}
