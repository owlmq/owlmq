package main

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type User struct {
	Username string
	Password string
	Role     string
}

var users = map[string]User{
	"admin":  {Username: "admin1", Password: "password", Role: "admin"},
	"user":   {Username: "user1", Password: "password", Role: "user"},
	"node":   {Username: "node1", Password: "password", Role: "node"},
	"plugin": {Username: "plugin", Password: "password", Role: "plugin"},
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var jwtKey = []byte("my_secret_key")

func GenerateJWT(username, role string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

type AuthService struct{}

func (a *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, exists := users[req.Username]
	if !exists || user.Password != req.Password {
		return nil, errors.New("invalid credentials")
	}
	token, err := GenerateJWT(user.Username, user.Role)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{Token: token}, nil
}

var methodRoleMap = map[string][]string{
	"/HelloService/SayHello":  {"user", "admin"},
	"/HelloService/AdminOnly": {"admin"},
}

func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errors.New("missing metadata")
		}

		tokens := md["authorization"]
		if len(tokens) == 0 {
			return nil, errors.New("missing token")
		}

		tokenString := tokens[0]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			return nil, errors.New("invalid token")
		}

		allowedRoles, exists := methodRoleMap[info.FullMethod]
		if !exists {
			return nil, errors.New("unauthorized method")
		}

		for _, role := range allowedRoles {
			if claims.Role == role {
				return handler(ctx, req)
			}
		}
		return nil, errors.New("unauthorized")
	}
}

type HelloService struct{}

func (h *HelloService) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{Message: "Hello, " + req.Name}, nil
}

func (h *HelloService) AdminOnly(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{Message: "Admin access granted"}, nil
}

func main() {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(AuthInterceptor()),
	)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC Server is running on :50051")
	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
