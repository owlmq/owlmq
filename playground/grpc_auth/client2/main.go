package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/owlmq/owlmq/proto/protobufs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// jwtCredentials satisfies the PerRPCCredentials interface.
type jwtCredentials struct{}

// GetRequestMetadata will be called on every RPC call and returns a map which is used to build the request metadata.
func (j *jwtCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzc0NjU3NzcsImlzcyI6MTczNzQ2NDg3Nywic3ViIjoiZHVtbXktdXNlciJ9.0jj1JU2ypdPxJ21Jqrafg_mMbUFK8lArxsVtUzXtVVE"

	// return metadata map for RPC call
	return map[string]string{
		"authorization": token,
	}, nil
}

func (j *jwtCredentials) RequireTransportSecurity() bool {
	return false
}

func main() {
	creds := &jwtCredentials{}

	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(creds),
	)

	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewOrgServiceClient(conn)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter Org ID: ")
		id, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("could not read input: %v", err)
		}
		// Remove the newline character
		id = id[:len(id)-1]

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, err := client.GetOrganisationByID(ctx, &pb.OrganisationIDRequest{Id: id})
		if err != nil {
			log.Fatalf("could not call protected method: %v", err)
		}
		log.Printf("Org: %s, Name: %s", r.GetId(), r.GetName())
	}
}
