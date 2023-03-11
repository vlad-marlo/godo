package grpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"github.com/vlad-marlo/godo/pkg/proto/api/v1/pb"
)

// Ping ...
func (s *Server) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	var resp pb.PingResponse
	if err := s.srv.Ping(ctx); err != nil {
		fieldErr, ok := err.(*fielderr.Error)
		if !ok {
			return nil, s.internal("ping", err)
		}
		return nil, status.Error(fieldErr.CodeGRPC(), fieldErr.Error())
	}
	return &resp, nil
}

// CreateUser ...
func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	u, err := s.srv.RegisterUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if fErr, ok := err.(*fielderr.Error); ok {
			return nil, fErr.ErrGRPC()
		}
		return nil, s.internal("register user", err)
	}
	return &pb.CreateUserResponse{
		Id:    u.ID.String(),
		Email: u.Email,
	}, nil
}

// CreateToken ...
func (s *Server) CreateToken(ctx context.Context, req *pb.CreateTokenRequest) (*pb.CreateTokenResponse, error) {
	t, err := s.srv.CreateToken(ctx, req.GetEmail(), req.GetPassword(), req.GetTokenType())
	if err != nil {
		if fErr, ok := err.(*fielderr.Error); ok {
			return nil, fErr.ErrGRPC()
		}
		return nil, s.internal("create token", err)
	}
	return &pb.CreateTokenResponse{
		Type:  t.TokenType,
		Token: t.AccessToken,
	}, nil
}

// internal log message and return grpc error with Internal code.
func (s *Server) internal(msg string, err error) error {
	s.logger.Error(fmt.Sprintf("grpc: Service: %s: got unexpected error", msg), zap.Error(err))
	return status.Errorf(codes.Internal, "unexpected error: %v", err)
}
