package grpc

import (
	"testing"

	"go.uber.org/zap"

	"github.com/vlad-marlo/godo/internal/config"
	"github.com/vlad-marlo/godo/pkg/proto/api/v1/pb"
)

// TestServer is helper for tests that creates testing server.
func TestServer(t testing.TB, srv Service) *Server {
	t.Helper()
	return &Server{
		UnimplementedGodoServer: pb.UnimplementedGodoServer{},
		srv:                     srv,
		cfg:                     config.New(),
		logger:                  zap.L(),
	}
}
