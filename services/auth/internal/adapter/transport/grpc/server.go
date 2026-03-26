package grpc

import (
	"auth/internal/usecase"
	"context"

	pb "github.com/Dimassin/articles-microservices/proto/auth"
)

type AuthGrpcServer struct {
	pb.UnimplementedAuthServiceServer
	authUsecase *usecase.AuthUsecase
}

func NewAuthGrpcServer(authUsecase *usecase.AuthUsecase) *AuthGrpcServer {
	return &AuthGrpcServer{
		authUsecase: authUsecase,
	}
}

func (s *AuthGrpcServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := s.authUsecase.ValidateToken(ctx, req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	return &pb.ValidateTokenResponse{
		UserId:   claims.UserID,
		Email:    claims.Email,
		Username: claims.Username,
		Valid:    true,
	}, nil
}
