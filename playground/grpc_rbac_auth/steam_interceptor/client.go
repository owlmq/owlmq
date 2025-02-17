package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Secret Key für JWT
var secretKey = []byte("supersecretkey")

// Erstellt ein neues JWT
func generateJWT(username, role string) (string, error) {
	claims := &Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)), // Token lebt 1 Minute
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// Client mit Stream und automatischer Token-Erneuerung
func startClient() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := NewChatServiceClient(conn)

	// Erstes JWT erzeugen
	token, err := generateJWT("user1", "user")
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}

	// gRPC Metadata mit JWT setzen
	md := metadata.New(map[string]string{"authorization": "Bearer " + token})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Streaming starten
	stream, err := client.Chat(ctx)
	if err != nil {
		log.Fatalf("Failed to start stream: %v", err)
	}

	// Regelmäßige Token-Erneuerung
	go func() {
		for {
			time.Sleep(30 * time.Second) // Erneuerung alle 30 Sekunden
			newToken, _ := generateJWT("user1", "user")
			fmt.Println("🔄 Neues Token gesendet!")

			// Neues Metadata mit aktualisiertem Token setzen
			newMD := metadata.New(map[string]string{"authorization": "Bearer " + newToken})
			stream.SetHeader(newMD)
		}
	}()

	// Nachrichten senden
	for i := 0; i < 10; i++ {
		err := stream.Send(&ChatMessage{Message: fmt.Sprintf("Nachricht %d", i)})
		if err != nil {
			log.Fatalf("Send failed: %v", err)
		}
		time.Sleep(10 * time.Second)
	}
}
