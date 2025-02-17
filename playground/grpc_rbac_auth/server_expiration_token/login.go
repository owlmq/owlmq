package main

import (
	"context"
	"fmt"
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

// gRPC-Methode für Login
func (s *Server) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	user, exists := users[req.Username]
	if !exists || user.Password != req.Password {
		return nil, fmt.Errorf("invalid username or password")
	}

	accessToken, refreshToken, err := GenerateTokens(user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
