package main

import (
	"log"
	"net"
	"os"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/chord"
	"github.com/owlmq/owlmq/config"
	"github.com/owlmq/owlmq/messaging"
	"github.com/owlmq/owlmq/replicator"
	"github.com/owlmq/owlmq/storage"
	"github.com/owlmq/owlmq/web"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Wrong usage of command. eg. ./src localhost:9000")
	}

	//setting up config
	c := config.New(os.Args[1])

	log.Printf("[NODEID:%s]: initializing\n", c.NodeID)

	//init layers
	sl, err := storage.New(storage.InMemory)
	if err != nil {
		panic(err)
	}
	cl := chord.New(sl)

	// replicator
	rep, _ := replicator.New(sl)
	go rep.StartReplicationRoutine()
	go rep.StartCleanupRoutine()

	// stabilize in the Background
	go cl.Stabilize()

	//messaging layer
	ml, _ := messaging.New()
	startServer(c, cl, ml)
}

func startServer(c *config.Config, cl *chord.Chord, ml messaging.MessagingLayer) {
	server := grpc.NewServer()
	owlmqServer := web.NewOwlmqServer(cl, ml)
	pb.RegisterOwlmqServer(server, owlmqServer)
	reflection.Register(server)

	log.Printf("[NODEID:%s]: now reachable on %s\n", c.NodeID, c.Hostname)
	listener, err := net.Listen("tcp", c.Hostname)
	if err != nil {
		panic(err)
	}
	if err := server.Serve(listener); err != nil {
		panic(err)
	}
}
