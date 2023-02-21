package grpc

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vlad-marlo/godo/internal/pkg/fielderr"
	"github.com/vlad-marlo/godo/pkg/proto/api/v1/pb"
)

const handlerField = "grpc_handler"

// Ping ...
func (s *Server) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	l := s.logger.With(zap.String(handlerField, "ping"))
	var resp pb.PingResponse
	if err := s.srv.Ping(ctx); err != nil {
		fieldErr, ok := err.(*fielderr.Error)
		if !ok {
			l.Error("grpc: Service: ping: got unexpected error", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "unexpected error: %v", err)
		}
		return nil, status.Error(fieldErr.CodeGRPC(), fieldErr.Error())
	}
	return &resp, nil
}
