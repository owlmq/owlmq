package main

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// gRPC-Methode zur Token-Erneuerung
func (s *Server) RefreshToken(ctx context.Context, req *RefreshRequest) (*RefreshResponse, error) {
	refreshToken := req.RefreshToken
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Neues Access-Token erstellen
	newAccessToken, _, err := GenerateTokens(claims.Username, claims.Role)
	if err != nil {
		return nil, err
	}

	return &RefreshResponse{AccessToken: newAccessToken}, nil
}
