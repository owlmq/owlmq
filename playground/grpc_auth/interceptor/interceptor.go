package interceptor

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	// Validator defines an interface for token validation. This is satisfied by our auth service.
	Validator interface {
		ValidateToken(ctx context.Context, token string) (string, error)
	}

	authInterceptor struct {
		validator Validator
	}
)

func NewAuthInterceptor(validator Validator) (*authInterceptor, error) {
	if validator == nil {
		return nil, errors.New("validator cannot be nil")
	}
	return &authInterceptor{validator: validator}, nil
}

func (i *authInterceptor) UnaryAuthMiddleware(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// get metadata object
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	// extract token from authorization header
	token := md["authorization"]
	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}

	// validate token and retrieve the userID
	userID, err := i.validator.ValidateToken(ctx, token[0])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// add our user ID to the context, so we can use it in our RPC handler
	ctx = context.WithValue(ctx, "user_id", userID)

	// call our handler
	return handler(ctx, req)
}
