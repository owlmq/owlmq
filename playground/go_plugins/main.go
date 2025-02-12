package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "path/to/generated/plugin" // Replace with the actual import path

	"google.golang.org/grpc"
)

// PluginServer implements the PluginService

type PluginServer struct {
	pb.UnimplementedPluginServiceServer
}

func (s *PluginServer) Run(ctx context.Context, req *pb.PluginRequest) (*pb.PluginResponse, error) {
	response := fmt.Sprintf("Hello Plugin received: %s", req.Input)
	return &pb.PluginResponse{Output: response}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterPluginServiceServer(grpcServer, &PluginServer{})

	log.Println("Hello Plugin Server is running on port 50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
