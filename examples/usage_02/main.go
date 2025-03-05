package main

import (
	"fmt"
	"log"
)

// This example demonstrates sending a message and later recving it
func main() {
	//TODO use grpc
	//connect
	api := Pb{}

	// ======== Create Queue
	qName := "abc-queue"
	if api.NewQueue(
		qName, // name
		false, // persistent
	) != nil {
		fmt.Println("Error")
	}

	// ======== Send
	//Produce msg to owlmq
	if api.Publish(
		"",            // exchange
		qName,         // routing key
		"hello world", // content
	) != nil {
		fmt.Println("Error")
	}

	// ======== RECV
	go func() {
		for {
			msg, err := api.Consume(
				qName, // queue
			)
			if err != nil {
				fmt.Println("Error")
			}
			log.Printf("Received a message: %s", msg)
		}
	}()
}

type Pb struct{}

func (p *Pb) NewQueue(name string, persistent bool) error {
	return nil
}
func (p *Pb) Publish(name string, routing_key string, content string) error {
	return nil
}
func (p *Pb) Consume(queue string) (string, error) {
	return "", nil
}
