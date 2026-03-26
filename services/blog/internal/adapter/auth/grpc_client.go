package auth

import (
	"blog/internal/domain"
	"context"

	pb "github.com/Dimassin/articles-microservices/proto/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCTokenValidator struct {
	client pb.AuthServiceClient
}

func NewGRPCTokenValidator(addr string) (*GRPCTokenValidator, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GRPCTokenValidator{
		client: pb.NewAuthServiceClient(conn),
	}, nil
}

func (v *GRPCTokenValidator) Validate(ctx context.Context, token string) (string, error) {
	resp, err := v.client.ValidateToken(ctx, &pb.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return "", err
	}

	if !resp.Valid {
		return "", domain.ErrInvalidToken
	}

	return resp.UserId, nil
}
