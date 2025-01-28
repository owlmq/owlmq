package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/owldb/owldb/playground/chord"
	"github.com/owldb/owldb/playground/utils"
	"github.com/owldb/owldb/playground/web"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Wrong usage of command. eg. ./src localhost:9000")
	}
	hostname := os.Args[1]
	//DEBUG
	log.Printf("[NODEID:%s]: initializing\n", utils.GenerateSHA1(hostname))

	//setting up context env
	ctx := context.Background()
	ctx = context.WithValue(ctx, "hostname", hostname)
	ctx = context.WithValue(ctx, "nodeID", utils.GenerateSHA1(hostname))
	ctx = context.WithValue(ctx, "predecessor", "")
	ctx = context.WithValue(ctx, "successor", "")

	//init chord layer with creation of the fingertable
	cl := chord.New(&ctx)

	//init fingertable and start the updater
	ctx = context.WithValue(ctx, "successor", cl.FindSuccessor(hostname, hostname))
	cl.GetFingerTable().StartEntryUpdater(&ctx)

	//create web layer
	web := web.New(&ctx, cl)

	log.Printf("[NODEID:%s]: now reachable on %s\n", utils.GenerateSHA1(hostname), hostname)
	log.Fatal(http.ListenAndServe(hostname, web.NewWebserver(&ctx)))
}
