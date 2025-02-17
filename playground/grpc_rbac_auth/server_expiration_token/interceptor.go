package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TODO refactor this into an own file / config file or db -> maybe this can be inserted into the key value store structure
var methodRoleMap = map[string][]string{
	"/HelloService/SayHello":  {"user", "admin"},
	"/HelloService/AdminOnly": {"admin"},
}

// gRPC Auth-Interceptor mit RBAC-Prüfung
func AuthInterceptor(ctx context.Context, method string) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata")
	}

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

	// RBAC-Prüfung
	if !isRoleAllowed(method, claims.Role) {
		return nil, fmt.Errorf("access denied for role: %s", claims.Role)
	}

	// Kontext mit Benutzerinformationen erweitern
	ctx = context.WithValue(ctx, "username", claims.Username)
	ctx = context.WithValue(ctx, "role", claims.Role)

	return ctx, nil
}

// Prüft, ob eine Rolle Zugriff auf eine gRPC-Methode hat
func isRoleAllowed(method, role string) bool {
	allowedRoles, exists := methodRoleMap[method]
	if !exists {
		return false
	}

	for _, r := range allowedRoles {
		if r == role {
			return true
		}
	}
	return false
}

// UnaryInterceptor für gRPC-Server
func UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	ctx, err := AuthInterceptor(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}
	return handler(ctx, req)
}
