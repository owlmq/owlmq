package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/owldb/owldb/proto/protobufs"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
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

		//get the token and set it to the metadata TODO find a better way
		md := metadata.Pairs("authorization", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Mzc0NjU3NzcsImlzcyI6MTczNzQ2NDg3Nywic3ViIjoiZHVtbXktdXNlciJ9.0jj1JU2ypdPxJ21Jqrafg_mMbUFK8lArxsVtUzXtVVE")
		ctx = metadata.NewOutgoingContext(ctx, md)

		r, err := client.GetOrganisationByID(ctx, &pb.OrganisationIDRequest{Id: id})
		if err != nil {
			log.Fatalf("could not call protected method: %v", err)
		}
		log.Printf("Org: %s, Name: %s", r.GetId(), r.GetName())
	}
}
