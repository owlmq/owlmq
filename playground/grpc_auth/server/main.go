package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/owlmq/owlmq/auth"
	"github.com/owlmq/owlmq/interceptor"
	pb "github.com/owlmq/owlmq/proto/protobufs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedOrgServiceServer
	orgs map[string]*pb.OrganisationResponse
}

func (s *server) GetOrganisationByID(ctx context.Context, req *pb.OrganisationIDRequest) (*pb.OrganisationResponse, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return nil, status.Error(codes.FailedPrecondition, "user id missing")
	}
	fmt.Println(userID)

	org, exists := s.orgs[req.GetId()]
	if !exists {
		return nil, fmt.Errorf("Org with ID %s not found", req.GetId())
	}
	return org, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// get JWT secret key from environment variable
	//jwtSecret, ok := os.LookupEnv("JWT_SECRET")
	//if !ok {
	//log.Fatal("JWT_SECRET must be provided")
	//}
	jwtSecret := "jwt-password"

	// initialise our auth service & interceptor
	authSvc, err := auth.NewService(jwtSecret)
	if err != nil {
		log.Fatalf("failed to initialize auth service: %v", err)
	}
	interceptor, err := interceptor.NewAuthInterceptor(authSvc)
	if err != nil {
		log.Fatalf("failed to initialize interceptor: %v", err)
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryAuthMiddleware),
	)
	srv := &server{
		orgs: map[string]*pb.OrganisationResponse{
			"1": {Id: "1", Name: "Organisation One"},
			"2": {Id: "2", Name: "Organisation Two"},
		},
	}
	pb.RegisterOrgServiceServer(s, srv)

	//TODO find a better way by using a login service at the server
	fmt.Println(authSvc.IssueToken("dummy-user"))

	fmt.Println("Server started on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
