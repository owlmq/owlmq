package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Secret Key für JWT
var secretKey = []byte("supersecretkey")

// JWT Claims Struktur
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Stream Interceptor mit JWT-Erneuerung
func StreamAuthInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := stream.Context()

	// Eingangsprüfung des Tokens
	newCtx, err := ValidateJWTFromMetadata(ctx)
	if err != nil {
		return err
	}

	// WrappedStream mit erweitertem Kontext
	wrappedStream := &wrappedStream{ServerStream: stream, ctx: newCtx}

	// Starte den Stream
	return handler(srv, wrappedStream)
}

// Validiert JWT aus den gRPC-Metadaten (erneuert falls nötig)
func ValidateJWTFromMetadata(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

	// Token aus den Metadata-Headern extrahieren
	authHeaders := md["authorization"]
	if len(authHeaders) == 0 {
		return nil, fmt.Errorf("missing authorization token")
	}

	tokenString := strings.TrimPrefix(authHeaders[0], "Bearer ")
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// **Check: Ist das Token bald abgelaufen?**
	if time.Until(claims.ExpiresAt.Time) < 30*time.Second {
		fmt.Println("Token fast abgelaufen, bitte erneuern!")
	}

	// Kontext mit Benutzerinformationen erweitern
	ctx = context.WithValue(ctx, "username", claims.Username)
	ctx = context.WithValue(ctx, "role", claims.Role)

	return ctx, nil
}

// WrappedStream für Token-Erneuerung
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Kontext-Funktion überschreiben
func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
